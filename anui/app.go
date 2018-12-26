package anui

import (
	"fmt"
	"image"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// RunLayer is a convenience that runs the given layer under a new standard
// terminal setup (as returned by anansi.OpenTerm).
func RunLayer(layer Layer, opts ...Option) error {
	term, err := anansi.OpenTerm()
	if err != nil {
		return err
	}
	return RunTermLayer(term, layer, opts...)
}

// RunTermLayer runs the given layer under an event loop running within the
// given terminal. The given terminal must not be active yet so that additional
// anansi.Context may be added to it.
//
// If the given terminal is standard, then standard log output will be diverted
// away from os.Stderr, and typcial terminal lifecycle signal handlers will be
// installed.
//
// Typical Ctrl-C halting and Ctrl-L redraw-forcing behavior is wrapped around
// the given layer. Furthermore the given layer's NeedsDraw() durations will be
// throttled so that no more than 120 draws-per-second are done.
func RunTermLayer(term *anansi.Term, layer Layer, opts ...Option) error {
	var opt Option
	if len(opts) == 0 {
		opt = DefaultOptions
	} else {
		opt = options(opts)
	}

	var app app
	app.Term = term
	app.Layer = layer
	app.Term.AddContext(&app.input, &app.screen)
	if err := app.init(opt); err != nil {
		return err
	}
	app.Layer = ClampNeedsDrawLayer(app.Layer, time.Second/120)
	return app.Term.RunWith(app.run)
}

// Option customizes RunTermLayer behavior.
type Option interface {
	apply(app *app) error
}

// DefaultOptions are used if no options are given to RunTermLayer.
var DefaultOptions = Options(
	StandardKeys,
)

// StandardKeys is an option that installs sensible default Ctrl-C (halt),
// Ctrl-Z (suspend), and Ctrl-L (force full redraw) handling.
var StandardKeys = Options(
	WithCtrlCHalt,
	WithCtrlZSuspend,
	WithCtrlLRedraw,
)

// WithContext returns an option that adds context to the terminal about to run
// a layer.
func WithContext(cs ...anansi.Context) Option {
	return optionFunc(func(app *app) error {
		app.Term.AddContext(cs...)
		return nil
	})
}

// WithLayerFunc returns an option that will call the given function, giving it
// a chance to wrap the layer.
func WithLayerFunc(f func(layer Layer) (Layer, error)) Option {
	return layerOptionFunc(f)
}

// WithTermLayerFunc returns an option that will call the given function, which
// may wrap the layer or manipulate the terminal. Any wrapper layer may retain
// the terminal for later use.
func WithTermLayerFunc(f func(layer Layer, term *anansi.Term) (Layer, error)) Option {
	return termOptionFunc(f)
}

// WithScreenLayerFunc returns an option that will call the given function,
// which may wrap the layer or manipulate the screen. Any wrapper layer may
// retain the screen for later use.
func WithScreenLayerFunc(f func(layer Layer, screen *anansi.TermScreen) (Layer, error)) Option {
	return screenOptionFunc(f)
}

type layerOptionFunc func(Layer) (Layer, error)
type termOptionFunc func(Layer, *anansi.Term) (Layer, error)
type screenOptionFunc func(Layer, *anansi.TermScreen) (Layer, error)

func (f layerOptionFunc) apply(app *app) (err error) {
	app.Layer, err = f(app.Layer)
	return err
}

func (f termOptionFunc) apply(app *app) (err error) {
	app.Layer, err = f(app.Layer, app.Term)
	return err
}

func (f screenOptionFunc) apply(app *app) (err error) {
	app.Layer, err = f(app.Layer, &app.screen)
	return err
}

// Options returns an option that applies each given option in order, stopping
// on the first one that fails.
func Options(opts ...Option) Option {
	if len(opts) == 0 {
		return nil
	}
	a := opts[0]
	for i := 1; i < len(opts); i++ {
		b := opts[i]
		if b == nil || b == Option(nil) {
			continue
		}
		if a == nil || a == Option(nil) {
			a = b
			continue
		}
		as, haveAs := a.(options)
		bs, haveBs := b.(options)
		if haveAs && haveBs {
			a = append(as[:len(as):len(as)], bs...)
		} else if haveAs {
			a = append(as[:len(as):len(as)], b)
		} else if haveBs {
			a = append(options{a}, bs...)
		} else {
			a = options{a, b}
		}
	}
	return a
}

type options []Option

func (opts options) apply(app *app) error {
	for _, opt := range opts {
		if opt != nil {
			if err := opt.apply(app); err != nil {
				return err
			}
		}
	}
	return nil
}

type optionFunc func(app *app) error

func (f optionFunc) apply(app *app) error { return f(app) }

type app struct {
	*anansi.Term
	Layer

	// virtual terminal screen to draw into
	screen anansi.TermScreen

	// event loop signals
	input  anansi.InputSignal
	halt   anansi.Signal
	resize anansi.Signal
	t      drawTimer
}

// init ialize term ancillaries and settings.
func (app *app) init(opt Option) error {
	// apply options
	if opt != nil {
		if err := opt.apply(app); err != nil {
			return err
		}
	}

	// handle SIGTERM and SIGINT when input is a terminal
	if anansi.IsStandardTermFile(app.Term.Input.File) {
		app.halt = anansi.Notify(syscall.SIGTERM, syscall.SIGINT)
		app.Term.AddContext(&app.halt)
	}

	// handle SIGWINCH when output is a terminal
	if anansi.IsStandardTermFile(app.Term.Output.File) {
		initLogs()
		app.resize = anansi.Notify(syscall.SIGWINCH)
		app.Term.AddContext(&app.resize)
	}

	// initialize fullscreen terminal control
	app.Term.AddMode(ansi.ModeAlternateScreen)
	if err := app.Term.SetRaw(true); err != nil {
		return err
	}
	return app.Term.SetEcho(false)
}

// run the layer in an event loop under the terminal's context.
func (app *app) run(term *anansi.Term) (err error) {
	app.resize.Send("initial screen resize")
	var needsDraw time.Duration
	for err == nil {
		if needsDraw == 0 {
			needsDraw = app.Layer.NeedsDraw()
		}
		if needsDraw != 0 {
			app.t.request(needsDraw)
			needsDraw = 0
		}

		select {

		case sig := <-app.halt.C:
			return anansi.SigErr(sig)

		case <-app.resize.C:
			app.screen.Screen = app.screen.Screen.Full()
			err = app.screen.SizeToTerm(term)
			needsDraw = 1

		case <-app.input.C:
			_, err = term.ReadAny()
			for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
				if _, herr := app.Layer.HandleInput(e, a); herr != nil {
					err = herr
					break
				}
			}

		case now := <-app.t.C:
			app.screen.Screen = drawLayerInto(app.Layer, app.screen.Screen, now)
			err = app.Flush(&app.screen)

		}
	}
	return err
}

