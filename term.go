package anansi

import (
	"os"
)

// NewTerm creates a new Term attached to the given file, and with optional
// associated context.
func NewTerm(f *os.File, cs ...Context) *Term {
	term := &Term{File: f}
	term.ctx = Contexts(&term.Attr, Contexts(cs...))
	return term
}

// Term combines a terminal file handle with attribute control and further
// Context-ual state.
type Term struct {
	Attr
	*os.File

	active bool
	ctx    Context
}

// With runs the given function within the terminal's context, activating it if
// necessary, and deactivating it if activation was necessary. If With Enter()
// context, then it calls context Exit() even after error or panic.
func (term *Term) With(within func(*Term) error) (err error) {
	if term.active {
		return within(term)
	}
	defer func() {
		if cerr := term.ctx.Exit(term); cerr == nil {
			term.active = false
		} else if err == nil {
			err = cerr
		}
	}()
	if err = term.ctx.Enter(term); err == nil {
		term.active = true
		err = within(term)
	}
	return err
}

// Without runs the given function outside the terminal's context, deactivating
// it if necessary, and reactivating it if deactivation was necessary.
// Reactivation is not done if an error or panic occurs.
func (term *Term) Without(outside func(*Term) error) (err error) {
	if !term.active {
		return outside(term)
	}
	if err = term.ctx.Exit(term); err == nil {
		term.active = false
		if err = outside(term); err == nil {
			if err = term.ctx.Enter(term); err == nil {
				term.active = true
			}
		}
	}
	return err
}
