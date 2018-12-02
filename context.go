package anansi

// Context provides a piece of terminal context setup and teardown logic.
//
// See Term.RunWith and Term.RunWithout for more detail.
type Context interface {
	// Enter is called to (re)establish terminal context at the start of the
	// first Term.RunWith and at the end of every Term.RunWithout.
	Enter(term *Term) error

	// Exit is called to restore original terminal context at the end of the
	// first Term.RunWith and at the start of Term.RunWithout.
	Exit(term *Term) error
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

func (tcs contexts) Enter(term *Term) error {
	for i := 0; i < len(tcs); i++ {
		if err := tcs[i].Enter(term); err != nil {
			return err
		}
	}
	return nil
}

func (tcs contexts) Exit(term *Term) (rerr error) {
	for i := len(tcs) - 1; i >= 0; i-- {
		if err := tcs[i].Exit(term); rerr == nil {
			rerr = err
		}
	}
	return rerr
}
