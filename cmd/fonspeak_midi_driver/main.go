package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sammyshear/adon-olam/internal/fonspeak_midi"
	"github.com/sammyshear/fonspeak"
)

func main() {
	// Define command-line flags
	midiPath := flag.String("midi", "", "Path to MIDI file (required)")
	ipaPath := flag.String("lyrics", "", "Path to lyrics text file with X-SAMPA syllables (required)")
	outPath := flag.String("out", "output.wav", "Output WAV file path")
	voice := flag.String("voice", "he", "Voice to use for synthesis (default: he)")
	maxHz := flag.Float64("maxhz", 500.0, "Maximum frequency cap in Hz (default: 500)")
	trackNo := flag.Int("track", 0, "MIDI track number to use (default: 0)")

	flag.Parse()

	// Validate required flags
	if *midiPath == "" || *ipaPath == "" {
		flag.Usage()
		log.Fatal("Error: Both -midi and -lyrics flags are required")
	}

	// Run the synthesis pipeline
	if err := runSynthesis(*midiPath, *ipaPath, *outPath, *voice, *maxHz, *trackNo); err != nil {
		log.Fatalf("Synthesis failed: %v", err)
	}

	fmt.Printf("Successfully generated speech to %s\n", *outPath)
}

func runSynthesis(midiPath, ipaPath, outPath, voice string, maxHz float64, trackNo int) error {
	// 1. Read MIDI file and extract monophonic melody
	fmt.Println("Reading MIDI file...")
	midiFile, err := os.Open(midiPath)
	if err != nil {
		return fmt.Errorf("failed to open MIDI file: %w", err)
	}
	defer midiFile.Close()

	notes, err := fonspeak_midi.ExtractMonophonicMelody(midiFile, trackNo)
	if err != nil {
		return fmt.Errorf("failed to extract melody: %w", err)
	}

	fmt.Printf("Extracted %d notes from MIDI track %d\n", len(notes), trackNo)

	// 2. Read X-SAMPA lyrics
	fmt.Println("Reading X-SAMPA lyrics...")
	lyricsFile, err := os.Open(ipaPath)
	if err != nil {
		return fmt.Errorf("failed to open lyrics file: %w", err)
	}
	defer lyricsFile.Close()

	lyricsContent, err := io.ReadAll(lyricsFile)
	if err != nil {
		return fmt.Errorf("failed to read lyrics file: %w", err)
	}

	// 3. Parse X-SAMPA syllables (space-separated)
	syllables := fonspeak_midi.IPAToXSAMPA(string(lyricsContent))
	fmt.Printf("Loaded %d syllables\n", len(syllables))

	if len(syllables) == 0 {
		return fmt.Errorf("no syllables found in lyrics file")
	}

	// 4. Compute global octave drop to cap maximum frequency
	maxFreq := fonspeak_midi.FindMaxFrequency(notes)
	octaveDrop := fonspeak_midi.ComputeGlobalOctaveDropFromHz(maxFreq, maxHz)

	if octaveDrop > 0 {
		fmt.Printf("Original max frequency: %.2f Hz\n", maxFreq)
		fmt.Printf("Applying global octave drop: %d octaves\n", octaveDrop)
		newMaxFreq := maxFreq / float64(uint(1)<<uint(octaveDrop))
		fmt.Printf("New max frequency: %.2f Hz\n", newMaxFreq)
	} else {
		fmt.Printf("Max frequency: %.2f Hz (no octave drop needed)\n", maxFreq)
	}

	// 5. Align syllables to melody
	// If more syllables than notes, repeat melody
	// If more notes than syllables, duplicate last syllable
	var alignedNotes []fonspeak_midi.Note
	var alignedSyllables []string

	if len(syllables) > len(notes) {
		// Repeat melody to cover all syllables
		alignedNotes = fonspeak_midi.RepeatMelodyToCoverSyllables(notes, len(syllables))
		alignedSyllables = syllables
		fmt.Printf("Repeated melody to match %d syllables\n", len(syllables))
	} else {
		// Use all notes and extend syllables if needed
		alignedNotes = notes
		alignedSyllables = fonspeak_midi.AlignSyllablesToMelody(syllables, len(notes))
		if len(alignedSyllables) > len(syllables) {
			fmt.Printf("Extended syllables with melisma (last syllable repeated %d times)\n",
				len(alignedSyllables)-len(syllables))
		}
	}

	fmt.Printf("Aligned to %d note-syllable pairs\n", len(alignedNotes))

	// 6. Build syllable parameters for fonspeak
	syllableList := make([]fonspeak.Params, 0, len(alignedNotes))

	for i, note := range alignedNotes {
		// Convert MIDI note to Hz with global octave drop
		pitchHz := fonspeak_midi.MIDINoteToHz(note.MIDINote, -octaveDrop)

		// Calculate WPM from duration
		wpm := fonspeak_midi.WPMFromDuration(note.Duration)

		syllableList = append(syllableList, fonspeak.Params{
			Syllable:   alignedSyllables[i],
			PitchShift: pitchHz,
			Voice:      voice,
			Wpm:        wpm,
		})
	}

	// 7. Synthesize speech with fonspeak
	fmt.Println("Synthesizing speech...")

	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	err = fonspeak.FonspeakPhrase(fonspeak.PhraseParams{
		Syllables: syllableList,
		WavFile:   &bufWriteCloser{Writer: bufWriter},
	}, 15)

	if err != nil {
		return fmt.Errorf("fonspeak synthesis failed: %w", err)
	}

	// 8. Write output file
	fmt.Println("Writing output file...")
	err = os.WriteFile(outPath, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// bufWriteCloser wraps bufio.Writer to implement WriteCloser
type bufWriteCloser struct {
	*bufio.Writer
}

func (bwc *bufWriteCloser) Close() error {
	return bwc.Flush()
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of fonspeak_midi_driver:\n")
		fmt.Fprintf(os.Stderr, "\nGenerates speech synthesis of Adon Olam lyrics to a MIDI melody.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  fonspeak_midi_driver -midi melody.mid -lyrics adon_olam_xsampa.txt -out output.wav\n")
		fmt.Fprintf(os.Stderr, "  fonspeak_midi_driver -midi melody.mid -lyrics adon_olam_xsampa.txt -track 1 -voice he -maxhz 500 -out result.wav\n")
	}
}
