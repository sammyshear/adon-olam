package fonspeak_midi

import (
	"fmt"
	"io"
	"sort"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// ExtractMonophonicMelody reads a MIDI file and extracts a monophonic melody
// from the specified track. If multiple notes occur simultaneously (chord),
// it selects the lowest pitch.
func ExtractMonophonicMelody(reader io.Reader, trackNo int) ([]Note, error) {
	var events []smf.TrackEvent

	// Read all track events from the MIDI file
	err := smf.ReadTracksFrom(reader).Do(func(te smf.TrackEvent) {
		events = append(events, te)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read MIDI tracks: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no MIDI events found")
	}

	// Group events by time to detect simultaneous notes (chords)
	type noteEvent struct {
		time     uint32
		key      uint8
		duration uint32
	}

	noteEvents := []noteEvent{}

	// Build a map of note-on to note-off for duration calculation
	for i, te := range events {
		if te.TrackNo == trackNo && te.Message.Is(midi.NoteOnMsg) {
			var channel, key, velocity uint8
			te.Message.GetNoteOn(&channel, &key, &velocity)

			// Skip if velocity is 0 (some MIDI files use note-on with velocity 0 as note-off)
			if velocity == 0 {
				continue
			}

			noteOnTime := te.AbsMicroSeconds

			// Search for corresponding NoteOff
			for j := i + 1; j < len(events); j++ {
				te2 := events[j]
				if te2.TrackNo == trackNo && te2.AbsMicroSeconds > noteOnTime {
					var matched bool

					// Check for NoteOff message
					if te2.Message.Is(midi.NoteOffMsg) {
						var channel2, keyOff, velocity2 uint8
						te2.Message.GetNoteOff(&channel2, &keyOff, &velocity2)
						if keyOff == key {
							matched = true
						}
					}

					// Also check for NoteOn with velocity 0 (alternative note-off)
					if te2.Message.Is(midi.NoteOnMsg) {
						var channel2, keyOn, velocity2 uint8
						te2.Message.GetNoteOn(&channel2, &keyOn, &velocity2)
						if keyOn == key && velocity2 == 0 {
							matched = true
						}
					}

					if matched {
						duration := te2.AbsMicroSeconds - noteOnTime // in microseconds
						noteEvents = append(noteEvents, noteEvent{
							time:     uint32(noteOnTime / 1000), // convert to milliseconds
							key:      key,
							duration: uint32(duration / 1000), // convert to milliseconds
						})
						break
					}
				}
			}
		}
	}

	if len(noteEvents) == 0 {
		return nil, fmt.Errorf("no notes found in track %d", trackNo)
	}

	// Sort by time
	sort.Slice(noteEvents, func(i, j int) bool {
		return noteEvents[i].time < noteEvents[j].time
	})

	// Collapse simultaneous notes (chords) to lowest pitch
	// Group notes that start within a small time window as simultaneous
	// 10ms threshold chosen to account for MIDI timing jitter while still detecting true chords
	const simultaneousThreshold = 10 // milliseconds

	result := []Note{}
	i := 0
	for i < len(noteEvents) {
		currentTime := noteEvents[i].time

		// Collect all notes that start at approximately the same time
		simultaneousNotes := []noteEvent{noteEvents[i]}
		j := i + 1
		for j < len(noteEvents) && noteEvents[j].time-currentTime <= simultaneousThreshold {
			simultaneousNotes = append(simultaneousNotes, noteEvents[j])
			j++
		}

		// Select the lowest pitch
		lowestNote := simultaneousNotes[0]
		for _, note := range simultaneousNotes[1:] {
			if note.key < lowestNote.key {
				lowestNote = note
			}
		}

		// Use the longest duration among simultaneous notes
		maxDuration := lowestNote.duration
		for _, note := range simultaneousNotes {
			if note.duration > maxDuration {
				maxDuration = note.duration
			}
		}

		result = append(result, Note{
			MIDINote: int(lowestNote.key),
			Duration: float64(maxDuration) / 1000.0, // convert to seconds
		})

		i = j
	}

	return result, nil
}
