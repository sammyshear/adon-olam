package fonspeak_midi

import (
	"math"
	"strings"
	"unicode"
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

// isVowelChar checks if a character is a vowel (including X-SAMPA vowel symbols)
func isVowelChar(r rune) bool {
	vowels := "aeiouæɑɔəɛɪʊʌAEIOUyY@69"
	return strings.ContainsRune(vowels, r)
}

// isVowelPhoneme checks if a phoneme string represents a vowel
func isVowelPhoneme(phoneme string) bool {
	if phoneme == "" {
		return false
	}
	for _, r := range phoneme {
		if unicode.IsLetter(r) || r == '@' {
			return isVowelChar(r)
		}
	}
	return false
}

// splitSyllable splits a syllable into onset (consonants before vowel), 
// nucleus (vowel), and coda (consonants after vowel)
// Returns onset, nucleus, coda strings
func splitSyllable(syllable string) (string, string, string) {
	if syllable == "" {
		return "", "", ""
	}
	
	// Find the first vowel position
	vowelStart := -1
	vowelEnd := -1
	
	for i, r := range syllable {
		if isVowelChar(r) {
			if vowelStart == -1 {
				vowelStart = i
			}
			vowelEnd = i + 1
		} else if vowelStart != -1 {
			// Found consonant after vowel, stop
			break
		}
	}
	
	// If no vowel found, treat entire syllable as onset
	if vowelStart == -1 {
		return syllable, "", ""
	}
	
	onset := syllable[:vowelStart]
	nucleus := syllable[vowelStart:vowelEnd]
	coda := syllable[vowelEnd:]
	
	return onset, nucleus, coda
}

// extendSyllableVowel extends a syllable by duplicating its vowel nucleus
// For example: "don" with count=5 becomes ["d", "o", "o", "o", "on"]
func extendSyllableVowel(syllable string, count int) []string {
	if count <= 1 {
		return []string{syllable}
	}
	
	onset, nucleus, coda := splitSyllable(syllable)
	
	// If no vowel found, just repeat the syllable
	if nucleus == "" {
		result := make([]string, count)
		for i := range result {
			result[i] = syllable
		}
		return result
	}
	
	result := make([]string, count)
	
	// First syllable: onset + vowel
	if onset != "" {
		result[0] = onset + nucleus
	} else {
		result[0] = nucleus
	}
	
	// Middle syllables: just the vowel
	for i := 1; i < count-1; i++ {
		result[i] = nucleus
	}
	
	// Last syllable: vowel + coda
	if coda != "" {
		result[count-1] = nucleus + coda
	} else {
		result[count-1] = nucleus
	}
	
	return result
}

// AlignSyllablesToMelody aligns syllables to notes
// If there are more notes than syllables, distributes syllables evenly across notes
// and extends each syllable's vowel across its assigned notes (melisma)
// If there are more syllables than notes, the melody should have been repeated already
func AlignSyllablesToMelody(syllables []string, noteCount int) []string {
	if len(syllables) == 0 {
		return syllables
	}

	if len(syllables) >= noteCount {
		return syllables[:noteCount]
	}

	// Distribute syllables evenly across notes with vowel extension
	result := make([]string, noteCount)
	
	notesPerSyllable := float64(noteCount) / float64(len(syllables))
	
	// For each syllable, determine how many notes it gets and extend its vowel
	noteIdx := 0
	for sylIdx, syllable := range syllables {
		// Calculate how many notes this syllable should span
		startNote := int(float64(sylIdx) * notesPerSyllable)
		endNote := int(float64(sylIdx+1) * notesPerSyllable)
		
		// Handle rounding for the last syllable
		if sylIdx == len(syllables)-1 {
			endNote = noteCount
		}
		
		notesForThisSyllable := endNote - startNote
		if notesForThisSyllable < 1 {
			notesForThisSyllable = 1
		}
		
		// Extend the syllable's vowel across the notes
		extendedSyllables := extendSyllableVowel(syllable, notesForThisSyllable)
		
		// Place the extended syllables into the result
		for _, extSyl := range extendedSyllables {
			if noteIdx < noteCount {
				result[noteIdx] = extSyl
				noteIdx++
			}
		}
	}

	return result
}
