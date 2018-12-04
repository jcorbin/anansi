package anansi

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Notify is a convenience constructor for Signal values.
func Notify(notify ...os.Signal) Signal {
	return Signal{notify, nil}
}

// Signal supports Term-contextual signal notification.
//
// Example:
//
// 	usr1 := anansi.Notify(syscall.SIGUSR1)
// 	anansi.NewTerm(in, out, &usr1).RunWith(func(term *anansi.Term) error {
// 		for sig := range usr1 {
// 			// TODO something
// 		}
// 	})
//
// See ExampleSignal for a more normative example.
type Signal struct {
	Notify []os.Signal
	C      chan os.Signal
}

// Enter calls Open, ensuring signal notification is started.
func (sig *Signal) Enter(term *Term) error { return sig.Open() }

// Exit is a no-op; NOTE signal notification is not stopped during temporary
// teardown (e.g when suspending).
func (sig *Signal) Exit(term *Term) error { return nil }

// Open allocates a signal channel (of capacity 1) if none has been allocated
// already, and then calls signal.Notify if sig.Notify is non-empty.
func (sig *Signal) Open() error {
	if sig.C == nil {
		sig.C = make(chan os.Signal, 1)
	}
	if len(sig.Notify) > 0 {
		signal.Notify(sig.C, sig.Notify...)
	}
	return nil
}

// Close stops notification any non-nil channel, and nils it out.
func (sig *Signal) Close() error {
	if sig.C != nil {
		signal.Stop(sig.C)
		sig.C = nil
	}
	return nil
}

// AsErr is a convenience for handling a termination signal set: it returns a
// SignalError if from any available signal on the channel, or nil if no signal
// is available.
func (sig *Signal) AsErr() error {
	select {
	case s := <-sig.C:
		return SigErr(s)
	default:
		return nil
	}
}

type syntheticSignal string

func (ss syntheticSignal) String() string { return string(ss) }
func (ss syntheticSignal) Signal()        {}

// Send a synthetic signal to the channel, e.g. to prime it so that it fires
// immediately to initialize a run loop.
func (sig *Signal) Send(mess string) {
	if sig.C == nil {
		sig.C = make(chan os.Signal, 1)
		sig.C <- syntheticSignal(mess)
		return
	}
	select {
	case sig.C <- syntheticSignal(mess):
	default:
	}
}

// SigErr is a convenience constructor for SignalError values.
func SigErr(sig os.Signal) error { return SignalError{sig} }

// SignalError supports passing signals as errors.
type SignalError struct{ Sig os.Signal }

func (sig SignalError) String() string { return sig.Sig.String() }
func (sig SignalError) Error() string  { return fmt.Sprintf("signal %v", sig.Sig.String()) }

// Signal implements the os.Signal interface, allowing SignalError to double
// both as an error, and a wrapped signal value.
func (sig SignalError) Signal() {}

// ExitCode returns the corresponding value that the program should exit with
// due to the wrapped signal.
func (sig SignalError) ExitCode() int {
	switch impl := sig.Sig.(type) {
	case syscall.Signal:
		return -int(impl)
	}
	return 1
}

var _ os.Signal = SignalError{}
