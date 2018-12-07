package anansi_test

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	// We need to handle these signals so that we restore terminal state
	// properly (raw mode and exit the alternate screen).
	halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT)

	// terminal resize signals
	resize = anansi.Notify(syscall.SIGWINCH)

	// input availability notification
	inputReady anansi.InputSignal

	// The virtual screen that will be our canvas.
	screen anansi.TermScreen
)

// logf prints a timestamped message to the virtual screen
func logf(mess string, args ...interface{}) {
	fmt.Fprintf(&screen, "t:%v ", time.Now())
	fmt.Fprintf(&screen, mess, args...)
	screen.WriteString("\r\n")
}

// An example of building a simple fullscreen terminal application with an
// async io loop.
func ExampleTermScreen_termapp() {
	term := anansi.NewTerm(os.Stdin, os.Stdout, &halt, &resize, &inputReady)
	term.SetRaw(true)
	term.AddMode(ansi.ModeAlternateScreen)
	resize.Send("initialize screen size")
	anansi.MustRun(term.RunWithFunc(run))
}

// run implements the main event loop under managed terminal context.
func run(term *anansi.Term) error {
	for {
		select {

		case sig := <-halt.C:
			return anansi.SigErr(sig)

		case <-resize.C:
			if err := screen.SizeToTerm(term); err != nil {
				return err
			}
			logf("resized:%v", screen.Bounds().Size())
			update(term)

		case <-inputReady.C:
			_, err := term.ReadAny()
			if uerr := update(term); err == nil {
				err = uerr
			}
			if err != nil {
				return err // likely io.EOF
			}

		}
	}
}

// update handles any available input then flushes the screen.
func update(term *anansi.Term) error {
	for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
		switch e {

		case 0x03: // stop on Ctrl-C
			return fmt.Errorf("read %v", e)

		case 0x0c: // clear screen on Ctrl-L
			screen.Clear()           // clear virtual contents
			screen.To(ansi.Pt(1, 1)) // cursor back to top
			screen.Invalidate()      // force full redraw
			logf("<clear>")

		default: // log input to screen
			if !e.IsEscape() {
				logf("rune:%q", rune(e))
			} else if a == nil {
				logf("escape:%v", e)
			} else {
				logf("escape:%v arg:%q", e, a)
			}

		}
	}
	return term.Flush(&screen)
}
