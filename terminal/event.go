package terminal

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jcorbin/anansi/termkey"
)

// Event is a terminal input event, either read from the input file, or
// delivered by a relevant signal.
//
// TODO event stolen from termbox; reconcile with tcell
type Event struct {
	Type          EventType // one of Event* constants
	termkey.Event           // EventKey and EventMouse
	Signal        os.Signal // EventSignal
	time.Time
}

// Modifier during a key or mouse event.
type Modifier = termkey.Modifier

// Key code during a key event.
type Key = termkey.Key

//go:generate sh -c "./scripts/copy_consts.sh ../termkey/key.go terminal | goimports >key.go"

// EventType type of an Event.
type EventType uint8

// Event types.
const (
	NoEvent EventType = iota
	KeyEvent
	MouseEvent
	ResizeEvent
	TickEvent
	RedrawEvent
	SignalEvent
	InterruptEvent
	EOFEvent

	FirstUserEvent
)

func (ev Event) String() string {
	switch ev.Type {
	case NoEvent:
		return "NoEvent"
	case KeyEvent:
		return fmt.Sprintf("KeyEvent(%s)", ev.keyString())
	case MouseEvent:
		return fmt.Sprintf("MouseEvent(%s)", ev.mouseString())
	case ResizeEvent:
		return "ResizeEvent"
	case TickEvent:
		return "TickEvent"
	case RedrawEvent:
		s := "RedrawEvent"
		var parts [2]string
		i := 0
		if ev.Key != 0 {
			parts[i] = "key=" + ev.keyString()
			i++
		}
		if ev.Signal != nil {
			parts[i] = "signal=" + ev.Signal.String()
			i++
		}
		if i > 0 {
			s = fmt.Sprintf("%s(%s)", s, strings.Join(parts[:i], " "))
		}
		return s
	case SignalEvent:
		return fmt.Sprintf("SignalEvent(%v)", ev.Signal.String())
	case InterruptEvent:
		return "InterruptEvent"
	case EOFEvent:
		return "EOFEvent"
	default:
		return fmt.Sprintf("UserEvent{Type:%d}", ev.Type)
	}
}

func (ev Event) keyString() string {
	var parts [4]string
	i := 0
	if ev.Mod != 0 {
		parts[i] = ev.Mod.String()
		i++
	}
	if ev.Key != 0 {
		parts[i] = ev.Key.String()
		i++
		if ev.Ch != 0 {
			parts[i] = "WITH_INVALID_CHAR"
			i++
		}
	}
	if ev.Ch != 0 {
		if strconv.IsPrint(ev.Ch) {
			parts[i] = string(ev.Ch)
		} else {
			s := strconv.QuoteRune(ev.Ch)
			parts[i] = s[1 : len(s)-1]
		}
		i++
	}
	switch i {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		return strings.Join(parts[:i], "+")
	}
}

func (ev Event) mouseString() string {
	parts := [6]string{
		ev.keyString(), "@<",
		strconv.Itoa(ev.X), ",",
		strconv.Itoa(ev.Y), ">",
	}
	return strings.Join(parts[:], "")
}

// EventFilter represents a piece of event processing middleware.
type EventFilter interface {
	FilterEvent(ev Event) (Event, error)
}

func chainEventFilter(a, b EventFilter) EventFilter {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	as, haveAs := a.(eventFilters)
	bs, haveBs := b.(eventFilters)
	if haveAs && haveBs {
		return append(as, bs...)
	} else if haveAs {
		return append(as, b)
	} else if haveBs {
		return append(eventFilters{a}, bs...)
	}
	return eventFilters{a, b}
}

type eventFilterFunc func(ev Event) (Event, error)

func (f eventFilterFunc) init(term *Terminal) error {
	term.EventFilter = chainEventFilter(term.EventFilter, f)
	return nil
}
func (f eventFilterFunc) FilterEvent(ev Event) (Event, error) { return f(ev) }

type eventFilters []EventFilter

func (evfs eventFilters) FilterEvent(ev Event) (Event, error) {
	for i := range evfs {
		ev, err := evfs[i].FilterEvent(ev)
		if err != nil || ev.Type != NoEvent {
			return ev, err
		}
	}
	return ev, nil
}
