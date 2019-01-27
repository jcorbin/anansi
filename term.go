package anansi

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// OpenTerm opens the standard terminal, attached to the controlling terminal.
// Prefers to existing os.Stdin and os.Stdout files if they're still attached,
// opens /dev/tty otherwise.
func OpenTerm() (*Term, error) {
	in, out, err := openTermFiles(os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}
	return NewTerm(in, out), nil
}

// openTermFiles opens /dev/tty if the given files are not terminals,
// returning a usable pair of terminal in/out file handles. It also redirects
// "log" package output to the Logs buffer if it hasn't already been done
// (e.g. by calling OpenLogFile).
func openTermFiles(in, out *os.File) (_, _ *os.File, rerr error) {
	if !IsTerminal(in) {
		f, err := os.OpenFile("/dev/tty", syscall.O_RDONLY, 0)
		if err != nil {
			return nil, nil, err
		}
		defer func() {
			if rerr != nil {
				in.Close()
			}
		}()
		in = f
	}
	if !IsTerminal(out) {
		f, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0)
		if err != nil {
			return nil, nil, err
		}
		out = f
	}
	return in, out, nil
}

// IsStandardTermFile returns true only if the given file's name corresponds to
// a standard process terminal file; that is if it's one of /dev/stdin,
// /dev/stdout, /dev/stderr, or /dev/tty.
func IsStandardTermFile(f *os.File) bool {
	switch f.Name() {
	case "/dev/stdin", "/dev/stdout", "/dev/stderr", "/dev/tty":
		return true
	}
	return false
}

// NewTerm constructs a new terminal attached the given file pair, and with the
// given context.
func NewTerm(in, out *os.File, cs ...Context) *Term {
	term := &Term{}
	term.Input.File = in
	term.Output.File = out
	term.ctx = Contexts(
		&term.Input,
		&term.Output,
		&term.Attr,
		&term.Mode,
		Contexts(cs...))
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

// RunWith runs the given function within the terminal's context, Enter()ing it
// if necessary, and Exit()ing it if Enter() was called after the given
// function returns. Exit() is called even if the within function returns an
// error or panics.
//
// If the context implements a `Close() error` method, then it will also be
// called immediately after Exit(). This allows a Context implementation to
// differentiate between temporary teardown, e.g. suspending under RunWithout,
// and final teardown as RunWith returns.
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

	if term.ctx == nil {
		term.ctx = Contexts(&term.Attr, &term.Mode)
	}

	if cl, ok := term.ctx.(interface{ Close() error }); ok {
		defer func() {
			if cerr := cl.Close(); err == nil {
				err = cerr
			}
		}()
	}

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

// ExitError may be implemented by an error to customize the exit code under
// MustRun.
type ExitError interface {
	error
	ExitCode() int
}

// MustRun is a useful wrapper for the outermost Term.RunWith: if the error
// value implements ExitError, and its ExitCode method returns non-0, it calls
// os.Exit; otherwise any non-nil error value is log.Fatal-ed.
func MustRun(err error) {
	if err != nil {
		if ex, ok := err.(ExitError); ok {
			log.Printf("exiting due to %v", ex)
			if ec := ex.ExitCode(); ec != 0 {
				os.Exit(ec)
			}
		}
		log.Fatalln(err)
	}
}
