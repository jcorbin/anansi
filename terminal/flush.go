package terminal

import (
	"runtime"
	"sync"
	"time"
)

func (out *Output) setWriteOption(wo writeObserver) {
	if fa, ok := out.writeObserver.(*FlushAfter); ok {
		fa.Stop()
	}
	if wo == nil {
		out.writeObserver = flushWhenFull{}
	} else {
		out.writeObserver = wo
	}
}

// FlushWhenFull causes a terminal's output buffer to prefer to flush rather
// than grow, similar to a bufio.Writer.
//
// TODO avoid writing large buffers and string indirectly, ability to pass
// through does not exist currently.
//
// NOTE mutually exclusive with any other Flush* options; the last one wins.
var FlushWhenFull Option = flushWhenFull{}

type flushWhenFull struct{}

func (fw flushWhenFull) init(term *Terminal) error {
	term.Output.setWriteOption(fw)
	return nil
}

func (fw flushWhenFull) preWrite(out *Output, n int) error {
	if m := out.buf.Len(); m > 0 && m+n >= out.buf.Cap() {
		return out.Flush()
	}
	return nil
}
func (fw flushWhenFull) postWrite(out *Output, n int) error {
	if m := out.buf.Len(); m > 0 && m == out.buf.Cap() {
		return out.Flush()
	}
	return nil
}

// FlushAfter implements an Option that causes a terminal to flush its output
// some specified time after the first write to it. The user should retain and
// lock their FlushAfter instance during their drawing update routine so that
// partial output does not get flushed. Example usage:
//
//	fa := terminal.FlushAfter{Duration: time.Second / 60}
//	term, err := terminal.Open(nil, nil, terminal.Options(&fa))
//	if term != nil {
//		defer term.Close()
//	}
//	var ev terminal.Event
//	for err == nil {
//		fa.Lock()                  // exclude flushing partial output while...
//		term.Discard()             // ... drop any undrawn output from last round
//		draw(term, ev)             // ... call term.Write* to draw new output
//		fa.Unlock()                // ... end exclusion
//		ev, err = term.ReadEvent() // block for next input event
//	}
//	// TODO err handling
//
// NOTE mutually exclusive with any other Flush* options; the last one wins.
type FlushAfter struct {
	sync.Mutex
	time.Duration

	out  *Output
	set  bool
	stop chan struct{}
	t    *time.Timer
}

func (fa *FlushAfter) init(term *Terminal) error {
	fa.out = &term.Output
	term.Output.setWriteOption(fa)
	return nil
}

func (fa *FlushAfter) preWrite(out *Output, n int) error {
	fa.out = out
	fa.Start()
	return nil
}
func (fa *FlushAfter) postWrite(out *Output, n int) error {
	return nil
}

// Start the flush timer, allocating and spawn its monitor goroutine if
// necessary. Should only be called by the user in a locked section.
func (fa *FlushAfter) Start() {
	if fa.t == nil {
		fa.t = time.NewTimer(fa.Duration)
		fa.stop = make(chan struct{})
		go fa.monitor(fa.t.C, fa.stop)
	} else if !fa.set {
		fa.t.Reset(fa.Duration)
	}
	fa.set = true
}

// Stop the flush timer and any monitor goroutine. Should only be called by the
// user in a locked section.
func (fa *FlushAfter) Stop() {
	if fa.stop != nil {
		close(fa.stop)
		fa.t.Stop()
		fa.t = nil
		fa.stop = nil
		fa.set = false
	}
}

// Cancel any flush timer, returning true if one was canceled; users should
// call this method after any manual terminal flush. Should only be called by
// the user in a locked section.
func (fa *FlushAfter) Cancel() bool {
	fa.set = false
	if fa.t == nil {
		return false
	}
	return fa.t.Stop()
}

func (fa *FlushAfter) monitor(ch <-chan time.Time, stop <-chan struct{}) {
	runtime.LockOSThread() // dedicate this thread to terminal writing
	done := false
	for !done {
		select {
		case <-stop:
			done = true
		case t := <-ch:
			if fa.flush(t) != nil {
				break
			}
		}
	}
	fa.Lock()
	defer fa.Unlock()
	if fa.t != nil && fa.t.C == ch {
		fa.t = nil
		fa.set = false
	}
	if fa.stop == stop {
		fa.stop = nil
	}
}

func (fa *FlushAfter) flush(_ time.Time) error {
	fa.Lock()
	defer fa.Unlock()
	fa.set = false
	return fa.out.Flush()
}
