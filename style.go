package anansi

import "github.com/jcorbin/anansi/ansi"

// Style allows styling of cell data during a drawing or rendering routine.
// Its eponymous method gets called for each cell as it is about to be
// rendered
//
// The Style method receives the point being rendered on screen or within a
// destination Grid. Its pr and pa arguments contain any prior rune and
// attribute data when drawing in a Grid, while the r and a arguments contain
// the source rune and attribute being drawn.
//
// Whatever rune and attribute values are returned by Style() will be the ones
// drawn/rendered.
//
// Style implementations are also called with all-zero argument values (i.e.
// an invalid ansi point) to probe any default rune/attr values. These default
// values are then used as fill values when rendering or drawing
type Style interface {
	Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr)
}

// Styles combines zero or more styles into a non-nil Style; if given none, it
// returns a no-op Style; if given many, it returns a Style that calls each in
// turn.
func Styles(ss ...Style) Style {
	var res styles
	for _, s := range ss {
		switch impl := s.(type) {
		case _noopStyle:
			continue
		case styles:
			res = append(res, impl...)
		default:
			res = append(res, s)
		}
	}
	switch len(res) {
	case 0:
		return NoopStyle
	case 1:
		return res[0]
	default:
		return res
	}
}

// StyleFunc is a convenience type alias for implementing Style.
type StyleFunc func(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr)

// Style calls the aliased function pointer
func (f StyleFunc) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	return f(p, pr, r, pa, a)
}

type _noopStyle struct{}

func (ns _noopStyle) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	return r, 0
}

// NoopStyle is a no-op style, used as a zero fill by Styles.
var NoopStyle Style = _noopStyle{}

type styles []Style

func (ss styles) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	for _, s := range ss {
		r, a = s.Style(p, pr, r, pa, a)
	}
	return r, a
}

// ZeroRuneStyle maps a fixed rune to 0. Useful for implementing transparency
// of rune values, e.g. space space characters or empty braille cells.
type ZeroRuneStyle rune

// Style replaces the passed rune with 0 if it equals the receiver.
func (es ZeroRuneStyle) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if r == rune(es) {
		r = 0
	}
	return r, a
}

// FillRuneStyle maps 0 runes to a fixed rune value. Useful for normalizing
// transparent rune values when rendering a base grid.
type FillRuneStyle rune

// Style replaces the passed rune with the receiver if the passed rune is 0.
func (fs FillRuneStyle) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if r == 0 {
		r = rune(fs)
	}
	return r, a
}

// DefaultRuneStyle maps 0 runes to a fixed rune value, only when their
// coresponding SGRAttr value is non-zero.
type DefaultRuneStyle rune

// Style replaces the passed rune with the receiver if the passed rune is 0
// and the passed attr is not.
func (fs DefaultRuneStyle) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if r == 0 && a != 0 {
		r = rune(fs)
	}
	return r, a
}

// AttrStyle implements a Style that returns a fixed ansi attr for any non-zero runes.
type AttrStyle ansi.SGRAttr

// Style replaces the passed attr with the receiver if the passed rune is non-0.
func (as AttrStyle) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if r != 0 {
		a = ansi.SGRAttr(as)
	}
	return r, a
}

var (
	// TransparentRunes is a style that stops 0 rune values from overwriting
	// any prior rune values when drawing.
	TransparentRunes Style = transparentRuneStyle{}

	// TransparentAttrBG is a style that stops 0 background attribute from
	// overwriting any prior attribute background value when drawing.
	TransparentAttrBG Style = transparentAttrStyle(ansi.SGRAttrBGMask)

	// TransparentAttrFG is a style that stops 0 foreground attribute from
	// overwriting any prior attribute foreground value when drawing. This
	// includes text attributes like bold and italics, not just foreground
	// color.
	TransparentAttrFG Style = transparentAttrStyle(ansi.SGRAttrFGMask | ansi.SGRAttrMask)

	// TransparentAttrBGFG is a style that acts as both TransparentBG and
	// TransparentAttrFG combined.
	TransparentAttrBGFG Style = transparentAttrStyle(ansi.SGRAttrBGMask | ansi.SGRAttrFGMask | ansi.SGRAttrMask)
)

type transparentRuneStyle struct{}
type transparentAttrStyle ansi.SGRAttr

func (tr transparentRuneStyle) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if r == 0 {
		r = pr
	}
	return r, a
}

func (ta transparentAttrStyle) Style(p ansi.Point, pr, r rune, pa, a ansi.SGRAttr) (rune, ansi.SGRAttr) {
	if a&ansi.SGRAttr(ta) == 0 {
		a |= pa & ansi.SGRAttr(ta)
	}
	return r, a
}
