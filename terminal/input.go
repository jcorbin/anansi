package terminal

import (
	"bytes"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jcorbin/anansi/terminfo"
	"github.com/jcorbin/anansi/termkey"
)

const (
	minRead        = 128
	signalCapacity = 16
)

// DecodeSignal decodes an os Signal into an Event.
func DecodeSignal(sig os.Signal) Event {
	switch sig {
	case syscall.SIGINT:
		return Event{Type: InterruptEvent}
	case syscall.SIGWINCH:
		return Event{Type: ResizeEvent}
	default:
		return Event{Type: SignalEvent, Signal: sig}
	}
}

// Decoder supports reading terminal input events from a file handle.
type Decoder struct {
	File *os.File

	info       *terminfo.Terminfo
	buf        bytes.Buffer
	err        error
	keyDecoder *termkey.Decoder
}

// Terminfo returns the currently set terminfo that the decoder will use; may
// be nil, in which case the first ReadEvents() call will try to load the
// correct terminfo based on environment.
func (dec *Decoder) Terminfo() *terminfo.Terminfo {
	return dec.info
}

// SetTerminfo sets the terminfo definitions that the encoder will use going
// forward.
func (dec *Decoder) SetTerminfo(info *terminfo.Terminfo) {
	dec.info = info
	dec.keyDecoder = termkey.NewDecoder(info)
}

