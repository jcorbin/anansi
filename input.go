package anansi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/jcorbin/anansi/ansi"
)

const defaultMinRead = 128

// NewInput creates an Input around the given file; the optional minRead
// argument defaults to 128.
func NewInput(f *os.File, minRead int) *Input {
	if minRead == 0 {
		minRead = defaultMinRead
	}
	in := &Input{
		file:    f,
		minRead: minRead,
	}
	return in
}

// Input supports reading terminal input from a file handle with a buffer for
// things like escape sequences. It supports both blocking and non-blocking
// reads. It is not safe to use Input in parallel from multiple goroutines,
// such users need to layer a lock around an Input.
type Input struct {
	file *os.File

	ateof    bool
	minRead  int
	nonblock bool
	buf      bytes.Buffer

	rec    io.Writer
	recTmp bytes.Buffer
}

var errNoFd = errors.New("non-blocking io not supported by underlying reader")

// SetRecording enables (or disables if nil-argument) input recording. When
// enabled, all read bytes are written to the given destination with added
// timing marks.
//
// Timing marks are encoded as an ANSI Application Program Command (APC)
// string; schematically:
//
// 	APC "recTime:" RFC3339Nano ST
//
// For example:
//
// 	"\x1b_recTime:12345\x1b\\any bytes read at that time"
//
// In specific: ReadAny() records the time it was called (even if no bytes were
// available to be read); ReadMore() records the time that its blocking read
// call returned non-zero bytes (modulo go scheduler lag of course).
//
// Any read error(s) are recorded similiarly:
//
// 	APC "recReadErr:" error-string ST
func (in *Input) SetRecording(dest io.Writer) {
	in.rec = dest
}

// IsRecording returns true only if a recording destination is in effect (as
// given to SetRecording).
func (in *Input) IsRecording(dest *os.File) bool {
	return in.rec != nil
}

const (
	recMarkStart = "\x1b_recTime:" // APC "recTime:"
	recMarkEnd   = "\x1b\\"        // ST
	recMarkSize  = 2 + 8 + len(time.RFC3339Nano) + 2

	recReadErrStart = "\x1b_recReadErr:" // APC "recReadErr:"
	recReadErrEnd   = "\x1b\\"           // ST
	recReadErrSize  = 2 + 11 + 256 + 2
)

type fdProvider interface {
	Fd() uintptr
}

// AtEOF returns true if the last input read returned io.EOF.
func (in *Input) AtEOF() bool {
	return in.ateof
}

// DecodeEscape tries to decode an ANSI escape sequence from the internal byte
// buffer; if the returned escape identifier is 0, then the user may proceed to
// call DecodeRune().
//
// NOTE any returned argument slice becomes invalid after the next call to
// DecodeEscape or DecodeRune; the caller must copy any bytes out if it needs
// to retain them.
func (in *Input) DecodeEscape() (e ansi.Escape, a []byte) {
	if in.buf.Len() == 0 {
		return 0, nil
	}
	e, a, n := ansi.DecodeEscape(in.buf.Bytes())
	if n > 0 {
		in.buf.Next(n)
	}
	return e, a
}

// DecodeRune tries to decode a complete non ANSI escape-sequence-related rune
// from the internal buffer, returning it and true if possible.
//
// Otherwise it returns 0 and false, not advancing the internal byte buffer
// beyond the control or partial rune so that DecodeEscape can have a chance to
// decode it with future input.
func (in *Input) DecodeRune() (rune, bool) {
	if in.buf.Len() == 0 {
		return 0, false
	}
	r, n := utf8.DecodeRune(in.buf.Bytes())
	if !in.ateof {
		switch r {
		case 0x90, 0x9B, 0x9D, 0x9E, 0x9F: // DCS, CSI, OSC, PM, APC
			return 0, false
		case utf8.RuneError:
		case 0x1B: // ESC
			if p := in.buf.Bytes(); len(p) == cap(p) {
				return 0, false
			}
		}
	}
	in.buf.Next(n)
	return r, true
}

