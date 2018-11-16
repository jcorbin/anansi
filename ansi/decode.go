package ansi

import (
	"errors"
	"fmt"
	"image"
	"unicode/utf8"
)

// DecodeEscape unpacks a UTF-8 encoded ANSI escape sequence at the beginning
// of p returning its Escape identifier, argument, and the number of bytes
// consumed by decoding it. If the returned escape identifier is non-zero, then
// it represents a complete escape sequence (perhaps malformed!).
//
// The bytes p[:n] have been processed after DecodeEscape() returns: they're
// either represented by the returned Escape value or by subsequent bytes
// remaining in p[n:]. The caller MUST NOT pass the bytes in p[:n] to
// DecodeEscape() again, and SHOULD NOT look at them itself.
//
// If the returned escape identifier is 0, the caller MAY proceed to decode a
// UTF-8 rune from p[n:]; if this rune turns out to be ESCape (U+001B), the
// caller MAY decide either to process it immediately, or whether to wait for
// additional input bytes which may complete an ESCape sequence.
func DecodeEscape(p []byte) (e Escape, arg []byte, n int) {
	r, m := decodeRune(p)
	if r == 0x1B {
		return decodeESC(p)
	}
	switch r {
	case 0x9B: // CSI
		if cse, csa, csn := decodeCSI(p[m:]); cse != 0 {
			return cse, csa, m + csn
		}
	case 0x90: // DCS
		// TODO stricter DCS state machine per vt100.net
		if sa, sn := decodeString(p[m:]); sn > 0 {
			return Escape(r), sa, m + sn
		}
	case 0x9D: // OSC
		if sa, sn := decodeString(p[m:]); sn > 0 {
			return Escape(r), sa, m + sn
		}
		// TODO linux compat handling for OSC
	case 0x9E, 0x9F: // PM, APC
		if sa, sn := decodeString(p[m:]); sn > 0 {
			return Escape(r), sa, m + sn
		}
	}
	if p[0] == 0x1B {
		// Encode translated C1 control character so that caller can act on it.
		utf8.EncodeRune(p[:m], r)
	}
	return 0, nil, 0
}

func decodeRune(p []byte) (rune, int) {
	r, m := utf8.DecodeRune(p)
	if r == 0x1B {
		if rr, mm := utf8.DecodeRune(p[m:]); 0x40 <= rr && rr <= 0x5F {
			// Uppercase: Translate it into a C1 control character.
			return 0x80 | rr&0x1f, m + mm
		}
	}
	return r, m
}

func decodeESC(p []byte) (e Escape, a []byte, n int) {
	// NOTE caller ensures p[0] == 0x1B
	n++ // count the escape byte as consumed
	ei, ai, ni := 0, 0, 1

	// shuffle bogus bytes out of an escape sequence so that the user can
	// process them (i.e. a control character or high rune)
	rshift := func(m int) {
		var tmp [4]byte
		copy(tmp[:], p[ni:ni+m])
		copy(p[ei+m:ni+m], p[ei:])
		copy(p[ei:], tmp[:m])
	}

	for {
		// ei : index of escape byte
		// ai : index of first arg byte; if 0, no arg
		// ni : index of next byte to process

		// TODO consider recognizing raw C1 control bytes, rather than letting
		// them result in a RuneError through utf8.DecodeRune

		// decode and process the next rune
		r, m := utf8.DecodeRune(p[ni:])
		switch {
		case r == utf8.RuneError: // may be fixed by more bytes, caller can choose
			return 0, nil, ei
		case r > 0xFF: // higher codepoint not part of an escape sequence
			rshift(m)
			return 0, nil, ei
		case 0x80 <= r && r <= 0xFF: // C1 and G1: Treat the same as their 7-bit counterparts
			r &= 0x7F
		}
		if m > 2 {
			panic(fmt.Sprintf("inconceivable: can't decode an ascii rune from %v bytes", m))
		}
		n += m

		// dispatch the rune, now known <= utf8.RuneSelf
		switch {
		case 0x30 <= r && r <= 0x7E: // End of an escape sequence.
			if ai != 0 {
				if ni-ai == 1 && 0x20 <= p[ai] && p[ai] <= 0x2F {
					// name the character selection block after its
					// intermediate byte, rather than its parameter
					return ESC(p[ai]), p[ni : ni+m], n
				}
				a = p[ai:ni]
			}
			return ESC(byte(r)), a, n

		case 0x20 <= r && r <= 0x2F: // Intermediate: Expect zero or more intermediates...
			if ai == 0 {
				ai = ni
			}
			ni += m
			continue

		case r == 0x7F: // Delete: Ignore it, and continue interpreting the ESCape sequence
			r = 0
			rshift(m)
			ei += m
			ni += m
			if ai != 0 {
				ai += m
			}
			continue

		case r == 0x18: // CANcel the current ESCape sequence
			// TODO could provide visibility
			return 0, nil, ni + m

		case r <= 0x1f: // C0 control: Interpret it first, then resume processing ESCape sequence.
			rshift(m)
			return 0, nil, ei // return normalized control character to user

		default:
			panic("inconceivable: exhaustive 7-bit switch wasn't")
		}
	}
}

