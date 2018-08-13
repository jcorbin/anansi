package ansi

import (
	"fmt"
	"unicode/utf8"
)

// DecodeEscape unpacks a UTF-8 encoded ANSI escape sequence at the beginning
// of p returning its Escape identifier, argument, and width in bytes. If the
// returned escape identifier is non-zero, then it represents a complete escape
// sequence (perhaps malformed!).
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
func DecodeEscape(p []byte) (e Escape, a []byte, n int) {
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
