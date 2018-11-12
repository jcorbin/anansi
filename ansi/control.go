package ansi

import (
	"fmt"
	"strconv"
)

const (
	numStaticBytes = 3
	numStaticInts  = 3
)

// Seq represents an escape sequence, led either by ESC or CSI control
// sequence for writing to some output. May only be constructed by any of the
// Escape.With family of methods.
type Seq struct {
	id Escape

	numBytes int
	numInts  int

	argBytes [numStaticBytes]byte
	argInts  [numStaticInts]int

	argExtraBytes []byte
	argExtraInts  []int
}

func (id Escape) seq() Seq {
	switch {
	case 0x0000 < id && id < 0x001F,
		0xEF00 < id && id < 0xEFFF:
		return Seq{id: id}
	}
	panic(fmt.Sprintf("not an Control or Escape rune: %U", id))
}

// With constructs an escape sequence with this identifier and given argument
// byte(s).
// Panics if the escape id is a normal non-Escape rune.
// See Seq.With for details.
func (id Escape) With(arg ...byte) Seq { return id.seq().With(arg...) }

// WithInts constructs an escape sequence with this identifier and the given
// integer argument(s).
// Panics if the escape id is a normal non-Escape rune.
// See Seq.WithInts for details.
func (id Escape) WithInts(args ...int) Seq { return id.seq().WithInts(args...) }

// WithPoint contstructs an escape sequence with an screen point component
// values added as integer arguments in column,row (Y,X) order.
func (id Escape) WithPoint(p Point) Seq { return id.WithInts(p.Y, p.X) }

// ID returns the sequence's Escape identifier.
func (seq Seq) ID() Escape { return seq.id }

// With returns a copy of the sequence with the given argument bytes added.
// Argument bytes will be written immediately after the ESCape identifier
// itself.
func (seq Seq) With(arg ...byte) Seq {
	if len(arg) == 0 {
		return seq
	}
	n := seq.numBytes
	if extraNeed := n + len(arg) - numStaticBytes; extraNeed > 0 {
		argExtraBytes := make([]byte, 0, extraNeed)
		if seq.argExtraBytes != nil {
			argExtraBytes = append(argExtraBytes, seq.argExtraBytes...)
		}
		seq.argExtraBytes = argExtraBytes
	}
	i := 0
	for ; i < len(arg) && n < numStaticBytes; i++ {
		seq.argBytes[n] = arg[i]
		n++
	}
	for ; i < len(arg); i++ {
		seq.argExtraBytes = append(seq.argExtraBytes, arg[i])
		n++
	}
	seq.numBytes = n
	return seq
}

// WithInts returns a copy of the sequence with the given integer arguments
// added. These integer arguments will be written after any byte and string
// arguments in base-10 form, separated by a ';' byte.
// Panics if the sequence identifier is not a CSI function.
func (seq Seq) WithInts(args ...int) Seq {
	if len(args) == 0 {
		return seq
	}
	if 0xEF80 >= seq.id || seq.id >= 0xEFFF {
		panic("may only provide integer arguments to a CSI-sequence")
	}
	n := seq.numInts
	if extraNeed := n + len(args) - numStaticInts; extraNeed > 0 {
		argExtraInts := make([]int, 0, extraNeed)
		if seq.argExtraInts != nil {
			argExtraInts = append(argExtraInts, seq.argExtraInts...)
		}
		seq.argExtraInts = argExtraInts
	}
	i := 0
	for ; i < len(args) && n < numStaticInts; i++ {
		seq.argInts[n] = args[i]
		n++
	}
	for ; i < len(args); i++ {
		seq.argExtraInts = append(seq.argExtraInts, args[i])
		n++
	}
	seq.numInts = n
	return seq
}

// WithPoint returns a copy of the sequence with the given screen point
// component values added as integer arguments in column,row (Y,X) order.
func (seq Seq) WithPoint(p Point) Seq { return seq.WithInts(p.Y, p.X) }

// AppendTo appends the escape code to the given byte slice.
func (id Escape) AppendTo(p []byte) []byte {
	// TODO stricter
	switch {
	case 0x0000 < id && id < 0x001F: // C0 controls
		return append(p, byte(id&0x1F))
	case 0x0080 < id && id < 0x009F: // C1 controls
		return append(p, '\x1b', byte(0x40|id&0x1F))
	case 0xEF20 < id && id < 0xEF7E: // ESC + byte
		return append(p, '\x1b', byte(id&0x7F))
	case 0xEF80 < id && id < 0xEFFF: // CSI + arg (if any)
		return append(p, '\x1b', '[', byte(id&0x7F))
	}
	return p
}

