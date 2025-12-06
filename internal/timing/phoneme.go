package timing

import (
	"strings"
	"unicode"
)

// isVowel checks if a character is a vowel (including X-SAMPA vowel symbols)
func isVowel(r rune) bool {
	// Common vowels in X-SAMPA and IPA
	vowels := "aeiouæɑɔəɛɪʊʌAEIOUyY@69"
	return strings.ContainsRune(vowels, r)
}

// ClassifyPhonemeKind determines if a phoneme text represents a vowel or consonant
func ClassifyPhonemeKind(phonemeText string) PhonemeKind {
	if phonemeText == "" {
		return Unknown
	}
	
	// Check if the phoneme starts with a vowel character
	// This is a heuristic that works for most X-SAMPA representations
	for _, r := range phonemeText {
		if unicode.IsLetter(r) || r == '@' {
			if isVowel(r) {
				return Vowel
			}
			// If we encounter a non-vowel letter first, it's likely a consonant
			return Consonant
		}
	}
	
	return Unknown
}

// ParseSyllableToPhonemes breaks a syllable into individual phonemes
// This is a simplified implementation that treats each character as a phoneme
// For production, you might want more sophisticated parsing
func ParseSyllableToPhonemes(syllableText string) []Phoneme {
	if syllableText == "" {
		return []Phoneme{}
	}
	
	phonemes := []Phoneme{}
	
	// Simple approach: split into character-level phonemes
	// For X-SAMPA, some phonemes are multi-character (e.g., "aI", "eI")
	// but for simplicity, we'll handle common cases
	
	i := 0
	for i < len(syllableText) {
		// Check for common multi-character X-SAMPA combinations
		var phonemeText string
		
		// Try two-character combinations first
		if i+1 < len(syllableText) {
			twoChar := syllableText[i:i+2]
			// Common X-SAMPA digraphs: aI, eI, OI, aU, @U, etc.
			if isMultiCharPhoneme(twoChar) {
				phonemeText = twoChar
				i += 2
			} else {
				phonemeText = string(syllableText[i])
				i++
			}
		} else {
			phonemeText = string(syllableText[i])
			i++
		}
		
		kind := ClassifyPhonemeKind(phonemeText)
		phonemes = append(phonemes, Phoneme{
			Text:         phonemeText,
			Kind:         kind,
			BaseDuration: 0, // Will be set later
			Duration:     0, // Will be computed
		})
	}
	
	return phonemes
}

// isMultiCharPhoneme checks if a two-character string is a known multi-character phoneme
func isMultiCharPhoneme(s string) bool {
	// Common X-SAMPA multi-character phonemes
	multiChar := []string{
		"aI", "eI", "OI", "aU", "@U", "I@", "E@", "U@", // diphthongs
		"tS", "dZ", "er", // affricates and r-colored vowels
	}
	
	for _, mc := range multiChar {
		if strings.EqualFold(s, mc) {
			return true
		}
	}
	
	return false
}

// ParseSyllable converts a syllable string into a Syllable structure with phonemes
func ParseSyllable(syllableText string) Syllable {
	return Syllable{
		Text:     syllableText,
		Phonemes: ParseSyllableToPhonemes(syllableText),
	}
}
