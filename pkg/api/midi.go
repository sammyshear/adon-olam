package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"gitlab.com/gomidi/midi/v2/smf"
)

var hertzTable = map[midi.Note]float64{
	midi.Note(midi.C(0)):  16.351,
	midi.Note(midi.Db(0)): 17.324,
	midi.Note(midi.D(0)):  18.354,
	midi.Note(midi.Eb(0)): 19.445,
	midi.Note(midi.E(0)):  20.601,
	midi.Note(midi.F(0)):  21.827,
	midi.Note(midi.Gb(0)): 23.124,
	midi.Note(midi.G(0)):  24.499,
	midi.Note(midi.Ab(0)): 25.956,
	midi.Note(midi.A(0)):  27.5,
	midi.Note(midi.Bb(0)): 29.135,
	midi.Note(midi.B(0)):  30.868,
	midi.Note(midi.C(1)):  32.703,
	midi.Note(midi.Db(1)): 34.648,
	midi.Note(midi.D(1)):  36.708,
	midi.Note(midi.Eb(1)): 38.891,
	midi.Note(midi.E(1)):  41.203,
	midi.Note(midi.F(1)):  43.654,
	midi.Note(midi.Gb(1)): 46.249,
	midi.Note(midi.G(1)):  48.999,
	midi.Note(midi.Ab(1)): 51.913,
	midi.Note(midi.A(1)):  55,
	midi.Note(midi.Bb(1)): 58.27,
	midi.Note(midi.B(1)):  61.735,
	midi.Note(midi.C(2)):  65.406,
	midi.Note(midi.Db(2)): 69.296,
	midi.Note(midi.D(2)):  73.416,
	midi.Note(midi.Eb(2)): 77.782,
	midi.Note(midi.E(2)):  82.407,
	midi.Note(midi.F(2)):  87.307,
	midi.Note(midi.Gb(2)): 92.499,
	midi.Note(midi.G(2)):  97.999,
	midi.Note(midi.Ab(2)): 103.826,
	midi.Note(midi.A(2)):  110,
	midi.Note(midi.Bb(2)): 116.541,
	midi.Note(midi.B(2)):  123.471,
	midi.Note(midi.C(3)):  130.813,
	midi.Note(midi.Db(3)): 138.591,
	midi.Note(midi.D(3)):  146.832,
	midi.Note(midi.Eb(3)): 155.563,
	midi.Note(midi.E(3)):  164.814,
	midi.Note(midi.F(3)):  174.614,
	midi.Note(midi.Gb(3)): 184.997,
	midi.Note(midi.G(3)):  195.998,
	midi.Note(midi.Ab(3)): 207.652,
	midi.Note(midi.A(3)):  220,
	midi.Note(midi.Bb(3)): 233.082,
	midi.Note(midi.B(3)):  246.942,
	midi.Note(midi.C(4)):  261.626,
	midi.Note(midi.Db(4)): 277.183,
	midi.Note(midi.D(4)):  293.665,
	midi.Note(midi.Eb(4)): 311.127,
	midi.Note(midi.E(4)):  329.628,
	midi.Note(midi.F(4)):  349.228,
	midi.Note(midi.Gb(4)): 369.994,
	midi.Note(midi.G(4)):  391.995,
	midi.Note(midi.Ab(4)): 415.305,
	midi.Note(midi.A(4)):  440,
	midi.Note(midi.Bb(4)): 466.164,
	midi.Note(midi.B(4)):  493.883,
	midi.Note(midi.C(5)):  523.251,
	midi.Note(midi.Db(5)): 554.365,
	midi.Note(midi.D(5)):  587.33,
	midi.Note(midi.Eb(5)): 622.254,
	midi.Note(midi.E(5)):  659.255,
	midi.Note(midi.F(5)):  698.456,
	midi.Note(midi.Gb(5)): 739.989,
	midi.Note(midi.G(5)):  783.991,
	midi.Note(midi.Ab(5)): 830.609,
	midi.Note(midi.A(5)):  880,
	midi.Note(midi.Bb(5)): 932.328,
	midi.Note(midi.B(5)):  987.767,
	midi.Note(midi.C(6)):  1046.502,
	midi.Note(midi.Db(6)): 1108.731,
	midi.Note(midi.D(6)):  1174.659,
	midi.Note(midi.Eb(6)): 1244.508,
	midi.Note(midi.E(6)):  1318.51,
	midi.Note(midi.F(6)):  1396.913,
	midi.Note(midi.Gb(6)): 1479.978,
	midi.Note(midi.G(6)):  1567.982,
	midi.Note(midi.Ab(6)): 1661.219,
	midi.Note(midi.A(6)):  1760,
	midi.Note(midi.Bb(6)): 1864.655,
	midi.Note(midi.B(6)):  1975.533,
	midi.Note(midi.C(7)):  2093.005,
	midi.Note(midi.Db(7)): 2217.461,
	midi.Note(midi.D(7)):  2349.318,
	midi.Note(midi.Eb(7)): 2489.016,
	midi.Note(midi.E(7)):  2637.021,
	midi.Note(midi.F(7)):  2793.826,
	midi.Note(midi.Gb(7)): 2959.955,
	midi.Note(midi.G(7)):  3135.964,
	midi.Note(midi.Ab(7)): 3322.438,
	midi.Note(midi.A(7)):  3520,
	midi.Note(midi.Bb(7)): 3729.31,
	midi.Note(midi.B(7)):  3951.066,
}

