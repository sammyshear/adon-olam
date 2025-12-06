package fonspeak_midi

import (
	"math"
)

// Note represents a musical note with pitch (MIDI number) and duration in seconds
type Note struct {
	MIDINote int     // MIDI note number (0-127)
	Duration float64 // Duration in seconds
}

// MIDINoteToHz converts a MIDI note number to frequency in Hz with optional octave shift
// Uses A4 = 440 Hz as reference (MIDI note 69)
// octaveShift: negative values shift down, positive values shift up (in octaves)
func MIDINoteToHz(midiNote int, octaveShift int) float64 {
	// Apply octave shift (each octave is 12 semitones)
	adjustedNote := midiNote + (octaveShift * 12)

	// Convert to Hz using equal temperament formula: f = 440 * 2^((n-69)/12)
	return 440.0 * math.Pow(2.0, float64(adjustedNote-69)/12.0)
}

// ComputeGlobalOctaveDropFromHz calculates how many octaves to drop
// to ensure the highest frequency is below or equal to capHz
func ComputeGlobalOctaveDropFromHz(maxHz float64, capHz float64) int {
	if maxHz <= capHz {
		return 0
	}

	// Calculate how many octaves we need to drop
	// Each octave divides frequency by 2
	octaveDrop := 0
	for maxHz > capHz {
		maxHz /= 2.0
		octaveDrop++
	}

	return octaveDrop
}

// FindMaxFrequency finds the highest frequency in a list of notes
func FindMaxFrequency(notes []Note) float64 {
	if len(notes) == 0 {
		return 0
	}

	maxHz := 0.0
	for _, note := range notes {
		hz := MIDINoteToHz(note.MIDINote, 0)
		if hz > maxHz {
			maxHz = hz
		}
	}

	return maxHz
}

// WPMFromDuration calculates words per minute from a note duration
// Assumes each syllable is roughly equivalent to a "word" for speech synthesis
// Duration is in seconds
// Returns WPM clamped between 30 and 300
func WPMFromDuration(seconds float64) int {
	if seconds <= 0 {
		return 160 // default reasonable rate
	}

	// WPM = 60 / duration_in_seconds (since one syllable per note)
	wpm := int(60.0 / seconds)

	// Clamp to reasonable range
	if wpm < 30 {
		wpm = 30
	}
	if wpm > 300 {
		wpm = 300
	}

	return wpm
}

// RepeatMelodyToCoverSyllables repeats the melody sequence until it covers all syllables
func RepeatMelodyToCoverSyllables(notes []Note, syllableCount int) []Note {
	if len(notes) == 0 {
		return notes
	}

	if len(notes) >= syllableCount {
		return notes[:syllableCount]
	}

	result := make([]Note, 0, syllableCount)
	for len(result) < syllableCount {
		for _, note := range notes {
			result = append(result, note)
			if len(result) >= syllableCount {
				break
			}
		}
	}

	return result
}

// AlignSyllablesToMelody aligns syllables to notes
// If there are more notes than syllables, distributes syllables evenly across notes
// Each syllable can be mapped to multiple consecutive notes (melisma)
// If there are more syllables than notes, the melody should have been repeated already
func AlignSyllablesToMelody(syllables []string, noteCount int) []string {
	if len(syllables) == 0 {
		return syllables
	}

	if len(syllables) >= noteCount {
		return syllables[:noteCount]
	}

	// Distribute syllables evenly across notes
	// Each syllable gets roughly noteCount/syllableCount notes
	result := make([]string, noteCount)
	
	notesPerSyllable := float64(noteCount) / float64(len(syllables))
	
	for i := 0; i < noteCount; i++ {
		// Determine which syllable this note belongs to
		syllableIdx := int(float64(i) / notesPerSyllable)
		if syllableIdx >= len(syllables) {
			syllableIdx = len(syllables) - 1
		}
		result[i] = syllables[syllableIdx]
	}

	return result
}
