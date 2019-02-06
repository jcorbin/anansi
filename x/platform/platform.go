package platform

import (
	"bytes"
	"flag"
	"image"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// MustRun call Run, calling os.Exit(1) if it returns a non-nil error.
func MustRun(in, out *os.File, run func(*Platform) error, opts ...Option) {
	anansi.MustRun(Run(in, out, run, opts...))
}

// Run is a convenience wrapper that calls the run function with a newly
// created Platform activated under a newly constructed anansi.Term.
func Run(in, out *os.File, run func(*Platform) error, opts ...Option) error {
	p, err := New(in, out, opts...)
	if err != nil {
		return err
	}
	return p.RunWith(run)
}

const defaultFrameRate = 60

// New creates a platform layer for running interactive fullscreen terminal
// applications.
func New(in, out *os.File, opts ...Option) (*Platform, error) {
	p := &Platform{}

	p.stop = anansi.Notify(syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	p.resize = anansi.Notify(syscall.SIGWINCH)

	p.term = anansi.NewTerm(in, out,
		&p.stop,
		&p.resize,
		&p.screen,
		&p.Config,
		&p.ticker,
		&p.bg,
	)

	p.sigio.C = make(chan os.Signal, 1)
	if err := p.term.Notify(p.sigio.C); err != nil {
		return nil, err
	}

	_ = p.term.SetRaw(true)
	p.term.AddMode(
		ansi.ModeAlternateScreen,
		ansi.ModeMouseSgrExt,   // TODO detection?
		ansi.ModeMouseBtnEvent, // TODO options?
		ansi.ModeMouseAnyEvent, // TODO options?
	)
	p.term.AddModeSeq(ansi.SoftReset, ansi.SGRReset) // TODO options?

	p.ticker.d = time.Second / defaultFrameRate

	timingPeriod := defaultFrameRate / 4
	p.FPSEstimate.data = make([]float64, defaultFrameRate)
	p.Timing.ts = make([]time.Time, timingPeriod)
	p.Timing.ds = make([]time.Duration, timingPeriod)
	p.Telemetry.coll.rusage.data = make([]rusageEntry, defaultFrameRate*10)
	p.bg.workers = append(p.bg.workers, &p.Telemetry.coll, &Logs)

	if !flag.Parsed() && !hasConfig(opts) {
		flagConfig := Config{}
		flagConfig.AddFlags(flag.CommandLine, "platform.")
		flag.Parse()
		if err := flagConfig.apply(p); err != nil {
			return nil, err
		}
	}

	if err := p.HUD.apply(p); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if err := opt.apply(p); err != nil {
			return nil, err
		}
	}

	if err := p.Config.setup(p); err != nil {
		return nil, err
	}

	return p, nil
}

// Platform is a high level abstraction for implementing frame-oriented
// interactive fullscreen terminal programs.
type Platform struct {
	Config

	term   *anansi.Term
	stop   anansi.Signal
	resize anansi.Signal
	sigio  anansi.Signal
	buf    anansi.Buffer
	events Events
	ticker Ticker

	recording *os.File
	replay    *replay
	bg        BackgroundWorkers

	State
	Time   time.Time // internal time (rewinds during replay)
	screen anansi.TermScreen

	Telemetry

	client Client

	HUD HUD
}

// State contains serializable Platform state.
type State struct {
	Paused   bool
	LastTime time.Time
	LastSize image.Point
}

// Client runs under a platform, processing input and generating output within
// each frame Context.
type Client interface {
	Update(*Context) error
}

// ClientFunc is a convenient way to implement a Client (e.g. for testing).
type ClientFunc func(*Context) error

// Update runs the aliased function.
func (f ClientFunc) Update(ctx *Context) error { return f(ctx) }

// Context manages frame input and output state within a Platform.
type Context struct {
	*Platform
	Time   time.Time
	Err    error
	Redraw bool
	Input  *Events
	Output *anansi.TermScreen
}

// IsReplayDone returns true if the error was due to a replay session finishing.
func IsReplayDone(err error) bool {
	return err == errReplayDone
}

// IsReplayStop returns true if the error was due to the user canceling a replay.
func IsReplayStop(err error) bool {
	return err == errReplayStop
}

// RunWith runs the given function under the platform anansi.Term; such
// function should call Platform.Run one or more times.
func (p *Platform) RunWith(run func(*Platform) error) error {
	return p.term.RunWith(func(_ *anansi.Term) error {
		return run(p)
	})
}

// Run a client under a platform. It loads client state from any active replay
// buffer, and then runs the client under a ticker loop.
func (p *Platform) Run(client Client) (err error) {
	p.client = client

	if p.replay != nil {
		if err := p.readState(bytes.NewReader(p.replay.cereal)); err != nil {
			return err
		}
		p.replay.cur = p.replay.input
	}

	log.Printf("running %T", p.client)
	defer func() {
		log.Printf("run done: %v", err)
	}()

	for p.Time = time.Now(); !p.Time.IsZero(); p.Time = p.ticker.Wait() {
		// update performance data
		if err := p.Telemetry.update(p); err != nil {
			return err
		}

		ctx := p.Context()

		// poll for signals
		for polling := true; polling; {
			select {
			case sig := <-p.stop.C:
				ctx.Err = errOr(ctx.Err, anansi.SigErr(sig))
			case <-p.resize.C:
				ctx.Err = errOr(ctx.Err, p.screen.SizeToTerm(p.term))
			case <-p.sigio.C:
				p.events.Clear()
				n, err := p.term.ReadAny()
				if n > 0 {
					p.events.DecodeInput(&p.term.Input)
				}
				ctx.Err = errOr(ctx.Err, err)
			default:
				polling = false
			}
		}

		// run current frame update
		if ctx.Update(); ctx.Err == nil {
			ctx.Err = p.term.Flush(ctx.Output)
		}

		// notify background workers
		if ctx.Err == nil {
			ctx.Err = p.bg.Notify()
		}

		if ctx.Err != nil {
			return ctx.Err
		}
	}
	return nil
}

// Context returns a new Context bound to the platform.
func (p *Platform) Context() Context {
	return Context{
		Platform: p,
		Input:    &p.events,
		Output:   &p.screen,
		Time:     p.Time,
	}
}

// Update runs a client round:
// - resets screen buffer
// - hides cursor (TODO)
// - processes user Ctrl-L to implement redraw flag
// - hands off to any active replay
// - re-reads terminal size on redraw
// - processes user Ctrl-R to toggle recording / replaying
// - runs the Platform client Update, under HUD Update
// - flushes screen buffer
func (ctx *Context) Update() {
	ctx.Output.Reset()
	outBounds := ctx.Output.Bounds()

	// Ctrl-L forces a size refresh
	ctx.Redraw = ctx.Input.CountRune('\x0c') > 0

	// Resize causes a redraw
	ctx.Redraw = ctx.Redraw ||
		outBounds.Size() == image.ZP ||
		outBounds.Size() != ctx.LastSize

	if ctx.Redraw {
		ctx.Output.Invalidate()
	}

	if ctx.replay != nil {
		if ctx.replay.update(ctx); ctx.replay != nil || ctx.Err != nil {
			return
		}
		// replay erased itself
	}

	if ctx.Redraw {
		ctx.Err = errOr(ctx.Err, ctx.Output.SizeToTerm(ctx.term))
	}

	// recording / replaying toggle on Ctrl-R
	if ctx.Input.CountRune('\x12')%2 == 1 {
		ctx.Err = errOr(ctx.Err, ctx.toggleRecRep())
	}

	ctx.Err = errOr(ctx.Err, ctx.runClient())
}

func (ctx *Context) runClient() error {
	if ctx.Paused {
		ctx.Time = ctx.Platform.LastTime
	}
	err := ctx.HUD.Update(ctx, ctx.client)
	ctx.Platform.LastTime = ctx.Time
	ctx.Platform.LastSize = ctx.Output.Bounds().Size()
	return err
}

// Suspend restores terminal context to pre-platform-run settings, suspends the
// current process, and then restores platform terminal context once resumed;
// returns any error preventing any of that.
func (p *Platform) Suspend() error { return p.term.Suspend() }

func errOr(a, b error) error {
	if a != nil {
		return a
	}
	return b
}
