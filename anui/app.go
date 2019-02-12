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
	return app.Term.RunWith(app.run)
}

// Option customizes RunTermLayer behavior.
type Option interface {
	apply(app *app) error
}

// DefaultDrawRate is the default number of layer draws per second to run at if
// specified by an other option.
const DefaultDrawRate = 30

// DefaultOptions are used if no options are given to RunTermLayer.
var DefaultOptions = Options(
	StandardKeys,
	WithDrawRate(DefaultDrawRate),
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

	appRunner
}

type appRunner interface {
	run(*app, *anansi.Term) error
}

// init ialize term ancillaries and settings.
func (app *app) init(opt Option) error {
	// apply options, with initial default sync draw rate
	opt = Options(WithDrawRate(DefaultDrawRate), opt)
	if err := opt.apply(app); err != nil {
		return err
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

func (app *app) run(term *anansi.Term) error {
	return app.appRunner.run(app, term)
}

// WithDrawRate sets the target draw rate for a layer application: draw times
// requested by Layer.NeedsDraw will be adjusted a consistent timing schedule
// that will target a time.Second/rate delay between calling Layer.Draw.
func WithDrawRate(rate int) Option {
	return &syncAppRunner{
		drawRate: rate,
	}
}

type syncAppRunner struct {
	// config
	drawRate     int
	drawInterval time.Duration

	// timer state
	nextDraw time.Time
	lastDraw time.Time
	timer    *time.Timer

	// controller state
	adjust        time.Duration
	p, i, d       float64
	timerErrRange [2]float64
	timerErr      float64
	timerErrI     float64
	timerErrD     float64
}

func (sar *syncAppRunner) apply(app *app) error {
	app.appRunner = sar
	return nil
}

func (sar *syncAppRunner) run(app *app, term *anansi.Term) error {
	if sar.drawRate == 0 {
		sar.drawRate = DefaultDrawRate
	}
	if sar.drawInterval == 0 {
		sar.drawInterval = time.Second / time.Duration(sar.drawRate)
	}
	if sar.p == 0 {
		sar.p = 1.0
		sar.i = 0.5
		sar.d = 0.25
		sar.timerErrRange[1] = float64(sar.drawInterval) / 2
		sar.timerErrRange[0] = -sar.timerErrRange[1]
	}

	sar.nextDraw = time.Now()
	sar.lastDraw = sar.nextDraw.Add(-sar.drawInterval)
	sar.timer = time.NewTimer(0)
	defer func() {
		sar.timer = nil
	}()

	for {
		select {

		// halt event, stop now
		case sig := <-app.halt.C:
			return anansi.SigErr(sig)

		// terminal resized, read the new size, and trigger a draw
		case <-app.resize.C:
			if err := app.screen.SizeToTerm(term); err != nil {
				return err
			}
			sar.scheduleDraw(1)

		// time to draw
		case now := <-sar.timer.C:
			sar.updateControl(now.Sub(sar.nextDraw))
			app.screen.Screen = drawLayerInto(app.Layer, app.screen.Screen, now)
			sar.lastDraw, sar.nextDraw = now, time.Time{}
			if err := term.Flush(&app.screen); err != nil {
				return err
			}
			sar.scheduleDraw(app.Layer.NeedsDraw())

		// input received, process it
		case <-app.input.C:
			_, err := term.ReadAny()
			for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
				if _, herr := app.Layer.HandleInput(e, a); herr != nil {
					err = herr
					break
				}
			}
			if err != nil {
				return err
			}
			sar.scheduleDraw(app.Layer.NeedsDraw())

		}
	}
}

func (sar *syncAppRunner) updateControl(drawTimeErr time.Duration) {
	timerErr := float64(drawTimeErr)
	if timerErr > sar.timerErrRange[1] {
		sar.timerErrD = 0
		sar.timerErrI = 0
		sar.timerErr = sar.timerErrRange[1]
	} else if timerErr < sar.timerErrRange[0] {
		sar.timerErrD = 0
		sar.timerErrI = 0
		sar.timerErr = sar.timerErrRange[0]
	} else {
		sar.timerErrD = timerErr - sar.timerErr
		sar.timerErrI += timerErr
		sar.timerErr = timerErr
		if sar.timerErrI > sar.timerErrRange[1] {
			sar.timerErrI = sar.timerErrRange[1]
		} else if sar.timerErrI < sar.timerErrRange[0] {
			sar.timerErrI = sar.timerErrRange[0]
		}
	}
	sar.adjust = time.Duration(
		sar.p*sar.timerErr +
			sar.i*sar.timerErrI +
			sar.d*sar.timerErrD)
}

func (sar *syncAppRunner) scheduleDraw(need time.Duration) {
	const minTimerSet = 10 * time.Microsecond

	if need == 0 {
		return
	}

	now := time.Now()

	// start of future draw containing needed draw
	drawsUntil := (need + sar.drawInterval - 1) / sar.drawInterval
	then := sar.lastDraw.Add(sar.drawInterval * drawsUntil)

	if timerSet := !sar.nextDraw.IsZero(); !(timerSet && then.Before(sar.nextDraw)) {
		until := then.Sub(now)

		// apply clamped control adjustment
		if sar.adjust > 0 {
			until -= sar.adjust
		}

		// trigger immediate draw if under minimum
		if until < minTimerSet {
			if timerSet {
				if !sar.timer.Stop() {
					<-sar.timer.C
				}
			}
			until = 0
		}

		sar.timer.Reset(until)
		sar.nextDraw = then
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
