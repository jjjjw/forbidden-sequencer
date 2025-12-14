package lib

// Scale represents a musical scale as semitone intervals from the root
type Scale []int

// Common scales
var (
	// MajorScale: whole, whole, half, whole, whole, whole, half
	MajorScale = Scale{0, 2, 4, 5, 7, 9, 11}

	// MinorScale: whole, half, whole, whole, half, whole, whole (natural minor)
	MinorScale = Scale{0, 2, 3, 5, 7, 8, 10}

	// MelodicMinorScale: whole, half, whole, whole, whole, whole, half (ascending melodic minor)
	MelodicMinorScale = Scale{0, 2, 3, 5, 7, 9, 11}
)

// NoteAt returns the MIDI note number for a given scale degree
// rootNote: the root MIDI note (e.g., 60 for middle C)
// degree: the scale degree (0-indexed, can be negative or > scale length)
// Returns: MIDI note number
//
// Examples:
//   MajorScale.NoteAt(60, 0) → 60 (C)
//   MajorScale.NoteAt(60, 2) → 64 (E)
//   MajorScale.NoteAt(60, 7) → 72 (C, one octave up)
//   MajorScale.NoteAt(60, -1) → 59 (B, one octave down)
func (s Scale) NoteAt(rootNote uint8, degree int) uint8 {
	scaleLen := len(s)
	if scaleLen == 0 {
		return rootNote
	}

	// Calculate which octave we're in and position within the scale
	octave := degree / scaleLen
	position := degree % scaleLen

	// Handle negative degrees
	if degree < 0 {
		octave = (degree - scaleLen + 1) / scaleLen
		position = degree - (octave * scaleLen)
	}

	// Calculate semitone offset: octaves + scale interval
	semitones := (octave * 12) + s[position]

	// Add to root note (with bounds checking)
	result := int(rootNote) + semitones
	if result < 0 {
		return 0
	}
	if result > 127 {
		return 127
	}

	return uint8(result)
}
