package fonspeak_midi

import (
	"strings"
)

// XSAMPAToSyllables parses X-SAMPA text into individual syllables
// The input should be space-separated syllables in X-SAMPA notation
var ipaToXSAMPAMap = map[string]string{
	"a":  "a",
	"o":  "o",
	"e":  "e",
	"i":  "i",
	"u":  "u",
	"@":  "@", // schwa
	"ɔ":  "O",
	"ə":  "@",
	"ɛ":  "E",
	"ɪ":  "I",
	"ʊ":  "U",
	"ɑ":  "A",
	"æ":  "{",
	"ʌ":  "V",
	"ɒ":  "Q",
	"ː":  ":", // long vowel marker
	"ʃ":  "S", // sh
	"ʒ":  "Z", // zh
	"tʃ": "tS",
	"dʒ": "dZ",
	"θ":  "T", // th in thin
	"ð":  "D", // th in this
	"ŋ":  "N", // ng
	"χ":  "X", // voiceless uvular fricative (Hebrew het)
	"ʔ":  "?", // glottal stop
	"j":  "j", // y sound
	"w":  "w",
	"r":  "r",
	"l":  "l",
	"m":  "m",
	"n":  "n",
	"p":  "p",
	"b":  "b",
	"t":  "t",
	"d":  "d",
	"k":  "k",
	"g":  "g",
	"f":  "f",
	"v":  "v",
	"s":  "s",
	"z":  "z",
	"h":  "h",
	"S":  "S", // already X-SAMPA
	"X":  "X", // already X-SAMPA
}

// IPAToXSAMPA parses X-SAMPA text into individual syllables
// The input should be space-separated syllables already in X-SAMPA format
// The function name is kept for compatibility, but it primarily splits the text
func IPAToXSAMPA(text string) []string {
	syllables := strings.Fields(text)
	return syllables
}

// ConvertIPAChar converts a single IPA character/sequence to X-SAMPA
func ConvertIPAChar(ipa string) string {
	if mapped, ok := ipaToXSAMPAMap[ipa]; ok {
		return mapped
	}
	// Return as-is if not in map
	return ipa
}
