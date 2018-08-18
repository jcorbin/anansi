package platform

import (
	"sync"
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
	if err := core.Error(); err != nil {
		return err
	}
	select {
	case core.w <- struct{}{}:
	default:
	}
	return nil
}
