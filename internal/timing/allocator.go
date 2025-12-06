package timing

import (
	"math"
	
	"github.com/sammyshear/adon-olam/internal/fonspeak_midi"
)

// AllocateDurations computes phoneme durations for note-syllable pairs
// based on the specified timing strategy
func AllocateDurations(notesWithSyllables []NoteWithSyllables, opts TimingOptions) []NoteWithSyllables {
	if opts.Strategy == LastPhoneme {
		return allocateLastPhoneme(notesWithSyllables, opts)
	}
	return allocatePerSyllable(notesWithSyllables, opts)
}

// allocateLastPhoneme implements the legacy strategy: put all extra time in the last phoneme
func allocateLastPhoneme(notesWithSyllables []NoteWithSyllables, opts TimingOptions) []NoteWithSyllables {
	result := make([]NoteWithSyllables, len(notesWithSyllables))
	
	for i, nws := range notesWithSyllables {
		result[i] = nws
		
		if len(nws.Syllables) == 0 {
			continue
		}
		
		// Calculate total base duration
		totalBaseDur := 0.0
		phonemeCount := 0
		for _, syl := range nws.Syllables {
			phonemeCount += len(syl.Phonemes)
		}
		
		// Set base durations for all phonemes
		for j := range result[i].Syllables {
			for k := range result[i].Syllables[j].Phonemes {
				result[i].Syllables[j].Phonemes[k].BaseDuration = opts.BasePhoneme
				result[i].Syllables[j].Phonemes[k].Duration = opts.BasePhoneme
				totalBaseDur += opts.BasePhoneme
			}
		}
		
		// Put all leftover time in the last phoneme
		noteDuration := nws.Note.Duration
		if noteDuration > totalBaseDur && phonemeCount > 0 {
			leftover := noteDuration - totalBaseDur
			lastSylIdx := len(result[i].Syllables) - 1
			lastPhonIdx := len(result[i].Syllables[lastSylIdx].Phonemes) - 1
			
			if lastPhonIdx >= 0 {
				result[i].Syllables[lastSylIdx].Phonemes[lastPhonIdx].Duration += leftover
			}
		}
	}
	
	return result
}

