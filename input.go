package anansi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/jcorbin/anansi/ansi"
)

const defaultMinReadSize = 128

// Input supports reading terminal input from a file handle with a buffer for
// things like escape sequences. It supports both blocking and non-blocking
// reads. It is not safe to use Input in parallel from multiple goroutines,
// such users need to layer a lock around an Input.
type Input struct {
	File        *os.File
	MinReadSize int

	oldFlags uintptr
	ateof    bool
	nonblock bool
	async    bool
	sigio    chan os.Signal
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

// Notify sets up async notification to the given channel, replacing (stopping
// notifications to) any prior channel passed to Notify(). If the passed
// channel is nil, async mode is disabled.
func (in *Input) Notify(sigio chan os.Signal) error {
	prior := in.sigio
	in.sigio = sigio
	if sigio != nil {
		signal.Notify(in.sigio, syscall.SIGIO)
	}
	if prior != nil {
		signal.Stop(prior)
	}
	return in.setAsync(sigio != nil)
}

// Decode decodes the next available ANSI escape sequence or UTF8 rune from the
// internal buffer filled by ReadMore or ReadAny. If the ok return value is
// false, then none can be decoded without first reading more input.
//
// The caller should call e.IsEscape() to tell the difference between
// escape-sequence-signifying runes, and normal ones. Normal runes may then be
// cast and handled ala `if !e.IsEscape() { r := rune(e) }`.
//
// NOTE any returned escape argument slice becomes invalid after the next call
// to Decode; the caller MUST copy any bytes out if it needs to retain them.
func (in *Input) Decode() (e ansi.Escape, a []byte, ok bool) {
	if e, a := in.decodeEscape(); e != 0 {
		return e, a, true
	}
	if r, ok := in.decodeRune(); ok {
		return ansi.Escape(r), nil, true
	}
	return 0, nil, false
}

func (in *Input) decodeEscape() (e ansi.Escape, a []byte) {
	if in.buf.Len() == 0 {
		return 0, nil
	}
	e, a, n := ansi.DecodeEscape(in.buf.Bytes())
	if n > 0 {
		in.buf.Next(n)
	}
	return e, a
}

func (in *Input) decodeRune() (rune, bool) {
	if in.buf.Len() == 0 {
		return 0, false
	}
	r, n := utf8.DecodeRune(in.buf.Bytes())
	if !in.ateof {
		switch r {
		case 0x90, 0x9B, 0x9D, 0x9E, 0x9F: // DCS, CSI, OSC, PM, APC
			return 0, false
		case 0x1B: // ESC
			if p := in.buf.Bytes(); len(p) == cap(p) && !in.ateof {
				return 0, false
			}
		}
	}
	in.buf.Next(n)
	return r, true
}

// ReadMore from the underlying file into the internal byte buffer; it
// may block until at least one new byte has been read. Returns the
// number of bytes read and any error.
func (in *Input) ReadMore() (int, error) {
	for {
		var frm InputFrame

		p := in.readBuf()
		n, err := in.File.Read(p)
		if in.rec != nil {
			frm.T = time.Now()
		}

		switch unwrapOSError(err) {
		case io.EOF:
			if in.ateof {
				return n, io.EOF
			}
			in.ateof = true
			if n > 0 {
				err = nil
			}

		case syscall.EWOULDBLOCK:
			in.ateof = false
			if n > 0 {
				return n, nil
			}
			if err := in.setNonblock(false); err != nil {
				return 0, err
			}

		case nil:
			in.ateof = false
		}

		if n > 0 {
			frm.B = p[:n]
			_, _ = in.buf.Write(frm.B)
		}

		if in.rec != nil {
			frm.E = err
			if werr := in.recordFrame(frm); err == nil {
				err = werr
			}
		}

		if n > 0 || err != nil {
			return n, err
		}
	}
}

// ReadAny reads any (and all!) available bytes from the underlying file into
// the internal byte buffer; uses non-blocking reads. Returns the number of
// bytes read and any error.
func (in *Input) ReadAny() (n int, err error) {
	if err = in.setNonblock(true); err != nil {
		return 0, err
	}
	in.ateof = false

	var frm InputFrame
	if in.rec != nil {
		frm.T = time.Now()
	}

	for err == nil {
		var m int
		p := in.readBuf()
		m, err = in.File.Read(p)
		if m == 0 {
			break
		}
		_, _ = in.buf.Write(p[:m])
		n += m
	}

	switch unwrapOSError(err) {
	case io.EOF:
		in.ateof = true

	case syscall.EWOULDBLOCK:
		in.ateof = false
		err = nil

	case nil:
		in.ateof = false
	}

	if n > 0 {
		p := in.buf.Bytes()
		frm.B = p[len(p)-n:]
	}

	if in.rec != nil {
		frm.E = err
		if werr := in.recordFrame(frm); err == nil {
			err = werr
		}
	}

	return n, err
}

// Enter gets the current fcntl flags for restoration during Exit(), and sets
// non-blocking/async modes if needed.
func (in *Input) Enter(term *Term) error {
	prior, err := in.setFlags()
	if err == nil {
		in.oldFlags = prior
	}
	return err
}

// Exit restores fcntl flags to their Enter() time value.
func (in *Input) Exit(term *Term) error {
	_, _, err := in.fcntl(syscall.F_SETFL, in.oldFlags)
	return err
}

// Close stops any signal notification setup by Notify().
func (in *Input) Close() error {
	if in.sigio != nil {
		signal.Stop(in.sigio)
		in.sigio = nil
	}
	return nil
}

func (in *Input) recordFrame(frm InputFrame) error {
	frm.writeIntoBuffer(&in.recTmp)
	_, err := in.recTmp.WriteTo(in.rec)
	in.recTmp.Reset()
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
	if in.MinReadSize == 0 {
		in.MinReadSize = defaultMinReadSize
	}
	in.buf.Grow(in.MinReadSize)
	p := in.buf.Bytes()
	p = p[len(p):cap(p)]
	return p
}

func (in *Input) setAsync(async bool) error {
	if async != in.async {
		in.async = async
		if _, err := in.setFlags(); err != nil {
			return err
		}
		if in.async {
			if _, _, err := in.fcntl(syscall.F_SETOWN, uintptr(syscall.Getpid())); err != nil && runtime.GOOS != "darwin" {
				return err
			}
		}
	}
	return nil
}

func (in *Input) setNonblock(nonblock bool) error {
	if nonblock != in.nonblock {
		in.nonblock = nonblock
		_, err := in.setFlags()
		return err
	}
	return nil
}

func (in *Input) setFlags() (prior uintptr, _ error) {
	flags, _, err := in.fcntl(syscall.F_GETFL, 0)
	if err != nil {
		return 0, err
	}
	prior = flags
	flags = in.buildFlags(flags)
	_, _, err = in.fcntl(syscall.F_SETFL, flags)
	return prior, err
}

func (in *Input) buildFlags(flags uintptr) uintptr {
	if in.nonblock {
		flags |= syscall.O_NONBLOCK
	}
	if in.async {
		flags |= syscall.O_ASYNC
	}
	return flags
}

func (in *Input) fcntl(a2, a3 uintptr) (r1, r2 uintptr, err error) {
	r1, r2, e := syscall.Syscall(syscall.SYS_FCNTL, in.File.Fd(), a2, a3)
	if e != 0 {
		return 0, 0, e
	}
	return r1, r2, nil
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
		in     = Input{File: f, MinReadSize: 1024}
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
		e, a, ok := in.Decode()
		if !ok {
			if _, err := in.ReadMore(); err == io.EOF {
				push()
				break
			} else if err != nil {
				return nil, err
			}
		}

		if e == 0x9F { // APC
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
		} else if e.IsEscape() {
			bs = e.AppendWith(bs, a...)
		} else {
			var tmp [4]byte
			n := utf8.EncodeRune(tmp[:], rune(e))
			bs = append(bs, tmp[:n]...)
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

func unwrapOSError(err error) error {
	for {
		switch val := err.(type) {
		case *os.PathError:
			err = val.Err
		case *os.LinkError:
			err = val.Err
		case *os.SyscallError:
			err = val.Err
		default:
			return err
		}
	}
}
