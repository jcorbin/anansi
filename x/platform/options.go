package platform

import (
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
		p.ticker.d = time.Second / time.Duration(fps)
		p.FPSEstimate.data = make([]float64, fps)
		p.Timing.ts = make([]time.Time, timingPeriod)
		p.Timing.ds = make([]time.Duration, timingPeriod)
		return nil
	})
}

func hasConfig(opts []Option) bool {
	for _, opt := range opts {
		if _, isConfig := opt.(Config); isConfig {
			return true
		}
	}
	return false
}
