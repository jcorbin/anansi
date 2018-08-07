package terminal

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
)

// Terminal supports interacting with a terminal:
// - in-band event reading
// - out-of-band event signaling
// - tracks cursor state combined with
// - an output buffer to at least coalesce writes (no front/back buffer
//   flipping is required or implied; the buffer serves as more of a command
//   queue)
type Terminal struct {
	Attr
	Processor
	Output

	active bool
	closed bool
	ctx    Context
}

// Open a terminal on the given input/output file pair (defaults to os.Stdin
// and os.Stdout) with the given option(s).
//
// If the user wants to process input, they should call term.Notify() shortly
// after Open() to start event processing.
func Open(in, out *os.File, opt Option) (*Terminal, error) {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	opt = Options(opt, DefaultTerminfo)

	term := &Terminal{}
	term.Decoder.File = in
	term.Output.File = out
	term.ctx = &term.Attr

	term.Processor.Init()
	term.Output.Init()

	if err := opt.init(term); err != nil {
		return nil, err
	}

	if err := term.ctx.Enter(term); err != nil {
		_ = term.Close()
		return nil, err
	}

	return term, nil
}

// Close resets the terminal, flushing any buffered output.
func (term *Terminal) Close() error {
	if term.closed {
		return errors.New("terminal already closed")
	}
	err := term.Processor.Close()
	if term.active {
		if cerr := term.ctx.Exit(term); err == nil {
			err = cerr
		}

		// TODO do this only if the cursor isn't homed on a new row (requires
		// cursor to have been parsing and following output all along...)?
		_, _ = term.WriteString("\r\n")

		if ferr := term.Flush(); err == nil {
			err = ferr
		}
	}
	term.closed = true
	return err
}

func (term *Terminal) closeOnPanic() {
	if e := recover(); e != nil {
		if !term.closed {
			_ = term.Close()
		}
		panic(e)
	}
}

// Enter the terminal's context if inactive.
func (term *Terminal) Enter() error {
	if !term.active {
		return term.ctx.Enter(term)
	}
	return nil
}

// Exit the terminal's context if active.
func (term *Terminal) Exit() error {
	if term.active {
		return term.ctx.Exit(term)
	}
	return nil
}

func (term *Terminal) without(f func() error) error {
	if !term.active {
		return f()
	}
	err := term.ctx.Exit(term)
	if err == nil {
		err = f()
	}
	if err == nil {
		err = term.ctx.Enter(term)
	}
	return err
}

// Suspend the terminal program: restore terminal state, send SIGTSTP, wait for
// SIGCONT, then re-setup terminal state once we're back. It returns any error
// encountered or the received SIGCONT signal for completeness on success.
func (term *Terminal) Suspend() (sig os.Signal, err error) {
	err = term.without(func() error {
		contCh := make(chan os.Signal, 1)
		signal.Notify(contCh, syscall.SIGCONT)
		err := syscall.Kill(0, syscall.SIGTSTP)
		if err == nil {
			sig = <-contCh
		}
		return err
	})
	return
}
