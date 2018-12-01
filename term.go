package anansi

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// NewTerm constructs a new terminal attached the given file pair, and with the
// given context.
func NewTerm(in, out *os.File, cs ...Context) *Term {
	term := &Term{}
	term.Input.File = in
	term.Output.File = out
	term.initContext()
	term.ctx = Contexts(term.ctx, Contexts(cs...))
	return term
}

// Term combines a terminal file handle with attribute control and further
// Context-ual state.
type Term struct {
	Attr
	Mode
	Input
	Output

	active bool
	under  bool
	ctx    Context
}

func (term *Term) initContext() {
	if term.ctx == nil {
		term.ctx = Contexts(
			&term.Input,
			&term.Output,
			&term.Attr,
			&term.Mode)
	}
}

// RunWith runs the given function within the terminal's context, Enter()ing it
// if necessary, and Exit()ing it if Enter() was called after the given
// function returns. Exit() is called even if the within function returns an
// error or panics.
func (term *Term) RunWith(within func(*Term) error) (err error) {
	if term.active {
		return within(term)
	}

	term.active = true
	defer func() {
		term.active = false
	}()

	if !term.under {
		term.under = true
		defer func() {
			term.under = false
		}()
	}

	term.initContext()

	defer func() {
		if cerr := term.ctx.Exit(term); err == nil {
			err = cerr
		}
	}()
	if err = term.ctx.Enter(term); err == nil {
		err = within(term)
	}
	return err
}

// RunWithout runs the given function without the terminal's context, Exit()ing
// it if necessary, and Enter()ing it if deactivation was necessary.
// Re-Enter() is not called is not done if a non-nil error is returned, or if
// the without function panics.
func (term *Term) RunWithout(without func(*Term) error) (err error) {
	if !term.active {
		return without(term)
	}
	if err = term.ctx.Exit(term); err == nil {
		term.active = false
		if err = without(term); err == nil {
			if err = term.ctx.Enter(term); err == nil {
				term.active = true
			}
		}
	}
	return err
}

// Suspend signals the process to stop, and blocks on its later restart. If the
// terminal is currently active, this is done under RunWithout to restore prior
// terminal state.
func (term *Term) Suspend() error {
	if term.active {
		return term.RunWithout((*Term).Suspend)
	}

	cont := make(chan os.Signal)
	signal.Notify(cont, syscall.SIGCONT)
	defer signal.Stop(cont)
	log.Printf("suspending")
	if err := syscall.Kill(0, syscall.SIGTSTP); err != nil {
		return err
	}
	sig := <-cont
	log.Printf("resumed, signal: %v", sig)
	return nil
}

// MustRun is a useful wrapper for the outermost Term.RunWith: it log.Fatals
// any non-nil error.
func MustRun(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