// ReadEvents reads events into the given slice, stopping either when there are
// no more buffered inputs bytes to parse, or the given events buffer is full.
// Reads and blocks from the underlying file until at least one event can be
// parsed. Returns the number of events read and any read error.
func (dec *Decoder) ReadEvents(evs []Event) (n int, _ error) {
	if dec.err != nil {
		return 0, dec.err
	}
	if dec.keyDecoder == nil {
		if dec.info == nil {
			panic("Decoder: no terminfo set")
		}
		dec.keyDecoder = termkey.NewDecoder(dec.info)
	}
	n = dec.decodeEvents(evs)
	for n == 0 {
		_, err := dec.readMore(minRead)
		n = dec.decodeEvents(evs)
		if err == io.EOF && n > 0 && n == len(evs) {
			return n, nil
		} else if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (dec *Decoder) decodeEvents(evs []Event) int {
	i := 0
	for i < len(evs) {
		buf := dec.buf.Bytes()
		if len(buf) == 0 {
			break
		}
		kev, n := dec.keyDecoder.Decode(buf)
		if n == 0 {
			break
		}
		dec.buf.Next(n)
		ev := Event{Type: KeyEvent, Event: kev}
		if kev.Key.IsMouse() {
			ev.Type = MouseEvent
		}
		evs[i] = ev
		i++
	}
	return i
}

func (dec *Decoder) readMore(n int) (int, error) {
	if dec.err != nil {
		return 0, dec.err
	}
	dec.buf.Grow(n)
	buf := dec.buf.Bytes()
	buf = buf[len(buf):cap(buf)]
	n, dec.err = dec.File.Read(buf)
	if n > 0 {
		_, _ = dec.buf.Write(buf[:n])
	}
	return n, dec.err
}

// Err returns any read error encountered so far; if this is non-nill, all
// future reads will fail, and the processor is dead.
func (dec *Decoder) Err() error { return dec.err }

// Processor combines a low level event decoder with optional event filtering
// middleware. It also supports concurrent processing of os signaled events and
// concurrent decoding of events.
type Processor struct {
	Decoder
	EventFilter
	Signals chan<- os.Signal

	signals chan os.Signal
	stop    chan struct{}
}

// Init ialize Processor state; TODO would be great to eliminate this method.
func (proc *Processor) Init() {
	proc.signals = make(chan os.Signal, signalCapacity)
	proc.Signals = proc.signals
	proc.stop = make(chan struct{})
}

// Close the processor (but not the underlying file handle); stops signal
// delivery and any concurrent signal/input handling.
func (proc *Processor) Close() error {
	signal.Stop(proc.signals)
	close(proc.stop)
	return nil
}

// DecodeSignal decodes an os.Signal into an Event, passing it through any
// event filter(s) first.
func (proc *Processor) DecodeSignal(sig os.Signal) (Event, error) {
	ev := DecodeSignal(sig)
	if ev.Type == NoEvent {
		return Event{}, nil
	}
	if proc.EventFilter != nil {
		return proc.FilterEvent(ev)
	}
	return ev, nil
}

// DecodeEvent reads an event with ReadEvent, and then passes it through any
// event filter(s).
//
// If the filter masks the event (returns EventNone with nil error), then
// DecodeEvent loops and reads another event.
//
// Returns the first unfiltered event read and any error.
func (proc *Processor) DecodeEvent() (Event, error) {
	var tmp [1]Event
	for {
		var ev Event
		n, err := proc.ReadEvents(tmp[:])
		if n > 0 {
			ev = tmp[0]
		}
		if err == nil && proc.EventFilter != nil {
			ev, err = proc.FilterEvent(ev)
		}
		if err != nil || ev.Type != NoEvent {
			return ev, err
		}
	}
}

// DecodeEvents reads an event batch with ReadEvents, and then applies any
// event filter(s) to the batch. Any events filtered to EventNone are elided
// from the batch. If a filter returns non-nil error, event filtering stops.
//
// Loops until at least one unfiltered event has been read, and so may do more
// than one round of blocking IO.
//
// Returns the number of unfiltered events and any error.
func (proc *Processor) DecodeEvents(evs []Event) (n int, _ error) {
	for {
		n, err := proc.ReadEvents(evs)
		if err != nil {
			return n, err
		}

		// filter events
		j := 0
		for i := 0; i < n; i++ {
			var err error
			ev := evs[i]
			if proc.EventFilter != nil {
				ev, err = proc.FilterEvent(ev)
			}
			if ev.Type != NoEvent {
				evs[j] = ev
				j++
			}
			if err != nil {
				for i++; i < n; i++ {
					evs[j] = evs[i]
					j++
				}
				return j, err
			}
		}

		// return only if any unfiltered
		if j > 0 {
			return j, nil
		}
	}
}

// ReadEvent reads one event from the input file; this may happen from
// previously read / in-buffer bytes, and may not necessarily block.
//
// The event is first passed through any event filter(s), returning their final
// Event/error pair. NOTE this means that it's possible for ReadEvent to return
// EventNone and nil error if the event was masked by a filter.
//
// NOTE this is a low level method, most users should use proc.DecodeEvent() instead.
func (proc *Processor) ReadEvent() (Event, error) {
	var tmp [1]Event
	n, err := proc.ReadEvents(tmp[:])
	var ev Event
	if n > 0 {
		ev = tmp[0]
	}
	if err == nil && proc.EventFilter != nil {
		ev, err = proc.FilterEvent(ev)
	}
	return ev, err
}

// ProcessSignals handles signals delivered to proc.Signals until either an
// error occurs (generated by an EventFilter) or the decoder is Close()ed.
//
// Sends to the given events channel are non-blocking so that signal events are
// dropped in the case of a backlog. This is also the semantics of the
// underlying `signal.Notify()` semantics, so ProcessSignals follows suit to
// process incoming os signals as fast as possible.
//
// NOTE signal processing mostly makes sense for the Decoder attached to
// os.Stdin; users probably don't want to use it for things like pseudo
// terminals; for example:
//	proc := MakeDecoder(os.Stdin)
//	signal.Notify(proc.Signals, ...)
//	events := make(chan<- Event, 42)   // TODO use your own answer
//	go proc.ProcessSignals(events)     // TODO do something with any error
//	...                                // TODO process events somehow
func (proc *Processor) ProcessSignals(events chan<- Event) error {
	for {
		select {
		case <-proc.stop:
			return nil
		case sig := <-proc.signals:
			if ev, err := proc.DecodeSignal(sig); err != nil {
				return err
			} else if ev.Type != NoEvent {
				select {
				case events <- ev:
				default:
				}
			}
		}
	}
}

// ProcessInput decodes events from the input file handle until an error
// occurs, which it returns.
//
// In contrast to ProcessSignals, sends on the events channel will block if
// full for reliable delivery of in-band events.
func (proc *Processor) ProcessInput(events chan<- Event) error {
	for {
		ev, err := proc.DecodeEvent()
		if err != nil {
			return err
		}
		select {
		case events <- ev:
		case <-proc.stop:
			return nil
		}
	}
}

// ProcessInputBatches decodes event batches from the input file handle until
// an error occurs, which it returns.
//
// Decoding doesn't begin until a free event buffer is received from the free
// channel; decoding then fills calls DecodeEvents() to fill the buffer, and
// sends it on the batches channel.
func (proc *Processor) ProcessInputBatches(batches chan<- []Event, free <-chan []Event) error {
	for {
		var evs []Event
		select {
		case evs = <-free:
			evs = evs[:cap(evs)]
		case <-proc.stop:
			return nil
		}
		// TODO while <-free was blocking above, significant time could have
		// elapsed; it'd be great if we could still do low-level reading
		// concurrently while waiting
		n, err := proc.DecodeEvents(evs)
		if err != nil {
			return err
		}
		select {
		case batches <- evs[:n]:
		case <-proc.stop:
			return nil
		}
	}
}
