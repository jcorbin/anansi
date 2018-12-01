package platform

import (
	"sync"

	"github.com/jcorbin/anansi"
)

// BackgroundWorker supports doing deferred work in between frames of a
// platform run loop. The Start() and Stop() methods are called before/after
// the run loop. The Notify() method is called at the end of the run loop,
// before going back to sleep to wait for the next frame tick.
//
// The astute reader will note that BackgroundWorker itself provides no support
// for synchronizing with the background work. Any such needs must be
// implemented in-situ by the BackgroundWorker; no generic support is provided
// to block the start of the next frame  to wait for prior triggered work to
// finish.
type BackgroundWorker interface {
	Start() error
	Stop() error
	Notify() error
}

// BackgroundWorkers implements an anansi.Context-ually managed collection of
// background workers.
type BackgroundWorkers struct {
	workers []BackgroundWorker
}

// Enter starts the background workers, stopping on and returning first error.
func (bg BackgroundWorkers) Enter(term *anansi.Term) error {
	for i := 0; i < len(bg.workers); i++ {
		if err := bg.workers[i].Start(); err != nil {
			return err
		}
	}
	return nil
}

// Exit stops all background workers, returning the first error, but stopping
// all regardless.
//
// TODO consider whether Close() error would be a better idea: don't gracefully
// stop background work, e.g. when preparing to suspend.
func (bg BackgroundWorkers) Exit(term *anansi.Term) (err error) {
	for i := len(bg.workers) - 1; i >= 0; i-- {
		if serr := bg.workers[i].Stop(); err == nil {
			err = serr
		}
	}
	return err
}

// Notify all background workers, stopping on and returning the first error.
func (bg BackgroundWorkers) Notify() error {
	for i := 0; i < len(bg.workers); i++ {
		if err := bg.workers[i].Notify(); err != nil {
			return err
		}
	}
	return nil
}

type bgWorkerCore struct {
	sync.Mutex
	w chan struct{}
	e chan error
}

func (core *bgWorkerCore) Start() error {
	core.w = make(chan struct{}, 1)
	core.e = make(chan error, 1)
	return nil
}

func (core *bgWorkerCore) Stop() error {
	if core.w != nil {
		close(core.w)
		core.w = nil
	}
	if core.e == nil {
		return nil
	}
	return <-core.e
}

func (core *bgWorkerCore) Error() error {
	select {
	case err := <-core.e:
		if err != nil {
			close(core.w)
			core.w = nil
			return err
		}
	default:
	}
	return nil
}

func (core *bgWorkerCore) Notify() error {
	if core.w != nil {
		if err := core.Error(); err != nil {
			return err
		}
		select {
		case core.w <- struct{}{}:
		default:
		}
	}
	return nil
}