// ClampNeedsDrawLayer returns a Layer that clamps any non-zero value returned
// by layer.NeedsDraw() to be no smaller than the given duration. The min
// argument defaults to time.Second/120 if zero.
func ClampNeedsDrawLayer(layer Layer, min time.Duration) Layer {
	if min == 0 {
		min = time.Second / 120
	}
	return clampNeedsDrawLayer{
		Layer: layer,
		min:   min,
	}
}

type clampNeedsDrawLayer struct {
	Layer
	min time.Duration
}

func (cl clampNeedsDrawLayer) NeedsDraw() time.Duration {
	d := cl.Layer.NeedsDraw()
	if d == 0 {
		return 0
	}

	if d < cl.min {
		return cl.min
	}
	return d
}

type drawTimer struct {
	C        <-chan time.Time
	deadline time.Time
	timer    *time.Timer
}

// request the timer to expire at most d time in the future, maybe sooner;
// ONLY IF d is positive, no-op if d is zero (or negative for that matter).
func (t *drawTimer) request(d time.Duration) {
	if d <= 0 {
		return
	}
	now := time.Now()
	if dd := t.deadline.Sub(now); dd > 0 && dd < d {
		return
	}
	t.deadline = now.Add(d)
	if t.timer == nil {
		t.timer = time.NewTimer(d)
		t.C = t.timer.C
	} else {
		t.timer.Reset(d)
	}
}

// cancel any timer set by request().
// Should not be called concurrently with request or receiving on C.
// TODO drop if not needed
func (t *drawTimer) cancel() {
	t.deadline = time.Time{}
	if !t.timer.Stop() {
		<-t.C
	}
}

// WithCtrlCHalt provides standard halt on Ctrl-C behavior, wrapping the layer
// so that it never sees Ctrl-C input. Applications that want to provide
// contextual handling of Ctrl-C should implement their own handling, and elide
// this option.
var WithCtrlCHalt = WithLayerFunc(func(layer Layer) (Layer, error) {
	return ctrlCLayer{
		Layer: layer,
		err:   fmt.Errorf("read %v", ansi.Escape(0x03)),
	}, nil
})

type ctrlCLayer struct {
	Layer
	err error
}

func (cc ctrlCLayer) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	switch e {
	case 0x03: // stop on Ctrl-C
		return true, cc.err
	default:
		return cc.Layer.HandleInput(e, a)
	}
}

// WithCtrlZSuspend provides standard suspend on Ctrl-Z behavior, wrapping the
// layer so that it never sees Ctrl-Z input. If an application needs to perform
// special preparation before suspending, it should do so by implementing
// anansi.Context and pass it with WithContext along with WithCtrlZSuspend.
var WithCtrlZSuspend = WithTermLayerFunc(func(layer Layer, term *anansi.Term) (Layer, error) {
	return ctrlZLayer{
		Layer: layer,
		term:  term,
	}, nil
})

type ctrlZLayer struct {
	Layer
	term *anansi.Term
}

func (cz ctrlZLayer) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	switch e {
	case '\x1a': // suspend on Ctrl-Z
		return true, cz.term.Suspend()
	default:
		return cz.Layer.HandleInput(e, a)
	}

}

// WithCtrlLRedraw provides standard force-full-redraw on Ctrl-L behavior,
// wrapping the layer so that it never sees Ctrl-L input. Redraw forcing is
// done by calling (*anansi.TermScreen).Invalidate.
var WithCtrlLRedraw = WithScreenLayerFunc(func(layer Layer, screen *anansi.TermScreen) (Layer, error) {
	return redrawLayer{
		Layer:  layer,
		screen: screen,
	}, nil
})

type redrawLayer struct {
	Layer
	screen *anansi.TermScreen
}

func (rl redrawLayer) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	switch e {
	case 0x0c: // force full redraw on Ctrl-L
		rl.screen.Invalidate()
		return true, nil
	default:
		return rl.Layer.HandleInput(e, a)
	}
}

func (rl redrawLayer) NeedsDraw() time.Duration {
	if rl.screen.Real.Rect.Size() == image.ZP {
		return time.Millisecond
	}
	return rl.Layer.NeedsDraw()
}