var text = []string{"אֲד", "וֹן", "עוֹ", "לָם", "אֲ", "שֶׁר", "מָ", "לַךְ,", "בְּטֶ", "רֶם", "כָּל", "יְצִיר", "נִבְ", "רָא.", "לְעֵת", "נַעֲ", "שָׂה", "בְחֶפְ", "צוֹ", "כֹּל,", "אֲזַי", "מֶלֶךְ", "שְׁמוֹ", "נִקְ", "רָא."}

func UploadMidi(w http.ResponseWriter, r *http.Request) {
	speechKey := os.Getenv("SPEECH_KEY")
	speechRegion := os.Getenv("SPEECH_REGION")
	r.ParseMultipartForm(10 << 20) // 10 MB
	file, header, err := r.FormFile("uploadFile")
	if err != nil {
		http.Error(w, fmt.Sprintf("upload error: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	trackNo, err := strconv.Atoi(r.FormValue("trackNo"))
	if err != nil {
		http.Error(w, "Error parsing track number.", http.StatusInternalServerError)
		return
	}

	re := regexp.MustCompile(`/(?i:^.*\.(mid|midi)$)/gm`)
	fileName := header.Filename

	if re.MatchString(header.Filename) {
		http.Error(w, "Unsupported file type, please upload a midi file.", http.StatusUnsupportedMediaType)
		return
	}

	var bpm float64
	var messages []smf.Message

	smf.ReadTracksFrom(file).Do(func(te smf.TrackEvent) {
		if te.Message.IsMeta() {
			fmt.Printf("[%v] @%vms %s\n", te.TrackNo, te.AbsMicroSeconds/1000, te.Message.String())
			if te.Message.Is(smf.MetaTempoMsg) {
				te.Message.GetMetaTempo(&bpm)
			}
		} else {
			if te.TrackNo == trackNo {
				if te.Message.IsPlayable() {

					var channel, key, velocity uint8
					te.Message.GetNoteOn(&channel, &key, &velocity)
					if key > 0 {
						messages = append(messages, te.Message)
					}
				}
			}
		}
	})

	duration := len(text) * int(bpm)
	ssml := fmt.Sprintf(`<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xmlns:mstts='http://www.w3.org/2001/mstts' xml:lang='he-IL'><voice xml:lang='he-IL' xml:gender='Female' name='he-IL-HilaNeural'><mstts:audioduration value='%dms' />`, duration)
	for i, message := range messages {
		if message.IsPlayable() && i < len(text) {
			var channel, key, velocity uint8
			message.GetNoteOn(&channel, &key, &velocity)
			fmt.Printf("%d\n", key)
			ssml += fmt.Sprintf(`<prosody pitch='%fHz'>%s</prosody>`, hertzTable[midi.Note(key)], text[i])
		}
	}
	ssml += `</voice></speak>`

	fmt.Printf("%s\n", ssml)

	tokenReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s.api.cognitive.microsoft.com/sts/v1.0/issueToken", speechRegion), nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize request to get auth token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	tokenReq.Header.Add("Ocp-Apim-Subscription-Key", speechKey)

	tokenRes, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get auth token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	defer tokenRes.Body.Close()
	bSpeechAuth, err := io.ReadAll(tokenRes.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get auth token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	speechAuth := string(bSpeechAuth)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", speechRegion), strings.NewReader(ssml))
	if err != nil {
		http.Error(w, fmt.Sprintf("Can't initialize request to Azure Speech: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", speechAuth))
	req.Header.Add("Content-Type", "application/ssml+xml")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("X-Microsoft-OutputFormat", "audio-16khz-64kbitrate-mono-mp3")
	req.Header.Add("User-Agent", "adon-olam-tune-generator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to make request to Azure Speech: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if res.StatusCode != 200 {
		http.Error(w, fmt.Sprintf("Request failed with status %s", res.Status), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	bMp3, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing bytes of returned audio file: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	uploadMp3(bMp3, fileName)

	fmt.Fprintf(w, "Uploaded File %s Successfully", fileName)
}

func uploadMp3(b []byte, fileName string) error {
	bucket := "mp3_bucket_adon_olam"
	object := fileName + ".mp3"
	file := bytes.NewReader(b)
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	o := client.Bucket(bucket).Object(object)

	o = o.If(storage.Conditions{DoesNotExist: true})

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}

	fmt.Printf("File %s uploaded successfully", fileName)
	return nil
}
