package anansi

import (
	"os"
)

// Context provides a piece of terminal context setup and teardown logic.
type Context interface {
	Enter(f *os.File) error
	Exit(f *os.File) error
}

// Contexts returns a Context that: calls all given context Enter()s in order,
// stopping on first failure; and calls all given context Exit()s in reverse
// order, returning the first error, but proceeding to all all remaining
// Exit() even under error.
func Contexts(cs ...Context) Context {
	if len(cs) == 0 {
		return nil
	}
	a := cs[0]
	for i := 1; i < len(cs); i++ {
		b := cs[i]
		if b == nil || b == Context(nil) {
			continue
		}
		if a == nil || a == Context(nil) {
			a = b
			continue
		}
		as, haveAs := a.(contexts)
		bs, haveBs := b.(contexts)
		if haveAs && haveBs {
			a = append(as, bs...)
		} else if haveAs {
			a = append(as, b)
		} else if haveBs {
			a = append(contexts{a}, bs...)
		} else {
			a = contexts{a, b}
		}
	}
	return a
}

type contexts []Context

func (tcs contexts) Enter(f *os.File) error {
	for i := 0; i < len(tcs); i++ {
		if err := tcs[i].Enter(f); err != nil {
			return err
		}
	}
	return nil
}

func (tcs contexts) Exit(f *os.File) (rerr error) {
	for i := len(tcs) - 1; i >= 0; i-- {
		if err := tcs[i].Exit(f); rerr == nil {
			rerr = err
		}
	}
	return rerr
}
