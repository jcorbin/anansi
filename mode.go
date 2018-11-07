package anansi

import "github.com/jcorbin/anansi/ansi"

// Modes combines one or more ansi modes into a Context that calls Set/Reset on
// Enter/Exit on the Term.File.
func Modes(ms ...ansi.Mode) ModeSeqs {
	var mss ModeSeqs
	return mss.AddMode(ms...)
}

// ModeSeqs holds a set/reset byte buffer to write to a terminal file during Enter/Exit.
type ModeSeqs struct {
	Set, Reset []byte
}

// Enter writes the modes' Set() string to the terminal's file.
func (mss ModeSeqs) Enter(term *Term) error {
	_, err := term.File.Write(mss.Set)
	return err
}

// Exit writes the modes' Reset() string to the terminal's file.
func (mss ModeSeqs) Exit(term *Term) error {
	_, err := term.File.Write(mss.Reset)
	return err
}

// AddMode adds one or more ansi modes's Set() and Reset() sequence pairs.
func (mss ModeSeqs) AddMode(ms ...ansi.Mode) ModeSeqs {
	for _, m := range ms {
		mss = mss.AddPair(m.Set(), m.Reset())
	}
	return mss
}

// AddPair appends as Set sequence and prepends a Reset sequence.
func (mss ModeSeqs) AddPair(set, reset ansi.Seq) ModeSeqs {
	var b [128]byte
	mss.Set = set.AppendTo(mss.Set)
	mss.Reset = append(reset.AppendTo(b[:0]), mss.Reset...)
	return mss
}

// AddSeq appends one or more ansi sequences to the set buffer and prepends
// them to the reset buffer.
func (mss ModeSeqs) AddSeq(seqs ...ansi.Seq) ModeSeqs {
	for _, seq := range seqs {
		n := len(mss.Set)
		mss.Set = seq.AppendTo(mss.Set)
		m := len(mss.Set)
		mss.Reset = append(mss.Set[n:m:m], mss.Reset...)
	}
	return mss
}
