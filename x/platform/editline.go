package platform

import (
	"unicode"
	"unicode/utf8"

	"github.com/jcorbin/anansi/ansi"
)

// EditLineState code.
type EditLineState uint8

// EditLineState constants.
const (
	EditLineReady EditLineState = iota
	EditLineDone
	EditLineCanceled
)

// EditLine provides basic line-editing .
type EditLine struct {
	State EditLineState
	Box   ansi.Rectangle
	Buf   []byte
	Cur   int
	View  int

	handleEvent editLineHandler
}

// Reset state.
func (edl *EditLine) Reset() {
	edl.State = EditLineReady
	edl.Box.Min = ansi.Pt(1, 1)
	edl.Box.Max = edl.Box.Min
	edl.Buf = edl.Buf[:0]
	edl.View = 0
	edl.Cur = 0
}

// Active returns true if the edit line is accepting user input.
func (edl *EditLine) Active() bool { return edl.State == EditLineReady }

// Done returns true if the user submitted input (e.g. hit <Enter>).
func (edl *EditLine) Done() bool { return edl.State == EditLineDone }

// Canceled returns true if the user canceled editing (e.g. hit <Esc>).
func (edl *EditLine) Canceled() bool { return edl.State == EditLineCanceled }

// Update processes user input and draws the edit line.
func (edl *EditLine) Update(ctx *Context) {
	if edl.State != EditLineReady {
		return
	}

	if edl.handleEvent == nil {
		edl.handleEvent = defaultBehavior
	}
	for eid := range ctx.Input.Type {
		edl.handleEvent(edl, ctx, eid)
	}

	if edl.State != EditLineReady {
		return
	}

	// TODO clamp to box, showing cursor
	n := utf8.RuneCount(edl.Buf)

	// ensure cursor past left rune offset
	if edl.Cur < edl.View {
		if edl.Cur > 0 {
			edl.View = edl.Cur - 1
		} else {
			edl.View = 0
		}
	}

	// ensure cursor within right rune offset
	hi := edl.View + n
	if w := edl.Box.Dx(); n > w {
		hi = edl.View + w
	}
	if edl.Cur >= hi {
		d := edl.Cur - hi
		if hi < n {
			d++
		}
		hi += d
		edl.View += d
	} else if reclaim := hi - n; reclaim > 0 {
		edl.View -= reclaim
		hi -= reclaim
	}
	lo := edl.View

	// TODO ellipses

	b := edl.Buf
	for i := 0; i < lo; i++ {
		_, n := utf8.DecodeRune(b)
		b = b[n:]
	}
	for i := hi; i < n; i++ {
		_, n := utf8.DecodeLastRune(b)
		b = b[:len(b)-n]
	}
	if len(b) > 0 {
		ctx.Output.To(edl.Box.Min)
		ctx.Output.Write(b)
	}

	ctx.Output.UserCursor.Visible = true
	ctx.Output.UserCursor.Point = ansi.Pt(edl.Box.Min.X+edl.Cur-edl.View, edl.Box.Min.Y)
	if ctx.Output.X == ctx.Output.UserCursor.X {
		ctx.Output.WriteRune(' ')
	}

	return
}

// TODO elaborate the monadic structure around this
type editLineHandler func(edl *EditLine, ctx *Context, eid int)

// TODO other behaviors to support advanced editing ala Emacs/Vi-syle

var defaultBehavior = editLineHandlers(
	// TODO accept bracketed paste
	(*EditLine).handleGraphicRune,
	(*EditLine).handleControlRune,
	(*EditLine).handleArrowKeys,
)

func editLineHandlers(hs ...editLineHandler) editLineHandler {
	return func(edl *EditLine, ctx *Context, eid int) {
		for i := 0; i < len(hs) && ctx.Input.Type[eid] != EventNone; i++ {
			hs[i](edl, ctx, eid)
		}
		return
	}
}

func (edl *EditLine) handleGraphicRune(ctx *Context, eid int) {
	if ctx.Input.Type[eid] != EventRune || !unicode.IsGraphic(ctx.Input.Rune(eid)) {
		return
	}
	var tmp [4]byte
	b := tmp[:utf8.EncodeRune(tmp[:], ctx.Input.Rune(eid))]
	n := utf8.RuneCount(edl.Buf)
	edl.Buf = append(edl.Buf, b...)
	if edl.Cur < n {
		off := 0
		in := edl.Buf
		in0 := in
		for i := 0; i < edl.Cur; i++ {
			off += n
			_, n := utf8.DecodeRune(in)
			in = in[n:]
		}
		copy(in[len(b):], in)
		copy(in, b)
		edl.Buf = in0
	}
	edl.Cur++
	ctx.Input.Type[eid] = EventNone
	return
}

func (edl *EditLine) handleControlRune(ctx *Context, eid int) {
	if ctx.Input.Type[eid] != EventRune {
		return
	}
	switch ctx.Input.Rune(eid) {
	case '\x0D': // <Enter> submits
		edl.State = EditLineDone
	case '\x1B': // ESC cancels
		edl.State = EditLineCanceled
	case '\x7F': // <Delete> backwards
		if edl.Cur > 0 {
			if edl.Cur < utf8.RuneCount(edl.Buf) {
				in := edl.Buf
				for j := 0; j < edl.Cur; j++ {
					_, n := utf8.DecodeRune(in)
					in = in[n:]
				}
				_, n := utf8.DecodeRune(in)
				copy(in, in[n:])
				edl.Buf = edl.Buf[:len(edl.Buf)-n]
			} else {
				_, n := utf8.DecodeLastRune(edl.Buf)
				edl.Buf = edl.Buf[:len(edl.Buf)-n]
			}
			edl.Cur--
		}
	default:
		return
	}
	ctx.Input.Type[eid] = EventNone
	return
}

func (edl *EditLine) handleArrowKeys(ctx *Context, eid int) {
	if ctx.Input.Type[eid] != EventEscape {
		return
	}
	e := ctx.Input.Escape(eid)
	if move, is := ansi.DecodeCursorCardinal(e.ID, e.Arg); is {
		edl.Cur += move.X
		if n := utf8.RuneCount(edl.Buf); edl.Cur > n {
			edl.Cur = n
		} else if edl.Cur < 0 {
			edl.Cur = 0
		}
		ctx.Input.Type[eid] = EventNone
	}
	return
}