// ReadMore from the underlying file into the internal byte buffer; it loops
// until at least one new byte has been read. Returns the number of bytes read
// and any error.
func (in *Input) ReadMore() (int, error) {
	// TODO opportunistically read in non-blocking mode if set, only
	//      transitioning to blocking if needed
	if err := in.setNonblock(false); err != nil {
		return 0, err
	}
	for {
		p := in.readBuf()
		n, err := in.file.Read(p)
		if ateof := err == io.EOF; in.ateof && ateof {
			// TODO if n > 0 the io.Reader is being misbehaved... do we care?
			return 0, io.EOF
		} else if in.ateof = ateof; ateof {
			err = nil
		}
		var frm InputFrame
		if in.rec != nil {
			frm.T = time.Now()
		}
		frm.E = err
		if n > 0 {
			frm.B = p[:n]
		}
		err = in.write(frm)
		if n > 0 || err != nil {
			return n, err
		}
	}
}

// ReadAny available bytes from the underlying file into the internal byte
// buffer; uses non-blocking reads. Returns the number of bytes read and any
// error.
func (in *Input) ReadAny() (int, error) {
	if err := in.setNonblock(true); err != nil {
		return 0, err
	}
	in.ateof = false
	var frm InputFrame
	if in.rec != nil {
		frm.T = time.Now()
	}
	p := in.readBuf()
	n, err := in.file.Read(p)
	if isEWouldBlock(err) {
		err = nil
	}
	if in.ateof = err == io.EOF; in.ateof {
		err = nil
	}
	frm.E = err
	if n > 0 {
		frm.B = p[:n]
	}
	err = in.write(frm)
	return n, err
}

// Enter retains the passed the terminal file handle if one isn't already,
// returns an error otherwise.
func (in *Input) Enter(term *Term) error {
	if in.file != nil {
		return errors.New("anansi.Input may only only be attached to one terminal")
	}
	if in.minRead == 0 {
		in.minRead = defaultMinRead
	}
	in.file = term.File
	return nil
}

// Exit clears the retained file handle (only if it's the same as the
// terminal's). Any non-blocking mode is cleared.
func (in *Input) Exit(term *Term) error {
	if in.file != term.File {
		return nil
	}
	in.nonblock = false
	err := in.setFlags()
	in.file = nil
	return err
}

func (in *Input) write(frm InputFrame) error {
	err := frm.E
	_, _ = in.buf.Write(frm.B)
	if in.rec != nil {
		frm.writeIntoBuffer(&in.recTmp)
		if _, werr := in.recTmp.WriteTo(in.rec); err == nil {
			err = werr
		}
		in.recTmp.Reset()
	}
	return err
}

func (frm InputFrame) writeIntoBuffer(buf *bytes.Buffer) {
	if frm.E == nil {
		buf.Grow(recMarkSize + len(frm.B))
		_, _ = buf.WriteString(recMarkStart)
		b := buf.Bytes()
		_, _ = buf.Write(frm.T.AppendFormat(b[len(b):], time.RFC3339Nano))
		_, _ = buf.WriteString(recMarkEnd)
	} else {
		buf.Grow(recMarkSize + recReadErrSize + len(frm.B))
		_, _ = buf.WriteString(recMarkStart)
		b := buf.Bytes()
		_, _ = buf.Write(frm.T.AppendFormat(b[len(b):], time.RFC3339Nano))
		_, _ = buf.WriteString(recMarkEnd)
		_, _ = buf.WriteString(recReadErrStart)
		b = buf.Bytes()
		_, _ = buf.WriteString(frm.E.Error())
		_, _ = buf.WriteString(recReadErrEnd)
	}
	_, _ = buf.Write(frm.B)
}

// readBuf returns a slice into the internal byte buffer with enough space to
// read at least n bytes.
func (in *Input) readBuf() []byte {
	in.buf.Grow(in.minRead)
	p := in.buf.Bytes()
	p = p[len(p):cap(p)]
	return p
}

func (in *Input) setNonblock(nonblock bool) error {
	if nonblock != in.nonblock {
		in.nonblock = nonblock
		return in.setFlags()
	}
	return nil
}

func (in *Input) setFlags() error {
	var flags uintptr
	if in.nonblock {
		flags |= syscall.O_NONBLOCK
	}
	return in.fcntl(syscall.F_SETFL, flags)
}

