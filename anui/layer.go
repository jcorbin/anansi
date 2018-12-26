package anui

import (
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// Layer is a composable user interface element.
type Layer interface {
	HandleInput(e ansi.Escape, a []byte) (handled bool, err error)
	Draw(sc anansi.Screen, now time.Time) anansi.Screen
	NeedsDraw() time.Duration
}

func drawLayerInto(layer Layer, sc anansi.Screen, now time.Time) anansi.Screen {
	sc.Clear()
	sc = layer.Draw(sc, now)
	sc = sc.Full()
	return sc
}

// Layers combines the given layer(s) into a single layer that dispatches
// HandleInput in order and Draw in reverse order. Its NeedsDraw method returns
// the smallest non-zero value from the constituent layers.
func Layers(ls ...Layer) Layer {
	if len(ls) == 0 {
		return nil
	}
	a := ls[0]
	for i := 1; i < len(ls); i++ {
		b := ls[i]
		if b == nil || b == Layer(nil) {
			continue
		}
		if a == nil || a == Layer(nil) {
			a = b
			continue
		}
		as, haveAs := a.(layers)
		bs, haveBs := b.(layers)
		if haveAs && haveBs {
			a = append(as, bs...)
		} else if haveAs {
			a = append(as, b)
		} else if haveBs {
			a = append(layers{a}, bs...)
		} else {
			a = layers{a, b}
		}
	}
	return a
}

type layers []Layer

func (ls layers) NeedsDraw() (d time.Duration) {
	for i := 0; i < len(ls); i++ {
		d = MinNeedsDraw(d, ls[i].NeedsDraw())
	}
	return d
}

// MinNeedsDraw returns the minimum non-zero duration from its arguments, or
// zero if no arg is non-zero.
func MinNeedsDraw(ds ...time.Duration) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	d := ds[0]
	for _, nd := range ds[1:] {
		if d == 0 || (nd > 0 && nd < d) {
			d = nd
		}
	}
	return d
}

func (ls layers) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	for i := 0; i < len(ls); i++ {
		if handled, err = ls[i].HandleInput(e, a); handled || err != nil {
			return handled, err
		}
	}
	return false, nil
}

func (ls layers) Draw(sc anansi.Screen, now time.Time) anansi.Screen {
	for i := len(ls) - 1; i >= 0; i-- {
		sc = ls[i].Draw(sc, now)
	}
	return sc
}

// DrawFunc is a layer that implements only drawing.
type DrawFunc func(sc anansi.Screen, now time.Time) anansi.Screen

// HandleInput is a no-op.
func (f DrawFunc) HandleInput(e ansi.Escape, a []byte) (bool, error) { return false, nil }

// Draw calls the function pointer.
func (f DrawFunc) Draw(sc anansi.Screen, now time.Time) anansi.Screen { return f(sc, now) }

// NeedsDraw is a no-op.
func (f DrawFunc) NeedsDraw() time.Duration { return 0 }

// HandleInputFunc is a layer that implements only input handling.
type HandleInputFunc func(e ansi.Escape, a []byte) (handled bool, err error)

// HandleInput calls the function pointer.
func (f HandleInputFunc) HandleInput(e ansi.Escape, a []byte) (bool, error) { return f(e, a) }

// Draw is a no-op.
func (f HandleInputFunc) Draw(sc anansi.Screen, now time.Time) anansi.Screen { return sc }

// NeedsDraw is a no-op.
func (f HandleInputFunc) NeedsDraw() time.Duration { return 0 }

// NeedsDrawFunc is a layer that implements only draw requesting.
type NeedsDrawFunc func() time.Duration

// HandleInput is a no-op.
func (f NeedsDrawFunc) HandleInput(e ansi.Escape, a []byte) (bool, error) { return false, nil }

// Draw is a no-op.
func (f NeedsDrawFunc) Draw(sc anansi.Screen, now time.Time) anansi.Screen { return sc }

// NeedsDraw calls the function pointer.
func (f NeedsDrawFunc) NeedsDraw() time.Duration { return f() }
