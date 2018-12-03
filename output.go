package anansi

import (
	"io"
	"os"
	"syscall"
	"time"
)

// Output supports writing buffered output from a io.WriterTo (implemented by
// both Cursor and Screen) into a file handle (presumably attached to a
// terminal). It is not safe to use Output in parallel from multiple
// goroutines, such users need to layer a lock around an Output.
type Output struct {
	File    *os.File
	Flushed int
	blocks  []time.Duration
}

// TrackStalls allocates a buffer for tracking stall times; otherwise Stalls()
// will always return nil. If output is used with a non-blocking file handle,
// and if a Flush() write encounters syscall.EWOULDBLOCK, then it switches the
// file handle into blocking mode, and performs a blocking write.
//
// Such blocking flush is counted as a "stall", and the total time spent
// performing FCNTL syscalls and the blocking is counted in any internal buffer
// (allocated by TrackStalls) for further collection and reporting. Once the
// buffer fills, such timing measurements cease to be taken, as if no buffer
// was available. Users interested in collecting these metrics should attempt
// to harvest data using the Stalls() method, and only process such data when
// it is full (len() == cap()).
func (out *Output) TrackStalls(n int) {
	if n == 0 {
		out.blocks = nil
	} else {
		out.blocks = make([]time.Duration, 0, n)
	}
}

// Stalls resets the stalls counter, returning the prior value; a stall happens
// when a flush must do a blocking write on an otherwise non-blocking
// underlying file. The caller must use the returned duration slice
// immediately, as it will be reused if full or if consume was true.
func (out *Output) Stalls(consume bool) []time.Duration {
	if out.blocks == nil {
		return nil
	}
	blocks := out.blocks
	if len(blocks) == cap(blocks) || consume {
		out.blocks = blocks[:0]
	}
	return blocks
}

// Enter is a no-op.
func (out *Output) Enter(term *Term) error { return nil }

// Exit is a no-op.
func (out *Output) Exit(term *Term) error { return nil }

// TODO should this be ReadFrom(r Reader) (n int64, err error) ?

// Flush calls the given io.Writerto on any active file handle. If EWOULDBLOCK
// occurs, it transitions the file into blocking mode, and restarts the write.
func (out *Output) Flush(wer io.WriterTo) error {
	if out.File == nil {
		return nil
	}
	out.Flushed = 0
	n, err := wer.WriteTo(out.File)
	out.Flushed += int(n)
	if unwrapOSError(err) == syscall.EWOULDBLOCK {
		return out.blockingFlush(wer)
	}
	return err
}

func (out *Output) blockingFlush(wer io.WriterTo) error {
	if out.blocks != nil {
		defer out.recordStall(time.Now())
	}
	flags, _, err := out.fcntl(syscall.F_GETFL, 0)
	if err != nil {
		return err
	}
	if _, _, err = out.fcntl(syscall.F_SETFL, flags & ^uintptr(syscall.O_NONBLOCK)); err != nil {
		return err
	}
	n, err := wer.WriteTo(out.File)
	out.Flushed += int(n)
	if _, _, ferr := out.fcntl(syscall.F_SETFL, flags); err == nil {
		err = ferr
	}
	return err
}

func (out *Output) recordStall(t0 time.Time) {
	t1 := time.Now()
	if len(out.blocks) < cap(out.blocks) {
		out.blocks = append(out.blocks, t1.Sub(t0))
	}
}

func (out *Output) fcntl(a2, a3 uintptr) (r1, r2 uintptr, err error) {
	r1, r2, e := syscall.Syscall(syscall.SYS_FCNTL, out.File.Fd(), a2, a3)
	if e != 0 {
		return 0, 0, e
	}
	return r1, r2, nil
}