func (in *Input) fcntl(a2, a3 uintptr) error {
	if _, _, e := syscall.Syscall(syscall.SYS_FCNTL, in.file.Fd(), a2, a3); e != 0 {
		return e
	}
	return nil
}

// InputReplay is a session of recorded input.
type InputReplay []InputFrame

// InputFrame is a frame of recorded input.
type InputFrame struct {
	T time.Time // input timestamp
	E error     // input read error
	B []byte    // read input
	M []byte    // pass through APC message
}

// InputError represets a recorderded read error.
type InputError []byte

func (ie InputError) String() string { return string(ie) }
func (ie InputError) Error() string  { return string(ie) }

// Duration returns the total elapsed time of the replay.
func (rs InputReplay) Duration() time.Duration {
	if len(rs) == 0 {
		return 0
	}
	return rs[len(rs)-1].T.Sub(rs[0].T)
}

func (frm InputFrame) String() string {
	var buf bytes.Buffer
	buf.Grow(1024)
	_, _ = buf.WriteString(frm.T.String())
	if frm.E != nil {
		if ie, ok := frm.E.(InputError); ok {
			_, _ = fmt.Fprintf(&buf, " err=%s", []byte(ie))
		} else {
			_, _ = buf.WriteString(" err=")
			_, _ = buf.WriteString(frm.E.Error())
		}
	}
	if len(frm.M) != 0 {
		_, _ = fmt.Fprintf(&buf, " mess=%q", frm.M)
	}
	if len(frm.B) == 0 {
		_, _ = buf.WriteString(" mark")
	} else {
		_, _ = fmt.Fprintf(&buf, " %q", frm.M)
	}
	return buf.String()
}

// ReadInputReplay reads an InputReplay from the provided file.
func ReadInputReplay(f *os.File) (InputReplay, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	estBytes := int(fi.Size())
	estFrames := estBytes / recMarkSize

	type protoFrame struct {
		InputFrame
		e [2]int
		b [2]int
		m [2]int
	}

	var (
		in     = NewInput(f, 1024)
		frames = make([]protoFrame, 0, estFrames)
		bs     = make([]byte, 0, estBytes)
		off    int
		prot   protoFrame
	)

	push := func() {
		if prot.m != [2]int{} || !prot.T.IsZero() || off < len(bs) {
			prot.b = [2]int{off, len(bs)}
			frames = append(frames, prot)
			prot = protoFrame{}
			off = len(bs)
		}
	}

	for {
		if e, a := in.DecodeEscape(); e == 0x9F { // APC
			switch {
			case bytes.HasPrefix(a, []byte("recTime:")):
				push()
				if parseErr := prot.T.UnmarshalText(a[8:]); parseErr != nil {
					return nil, fmt.Errorf("invalid recTime value %q", a[8:])
				}
			case bytes.HasPrefix(a, []byte("recReadErr:")):
				bs = append(bs, a[11:]...)
				prot.m = [2]int{off, len(bs)}
				off = len(bs)
				push()
			default:
				push()
				bs = append(bs, a...)
				prot.m = [2]int{off, len(bs)}
				off = len(bs)
			}
		} else if e != 0 {
			bs = e.AppendWith(bs, a...)
		} else if r, ok := in.DecodeRune(); ok {
			var tmp [4]byte
			n := utf8.EncodeRune(tmp[:], r)
			bs = append(bs, tmp[:n]...)
		} else if _, err := in.ReadMore(); err == io.EOF {
			push()
			break
		} else if err != nil {
			return nil, err
		}
	}

	result := make(InputReplay, len(frames))
	for i, frm := range frames {
		if frm.e != [2]int{} {
			prot.E = InputError(bs[frm.e[0]:frm.e[1]])
		}
		if frm.b != [2]int{} {
			frm.B = bs[frm.b[0]:frm.b[1]]
		}
		if frm.m != [2]int{} {
			frm.M = bs[frm.m[0]:frm.m[1]]
		}
		result[i] = frm.InputFrame
	}
	return result, nil
}

func isEWouldBlock(err error) bool {
	switch val := err.(type) {
	case *os.PathError:
		err = val.Err
	case *os.LinkError:
		err = val.Err
	case *os.SyscallError:
		err = val.Err
	}
	return err == syscall.EWOULDBLOCK
}
