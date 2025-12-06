package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sammyshear/adon-olam/internal/fonspeak_midi"
	"github.com/sammyshear/adon-olam/internal/timing"
	"github.com/sammyshear/fonspeak"
)

type BufWriteCloser struct {
	*bufio.Writer
}

func (bwc *BufWriteCloser) Close() error {
	return bwc.Flush()
}

// syllables contains the Adon Olam lyrics in X-SAMPA format
var syllables = []string{"a", "don", "o", "l@m", "aS", "er", "ma", "laX", "b@", "ter", "em", "kol", "je", "tsir", "niv", "ra", "l@", "et", "na:", "sa", "veX", "ef", "tso", "kol", "az", "ai", "mel", "eX", "Se", "mo", "nik", "ra", "ve", "aX", "a", "rei", "kix", "lot", "ha", "kol", "l@", "va", "do", "jim", "loX", "no", "ra", "v@", "hu", "ha", "ja", "v@", "hu", "ho", "ve", "v@", "hu", "ji", "je", "bet", "if", "ar", "a", "v@", "hu", "eX", "ad", "v@", "ein", "Se", "ni", "l@", "ham", "Sil", "lo", "l@", "haX", "bi", "ra", "bli", "re", "Sit", "bli", "taX", "lit", "v@", "lo", "ha", "oz", "v@", "ham", "mis", "rah", "v@", "hu", "el", "i", "v@", "Xai", "go", "al", "i", "v@", "tsur", "Xev", "li", "b@", "et", "tsa", "ra", "v@", "hu", "nis", "si", "u", "ma", "nos", "li", "m@", "nat", "ko", "si", "b@", "jom", "ek", "ra", "b@", "ja", "do", "af", "kid", "ru", "Xi", "b@", "et", "iS", "an", "v@", "a", "ir", "a", "v@", "im", "ru", "Xi", "g@", "vi", "ja", "ti", "ad", "on", "ai", "li", "v@", "lo", "ir", "a"}

type channel struct {
	requestID       string
	file            multipart.File
	header          *multipart.FileHeader
	statusURL       string
	trackNo         int
	timingStrategy  string
}

func UploadMidiHandler(ch chan channel) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		statusURL := "/api/status/"
		requestID := generateRequestID()
		storeStatus(requestID, JobStatus{State: "NEW"})

		statusURL += requestID

		r.ParseMultipartForm(10 << 20) // 10 MB
		file, header, err := r.FormFile("uploadFile")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		trackNo, err := strconv.Atoi(r.FormValue("trackNo"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		
		// Get timing strategy from form, default to per-syllable
		timingStrategyVal := r.FormValue("timingStrategy")
		if timingStrategyVal == "" {
			timingStrategyVal = "per-syllable"
		}
		
		ch <- channel{requestID, file, header, statusURL, trackNo, timingStrategyVal}

		w.Header().Add("X-Status-URL", statusURL)

		w.WriteHeader(http.StatusAccepted)

		fmt.Fprintf(w, `
    <div hx-trigger="done" hx-get="%s" hx-swap="outerHTML" hx-target="this">
      <h3 role="status" id="pblabel" tabindex="-1" autofocus>Accepted, Running Operation</h3>
      <div hx-trigger="every 1s" hx-swap="none" hx-get="%s/tick"></div>
    </div>`, statusURL, statusURL)
	}
}

