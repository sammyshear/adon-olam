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
	midi.Note(midi.C(2)):  32.703,
	midi.Note(midi.Db(2)): 34.648,
	midi.Note(midi.D(2)):  36.708,
	midi.Note(midi.Eb(2)): 38.891,
	midi.Note(midi.E(2)):  41.203,
	midi.Note(midi.F(2)):  43.654,
	midi.Note(midi.Gb(2)): 46.249,
	midi.Note(midi.G(2)):  48.999,
	midi.Note(midi.Ab(2)): 51.913,
	midi.Note(midi.A(2)):  55,
	midi.Note(midi.Bb(2)): 58.27,
	midi.Note(midi.B(2)):  61.735,
	midi.Note(midi.C(3)):  32.703,
	midi.Note(midi.Db(3)): 34.648,
	midi.Note(midi.D(3)):  36.708,
	midi.Note(midi.Eb(3)): 38.891,
	midi.Note(midi.E(3)):  41.203,
	midi.Note(midi.F(3)):  43.654,
	midi.Note(midi.Gb(3)): 46.249,
	midi.Note(midi.G(3)):  48.999,
	midi.Note(midi.Ab(3)): 51.913,
	midi.Note(midi.A(3)):  55,
	midi.Note(midi.Bb(3)): 58.27,
	midi.Note(midi.B(3)):  61.735,
	midi.Note(midi.C(4)):  32.703,
	midi.Note(midi.Db(4)): 34.648,
	midi.Note(midi.D(4)):  36.708,
	midi.Note(midi.Eb(4)): 38.891,
	midi.Note(midi.E(4)):  41.203,
	midi.Note(midi.F(4)):  43.654,
	midi.Note(midi.Gb(4)): 46.249,
	midi.Note(midi.G(4)):  48.999,
	midi.Note(midi.Ab(4)): 51.913,
	midi.Note(midi.A(4)):  55,
	midi.Note(midi.Bb(4)): 58.27,
	midi.Note(midi.B(4)):  61.735,
	midi.Note(midi.C(5)):  32.703,
	midi.Note(midi.Db(5)): 34.648,
	midi.Note(midi.D(5)):  36.708,
	midi.Note(midi.Eb(5)): 38.891,
	midi.Note(midi.E(5)):  41.203,
	midi.Note(midi.F(5)):  43.654,
	midi.Note(midi.Gb(5)): 46.249,
	midi.Note(midi.G(5)):  48.999,
	midi.Note(midi.Ab(5)): 51.913,
	midi.Note(midi.A(5)):  55,
	midi.Note(midi.Bb(5)): 58.27,
	midi.Note(midi.B(5)):  61.735,
	midi.Note(midi.C(6)):  32.703,
	midi.Note(midi.Db(6)): 34.648,
	midi.Note(midi.D(6)):  36.708,
	midi.Note(midi.Eb(6)): 38.891,
	midi.Note(midi.E(6)):  41.203,
	midi.Note(midi.F(6)):  43.654,
	midi.Note(midi.Gb(6)): 46.249,
	midi.Note(midi.G(6)):  48.999,
	midi.Note(midi.Ab(6)): 51.913,
	midi.Note(midi.A(6)):  55,
	midi.Note(midi.Bb(6)): 58.27,
	midi.Note(midi.B(6)):  61.735,
	midi.Note(midi.C(7)):  32.703,
	midi.Note(midi.Db(7)): 34.648,
	midi.Note(midi.D(7)):  36.708,
	midi.Note(midi.Eb(7)): 38.891,
	midi.Note(midi.E(7)):  41.203,
	midi.Note(midi.F(7)):  43.654,
	midi.Note(midi.Gb(7)): 46.249,
	midi.Note(midi.G(7)):  48.999,
	midi.Note(midi.Ab(7)): 51.913,
	midi.Note(midi.A(7)):  55,
	midi.Note(midi.Bb(7)): 58.27,
	midi.Note(midi.B(7)):  61.735,
}

