package anansi

import (
	"fmt"

	"github.com/jcorbin/anansi/ansi"
)

// Mode holds a set/reset byte buffer to write to a terminal file during
// Enter/Exit. Primary useful for set/reset mode control sequences.
type Mode struct {
	Set, Reset []byte

	// TODO support active use with an *os.File
}

// Enter writes the modes' Set() string to the terminal's output file.
func (mode *Mode) Enter(term *Term) error {
	if _, err := term.Output.File.Write(mode.Set); err != nil {
		return fmt.Errorf("failed to write mode set string: %v", err)
	}
	return nil
}

// Exit writes the modes' Reset() string to the terminal's output file.
func (mode *Mode) Exit(term *Term) error {
	if _, err := term.Output.File.Write(mode.Reset); err != nil {
		return fmt.Errorf("failed to write mode reset string: %v", err)
	}
	return nil
}

// AddMode adds Set/Reset pairs from one or more ansi.Mode values.
func (mode *Mode) AddMode(ms ...ansi.Mode) {
	for _, m := range ms {
		mode.AddModePair(m.Set(), m.Reset())
	}
}

// AddModePair adds the byte representation of the given set/reset sequences
// to the mode's Set/Reset byte buffers; appends to Set, prepends to Reset.
func (mode *Mode) AddModePair(set, reset ansi.Seq) {
	var b [128]byte
	mode.Set = set.AppendTo(mode.Set)
	mode.Reset = append(reset.AppendTo(b[:0]), mode.Reset...)
}

// AddModeSeq appends one or more ansi sequences to the set buffer and
// prepends them to the reset buffer.
func (mode *Mode) AddModeSeq(seqs ...ansi.Seq) {
	for _, seq := range seqs {
		n := len(mode.Set)
		mode.Set = seq.AppendTo(mode.Set)
		m := len(mode.Set)
		mode.Reset = append(mode.Set[n:m:m], mode.Reset...)
	}
}
