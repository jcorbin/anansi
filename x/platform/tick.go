package platform

import (
	"time"

	"github.com/jcorbin/anansi"
)

// Ticker implements a contextual time.Ticker, start/stopping it during
// terminal enter/exit.
type Ticker struct {
	d time.Duration
	t *time.Ticker
	c chan struct{}
	// TODO useful to indirect t.C so that Enter can provide an immediate initial tick?
}

// Enter starts a new ticker, after stopping any prior one for good measure;
// always returns nil error.
func (ct *Ticker) Enter(term *anansi.Term) error {
	_ = ct.Exit(term)
	if ct.d == 0 {
		ct.d = time.Second / defaultFrameRate
	}
	ct.t = time.NewTicker(ct.d)
	ct.c = make(chan struct{})
	return nil
}

// Exit stops any running ticker; always returns nil error.
func (ct *Ticker) Exit(term *anansi.Term) error {
	if ct.t != nil {
		ct.t.Stop()
		ct.t = nil
	}
	if ct.c != nil {
		close(ct.c)
		ct.c = nil
	}
	return nil
}

// Wait blocks for the next ticker time, returning zero time if the Ticker
// isn't active, or is Exit-ed first.
func (ct *Ticker) Wait() time.Time {
	var tc <-chan time.Time
	if ct.t != nil {
		tc = ct.t.C
	}
	select {
	case t := <-tc:
		return t
	case <-ct.c:
	}
	return time.Time{}
}
