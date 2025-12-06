package timing

import (
	"math"
	"testing"
	
	"github.com/sammyshear/adon-olam/internal/fonspeak_midi"
)

func TestClassifyPhonemeKind(t *testing.T) {
	tests := []struct {
		name     string
		phoneme  string
		wantKind PhonemeKind
	}{
		{"vowel a", "a", Vowel},
		{"vowel e", "e", Vowel},
		{"vowel i", "i", Vowel},
		{"vowel o", "o", Vowel},
		{"vowel u", "u", Vowel},
		{"vowel @", "@", Vowel},
		{"consonant b", "b", Consonant},
		{"consonant d", "d", Consonant},
		{"consonant n", "n", Consonant},
		{"consonant S", "S", Consonant},
		{"consonant X", "X", Consonant},
		{"empty string", "", Unknown},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyPhonemeKind(tt.phoneme)
			if got != tt.wantKind {
				t.Errorf("ClassifyPhonemeKind(%q) = %v, want %v", tt.phoneme, got, tt.wantKind)
			}
		})
	}
}

func TestParseSyllableToPhonemes(t *testing.T) {
	tests := []struct {
		name        string
		syllable    string
		wantCount   int
		checkFirst  PhonemeKind // Kind of first phoneme
	}{
		{"single vowel", "a", 1, Vowel},
		{"consonant-vowel", "ba", 2, Consonant},
		{"vowel-consonant", "ab", 2, Vowel},
		{"complex syllable", "don", 3, Consonant},
		{"xsampa with @", "l@m", 3, Consonant},
		{"empty string", "", 0, Unknown},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSyllableToPhonemes(tt.syllable)
			if len(got) != tt.wantCount {
				t.Errorf("ParseSyllableToPhonemes(%q) returned %d phonemes, want %d", tt.syllable, len(got), tt.wantCount)
			}
			if len(got) > 0 && tt.wantCount > 0 && got[0].Kind != tt.checkFirst {
				t.Errorf("First phoneme kind = %v, want %v", got[0].Kind, tt.checkFirst)
			}
		})
	}
}

func TestParseSyllable(t *testing.T) {
	syl := ParseSyllable("don")
	if syl.Text != "don" {
		t.Errorf("Syllable text = %q, want %q", syl.Text, "don")
	}
	if len(syl.Phonemes) == 0 {
		t.Error("Expected phonemes to be parsed")
	}
}

