package fonspeak_midi

import (
	"math"
	"testing"
)

func TestMIDINoteToHz(t *testing.T) {
	tests := []struct {
		name        string
		midiNote    int
		octaveShift int
		want        float64
	}{
		{"A4 (440Hz)", 69, 0, 440.0},
		{"A4 down 1 octave", 69, -1, 220.0},
		{"A4 down 2 octaves", 69, -2, 110.0},
		{"C4 (middle C)", 60, 0, 261.63},
		{"C5", 72, 0, 523.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MIDINoteToHz(tt.midiNote, tt.octaveShift)
			if math.Abs(got-tt.want) > 0.1 {
				t.Errorf("MIDINoteToHz(%d, %d) = %.2f, want %.2f", tt.midiNote, tt.octaveShift, got, tt.want)
			}
		})
	}
}

func TestComputeGlobalOctaveDropFromHz(t *testing.T) {
	tests := []struct {
		name  string
		maxHz float64
		capHz float64
		want  int
	}{
		{"No drop needed", 440.0, 500.0, 0},
		{"One octave drop", 880.0, 500.0, 1},
		{"Two octave drop", 1760.0, 500.0, 2},
		{"Already below cap", 300.0, 500.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeGlobalOctaveDropFromHz(tt.maxHz, tt.capHz)
			if got != tt.want {
				t.Errorf("ComputeGlobalOctaveDropFromHz(%.2f, %.2f) = %d, want %d", tt.maxHz, tt.capHz, got, tt.want)
			}
		})
	}
}

func TestWPMFromDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
		want     int
	}{
		{"0.5 second note", 0.5, 120},
		{"1 second note", 1.0, 60},
		{"0.2 second note", 0.2, 300}, // clamped to 300
		{"2 second note", 2.0, 30},    // clamped to 30
		{"very fast", 0.1, 300},       // clamped
		{"very slow", 5.0, 30},        // clamped
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WPMFromDuration(tt.duration)
			if got != tt.want {
				t.Errorf("WPMFromDuration(%.2f) = %d, want %d", tt.duration, got, tt.want)
			}
		})
	}
}

func TestRepeatMelodyToCoverSyllables(t *testing.T) {
	notes := []Note{
		{MIDINote: 60, Duration: 0.5},
		{MIDINote: 62, Duration: 0.5},
		{MIDINote: 64, Duration: 0.5},
	}

	tests := []struct {
		name          string
		syllableCount int
		wantLen       int
	}{
		{"Exact match", 3, 3},
		{"Need repetition", 7, 7},
		{"Fewer syllables", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RepeatMelodyToCoverSyllables(notes, tt.syllableCount)
			if len(got) != tt.wantLen {
				t.Errorf("RepeatMelodyToCoverSyllables() returned %d notes, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestAlignSyllablesToMelody(t *testing.T) {
	syllables := []string{"a", "b", "c"}

	tests := []struct {
		name      string
		noteCount int
		wantLen   int
		wantLast  string
	}{
		{"Exact match", 3, 3, "c"},
		{"More notes (melisma)", 5, 5, "c"}, // last syllable repeated
		{"Fewer notes", 2, 2, "b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlignSyllablesToMelody(syllables, tt.noteCount)
			if len(got) != tt.wantLen {
				t.Errorf("AlignSyllablesToMelody() returned %d syllables, want %d", len(got), tt.wantLen)
			}
			if got[len(got)-1] != tt.wantLast {
				t.Errorf("Last syllable = %s, want %s", got[len(got)-1], tt.wantLast)
			}
		})
	}
}

func TestIPAToXSAMPA(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{"Single syllable", "a", 1},
		{"Multiple syllables", "a don o l@m", 4},
		{"Empty string", "", 0},
		{"Multiple spaces", "a  b   c", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IPAToXSAMPA(tt.input)
			if len(got) != tt.wantLen {
				t.Errorf("IPAToXSAMPA(%q) returned %d syllables, want %d", tt.input, len(got), tt.wantLen)
			}
		})
	}
}
