# Adon Olam Tune Generator

A TTS singer of Adon Olam to the tune of a provided MIDI file and track. Uses [fonspeak](https://github.com/sammyshear/fonspeak) under the hood to synthesize at precise pitches.

## Features

- Reads MIDI files and extracts monophonic melodies
- Collapses polyphonic tracks to monophonic by selecting the lowest pitch
- Applies global octave transposition to keep pitches within synthesizable range
- Aligns IPA syllables to musical notes
- Synthesizes speech with precise pitch control using fonspeak

## Usage

### Web Interface

Run the web server:
```bash
go run cmd/main.go
```

Then navigate to http://localhost:8080 to upload MIDI files and generate speech.

### Command Line Interface

The repository includes a CLI tool for MIDI-driven speech synthesis:

```bash
# Build the CLI tool
go build -o bin/fonspeak_midi_driver ./cmd/fonspeak_midi_driver

# Run with default settings
./bin/fonspeak_midi_driver -midi melody.mid -lyrics examples/adon_olam_xsampa.txt -out output.wav

# Specify track number and other options
./bin/fonspeak_midi_driver -midi melody.mid -lyrics examples/adon_olam_xsampa.txt -track 1 -voice he -maxhz 500 -out result.wav
```

#### CLI Flags

- `-midi` (required): Path to MIDI file
- `-lyrics` (required): Path to lyrics text file with space-separated syllables in X-SAMPA format
- `-out`: Output WAV file path (default: "output.wav")
- `-voice`: Voice to use for synthesis (default: "he")
- `-maxhz`: Maximum frequency cap in Hz (default: 500)
- `-track`: MIDI track number to use (default: 0)

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
4. **Syllable Alignment**: 
   - If more syllables than notes: repeats the melody to cover all syllables
   - If more notes than syllables: duplicates the last syllable (melisma)
5. **Synthesis**: Calls fonspeak for each note with precise pitch (Hz) and WPM calculated from note duration

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
