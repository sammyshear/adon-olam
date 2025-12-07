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

// ExtendSyllableVowel extends a syllable by duplicating its vowel nucleus
// For example: "don" with count=5 becomes ["d", "o", "o", "o", "on"]
// This creates multiple syllable variations where the vowel is progressively repeated
func ExtendSyllableVowel(syllableText string, count int) []string {
	if count <= 1 {
		return []string{syllableText}
	}
	
	phonemes := ParseSyllableToPhonemes(syllableText)
	if len(phonemes) == 0 {
		return []string{syllableText}
	}
	
	// Find the first vowel (nucleus) position
	vowelIdx := -1
	for i, ph := range phonemes {
		if ph.Kind == Vowel {
			vowelIdx = i
			break
		}
	}
	
	// If no vowel found, just repeat the syllable as-is
	if vowelIdx == -1 {
		result := make([]string, count)
		for i := range result {
			result[i] = syllableText
		}
		return result
	}
	
	// Split into onset (before vowel), nucleus (vowel), and coda (after vowel)
	var onset, nucleus, coda string
	
	// Build onset (consonants before the vowel)
	for i := 0; i < vowelIdx; i++ {
		onset += phonemes[i].Text
	}
	
	// Get the vowel
	nucleus = phonemes[vowelIdx].Text
	
	// Build coda (consonants after the vowel)
	for i := vowelIdx + 1; i < len(phonemes); i++ {
		coda += phonemes[i].Text
	}
	
	// Generate extended syllables
	result := make([]string, count)
	
	// First syllable: onset + vowel (no coda yet)
	if onset != "" {
		result[0] = onset + nucleus
	} else {
		result[0] = nucleus
	}
	
	// Middle syllables: just the vowel repeated
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
