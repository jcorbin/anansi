package anui

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// Logs is an in-memory buffer of all logs written through the standard "log"
// package.
var Logs bytes.Buffer

var logsSetup bool

// WithOpenLogFile initializes Logs buffer capture, optionally calling
// OpenLogFile with the given name (if non-empty). Once log output has been
// set, it defers restoring log output to os.Stderr, then calls the given
// function.
func WithOpenLogFile(name string, f func() error) error {
	if name == "" {
		initLogs()
	} else if err := OpenLogFile(name); err != nil {
		return err
	}
	defer log.SetOutput(os.Stderr)
	return f()
}

// OpenLogFile creates a file with the given name, and sets the "log" package
// output to be an io.MultiWriter to it and the Logs buffer.
func OpenLogFile(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("failed to create logfile %q: %v", name, err)
	}
	log.SetOutput(io.MultiWriter(
		&Logs,
		f,
	))
	logsSetup = true
	return nil
}

func initLogs() {
	if !logsSetup {
		log.SetOutput(&Logs)
		logsSetup = true
	}
}

// TODO cap the buffer, load from file if scroll past..

// LogLayer implements a layer for displaying in-memory buffered Logs.
type LogLayer struct {
	SubScreen func(sc anansi.Screen, numLines int) anansi.Screen
	lastLen   int
}

var _ Layer = (*LogLayer)(nil)

//HandleInput is a no-op.
func (ll LogLayer) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	// TODO support scrolling
	return false, nil
}

// Draw overlays the tail of buffered Logs content into the screen.
// If LogLayer.SubScreen is not nil, it is used to target a sub-screen.
func (ll *LogLayer) Draw(sc anansi.Screen, now time.Time) anansi.Screen {
	lb := Logs.Bytes()
	numLines := bytes.Count(lb, []byte("\n"))
	// if len(lb) > 0 { numLines++ }

	area := sc
	if ll.SubScreen != nil {
		area = ll.SubScreen(area, numLines)
	}

	height := sc.Bounds().Dy()

	off := len(lb)
	for i := 0; i < height; i++ {
		b := lb[:off]
		i := bytes.LastIndexByte(b, '\n')
		if i < 0 {
			off = 0
			break
		}
		off -= len(b) - i
	}
	for off < len(lb) && lb[off] == '\n' {
		off++
	}

	anansi.Process(&area, lb[off:])

	ll.lastLen = len(lb)
	return sc
}

// NeedsDraw returns non-zero if more logs have been written since last Draw.
func (ll LogLayer) NeedsDraw() time.Duration {
	if Logs.Len() > ll.lastLen {
		return time.Millisecond
	}
	return 0
}

// BottomNLines returns a function that returns a bottom-aligned sub screen of
// at most n lines within the grid it's passed.
func BottomNLines(n int) func(sc anansi.Screen, numLines int) anansi.Screen {
	return func(sc anansi.Screen, numLines int) anansi.Screen {
		if numLines > n {
			numLines = n
		}
		return sc.SubAt(ansi.Pt(
			1, sc.Bounds().Dy()-numLines,
		))
	}
}
