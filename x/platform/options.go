package platform

import (
	"flag"
	"time"
)

// Option customizes Platform's behavior.
type Option interface {
	apply(*Platform) error
}

type optionFunc func(*Platform) error

func (f optionFunc) apply(p *Platform) error { return f(p) }

// FrameRate changes the platform's Frames-Per-Second rate, which defaults to 60.
func FrameRate(fps int) Option {
	return optionFunc(func(p *Platform) error {
		timingPeriod := fps / 4
		p.ticks.SetGoal(fps)
		p.FPSEstimate.data = make([]float64, fps)
		p.Timing.ts = make([]time.Time, timingPeriod)
		p.Timing.ds = make([]time.Duration, timingPeriod)
		return nil
	})
}

type _platformFlags struct{ Config }

func (pf *_platformFlags) init() {
	flag.StringVar(&pf.LogFileName, "platform.logfile", "", "write logs to a file (in addition to in-memory buffer)")
	flag.StringVar(&pf.CPUProfileName, "platform.cpuprofile", "", "enables platform cpu profiling")
	flag.StringVar(&pf.MemProfileName, "platform.memprofile", "", "enables platform memory profiling")
	flag.StringVar(&pf.TraceFileName, "platform.tracefile", "", "enables platform execution tracing")

	flag.BoolVar(&pf.StartTiming, "platform.timing", false, "measure timing from the beginning")
	flag.BoolVar(&pf.LogTiming, "platform.timing.log", false, "measure and log timing from the beginning")
}

var platformFlags = _platformFlags{}

func init() {
	platformFlags.init()
}

func (pf *_platformFlags) apply(p *Platform) error {
	if !flag.Parsed() {
		flag.Parse()
	}
	return pf.Config.apply(p)
}
