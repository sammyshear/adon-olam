package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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
	"github.com/sammyshear/fonspeak"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"gitlab.com/gomidi/midi/v2/smf"
)

type BufWriteCloser struct {
	*bufio.Writer
}

func (bwc *BufWriteCloser) Close() error {
	return bwc.Flush()
}

var hertzTable = map[midi.Note]float64{
	midi.Note(midi.C(1)):  32.70,
	midi.Note(midi.Db(1)): 34.65,
	midi.Note(midi.D(1)):  36.71,
	midi.Note(midi.Eb(1)): 38.89,
	midi.Note(midi.E(1)):  41.20,
	midi.Note(midi.F(1)):  43.65,
	midi.Note(midi.Gb(1)): 46.25,
	midi.Note(midi.G(1)):  49.00,
	midi.Note(midi.Ab(1)): 51.91,
	midi.Note(midi.A(1)):  55.00,
	midi.Note(midi.Bb(1)): 58.27,
	midi.Note(midi.B(1)):  61.74,
	midi.Note(midi.C(2)):  65.41,
	midi.Note(midi.Db(2)): 69.30,
	midi.Note(midi.D(2)):  73.42,
	midi.Note(midi.Eb(2)): 77.78,
	midi.Note(midi.E(2)):  82.41,
	midi.Note(midi.F(2)):  87.31,
	midi.Note(midi.Gb(2)): 92.50,
	midi.Note(midi.G(2)):  98.00,
	midi.Note(midi.Ab(2)): 103.83,
	midi.Note(midi.A(2)):  110.00,
	midi.Note(midi.Bb(2)): 116.54,
	midi.Note(midi.B(2)):  123.47,
	midi.Note(midi.C(3)):  130.81,
	midi.Note(midi.Db(3)): 138.59,
	midi.Note(midi.D(3)):  146.83,
	midi.Note(midi.Eb(3)): 155.56,
	midi.Note(midi.E(3)):  164.81,
	midi.Note(midi.F(3)):  174.61,
	midi.Note(midi.Gb(3)): 185.00,
	midi.Note(midi.G(3)):  196.00,
	midi.Note(midi.Ab(3)): 207.65,
	midi.Note(midi.A(3)):  220.00,
	midi.Note(midi.Bb(3)): 233.08,
	midi.Note(midi.B(3)):  246.94,
	midi.Note(midi.C(4)):  261.63,
	midi.Note(midi.Db(4)): 277.18,
	midi.Note(midi.D(4)):  293.66,
	midi.Note(midi.Eb(4)): 311.13,
	midi.Note(midi.E(4)):  329.63,
	midi.Note(midi.F(4)):  349.23,
	midi.Note(midi.Gb(4)): 369.99,
	midi.Note(midi.G(4)):  392.00,
	midi.Note(midi.Ab(4)): 415.30,
	midi.Note(midi.A(4)):  440.00,
	midi.Note(midi.Bb(4)): 466.16,
	midi.Note(midi.B(4)):  493.88,
	midi.Note(midi.C(5)):  523.25,
	midi.Note(midi.Db(5)): 554.37,
	midi.Note(midi.D(5)):  587.33,
	midi.Note(midi.Eb(5)): 622.25,
	midi.Note(midi.E(5)):  659.26,
	midi.Note(midi.F(5)):  698.46,
	midi.Note(midi.Gb(5)): 739.99,
	midi.Note(midi.G(5)):  783.99,
	midi.Note(midi.Ab(5)): 830.61,
	midi.Note(midi.A(5)):  880.00,
	midi.Note(midi.Bb(5)): 932.33,
	midi.Note(midi.B(5)):  987.77,
	midi.Note(midi.C(6)):  1046.50,
	midi.Note(midi.Db(6)): 1108.73,
	midi.Note(midi.D(6)):  1174.66,
	midi.Note(midi.Eb(6)): 1244.51,
	midi.Note(midi.E(6)):  1318.51,
	midi.Note(midi.F(6)):  1396.91,
	midi.Note(midi.Gb(6)): 1479.98,
	midi.Note(midi.G(6)):  1567.98,
	midi.Note(midi.Ab(6)): 1661.22,
	midi.Note(midi.A(6)):  1760.00,
	midi.Note(midi.Bb(6)): 1864.66,
	midi.Note(midi.B(6)):  1975.53,
	midi.Note(midi.C(7)):  2093.00,
	midi.Note(midi.Db(7)): 2217.46,
	midi.Note(midi.D(7)):  2349.32,
	midi.Note(midi.Eb(7)): 2489.02,
	midi.Note(midi.E(7)):  2637.02,
	midi.Note(midi.F(7)):  2793.83,
	midi.Note(midi.Gb(7)): 2959.96,
	midi.Note(midi.G(7)):  3135.96,
	midi.Note(midi.Ab(7)): 3322.44,
	midi.Note(midi.A(7)):  3520.00,
	midi.Note(midi.Bb(7)): 3729.31,
	midi.Note(midi.B(7)):  3951.07,
}