const (
	ipa = "a.don. o.lam. a.ʃeʁ. ma.lax. be.teʁ.em. kol. je.tsiʁ. niv.ʁa. le.et. naa.sah. vex.ef.tso .kol. a.zai .melex. ʃe.mo. nik.ʁa. ve.ax.a.ʁei. kix.lot. hak.kol. le.vad.do. jim.lox. no.ʁa. ve.hu ha.jah. ve.hu. ho.veh. ve.hu. jih.jeh. bet.if.a.ʁah. ve.hu. e.xad. ve.ein. ʃe.ni. le.ham.ʃil. lo. lehax.bi.ʁah. be.li. ʁe.ʃit. be.li. tax.lit. ve.lo. ha.oz. ve.ham.mis.ʁah. ve.hu. e.li. ve.xai. go.a.li. ve.tsuʁ. xev.li. be.et. tsa.ʁah. ve.hu. nis.si. u.ma.nos. li. me.nat. ko.si. be.jom. ek.ʁa. beja.do. af.kid. ʁu.xi. beet. i.ʃan. ve.a.i.ʁah. ve.im .ʁu.xi. ge.vij.ja.ti. ad.o.nai. li. ve.lo. i.ʁa."
	t   = `
אֲד.וֹן. עוֹ.לָם. אֲ.שֶׁר. מָ.לַךְ.
בְּ.טֶ.רֶם. כָּל. יְ.צִיר. נִבְ.רָא.
לְ.עֵת. נַעֲ.שָׂה. בְחֶ.פְצוֹ. כֹּל.
אֲ.זַי. מֶ.לֶךְ. שְׁ.מוֹ. נִ.קְרָא.
וְ.אַ.חֲ.רֵי. כִּכְ.לוֹת. הַ.כֹּל.
לְ.בַדּ.וֹ. יִמְ.לוֹךְ. נוֹ.רָא.
וְ.הוּא. הָ.יָה. וְ.הוּא. הֹ.וֶה.
וְ.הוּא. יִהְ.יֶה. בְּ.תִ.פְאָ.רָה.
וְ.הוּא. אֶ.חָד. וְ.אֵין. שֵׁ.נִי.
לְ.הַמְ.שִׁיל. לוֹ. לְ.הַחְ.בִּי.רָה.
בְּלִי. רֵא.שִׁית. בְּלִי. תַכְ.לִית.
וְ.לוֹ. הָ.עֹז. וְ.הַמִּשְׂ.רָה.
וְ.הוּא. אֵ.לִי. וְ.חַי. גֹּ.אֲלִי.
וְ.צוּר. חֶ.בְלִי. בְּ.עֵת. צָ.רָה.
וְ.הוּא. נִ.סִּ.י. וּ.מָנ.וֹס. לִי.
מְ.נָת. כּוֹ.סִי. בְּ.יוֹם. אֶ.קְרָא.
בְּ.יָד.וֹ. אַפְ.קִיד. רוּ.חִי.
בְּ.עֵת. אִי.שַׁן. וְ.אָעִי.רָה.
וְ.עִם. רוּ.חִי. גְּוִ.יָּ.תִי.
יְיָ. לִי. וְ.לֹא. אִי.רָא`
)

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

	text := strings.Split(strings.ReplaceAll(ipa, " ", ""), ".")
	text2 := strings.Split(strings.ReplaceAll(strings.ReplaceAll(t, "\n", " "), " ", ""), ".")

	ssml := `<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xmlns:mstts='http://www.w3.org/2001/mstts' xml:lang='he-IL'><voice name='he-IL-AvriNeural'>`
	noteStart := time.Now()
	restStart := time.Now()
	i := 0
	for _, message := range messages {
		if i < len(text2) {
			if message.Is(midi.NoteOnMsg) {
				var channel, key, velocity uint8
				message.GetNoteOn(&channel, &key, &velocity)
				if velocity > 0 {
					noteStart = time.Now()
					restDur := time.Since(restStart)
					ssml += fmt.Sprintf("<mstts:silence type='tailing' value='%dms'/>", restDur.Abs().Milliseconds())
				} else {
					restStart = time.Now()
					noteDur := time.Since(noteStart)
					length := "medium"
					if noteDur.Microseconds() < 2 {
						length = "x-fast"
					} else if noteDur.Microseconds() > 2 && noteDur.Microseconds() < 5 {
						length = "fast"
					} else if noteDur.Seconds() > 5 && noteDur.Microseconds() < 8 {
						length = "medium"
					} else if noteDur.Microseconds() > 8 && noteDur.Microseconds() < 12 {
						length = "slow"
					} else if noteDur.Seconds() > 12 {
						length = "x-slow"
					}
					ssml += fmt.Sprintf(`<prosody pitch='%fHz' rate='%s'><phoneme alphabet='ipa' ph='%s'>%s</phoneme></prosody>`, hertzTable[midi.Note(key)], length, text[i], text2[i])
					i++
				}
			}
		}
	}
	ssml += `</voice></speak>`

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

	// o = o.If(storage.Conditions{DoesNotExist: true})

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
