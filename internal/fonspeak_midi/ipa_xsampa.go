package fonspeak_midi

import (
	"strings"
)

// IPAToXSAMPA parses X-SAMPA text into individual syllables
// The input should be space-separated syllables already in X-SAMPA format
// The function name is kept for compatibility with the original design,
// but it primarily splits the text into individual syllables
func IPAToXSAMPA(text string) []string {
	syllables := strings.Fields(text)
	return syllables
}
