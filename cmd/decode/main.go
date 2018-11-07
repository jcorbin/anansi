package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	rawMode   = flag.Bool("raw", false, "enable terminal raw mode")
	mouseMode = flag.Bool("mouse", false, "enable terminal mouse reporting")
)

func main() {
	flag.Parse()

	var mode anansi.Mode
	term := anansi.NewTerm(os.Stdout, &mode)

	if *mouseMode {
		mode.AddMode(
			ansi.ModeMouseSgrExt,
			ansi.ModeMouseBtnEvent,
			ansi.ModeMouseAnyEvent,
		)
	}

	term.SetEcho(!*rawMode)
	term.SetRaw(*rawMode)

	switch err := term.RunWith(run); err {
	case nil:
	case io.EOF:
		fmt.Println(err)
	default:
		log.Fatal(err)
	}
}

func run(term *anansi.Term) error {
	const minRead = 128
	var buf bytes.Buffer
	for {
		// read more input...
		buf.Grow(minRead)
		p := buf.Bytes()
		p = p[len(p):cap(p)]
		n, err := os.Stdin.Read(p)
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		_, _ = buf.Write(p[:n])

		// ...and process it
		if err := process(&buf); err != nil {
			return err
		}
	}
}

func process(buf *bytes.Buffer) error {
	for buf.Len() > 0 {
		e, a, n := ansi.DecodeEscape(buf.Bytes())
		if n > 0 {
			buf.Next(n)
		}

		// have a complete escape sequence
		if e != 0 {
			handleEscape(e, a)
			continue
		}

		// Decode a rune...
		switch r, n := utf8.DecodeRune(buf.Bytes()); r {

		case 0x90, 0x9D, 0x9E, 0x9F: // DCS, OSC, PM, APC
			return nil // ... need more bytes to complete a partial string.

		case 0x9B: // CSI
			return nil // ... need more bytes to complete a partial control sequence.

		case 0x1B: // ESC
			if p := buf.Bytes(); len(p) == cap(p) {
				return nil // ... need more bytes to determine if an escape sequence can be decoded.
			}
			fallthrough // ... literal ESC

		default:
			buf.Next(n)
			handleRune(r)
		}
	}
	return nil
}

func handleEscape(e ansi.Escape, a []byte) {
	switch e {
	case ansi.CSI('M'), ansi.CSI('m'):
		btn, pt, err := ansi.DecodeXtermExtendedMouse(e, a)
		if err != nil {
			fmt.Printf("invalid mouse: %v %q err:%v", e, a, err)
		} else {
			fmt.Printf("mouse(%v@%v)", btn, pt)
		}

	default:
		fmt.Print(e)
		if len(a) > 0 {
			fmt.Printf(" %q", a)
		}
	}
	fmt.Printf("\r\n")
}

func handleRune(r rune) {
	switch {
	// panic on ^C
	case r == 0x03:
		panic("goodbye")

	// print C0 controls phonetically
	case r < 0x20, r == 0x7f:
		fmt.Printf("^%s", string(0x40^r))

	// print C1 controls mnemonically
	case 0x80 <= r && r <= 0x9f:
		fmt.Print(ansi.C1Names[r&0x1f])

	// print normal rune
	default:
		fmt.Print(string(r))
	}
}
