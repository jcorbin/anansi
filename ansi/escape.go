package ansi

import "fmt"

// Escape identifies an ANSI control code, escape sequence, or control sequence
// as a Unicode codepoint:
// - U+0000-U+001F: C0 controls
// - U+0080-U+009F: C1 controls
// - U+EF20-U+EF2F: character set selection functions
// - U+EF30-U+EF3F: private ESCape-sequence functions
// - U+EF40-U+EF5F: non-standard ESCape-sequence functions
// - U+EF60-U+EF7E: standard ESCape-sequence functions
// -        U+EF7F: malformed ESC sequence
// - U+EFC0-U+EFFE: CSI functions
// -        U+EFFF: malformed CSI sequence
type Escape rune

// ESC returns an ESCape sequence identifier named by the given byte.
func ESC(b byte) Escape { return Escape(0xEF00 | 0x7F&rune(b)) }

// CSI returns a CSI control sequence identifier named by the given byte.
func CSI(b byte) Escape { return Escape(0xEF80 | rune(b)) }

// ESC returns the byte name of the ESCape sequence identified by this escape
// value, if any; returns 0 false otherwise.
func (id Escape) ESC() (byte, bool) {
	if 0xEF00 < id && id < 0xEF7F {
		return byte(id & 0x7F), true
	}
	return 0, false
}

// CSI returns the byte name of the CSI control sequence identified by this
// escape value, if any; returns 0 and false otherwise.
func (id Escape) CSI() (byte, bool) {
	if 0xEF80 < id && id < 0xEFFF {
		return byte(id & 0x7F), true
	}
	return 0, false
}

// C1Names provides representation names for the C1 extended-ASCII control
// block.
var C1Names = []string{
	"<RES@>",
	"<RESA>",
	"<RESB>",
	"<RESC>",
	"<IND>",
	"<NEL>",
	"<SSA>",
	"<ESA>",
	"<HTS>",
	"<HTJ>",
	"<VTS>",
	"<PLD>",
	"<PLU>",
	"<RI>",
	"<SS2>",
	"<SS3>",
	"<DCS>",
	"<PU1>",
	"<PU2>",
	"<STS>",
	"<CCH>",
	"<MW>",
	"<SPA>",
	"<EPA>",
	"<RESX>",
	"<RESY>",
	"<RESZ>",
	"<CSI>",
	"<ST>",
	"<OSC>",
	"<PM>",
	"<APC>",
}

// String returns a string representation of the identified control, escape
// sequence, or control sequence: C0 controls are represented phonetically, C1
// controls are represented mnemonically, escape sequences are "ESC+b", control
// sequences are "CSI+b", and the two malformed sentinel codepoints are
// "ESC+INVALID" and "CSI+INVALID" respectively. All other codepoints (albeit
// invalid Escape values) are represented using normal "U+XXXX" notation.
func (id Escape) String() string {
	switch {
	case id <= 0x1F:
		return "^" + string(byte(0x40^id))
	case id == 0x7F:
		return "^?"
	case 0x80 <= id && id <= 0x9F:
		return C1Names[id&^0x80]
	case 0xEF20 <= id && id <= 0xEF7E:
		return fmt.Sprintf("ESC+%s", string(byte(id)))
	case 0xEFC0 <= id && id <= 0xEFFE:
		return fmt.Sprintf("CSI+%s", string(byte(0x7f&id)))
	case id == 0xEF7F:
		return "ESC+INVALID"
	case id == 0xEFFF:
		return "CSI+INVALID"
	default:
		return fmt.Sprintf("%U", rune(id))
	}
}
