package terminal

import (
	"fmt"
	"os"

	"github.com/jcorbin/anansi/terminfo"
)

// Options creates a compound option from 0 or more options (returns nil in the
// 0 case).
func Options(opts ...Option) Option {
	if len(opts) == 0 {
		return nil
	}
	a := opts[0]
	opts = opts[1:]
	for len(opts) > 0 {
		b := opts[0]
		opts = opts[1:]
		if a == nil {
			a = b
			continue
		} else if b == nil {
			continue
		}
		as, haveAs := a.(options)
		bs, haveBs := b.(options)
		if haveAs && haveBs {
			a = append(as, bs...)
		} else if haveAs {
			a = append(as, b)
		} else if haveBs {
			a = append(options{a}, bs)
		} else {
			a = options{a, b}
		}
	}
	return a
}

// Option is an opaque option to pass to Open().
type Option interface {
	// init gets called while initializing internal terminal state; should not
	// manipulate external resources, but instead wire up a further lifecycle
	// option.
	init(term *Terminal) error
}

type optionFunc func(*Terminal) error

func (f optionFunc) init(term *Terminal) error { return f(term) }

type options []Option

func (os options) init(term *Terminal) error {
	for i := range os {
		if err := os[i].init(term); err != nil {
			return err
		}
	}
	return nil
}

// DefaultTerminfo loads default terminfo based on the TERM environment
// variable; basically it uses terminfo.Load(os.Getenv("TERM")).
var DefaultTerminfo = optionFunc(func(term *Terminal) error {
	if term.Decoder.Terminfo() != nil {
		return nil
	}
	info, err := terminfo.Load(os.Getenv("TERM"))
	if err == nil {
		term.Decoder.SetTerminfo(info)
	}
	return err
})

// Terminfo provides a terminfo definition selected explicitly, rather than
// relying on the Decoder's default loading mechanism.
func Terminfo(info *terminfo.Terminfo) Option {
	return optionFunc(func(term *Terminal) error {
		term.Decoder.SetTerminfo(info)
		return nil
	})
}

// With provides Context and EventFilter values to be attached to the terminal
// during open. This should only be used for external implementations of
// Context and EventFilter, as all standard implementations provided by the
// terminal package implement Option.
func With(args ...interface{}) Option {
	var eventFilter EventFilter
	var context Context
	for _, arg := range args {
		any := false
		if ef, ok := arg.(EventFilter); ok {
			eventFilter = chainEventFilter(eventFilter, ef)
			any = true
		}
		if ctx, ok := arg.(Context); ok {
			context = chainTermContext(context, ctx)
			any = true
		}
		if !any {
			panic(fmt.Sprintf("unsupported terminal.With arg type %T; must implement Context or EventFilter", arg))
		}
	}
	return optionFunc(func(term *Terminal) error {
		term.EventFilter = chainEventFilter(term.EventFilter, eventFilter)
		term.ctx = chainTermContext(term.ctx, context)
		return nil
	})
}
