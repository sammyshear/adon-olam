# Adon Olam Tune Generator

A TTS singer of Adon Olam to the tune of a provided MIDI file and track. Uses [fonspeak](https://github.com/sammyshear/fonspeak) under the hood to synthesize at precise pitches.

## Features

- Reads MIDI files and extracts monophonic melodies
- Collapses polyphonic tracks to monophonic by selecting the lowest pitch
- Applies global octave transposition to keep pitches within synthesizable range
- Aligns IPA syllables to musical notes
- **Intelligent syllable-aware phoneme timing** that distributes note durations naturally across syllables
- Synthesizes speech with precise pitch control using fonspeak

## Usage

### Web Interface

Run the web server:
```bash
go run cmd/main.go
```

Then navigate to http://localhost:8080 to upload MIDI files and generate speech.

**Timing Strategy:** The web interface includes a dropdown to select the timing strategy:
- **Per-Syllable (Recommended)**: Intelligently distributes note duration across syllables, prioritizing vowel lengthening for more natural-sounding speech
- **Last-Phoneme (Legacy)**: Places all extra duration at the end of the last phoneme

### Command Line Interface

The repository includes a CLI tool for MIDI-driven speech synthesis:

```bash
# Build the CLI tool
go build -o bin/fonspeak_midi_driver ./cmd/fonspeak_midi_driver

# Run with default settings (uses per-syllable timing strategy)
./bin/fonspeak_midi_driver -midi melody.mid -lyrics examples/adon_olam_xsampa.txt -out output.wav

# Specify track number and other options
./bin/fonspeak_midi_driver -midi melody.mid -lyrics examples/adon_olam_xsampa.txt -track 1 -voice he -maxhz 500 -out result.wav

# Use legacy timing strategy
./bin/fonspeak_midi_driver -midi melody.mid -lyrics examples/adon_olam_xsampa.txt -timing-strategy last-phoneme -out legacy.wav
```

#### CLI Flags

- `-midi` (required): Path to MIDI file
- `-lyrics` (required): Path to lyrics text file with space-separated syllables in X-SAMPA format
- `-out`: Output WAV file path (default: "output.wav")
- `-voice`: Voice to use for synthesis (default: "he")
- `-maxhz`: Maximum frequency cap in Hz (default: 500)
- `-track`: MIDI track number to use (default: 0)
- `-timing-strategy`: Timing strategy for phoneme duration allocation (default: "per-syllable")
  - `per-syllable`: Intelligently distributes duration across syllables, prioritizing vowel lengthening (recommended)
  - `last-phoneme`: Legacy behavior that puts extra duration in the last phoneme

#### Lyrics Text Format

The lyrics text file should contain space-separated syllables in X-SAMPA format. An example file is provided at `examples/adon_olam_xsampa.txt`.

Example content:
```
a don o l@m aS er ma laX b@ ter em kol je tsir niv ra...
```

### How It Works

1. **MIDI Reading**: Extracts notes from the specified MIDI track, including pitch (MIDI note number) and duration
2. **Monophonic Collapse**: If multiple notes occur simultaneously (chords), selects the lowest pitch
3. **Global Octave Cap**: Calculates the highest pitch in the melody and applies octave transposition (down) so the highest pitch is ≤ maxhz (default 500 Hz)
4. **Syllable Alignment & Vowel Extension**: 
   - If more syllables than notes: repeats the melody to cover all syllables
   - If more notes than syllables: distributes syllables evenly across notes with **vowel-only extension**
   - When a syllable spans multiple notes (melisma), only the vowel nucleus is duplicated, preserving consonants at boundaries
   - Example: "don" over 5 notes becomes ["do", "o", "o", "o", "on"] (d-o-o-o-on), not ["don", "don", "don", "don", "don"]
5. **Intelligent Timing Allocation** (new):
   - Breaks each syllable into phonemes (consonants and vowels)
   - Distributes the MIDI note duration across the syllable's phonemes
   - Prioritizes lengthening vowels to create more natural-sounding speech
   - Respects minimum and maximum duration bounds for different phoneme types
6. **Synthesis**: Calls fonspeak for each note with precise pitch (Hz) and WPM calculated from the intelligent timing allocation

### Timing Strategies Explained

#### Per-Syllable Strategy (Default)

The per-syllable timing strategy intelligently distributes MIDI note durations across syllables to create more natural-sounding speech:

- **Vowel-Only Extension**: When a syllable spans multiple notes, only the vowel nucleus is duplicated while consonants remain at syllable boundaries
- **Natural Distribution**: Syllables are distributed evenly across notes with intelligent vowel extension for melisma
- **Bounds Enforcement**: Respects minimum and maximum duration constraints for vowels (50ms-1000ms) and consonants (30ms-200ms)
- **Smart Handling**: Handles edge cases like syllables without clear vowels, extremely short or long notes

**Example 1**: For a 1-second note mapped to syllable "ba":
- Legacy approach: "b" gets base duration (~80ms), "a" gets all remaining time (~920ms)
- Per-syllable approach: "b" gets minimum (~30ms), "a" gets the rest distributed naturally (~970ms)

**Example 2**: Syllable "don" over 5 notes (vowel-only extension):
- Result: ["do", "o", "o", "o", "on"] (d-o-o-o-on)
- Onset "d" stays with first note, vowel "o" repeated in middle, coda "n" with last note
- Consonants are NOT duplicated, only the vowel is extended

**Example 3**: Two syllables ["a", "don"] over 6 notes:
- Notes 1-3: ["a", "a", "a"] (vowel "a" extended)
- Notes 4-6: ["do", "o", "on"] (vowel "o" extended, consonants at boundaries)

#### Last-Phoneme Strategy (Legacy)

This maintains backward compatibility with the original behavior:
- All phonemes get a base duration (80ms by default)
- Any remaining time is added entirely to the last phoneme
- Simple but can sound unnatural with long notes

## Development

### Prerequisites

- Go 1.24+
- Praat (for fonspeak pitch analysis)

### Build

```bash
# Install dependencies
go mod download

# Build web server
go build -o bin/app ./cmd/main.go

# Build CLI tool
go build -o bin/fonspeak_midi_driver ./cmd/fonspeak_midi_driver
```

### Test

```bash
# Run with a sample MIDI file (you need to provide your own)
./bin/fonspeak_midi_driver -midi your_melody.mid -lyrics examples/adon_olam_xsampa.txt -out test.wav
```