func decodeCSI(p []byte) (e Escape, a []byte, n int) {
	// 1. It starts with `CSI`, the Control Sequence Introducer.

	// TODO this could be stricter per the vt100.net state diagram
	// TODO compat CSI M CbCxCy

	ni, ai := 0, -1

	for ; ni < len(p); ni++ {
		switch c := p[ni]; {
		case 0x30 <= c && c <= 0x3F:
			// 2. It contains any number of parameter characters: `0123456789:;<=>?`.
			if ai == -1 {
				ai = ni
			}
		case 0x40 <= c && c <= 0x7E:
			// 3. It terminates with an alphabetic character.
			goto term

		case 0x20 <= c && c <= 0x2F:
			// 4. Intermediate characters (if any) immediately precede the terminator.
			if ai == -1 {
				ai = ni
			}
			ni++
			goto intermed
		}
	}
	return 0, nil, 0

intermed:
	for ; ni < len(p); ni++ {
		switch c := p[ni]; {
		case 0x20 <= c && c <= 0x2F:
			// 4. Intermediate characters (if any) immediately precede the terminator.
		case 0x40 <= c && c <= 0x7E:
			// 3. It terminates with an alphabetic character.
			goto term
		}
	}
	return 0, nil, 0

term:
	if ai >= 0 {
		a = p[ai:ni]
	}
	return CSI(p[ni]), a, ni + 1
}

func decodeString(p []byte) (a []byte, n int) {
	r, m := decodeRune(p)
	for {
		switch r {
		case utf8.RuneError:
			return nil, 0
		case 0x9C:
			return p[:n], n + m
		}
		n += m
		r, m = decodeRune(p[n:])
	}
}

var (
	errRange       = errors.New("value out of range")
	errSyntax      = errors.New("invalid syntax")
	errSGRInvalid  = errors.New("invalid sgr code")
	errModeInvalid = errors.New("invalid ansi mode")
)

// DecodeNumber decodes a signed base-10 encoded number from the beginning of
// the given byte buffer. If the first byte is ';', it is skipped. Returns the
// decode number and the number of bytes decoded, or a non-nil decode error
// (either errRange or errSyntax).
func DecodeNumber(p []byte) (r, n int, _ error) {
	// a specialized copy of strconv.ParseInt.

	// Empty string bad.
	if len(p) == 0 {
		return 0, 0, errSyntax
	}

	// Pick off leading sign.
	neg := false
	if p[0] == ';' {
		p = p[1:]
		n++
		if len(p) == 0 {
			return 0, 0, errSyntax
		}
	}

	switch p[0] {
	case '-':
		neg = true
		fallthrough
	case '+':
		p = p[1:]
		n++
		if len(p) == 0 {
			return 0, 0, errSyntax
		}
	}

	// Convert unsigned and check range.
	const (
		bitSize   = 32
		maxUint64 = (1<<64 - 1)
		cutoff    = maxUint64/10 + 1
		maxVal    = uint64(1)<<uint(bitSize) - 1
	)
	var un uint64

	for _, c := range p {
		if c < '0' || '9' < c {
			break
		}
		d := c - '0'
		n++
		if un < cutoff {
			un *= 10
			if un1 := un + uint64(d); !(un1 < un || un1 > maxVal) {
				un = un1
				continue
			} // else un+v overflows
		} // else un*10 overflows
		if !neg && maxVal >= cutoff {
			return int(cutoff - 1), n, errRange
		}
		if neg && maxVal > cutoff {
			return -int(cutoff), n, errRange
		}
	}

	if neg {
		return -int(un), n, nil
	}
	return int(un), n, nil
}

// DecodePoint decose a screen point, e.g. from a CUP sequence, into an 1,1
// origin-relative Point value.
func DecodePoint(a []byte) (p Point, n int, err error) {
	p.Y, n, err = DecodeNumber(a)
	if err == nil {
		var m int
		p.X, m, err = DecodeNumber(a[n:])
		n += m
	}
	return p, n, err
}

