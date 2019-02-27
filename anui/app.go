package anui

import (
	"fmt"
	"image"
	"math"
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
	WithSyncDrawRate(DefaultDrawRate),
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
	opt = Options(WithSyncDrawRate(DefaultDrawRate), opt)
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

// WithSyncDrawRate sets the target draw rate for a low draw rate /
// sporadically animated layer application.
//
// All input handling, drawing, and output rendering is done in a synchronous
// loop on a single goroutine.
//
// Draw timing is managed through a dynamically controlled time.Timer.
//
// XXX This is a sensible default for a typical terminal application that does not
// make heavy use of animation. Such applications do not need to draw
// frequently, primarily drawing in response to direct user input.
func WithSyncDrawRate(rate int) Option {
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
		sar.p = 0.0625
		sar.i = 0.25
		sar.d = 0.125
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
	sar.timerErrD = clampToRange(timerErr-sar.timerErr, sar.timerErrRange)
	sar.timerErrI = clampToRange(sar.timerErrI+timerErr, sar.timerErrRange)
	sar.timerErr = clampToRange(timerErr, sar.timerErrRange)
	sar.adjust = time.Duration(math.Floor(clampToRange(
		sar.p*sar.timerErr+
			sar.i*sar.timerErrI+
			sar.d*sar.timerErrD,
		sar.timerErrRange)))
}

func (sar *syncAppRunner) scheduleDraw(need time.Duration) {
	const minTimerSet = 10 * time.Microsecond

	if need == 0 {
		return
	}

	// needed draw deadline
	now := time.Now()
	deadline := now.Add(need)

	// start of future draw containing deadline
	drawsUntil := (deadline.Sub(sar.lastDraw) + sar.drawInterval - 1) / sar.drawInterval
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

func clampToRange(val float64, valRange [2]float64) float64 {
	if val < valRange[0] {
		return valRange[0]
	}
	if val > valRange[1] {
		return valRange[1]
	}
	return val
}

// WithAsyncDrawRate sets the target draw rate for a high draw rate / heavily
// animated layer application.
//
// Input handling and drawing are done in a synchronous loop on a goroutine,
// but output rendering is driven asynchronously in an ancillary goroutine.
//
// Render timing is driven directly by a time.Ticker in the ancillary, which in
// turn drives Draw timing on the primary goroutine.
//
// XXX This is a sensible default for an intensive application, e.g. one that
// employs animation, needing to draw every frame irrespective of user input.
func WithAsyncDrawRate(rate int) Option {
	return &asyncAppRunner{
		drawRate: rate,
	}
}

type asyncAppRunner struct {
	drawRate     int
	drawInterval time.Duration
}

func (aar *asyncAppRunner) apply(app *app) error {
	app.appRunner = aar
	return nil
}

func (aar *asyncAppRunner) run(app *app, term *anansi.Term) error {
	if aar.drawRate == 0 {
		aar.drawRate = DefaultDrawRate
	}
	if aar.drawInterval == 0 {
		aar.drawInterval = time.Second / time.Duration(aar.drawRate)
	}

	type drawReq struct {
		anansi.Screen
		force    bool
		drawTime time.Time
		nextDraw time.Time
	}

	type flushReq struct {
		drawReq
		skip bool
	}

	toDraw := make(chan drawReq, 1)
	flushErr := make(chan error, 1)
	defer func() {
		for range toDraw {
		}
	}()

	toFlush := make(chan flushReq, 1)
	defer close(toFlush)

	// runtime.LockOSThread()
	// defer runtime.UnlockOSThread()

	go func() {
		// runtime.LockOSThread()
		defer close(toDraw)
		defer close(flushErr)

		ticker := time.NewTicker(aar.drawInterval)
		defer ticker.Stop()

		toDraw <- drawReq{
			Screen:   app.screen.Screen,
			nextDraw: time.Now().Add(aar.drawInterval),
			force:    true,
		}
		app.screen.Screen = anansi.Screen{}

		// TODO less skipped frame on resize?

		for req := range toFlush {
			resp := drawReq{
				Screen: req.Screen,
			}
			var err error
			app.screen.Screen = req.Screen

			select {

			case <-app.resize.C:
				err = app.screen.SizeToTerm(term)
				resp.Screen = app.screen.Screen
				resp.force = true

			case drawTime := <-ticker.C:
				if !req.skip {
					app.screen.UserCursor = req.Screen.Cursor
					err = term.Flush(&app.screen)
				}
				resp.drawTime = drawTime // TODO should this be the new now?
				resp.nextDraw = drawTime.Add(aar.drawInterval)

			}

			app.screen.Screen = anansi.Screen{}
			if err != nil {
				flushErr <- err
				return
			}
			toDraw <- resp
		}
	}()

	nextDraw := time.Now()

	for {
		select {
		// halt event, stop now
		case sig := <-app.halt.C:
			return anansi.SigErr(sig)

		// output flush error
		case err := <-flushErr:
			return err

		// time to draw
		case req := <-toDraw:
			resp := flushReq{
				drawReq: req,
				skip:    true,
			}
			if req.force || req.nextDraw.After(nextDraw) {
				sc := req.Screen
				sc = sc.Full()
				sc.Clear()
				sc = app.Layer.Draw(sc, req.drawTime)
				if then := time.Now().Add(app.Layer.NeedsDraw()); then.After(nextDraw) {
					nextDraw = then
				}
				resp.skip = false
				resp.Screen = sc.Full()
			}
			toFlush <- resp

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
			if then := time.Now().Add(app.Layer.NeedsDraw()); then.After(nextDraw) {
				nextDraw = then
			}
		}
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
