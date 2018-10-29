package platform

import (
	"fmt"
	"image"
	"log"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// Events holds a queue of input events that were available at the start of the
// current frame's time window.
type Events struct {
	Type []EventType

	input *anansi.Input
	esc   []ansi.Escape
	arg   [][]byte
	mouse []Mouse
}

// EventType is the type of an entry in Events.
type EventType uint8

// Type constants for Events.
const (
	EventNone EventType = iota
	EventEscape
	EventRune
	EventMouse
	// EventKey TODO key code translation
)

// Escape represents ansi escape sequence data stored in an Events queue.
type Escape struct {
	ID  ansi.Escape
	Arg []byte
}

// Mouse represents mouse data stored in an Events queue.
type Mouse struct {
	State ansi.MouseState
	ansi.Point
}

// ZM is a convenience name for the zero value of Mouse.
var ZM Mouse

// HasTerminal returns true if the given terminal rune is in the event queue,
// striking it and truncating any events after it.
func (es *Events) HasTerminal(r rune) bool {
	for i := 0; i < len(es.Type); i++ {
		if es.Type[i] == EventRune && es.esc[i] == ansi.Escape(r) {
			for ; i < len(es.Type); i++ {
				es.Type[i] = EventNone
			}
			return true
		}
	}
	return false
}

// CountRune counts occurrences of the given rune, striking them out.
func (es *Events) CountRune(r rune) (n int) {
	for i := 0; i < len(es.Type); i++ {
		if es.Type[i] == EventRune && es.esc[i] == ansi.Escape(r) {
			es.Type[i] = EventNone
			n++
		}
	}
	return n
}

// CountPressesIn counts mouse presses of the given button within the given
// rectangle, striking them out.
func (es *Events) CountPressesIn(box ansi.Rectangle, buttonID uint8) (n int) {
	for id, kind := range es.Type {
		if kind == EventMouse {
			if sid, pressed := es.mouse[id].State.IsPress(); pressed && sid == buttonID {
				if es.mouse[id].Point.In(box) {
					n++
					es.Type[id] = EventNone
				}
			}
		}
	}
	return n
}

// AnyPressesOutside returns true if there are any mouse presses outside the
// given rectangle.
func (es *Events) AnyPressesOutside(box ansi.Rectangle) bool {
	for id, kind := range es.Type {
		if kind == EventMouse {
			if _, pressed := es.mouse[id].State.IsPress(); pressed {
				if !es.mouse[id].Point.In(box) {
					return true
				}
			}
		}
	}
	return false
}

// TotalScrollIn counts total mouse scroll delta within the given rectangle,
// striking out all such events.
func (es *Events) TotalScrollIn(box ansi.Rectangle) (n int) {
	for id, kind := range es.Type {
		if kind == EventMouse && es.mouse[id].Point.In(box) {
			switch es.mouse[id].State.ButtonID() {
			case 4: // wheel-up
				n--
				es.Type[id] = EventNone
			case 5: // wheel-down
				n++
				es.Type[id] = EventNone
			}
		}
	}
	return n
}

// TotalCursorMovement returns the total cursor movement delta (e.g. from arrow
// keys) striking out all such cursor movement events. Does not recognize
// cursor line movements (CNL and CPL).
func (es *Events) TotalCursorMovement() (move image.Point) {
	for id, kind := range es.Type {
		if kind == EventEscape {
			if d, isMove := ansi.DecodeCursorCardinal(es.esc[id], es.arg[id]); isMove {
				move = move.Add(d)
				es.Type[id] = EventNone
			}
		}
	}
	return move
}

// LastMouse returns the last mouse event, striking all mouse events out
// (including the last!) only if consume is true.
func (es *Events) LastMouse(consume bool) (m Mouse, have bool) {
	for id, kind := range es.Type {
		if kind == EventMouse {
			m = es.mouse[id]
			have = true
			if consume {
				es.Type[id] = EventNone
			}
		}
	}
	return m, have
}

func (e Escape) String() string { return fmt.Sprintf("%v %s", e.ID, e.Arg) }
func (m Mouse) String() string  { return fmt.Sprintf("%v@%v", m.State, m.Point) }

// Escape returns any ansi escape sequence data for the given event id.
func (es *Events) Escape(id int) Escape { return Escape{es.esc[id], es.arg[id]} }

// Mouse returns any mouse event data for the given event id.
func (es *Events) Mouse(id int) Mouse { return es.mouse[id] }

// Rune returns the event's rune (maybe an ansi.Escape PUA range rune).
func (es *Events) Rune(id int) rune { return rune(es.esc[id]) }

// Clear the event queue.
func (es *Events) Clear() {
	es.Type = es.Type[:0]
	es.esc = es.esc[:0]
	es.arg = es.arg[:0]
	es.mouse = es.mouse[:0]
}

// Load clears the event queue, and then parses from the given byte slice;
// useful for replays and testing.
func (es *Events) Load(b []byte) {
	es.Clear()
	for len(b) > 0 {
		e, a, n := ansi.DecodeEscape(b)
		b = b[n:]
		if e != 0 {
			es.add(e, a, 0)
		} else {
			r, n := utf8.DecodeRune(b)
			b = b[n:]
			es.add(0, nil, r)
		}
	}
}

// Poll clears the event queue, polls for input, and then parses as many input
// bytes as possible.
func (es *Events) Poll() error {
	es.Clear()
	if n, err := es.input.ReadAny(); n == 0 && err != nil {
		return err
	}
	for {
		e, a := es.input.DecodeEscape()
		if e != 0 {
			es.add(e, a, 0)
		} else if r, ok := es.input.DecodeRune(); ok {
			es.add(0, nil, r)
		} else {
			return nil
		}
	}
}

func (es *Events) add(e ansi.Escape, a []byte, r rune) {
	kind := EventEscape
	m := Mouse{}

	if e == 0 {
		kind = EventRune
		e = ansi.Escape(r)
	}

	switch e {
	case ansi.CSI('M'), ansi.CSI('m'):
		var err error
		if m.State, m.Point, err = ansi.DecodeXtermExtendedMouse(e, a); err != nil {
			log.Printf("mouse control: decode error %v %s : %v", e, a, err)
		} else if m.State != 0 || m.Point.Valid() {
			kind = EventMouse
		}

		// TODO map special keys to eventKey
	}

	es.Type = append(es.Type, kind)
	es.esc = append(es.esc, e)
	es.arg = append(es.arg, a)
	es.mouse = append(es.mouse, m)
}
