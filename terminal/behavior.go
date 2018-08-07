package terminal

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// FullscreenApp encapsulates the expected default behavior of a standard
// fullscreen terminal application:
// - drawing to a raw mode terminal with the cursor hidden
// - synthesize SIGWINCH into ResizeEvent
// - synthesize KeyCtrlL into RedrawEvent
// - synthesize SIGINT into InterruptEvent
// - synthesize KeyCtrlC into InterruptEvent
// - synthesize SIGTERM into ErrTerm
// - suspend when Ctrl-Z is pressed
var FullscreenApp = Options(
	HandleCtrlC,
	HandleCtrlL,
	// TODO HandleKey( KeyCtrlBackslash, send SIGQUIT ) ?
	// TODO enter alternate screen
	HandleSIGINT,
	HandleSIGTERM,
	HandleSIGWINCH,
	SuspendOn(KeyCtrlZ),
	RawMode,
	HiddenCursor,
)

// HandleSIGWINCH by turning it into ResizeEvent.
var HandleSIGWINCH Option = HandleSignal(syscall.SIGWINCH, func(ev Event) (Event, error) {
	return Event{Type: ResizeEvent}, nil
})

// HandleSIGINT by turning it into InterruptEvent.
var HandleSIGINT Option = HandleSignal(syscall.SIGINT, func(ev Event) (Event, error) {
	return Event{Type: InterruptEvent}, nil
})

// HandleSIGTERM by turning it into ErrTerm.
var HandleSIGTERM Option = HandleSignal(syscall.SIGTERM, func(ev Event) (Event, error) {
	return Event{}, ErrTerm
})

// HandleKey creates an option that adds a key handling event filter.
func HandleKey(
	key Key,
	handle func(ev Event) (Event, error),
) Option {
	return keyHandler{key: key, handle: handle}
}

type keyHandler struct {
	key    Key
	handle func(ev Event) (Event, error)
}

func (kh keyHandler) init(term *Terminal) error {
	term.EventFilter = chainEventFilter(term.EventFilter, kh)
	return nil
}

func (kh keyHandler) FilterEvent(ev Event) (Event, error) {
	if ev.Type == KeyEvent && ev.Key == kh.key {
		return kh.handle(ev)
	}
	return Event{}, nil
}

// HandleCtrlC by by turning it into InterruptEvent.
var HandleCtrlC = HandleKey(KeyCtrlC, func(ev Event) (Event, error) {
	return Event{Type: InterruptEvent}, nil
})

// HandleCtrlL by by turning it into RedrawEvent.
var HandleCtrlL = HandleKey(KeyCtrlL, func(ev Event) (Event, error) {
	ev.Type = RedrawEvent
	return ev, nil
})

// HandleSignal creates an option that adds a signal handling event filter
// during terminal lifecycle.
func HandleSignal(
	signal os.Signal,
	handle func(ev Event) (Event, error),
) Option {
	return signalHandler{signal: signal, handle: handle}
}

type signalHandler struct {
	signal os.Signal
	handle func(ev Event) (Event, error)
	active bool
}

func (sh signalHandler) init(term *Terminal) error {
	term.EventFilter = chainEventFilter(term.EventFilter, &sh)
	term.ctx = chainTermContext(term.ctx, &sh)
	return nil
}

func (sh *signalHandler) Enter(term *Terminal) error {
	if !sh.active {
		sh.active = true
		signal.Notify(term.Processor.Signals, sh.signal)
	}
	return nil
}

func (sh *signalHandler) Exit(term *Terminal) error {
	// TODO support (optional) deregistration (e.g. when suspending)?
	return nil
}

func (sh *signalHandler) FilterEvent(ev Event) (Event, error) {
	if ev.Type == SignalEvent && ev.Signal == sh.signal {
		return sh.handle(ev)
	}
	return Event{}, nil
}

// SuspendOn creates an Option that calls Terminal.Suspend() when the specified
// key(s) are pressed. The corresponding KeyEvents are filtered out, never seen
// by the client.
func SuspendOn(keys ...Key) Option {
	return suspendOn{keys: keys}
}

type suspendOn struct {
	keys []Key
	pend func() (os.Signal, error)
}

func (sus suspendOn) init(term *Terminal) error {
	sus.pend = term.Suspend
	term.EventFilter = chainEventFilter(term.EventFilter, &sus)
	return nil
}

func (sus suspendOn) FilterEvent(ev Event) (Event, error) {
	if ev.Type == KeyEvent {
		for i := range sus.keys {
			if ev.Key == sus.keys[i] {
				log.Printf("suspending on %v", ev)
				sig, err := sus.pend()
				if err == nil {
					ev.Type = RedrawEvent
					ev.Signal = sig
				}
				return ev, err
			}
		}
	}
	return ev, nil
}