// allocatePerSyllable implements the intelligent per-syllable strategy
// that distributes duration across vowels preferentially
func allocatePerSyllable(notesWithSyllables []NoteWithSyllables, opts TimingOptions) []NoteWithSyllables {
	result := make([]NoteWithSyllables, len(notesWithSyllables))
	
	for i, nws := range notesWithSyllables {
		result[i] = nws
		
		if len(nws.Syllables) == 0 {
			continue
		}
		
		noteDuration := nws.Note.Duration
		
		// Count vowels and consonants
		vowelCount := 0
		consonantCount := 0
		for _, syl := range nws.Syllables {
			for _, ph := range syl.Phonemes {
				if ph.Kind == Vowel {
					vowelCount++
				} else {
					consonantCount++
				}
			}
		}
		
		totalPhonemes := vowelCount + consonantCount
		if totalPhonemes == 0 {
			continue
		}
		
		// Set base durations
		for j := range result[i].Syllables {
			for k := range result[i].Syllables[j].Phonemes {
				if result[i].Syllables[j].Phonemes[k].Kind == Vowel {
					result[i].Syllables[j].Phonemes[k].BaseDuration = opts.MinVowelDur
					result[i].Syllables[j].Phonemes[k].Duration = opts.MinVowelDur
				} else {
					result[i].Syllables[j].Phonemes[k].BaseDuration = opts.MinConsonant
					result[i].Syllables[j].Phonemes[k].Duration = opts.MinConsonant
				}
			}
		}
		
		// Calculate total base duration
		totalBaseDur := float64(vowelCount)*opts.MinVowelDur + float64(consonantCount)*opts.MinConsonant
		
		// Distribute remaining duration
		if noteDuration > totalBaseDur {
			remaining := noteDuration - totalBaseDur
			
			// First pass: distribute to vowels proportionally
			if vowelCount > 0 {
				vowelAllocation := remaining
				
				// If we need to reserve some for consonants, do so
				// But prioritize vowels
				perVowel := vowelAllocation / float64(vowelCount)
				
				for j := range result[i].Syllables {
					for k := range result[i].Syllables[j].Phonemes {
						if result[i].Syllables[j].Phonemes[k].Kind == Vowel {
							newDur := result[i].Syllables[j].Phonemes[k].Duration + perVowel
							// Clamp to max vowel duration
							newDur = math.Min(newDur, opts.MaxVowelDur)
							result[i].Syllables[j].Phonemes[k].Duration = newDur
						}
					}
				}
				
				// Recalculate actual allocated duration
				actualAllocated := 0.0
				for j := range result[i].Syllables {
					for k := range result[i].Syllables[j].Phonemes {
						actualAllocated += result[i].Syllables[j].Phonemes[k].Duration
					}
				}
				
				// If there's still leftover (because vowels hit max), distribute to consonants
				if noteDuration > actualAllocated && consonantCount > 0 {
					remaining = noteDuration - actualAllocated
					perConsonant := remaining / float64(consonantCount)
					
					for j := range result[i].Syllables {
						for k := range result[i].Syllables[j].Phonemes {
							if result[i].Syllables[j].Phonemes[k].Kind == Consonant {
								newDur := result[i].Syllables[j].Phonemes[k].Duration + perConsonant
								// Clamp to max consonant duration
								newDur = math.Min(newDur, opts.MaxConsonant)
								result[i].Syllables[j].Phonemes[k].Duration = newDur
							}
						}
					}
				}
			} else {
				// No vowels, distribute to consonants
				if consonantCount > 0 {
					perConsonant := remaining / float64(consonantCount)
					
					for j := range result[i].Syllables {
						for k := range result[i].Syllables[j].Phonemes {
							newDur := result[i].Syllables[j].Phonemes[k].Duration + perConsonant
							newDur = math.Min(newDur, opts.MaxConsonant)
							result[i].Syllables[j].Phonemes[k].Duration = newDur
						}
					}
				}
			}
		} else if noteDuration < totalBaseDur {
			// Note is shorter than minimum durations - scale everything down proportionally
			scaleFactor := noteDuration / totalBaseDur
			for j := range result[i].Syllables {
				for k := range result[i].Syllables[j].Phonemes {
					result[i].Syllables[j].Phonemes[k].Duration *= scaleFactor
				}
			}
		}
	}
	
	return result
}

// ComputeWPMFromPhonemes calculates an appropriate WPM based on total phoneme duration
// This is used to maintain compatibility with fonspeak's WPM-based interface
func ComputeWPMFromPhonemes(syllables []Syllable) int {
	if len(syllables) == 0 {
		return 160 // default
	}
	
	totalDuration := 0.0
	for _, syl := range syllables {
		for _, ph := range syl.Phonemes {
			totalDuration += ph.Duration
		}
	}
	
	if totalDuration <= 0 {
		return 160
	}
	
	// WPM = 60 / duration_per_syllable
	// For multiple syllables, use total duration
	wpm := int(60.0 * float64(len(syllables)) / totalDuration)
	
	// Clamp to reasonable range
	if wpm < 30 {
		wpm = 30
	}
	if wpm > 300 {
		wpm = 300
	}
	
	return wpm
}

// PrepareNotesWithSyllables converts raw notes and syllable strings into NoteWithSyllables
// This handles the alignment of syllables to notes
func PrepareNotesWithSyllables(notes []fonspeak_midi.Note, syllableTexts []string) []NoteWithSyllables {
	if len(notes) == 0 || len(syllableTexts) == 0 {
		return []NoteWithSyllables{}
	}
	
	result := make([]NoteWithSyllables, len(notes))
	
	for i, note := range notes {
		// Map one syllable to one note (for now)
		// In the future, this could be more sophisticated
		if i < len(syllableTexts) {
			syl := ParseSyllable(syllableTexts[i])
			result[i] = NoteWithSyllables{
				Note:      note,
				Syllables: []Syllable{syl},
			}
		} else {
			// No more syllables, use empty
			result[i] = NoteWithSyllables{
				Note:      note,
				Syllables: []Syllable{},
			}
		}
	}
	
	return result
}