func TestAllocateLastPhoneme(t *testing.T) {
	// Create a simple note with syllables
	note := fonspeak_midi.Note{
		MIDINote: 60,
		Duration: 1.0, // 1 second
	}
	
	syl := ParseSyllable("ba") // consonant + vowel = 2 phonemes
	nws := []NoteWithSyllables{
		{
			Note:      note,
			Syllables: []Syllable{syl},
		},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = LastPhoneme
	
	result := AllocateDurations(nws, opts)
	
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	
	if len(result[0].Syllables) != 1 {
		t.Fatalf("Expected 1 syllable, got %d", len(result[0].Syllables))
	}
	
	phonemes := result[0].Syllables[0].Phonemes
	if len(phonemes) != 2 {
		t.Fatalf("Expected 2 phonemes, got %d", len(phonemes))
	}
	
	// Check that last phoneme has most of the duration
	lastPhoneme := phonemes[len(phonemes)-1]
	if lastPhoneme.Duration <= opts.BasePhoneme {
		t.Errorf("Last phoneme duration = %.3f, expected it to be > %.3f (base)", lastPhoneme.Duration, opts.BasePhoneme)
	}
	
	// Check total duration approximately matches note duration
	totalDur := 0.0
	for _, ph := range phonemes {
		totalDur += ph.Duration
	}
	
	if math.Abs(totalDur-note.Duration) > 0.01 {
		t.Errorf("Total duration = %.3f, want %.3f", totalDur, note.Duration)
	}
}

func TestAllocatePerSyllable_SingleVowel(t *testing.T) {
	// Single note, single syllable with a vowel
	note := fonspeak_midi.Note{
		MIDINote: 60,
		Duration: 0.5, // 500ms
	}
	
	syl := ParseSyllable("a") // single vowel
	nws := []NoteWithSyllables{
		{
			Note:      note,
			Syllables: []Syllable{syl},
		},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = PerSyllable
	
	result := AllocateDurations(nws, opts)
	
	if len(result[0].Syllables[0].Phonemes) != 1 {
		t.Fatalf("Expected 1 phoneme")
	}
	
	phoneme := result[0].Syllables[0].Phonemes[0]
	
	// Vowel should be lengthened to cover most of the note duration
	if phoneme.Kind != Vowel {
		t.Error("Expected phoneme to be a vowel")
	}
	
	// Should be close to note duration (allowing for some margin)
	if math.Abs(phoneme.Duration-note.Duration) > 0.1 {
		t.Errorf("Vowel duration = %.3f, expected close to note duration %.3f", phoneme.Duration, note.Duration)
	}
}

func TestAllocatePerSyllable_ConsonantVowel(t *testing.T) {
	// Single note with consonant-vowel syllable
	note := fonspeak_midi.Note{
		MIDINote: 60,
		Duration: 0.6, // 600ms
	}
	
	syl := ParseSyllable("ba") // consonant + vowel
	nws := []NoteWithSyllables{
		{
			Note:      note,
			Syllables: []Syllable{syl},
		},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = PerSyllable
	
	result := AllocateDurations(nws, opts)
	
	phonemes := result[0].Syllables[0].Phonemes
	if len(phonemes) != 2 {
		t.Fatalf("Expected 2 phonemes, got %d", len(phonemes))
	}
	
	consonant := phonemes[0]
	vowel := phonemes[1]
	
	// Vowel should be longer than consonant
	if vowel.Duration <= consonant.Duration {
		t.Errorf("Vowel duration %.3f should be > consonant duration %.3f", vowel.Duration, consonant.Duration)
	}
	
	// Total should match note duration
	totalDur := consonant.Duration + vowel.Duration
	if math.Abs(totalDur-note.Duration) > 0.01 {
		t.Errorf("Total duration = %.3f, want %.3f", totalDur, note.Duration)
	}
}

func TestAllocatePerSyllable_MultiSyllable(t *testing.T) {
	// One note with multiple syllables (each should get proportional time)
	note := fonspeak_midi.Note{
		MIDINote: 60,
		Duration: 1.0, // 1 second
	}
	
	syl1 := ParseSyllable("ba")
	syl2 := ParseSyllable("na")
	nws := []NoteWithSyllables{
		{
			Note:      note,
			Syllables: []Syllable{syl1, syl2},
		},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = PerSyllable
	
	result := AllocateDurations(nws, opts)
	
	// Calculate total duration across both syllables
	totalDur := 0.0
	vowelDur := 0.0
	for _, syl := range result[0].Syllables {
		for _, ph := range syl.Phonemes {
			totalDur += ph.Duration
			if ph.Kind == Vowel {
				vowelDur += ph.Duration
			}
		}
	}
	
	// Total should approximately match note duration
	if math.Abs(totalDur-note.Duration) > 0.01 {
		t.Errorf("Total duration = %.3f, want %.3f", totalDur, note.Duration)
	}
	
	// Vowels should have gotten most of the extra duration
	// We have 2 vowels and 2 consonants
	// With min vowel 0.05 and min consonant 0.03: total base = 2*0.05 + 2*0.03 = 0.16
	// Extra = 1.0 - 0.16 = 0.84, should go mostly to vowels
	if vowelDur < 0.5 {
		t.Errorf("Vowel total duration = %.3f, expected most of the note duration", vowelDur)
	}
}

func TestAllocatePerSyllable_MultipleNotes(t *testing.T) {
	// Multiple notes with syllables
	note1 := fonspeak_midi.Note{MIDINote: 60, Duration: 0.5}
	note2 := fonspeak_midi.Note{MIDINote: 62, Duration: 0.8}
	
	syl1 := ParseSyllable("a")
	syl2 := ParseSyllable("don")
	
	nws := []NoteWithSyllables{
		{Note: note1, Syllables: []Syllable{syl1}},
		{Note: note2, Syllables: []Syllable{syl2}},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = PerSyllable
	
	result := AllocateDurations(nws, opts)
	
	// Check first note
	dur1 := 0.0
	for _, ph := range result[0].Syllables[0].Phonemes {
		dur1 += ph.Duration
	}
	
	if math.Abs(dur1-note1.Duration) > 0.01 {
		t.Errorf("Note 1 total duration = %.3f, want %.3f", dur1, note1.Duration)
	}
	
	// Check second note
	dur2 := 0.0
	for _, ph := range result[1].Syllables[0].Phonemes {
		dur2 += ph.Duration
	}
	
	if math.Abs(dur2-note2.Duration) > 0.01 {
		t.Errorf("Note 2 total duration = %.3f, want %.3f", dur2, note2.Duration)
	}
}

func TestAllocatePerSyllable_NoVowels(t *testing.T) {
	// Edge case: syllable with no clear vowels (all consonants)
	note := fonspeak_midi.Note{
		MIDINote: 60,
		Duration: 0.3,
	}
	
	// Create a syllable manually with only consonants
	syl := Syllable{
		Text: "bcd",
		Phonemes: []Phoneme{
			{Text: "b", Kind: Consonant},
			{Text: "c", Kind: Consonant},
			{Text: "d", Kind: Consonant},
		},
	}
	
	nws := []NoteWithSyllables{
		{
			Note:      note,
			Syllables: []Syllable{syl},
		},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = PerSyllable
	
	result := AllocateDurations(nws, opts)
	
	// Should distribute among consonants
	totalDur := 0.0
	for _, ph := range result[0].Syllables[0].Phonemes {
		totalDur += ph.Duration
	}
	
	if math.Abs(totalDur-note.Duration) > 0.01 {
		t.Errorf("Total duration = %.3f, want %.3f", totalDur, note.Duration)
	}
}

func TestAllocatePerSyllable_VeryShortNote(t *testing.T) {
	// Edge case: note shorter than minimum phoneme durations
	note := fonspeak_midi.Note{
		MIDINote: 60,
		Duration: 0.04, // 40ms, less than min vowel (50ms)
	}
	
	syl := ParseSyllable("a")
	nws := []NoteWithSyllables{
		{
			Note:      note,
			Syllables: []Syllable{syl},
		},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = PerSyllable
	
	result := AllocateDurations(nws, opts)
	
	totalDur := 0.0
	for _, ph := range result[0].Syllables[0].Phonemes {
		totalDur += ph.Duration
	}
	
	// Should scale down to match note duration
	if math.Abs(totalDur-note.Duration) > 0.001 {
		t.Errorf("Total duration = %.3f, want %.3f", totalDur, note.Duration)
	}
}

func TestAllocatePerSyllable_VeryLongNote(t *testing.T) {
	// Edge case: very long note duration
	note := fonspeak_midi.Note{
		MIDINote: 60,
		Duration: 3.0, // 3 seconds
	}
	
	syl := ParseSyllable("a")
	nws := []NoteWithSyllables{
		{
			Note:      note,
			Syllables: []Syllable{syl},
		},
	}
	
	opts := DefaultTimingOptions()
	opts.Strategy = PerSyllable
	
	result := AllocateDurations(nws, opts)
	
	vowelDur := result[0].Syllables[0].Phonemes[0].Duration
	
	// Should be clamped to MaxVowelDur
	if vowelDur > opts.MaxVowelDur {
		t.Errorf("Vowel duration = %.3f, should be clamped to max %.3f", vowelDur, opts.MaxVowelDur)
	}
}

func TestComputeWPMFromPhonemes(t *testing.T) {
	// Create syllables with known durations
	syl := Syllable{
		Text: "ba",
		Phonemes: []Phoneme{
			{Text: "b", Kind: Consonant, Duration: 0.1},
			{Text: "a", Kind: Vowel, Duration: 0.4},
		},
	}
	
	// Total duration = 0.5 seconds for 1 syllable
	// WPM = 60 * 1 / 0.5 = 120
	wpm := ComputeWPMFromPhonemes([]Syllable{syl})
	
	if wpm != 120 {
		t.Errorf("ComputeWPMFromPhonemes() = %d, want 120", wpm)
	}
}

func TestComputeWPMFromPhonemes_Empty(t *testing.T) {
	wpm := ComputeWPMFromPhonemes([]Syllable{})
	
	if wpm != 160 {
		t.Errorf("ComputeWPMFromPhonemes(empty) = %d, want default 160", wpm)
	}
}

func TestPrepareNotesWithSyllables(t *testing.T) {
	notes := []fonspeak_midi.Note{
		{MIDINote: 60, Duration: 0.5},
		{MIDINote: 62, Duration: 0.6},
	}
	
	syllables := []string{"a", "don"}
	
	result := PrepareNotesWithSyllables(notes, syllables)
	
	if len(result) != 2 {
		t.Fatalf("Expected 2 note-syllable pairs, got %d", len(result))
	}
	
	if result[0].Syllables[0].Text != "a" {
		t.Errorf("First syllable text = %q, want %q", result[0].Syllables[0].Text, "a")
	}
	
	if result[1].Syllables[0].Text != "don" {
		t.Errorf("Second syllable text = %q, want %q", result[1].Syllables[0].Text, "don")
	}
}

func TestPrepareNotesWithSyllables_MoreNotesThanSyllables(t *testing.T) {
	notes := []fonspeak_midi.Note{
		{MIDINote: 60, Duration: 0.5},
		{MIDINote: 62, Duration: 0.6},
		{MIDINote: 64, Duration: 0.7},
	}
	
	syllables := []string{"a"}
	
	result := PrepareNotesWithSyllables(notes, syllables)
	
	if len(result) != 3 {
		t.Fatalf("Expected 3 note-syllable pairs, got %d", len(result))
	}
	
	// First should have syllable
	if len(result[0].Syllables) == 0 || result[0].Syllables[0].Text != "a" {
		t.Error("First note should have syllable 'a'")
	}
	
	// Others should have empty syllables
	if len(result[1].Syllables) != 0 {
		t.Error("Second note should have no syllables")
	}
	
	if len(result[2].Syllables) != 0 {
		t.Error("Third note should have no syllables")
	}
}
