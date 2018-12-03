package platform

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var errNoTerm = errors.New("platform not attached to a terminal")

// MustRun call Run, calling os.Exit(1) if it returns a non-nil error.
func MustRun(f *os.File, run func(*Platform) error, opts ...Option) {
	if err := Run(f, run, opts...); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Run is a convenience wrapper that calls the run function with a newly
// created Platform activated under a newly constructed anansi.Term.
func Run(f *os.File, run func(*Platform) error, opts ...Option) error {
	p, err := New(opts...)
	if err != nil {
		return err
	}
	return anansi.NewTerm(f, p).RunWith(func(_ *anansi.Term) error {
		return run(p)
	})
}

const defaultFrameRate = 60

// New creates a platform layer for running interactive fullscreen terminal
// applications.
func New(opts ...Option) (*Platform, error) {
	p := &Platform{}

	p.termContext = anansi.Contexts(
		&p.input,
		&p.output,
		&p.Config,
		&p.ticker,
		&p.bg,
	)

	p.mode.AddMode(
		ansi.ModeAlternateScreen,
		ansi.ModeMouseSgrExt,   // TODO detection?
		ansi.ModeMouseBtnEvent, // TODO options?
		ansi.ModeMouseAnyEvent, // TODO options?
	)
	p.mode.AddModeSeq(ansi.SoftReset, ansi.SGRReset) // TODO options?

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

	term *anansi.Term

	termContext anansi.Context
	buf         anansi.Buffer

	mode   anansi.Mode
	input  anansi.Input
	output anansi.Output

	events Events
	ticker Ticker

	recording *os.File
	replay    *replay
	bg        BackgroundWorkers

	State
	Time   time.Time // internal time (rewinds during replay)
	screen anansi.Screen

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
	Output *anansi.Screen
}

// IsReplayDone returns true if the error was due to a replay session finishing.
func IsReplayDone(err error) bool {
	return err == errReplayDone
}

// IsReplayStop returns true if the error was due to the user canceling a replay.
func IsReplayStop(err error) bool {
	return err == errReplayStop
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

	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer signal.Stop(stopSig)

	resizeSig := make(chan os.Signal, 1)
	signal.Notify(resizeSig, syscall.SIGWINCH)
	defer signal.Stop(resizeSig)

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

		// poll for events and input
		for polling := true; polling; {
			select {
			case sig := <-stopSig:
				ctx.Err = errOr(ctx.Err, signalError{sig})
			case <-resizeSig:
				ctx.Err = errOr(ctx.Err, p.readSize())
			default:
				p.events.Clear()
				n, err := p.input.ReadAny()
				if n > 0 {
					p.events.DecodeInput(&p.input)
				}
				ctx.Err = errOr(ctx.Err, err)
				polling = false
			}
		}

		// run current frame update
		if ctx.Update(); ctx.Err == nil {
			ctx.Err = p.output.Flush(ctx.Output)
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

func (p *Platform) readSize() error {
	if p.term == nil {
		return errNoTerm
	}
	sz, err := p.term.Size()
	if err == nil {
		if p.screen.Resize(sz) && p.recording != nil {
			err = p.recordSize()
		}
	}
	return err
}

// Enter applies terminal context, including raw mode and ansi mode sequences,
// wires up input, output, and initializes the tick controller.
func (p *Platform) Enter(term *anansi.Term) error {
	if p.term != nil {
		return errors.New("Platform may only be used under a single terminal")
	}
	p.term = term
	if err := p.readSize(); err != nil {
		return fmt.Errorf("initial term size request failed: %v", err)
	}
	if err := p.term.SetRaw(true); err != nil {
		return err
	}

	p.buf.Write(p.mode.Set)
	if p.buf.Len() > 0 {
		if _, err := p.buf.WriteTo(term.File); err != nil {
			return err
		}
	}

	if err := p.termContext.Enter(term); err != nil {
		return err
	}

	return nil
}

// Exit tears down everything that Enter setup.
func (p *Platform) Exit(term *anansi.Term) (err error) {
	if term != p.term {
		return nil
	}

	p.buf.WriteSGR(p.screen.CursorState.MergeSGR(0))
	p.buf.WriteSeq(p.screen.CursorState.Show())
	p.buf.Write(p.mode.Reset)
	if p.buf.Len() > 0 {
		_, err = p.buf.WriteTo(term.File)
	}

	err = errOr(err, p.termContext.Exit(term))
	p.screen.Resize(image.ZP)
	p.term = nil
	return err
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
		ctx.Err = errOr(ctx.Err, ctx.Platform.readSize())
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

type signalError struct{ sig os.Signal }

func (se signalError) String() string { return fmt.Sprintf("signal %v", se.sig) }
func (se signalError) Error() string  { return se.String() }

func errOr(a, b error) error {
	if a != nil {
		return a
	}
	return b
}