// DecodeSGR decodes an SGR attribute value from the given byte buffer; if
// non-nil error is returned, then n indicates the index of the offending byte.
func DecodeSGR(a []byte) (attr SGRAttr, n int, _ error) {
	for n < len(a) {
		switch a[n] {
		case ';':
			n++
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8':
			if m := n + 1; m == len(a) || a[m] == ';' {
				switch at := []SGRAttr{
					SGRAttrClear,
					SGRAttrBold,
					SGRAttrDim,
					SGRAttrItalic,
					SGRAttrUnderscore,
					0, // slow blink not supported
					0, // fast blink not supported
					SGRAttrNegative,
					0, // concealed not supported
				}[a[n]-'0']; at {
				case 0:
				case SGRAttrClear:
					attr = SGRAttrClear
				default:
					attr |= at
				}
				if m < len(a) {
					n = m + 1
				} else {
					n++
				}
				continue
			}
		}

		switch a[n] {

		case '3':
			n++
			c, m, err := decodeSGRColor(a[n:])
			n += m
			if err != nil {
				return attr, n, err
			}
			attr = attr.SansFG() | c.FG()

		case '4':
			n++
			c, m, err := decodeSGRColor(a[n:])
			n += m
			if err != nil {
				return attr, n, err
			}
			attr = attr.SansBG() | c.BG()

		case '9':
			n++
			c, m, err := decodeSGRBrightColor(a[n:])
			n += m
			if err != nil {
				return attr, n, err
			}
			attr = attr.SansFG() | c.FG()

		case '1':
			if n++; a[n] != '0' || n == len(a)-1 {
				return attr, n, errSGRInvalid
			}
			n++
			c, m, err := decodeSGRBrightColor(a[n:])
			n += m
			if err != nil {
				return attr, n, err
			}
			attr = attr.SansBG() | c.BG()

		default:
			return attr, n, errSGRInvalid
		}
	}
	return attr, n, nil
}

func decodeSGRColor(a []byte) (c SGRColor, n int, _ error) {
	if len(a) == 0 {
		return c, n, errSGRInvalid
	}

	switch a[0] {
	case '8':
		n++
		c, m, err := decodeSGRExtendedColor(a[n:])
		return c, n + m, err

	default:
		c, m, err := decodeSGRClassicColor(a[n:])
		return c, n + m, err
	}
}

func decodeSGRClassicColor(a []byte) (c SGRColor, n int, _ error) {
	if len(a) > 0 {
		switch b := a[0]; b {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			if n++; n == len(a) || a[n] == ';' {
				return SGRColor(b - '0'), n, nil
			}
		}
	}
	return c, n, errSGRInvalid
}

func decodeSGRBrightColor(a []byte) (c SGRColor, n int, _ error) {
	if len(a) > 0 {
		switch b := a[0]; b {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			if n++; n == len(a) || a[n] == ';' {
				return SGRColor(8 + b - '0'), n, nil
			}
		}
	}
	return c, n, errSGRInvalid
}

func decodeSGRExtendedColor(a []byte) (c SGRColor, n int, _ error) {
	if len(a) > 1 && a[n] == ';' {
		n++
		switch a[n] {
		case '2':
			n++
			c, m, err := decodeSGRRGBColor(a[n:])
			return c, n + m, err
		case '5':
			n++
			c, m, err := decodeSGRColorNumber(a[n:])
			return c, n + m, err
		}

	}
	return c, n, errSGRInvalid
}

func decodeSGRColorNumber(a []byte) (c SGRColor, n int, _ error) {
	if len(a) > 1 && a[n] == ';' {
		n++
		cn, m, err := decodeUint8(a[n:])
		n += m
		if err == nil {
			c = SGRColor(cn)
		}
		return c, n, err
	}
	return c, n, errSGRInvalid
}

func decodeSGRRGBColor(a []byte) (c SGRColor, n int, _ error) {
	if len(a) > 1 && a[n] == ';' {
		n++
		r, m, err := decodeUint8(a[n:])
		n += m
		if err != nil {
			return c, n, err
		}
		g, m, err := decodeUint8(a[n:])
		n += m
		if err != nil {
			return c, n, err
		}
		b, m, err := decodeUint8(a[n:])
		n += m
		if err != nil {
			return c, n, err
		}
		return RGB(r, g, b), n, nil
	}
	return c, n, errSGRInvalid
}

func decodeUint8(a []byte) (r uint8, n int, _ error) {
	if len(a) == 0 {
		return r, n, errSGRInvalid
	}
	var v uint16
	for n < len(a) {
		if a[n] == ';' {
			n++
			break
		} else if '0' <= a[n] && a[n] <= '9' {
			if v = 10*v + uint16(a[n]-'0'); v > 0xFF {
				return 0, n, errSGRInvalid
			}
			n++
		} else {
			return 0, n, errSGRInvalid
		}
	}
	return uint8(v), n, nil
}

// DecodeMode decodes a single mode parameter from escape argument bytes.
func DecodeMode(private bool, a []byte) (mode Mode, n int, _ error) {
	if len(a) == 0 {
		return mode, n, errModeInvalid
	}
	r, m, err := DecodeNumber(a[n:])
	n += m
	if err == nil {
		mode = Mode(r)
		if private {
			mode |= ModePrivate
		}
	}
	return mode, n, err
}

// DecodeCursorCardinal decodes a cardinal cursor move, one of: CUU, CUD, CUF, or CUB.
func DecodeCursorCardinal(id Escape, a []byte) (d image.Point, _ bool) {
	switch id {
	case CUU: // CUrsor Up
		d.Y = -1
	case CUD: // CUrsor Down
		d.Y = 1
	case CUF: // CUrsor Forward
		d.X = 1
	case CUB: // CUrsor Backward
		d.X = -1
	default:
		return image.ZP, false
	}
	if len(a) > 0 {
		if n, _, err := DecodeNumber(a); err == nil {
			d = d.Mul(n)
		}
	}
	return d, true
}
