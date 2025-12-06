package timing

import "github.com/sammyshear/adon-olam/internal/fonspeak_midi"

// TimingStrategy defines how phoneme durations are allocated
type TimingStrategy string

const (
	// PerSyllable distributes duration across syllables, prioritizing vowels
	PerSyllable TimingStrategy = "per-syllable"
	// LastPhoneme puts all extra duration in the last phoneme (legacy behavior)
	LastPhoneme TimingStrategy = "last-phoneme"
)

// PhonemeKind classifies phoneme types
type PhonemeKind int

const (
	Consonant PhonemeKind = iota
	Vowel
	Unknown
)

// Phoneme represents a single phoneme with its properties
type Phoneme struct {
	Text         string      // The phoneme text (X-SAMPA format)
	Kind         PhonemeKind // Consonant or Vowel
	BaseDuration float64     // Initial/minimum duration in seconds
	Duration     float64     // Allocated duration in seconds (will be computed)
}

// Syllable represents a syllable broken into onset, nucleus, and coda
type Syllable struct {
	Text    string    // Original syllable text
	Phonemes []Phoneme // All phonemes in the syllable
	// For simple implementation, we don't strictly separate onset/nucleus/coda
	// but identify vowels for lengthening
}

// TimingOptions configures the timing allocation algorithm
type TimingOptions struct {
	Strategy      TimingStrategy
	MinVowelDur   float64 // Minimum vowel duration in seconds
	MaxVowelDur   float64 // Maximum vowel duration in seconds
	MinConsonant  float64 // Minimum consonant duration in seconds
	MaxConsonant  float64 // Maximum consonant duration in seconds
	BasePhoneme   float64 // Base phoneme duration in seconds
}

// DefaultTimingOptions returns sensible defaults
func DefaultTimingOptions() TimingOptions {
	return TimingOptions{
		Strategy:     PerSyllable,
		MinVowelDur:  0.05,  // 50ms minimum for vowels
		MaxVowelDur:  1.0,   // 1 second maximum for vowels
		MinConsonant: 0.03,  // 30ms minimum for consonants
		MaxConsonant: 0.2,   // 200ms maximum for consonants
		BasePhoneme:  0.08,  // 80ms base duration per phoneme
	}
}

// NoteWithSyllables pairs a MIDI note with the syllables that should be sung/spoken
// during that note's duration. Used by the timing allocation algorithm to compute
// phoneme-level durations.
type NoteWithSyllables struct {
	Note      fonspeak_midi.Note
	Syllables []Syllable
}