var syllables = []string{"a", "don", "o", "l@m", "aS", "er", "ma", "laX", "b@", "ter", "em", "kol", "je", "tsir", "niv", "ra", "l@", "et", "na:", "sa", "veX", "ef", "tso", "kol", "az", "ai", "mel", "eX", "Se", "mo", "nik", "ra", "ve", "aX", "a", "rei", "kix", "lot", "ha", "kol", "l@", "va", "do", "jim", "loX", "no", "ra", "v@", "hu", "ha", "ja", "v@", "hu", "ho", "ve", "v@", "hu", "ji", "je", "bet", "if", "ar", "a", "v@", "hu", "eX", "ad", "v@", "ein", "Se", "ni", "l@", "ham", "Sil", "lo", "l@", "haX", "bi", "ra", "bli", "re", "Sit", "bli", "taX", "lit", "v@", "lo", "ha", "oz", "v@", "ham", "mis", "rah", "v@", "hu", "el", "i", "v@", "Xai", "go", "al", "i", "v@", "tsur", "Xev", "li", "b@", "et", "tsa", "ra", "v@", "hu", "nis", "si", "u", "ma", "nos", "li", "m@", "nat", "ko", "si", "b@", "jom", "ek", "ra", "b@", "ja", "do", "af", "kid", "ru", "Xi", "b@", "et", "iS", "an", "v@", "a", "ir", "a", "v@", "im", "ru", "Xi", "g@", "vi", "ja", "ti", "ad", "on", "ai", "li", "v@", "lo", "ir", "a"}

type channel struct {
	requestID string
	file      multipart.File
	header    *multipart.FileHeader
	statusURL string
	trackNo   int
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
		ch <- channel{requestID, file, header, statusURL, trackNo}

		w.Header().Add("X-Status-URL", statusURL)

		w.WriteHeader(http.StatusAccepted)

		fmt.Fprintf(w, `
    <div hx-trigger="done" hx-get="%s" hx-swap="outerHTML" hx-target="this">
      <h3 role="status" id="pblabel" tabindex="-1" autofocus>Accepted, Running Operation</h3>
      <div hx-trigger="every 1s" hx-swap="none" hx-get="%s/tick"></div>
    </div>`, statusURL, statusURL)
	}
}

func speechMaker(pitchList []float64, w io.WriteCloser) error {
	syllableList := make([]fonspeak.Params, 0)
	for i, pitch := range pitchList {
		syllableList = append(syllableList, fonspeak.Params{Syllable: syllables[i], PitchShift: pitch, Voice: "he", Wpm: 160})
	}

	err := fonspeak.FonspeakPhrase(fonspeak.PhraseParams{
		Syllables: syllableList,
		WavFile:   w,
	}, 15)

	return err
}

func uploadMidiProcessor(ch chan channel, wg *sync.WaitGroup) {
	_ = wg
	for c := range ch {
		id := c.requestID
		file := c.file
		header := c.header
		trackNo := c.trackNo
		statusURL := c.statusURL
		defer file.Close()

		re := regexp.MustCompile(`/(?i:^.*\.(mid|midi)$)/gm`)
		fileName := header.Filename

		if re.MatchString(header.Filename) {
			storeStatus(id, JobStatus{
				State:   "ERRORED",
				Message: "Not a midi file.",
				JobUrl:  statusURL,
			})
			return
		}

		var messages []smf.Message

		smf.ReadTracksFrom(file).Do(func(te smf.TrackEvent) {
			if te.Message.IsMeta() {
				fmt.Printf("[%v] @%vms %s\n", te.TrackNo, te.AbsMicroSeconds/1000, te.Message.String())
			} else {
				if te.TrackNo == trackNo {
					if te.Message.Is(midi.NoteOnMsg) {
						messages = append(messages, te.Message)
					}
				}
			}
		})

		pitchList := make([]float64, 157)

		for len(messages) < len(pitchList) {
			messages = append(messages, messages...)
		}

		for i, message := range messages {
			var channel, key, velocity uint8
			message.GetNoteOn(&channel, &key, &velocity)
			if i < 157 {
				pitchList[i] = hertzTable[midi.Note(key)]
			} else {
				break
			}
		}

		var buf bytes.Buffer

		w := &BufWriteCloser{
			Writer: bufio.NewWriter(&buf),
		}
		defer w.Close()

		err := speechMaker(pitchList, w)
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