// AppendWith appends the escape code and any given argument bytes to the given
// byte slice.
func (id Escape) AppendWith(p []byte, arg ...byte) []byte {
	// TODO stricter
	switch {
	case 0x0000 < id && id <= 0x001F: // C0 controls
		return append(p, byte(id&0x1F))
	case 0x0080 < id && id <= 0x009F: // C1 controls
		return append(p, '\x1b', byte(0x40|id&0x1F))
	case 0xEF20 < id && id < 0xEF7E: // ESC + byte
		return append(append(append(p, '\x1b'), arg...), byte(id&0x7F))
	case 0xEF80 < id && id < 0xEFFF: // CSI + arg (if any)
		return append(append(append(p, '\x1b', '['), arg...), byte(id&0x7F))
	}
	return p
}

// AppendTo writes the control sequence into the given byte buffer.
func (seq Seq) AppendTo(p []byte) []byte {
	if seq.id == 0 {
		return p
	}
	switch id := seq.id; {
	case id == 0:
	case 0x0000 < id && id < 0x001F: // C0 controls
		p = append(p, byte(id))
		p = seq.appendArgBytes(p)
	case 0xEF80 < id && id < 0xEFFF: // CSI
		p = append(p, "\x1b["...)
		p = seq.appendArgBytes(p)
		p = seq.appendArgNums(p)
		p = append(p, byte(id&0x7F))
	case 0xEF00 < id && id < 0xEF7F: // ESC
		p = append(p, '\x1b')
		p = seq.appendArgBytes(p)
		p = append(p, byte(id&0x7F))
	case 0xEF20 < id && id < 0xEF2F: // ESC character set control
		// NOTE character set selection sequences are special, in that they're
		// always a 3 byte sequence, and identified by the first
		// (intermediate range) byte after the ESC
		p = append(p, '\x1b', byte(id&0x7F), seq.argBytes[0])
	default:
		panic("inconceivable: should not be able to construct a Seq like that")
	}
	return p
}

func (seq Seq) appendArgBytes(p []byte) []byte {
	switch n := seq.numBytes; n {
	case 0:
		return p
	case 1:
		return append(p, seq.argBytes[0])
	case 2:
		return append(p, seq.argBytes[:2]...)
	case 3:
		return append(p, seq.argBytes[:3]...)
		// NOTE need to add more cases if we increase numStaticBytes
	}
	p = append(p, seq.argBytes[:3]...)
	return append(p, seq.argExtraBytes...)
}

func (seq Seq) appendArgNums(p []byte) []byte {
	ni := seq.numInts
	if ni == 0 {
		return p
	}
	p = strconv.AppendInt(p, int64(seq.argInts[0]), 10)
	i := 1
	for ; i < ni && i < numStaticInts; i++ {
		p = append(p, ';')
		p = strconv.AppendInt(p, int64(seq.argInts[i]), 10)
	}
	for ; i < ni; i++ {
		p = append(p, ';')
		p = strconv.AppendInt(p, int64(seq.argExtraInts[i-numStaticInts]), 10)
	}
	return p
}

// Size returns the number of bytes required to encode the escape.
func (id Escape) Size() int {
	switch {
	case 0x0000 < id && id <= 0x001F: // C0 controls
		return 1
	case 0x0080 < id && id <= 0x009F: // C1 controls
		return 2
	case 0xEF20 < id && id < 0xEF7E: // ESC + byte
		return 2
	case 0xEF80 < id && id < 0xEFFF: // CSI + arg (if any)
		return 3
	}
	return 0
}

// Size returns the number of bytes required to encode the escape sequence.
func (seq Seq) Size() int {
	if seq.id == 0 {
		return 0
	}
	return 4 + seq.numBytes + 10*seq.numInts
}

func (seq Seq) String() string {
	if seq.id == 0 && seq.numBytes == 0 && seq.numInts == 0 {
		return ""
	}
	p := make([]byte, 0, seq.numBytes+10*seq.numInts)
	p = seq.appendArgBytes(p)
	p = seq.appendArgNums(p)
	return fmt.Sprintf("%v%q", seq.id, p)
}