func uploadMidiProcessor(ch chan channel, wg *sync.WaitGroup) {
	_ = wg
	for c := range ch {
		id := c.requestID
		file := c.file
		header := c.header
		trackNo := c.trackNo
		timingStrategyStr := c.timingStrategy
		statusURL := c.statusURL
		defer file.Close()

		re := regexp.MustCompile(`(?i)^.*\.(mid|midi)$`)
		fileName := header.Filename

		if !re.MatchString(header.Filename) {
			storeStatus(id, JobStatus{
				State:   "ERRORED",
				Message: "Not a midi file.",
				JobURL:  statusURL,
			})
			return
		}

		// Extract monophonic melody using the new fonspeak_midi package
		notes, err := fonspeak_midi.ExtractMonophonicMelody(file, trackNo)
		if err != nil {
			storeStatus(id, JobStatus{
				State:   "ERRORED",
				Message: fmt.Sprintf("Failed to extract melody: %v", err),
				JobURL:  statusURL,
			})
			return
		}

		if len(notes) == 0 {
			storeStatus(id, JobStatus{
				State:   "ERRORED",
				Message: "No notes found in the specified track",
				JobURL:  statusURL,
			})
			return
		}

		// Apply global octave cap (max 500 Hz)
		const maxHz = 500.0
		maxFreq := fonspeak_midi.FindMaxFrequency(notes)
		octaveDrop := fonspeak_midi.ComputeGlobalOctaveDropFromHz(maxFreq, maxHz)

		// Align notes to syllables (157 syllables for Adon Olam)
		syllableCount := len(syllables)
		var alignedNotes []fonspeak_midi.Note
		var alignedSyllables []string

		if syllableCount > len(notes) {
			// Repeat melody to cover all syllables
			alignedNotes = fonspeak_midi.RepeatMelodyToCoverSyllables(notes, syllableCount)
			alignedSyllables = syllables
		} else {
			// Use all notes and extend syllables if needed
			alignedNotes = notes
			alignedSyllables = fonspeak_midi.AlignSyllablesToMelody(syllables, len(notes))
		}

		// Apply timing strategy to compute phoneme durations
		var timingStrat timing.TimingStrategy
		switch timingStrategyStr {
		case "last-phoneme":
			timingStrat = timing.LastPhoneme
		case "per-syllable":
			timingStrat = timing.PerSyllable
		default:
			// Default to per-syllable if invalid
			timingStrat = timing.PerSyllable
		}
		
		// Set up timing options
		timingOpts := timing.DefaultTimingOptions()
		timingOpts.Strategy = timingStrat
		
		// Prepare notes with syllables for timing allocation
		notesWithSyllables := timing.PrepareNotesWithSyllables(alignedNotes, alignedSyllables)
		
		// Allocate durations using the timing module
		notesWithSyllables = timing.AllocateDurations(notesWithSyllables, timingOpts)

		// Build syllable parameters for fonspeak
		syllableList := make([]fonspeak.Params, 0, len(alignedNotes))

		for i, nws := range notesWithSyllables {
			// Convert MIDI note to Hz with global octave drop
			pitchHz := fonspeak_midi.MIDINoteToHz(nws.Note.MIDINote, -octaveDrop)

			// Calculate WPM from phoneme durations (computed by timing module)
			wpm := timing.ComputeWPMFromPhonemes(nws.Syllables)
			
			// If no syllables, fall back to default
			if len(nws.Syllables) == 0 {
				wpm = fonspeak_midi.WPMFromDuration(nws.Note.Duration)
			}

			syllableList = append(syllableList, fonspeak.Params{
				Syllable:   alignedSyllables[i],
				PitchShift: pitchHz,
				Voice:      "he",
				Wpm:        wpm,
			})
		}

		// Synthesize speech
		var buf bytes.Buffer
		w := &BufWriteCloser{
			Writer: bufio.NewWriter(&buf),
		}
		defer w.Close()

		err = fonspeak.FonspeakPhrase(fonspeak.PhraseParams{
			Syllables: syllableList,
			WavFile:   w,
		}, 15)

		if err != nil {
			storeStatus(id, JobStatus{
				State:   "ERRORED",
				Message: err.Error(),
				JobURL:  statusURL,
			})
			return
		}

		uri, err := uploadWav(buf.Bytes(), fileName)
		if err != nil {
			storeStatus(id, JobStatus{
				State:   "ERRORED",
				Message: err.Error(),
				JobURL:  statusURL,
			})
			return
		}

		storeStatus(id, JobStatus{
			State:   "COMPLETED",
			Message: fmt.Sprintf("<audio controls><source src='%s' type='audio/wave' /></audio>", uri),
			JobURL:  statusURL,
		})
	}
}

func uploadWav(b []byte, fileName string) (string, error) {
	bucket := os.Getenv("MINIO_DEFAULT_BUCKETS")
	endpoint := os.Getenv("MINIO_ENDPOINT")
	secure := os.Getenv("MINIO_SECURE") == "true"
	log.Println(bucket)
	log.Println(endpoint)
	object := fileName + ".wav"
	log.Println(object)
	file := bytes.NewReader(b)
	ctx := context.Background()
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewEnvMinio(),
		Secure: secure,
	})
	if err != nil {
		return "", fmt.Errorf("minio.New: %w", err)
	}

	found, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return "", fmt.Errorf("bucket doesn't exist: %w", err)
	}

	if !found {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return "", err
		}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	_, err = client.PutObject(ctx, bucket, object, file, file.Size(), minio.PutObjectOptions{ContentType: "audio/wav"})
	if err != nil {
		return "", err
	}

	uri, err := client.PresignedGetObject(ctx, bucket, object, time.Hour, url.Values{"ContentType": {"audio/wav"}})
	if err != nil {
		return "", err
	}

	fmt.Printf("File %s uploaded successfully", fileName)
	return uri.String(), err
}
