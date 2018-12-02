package anansi

import (
	"os"
	"os/signal"
)

// Notify is a convenience constructor for Signal values.
func Notify(notify ...os.Signal) Signal {
	return Signal{notify, nil}
}

// Signal supports a Term-contextual signal notification.
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
