package ansi_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"unicode/utf8"

	"github.com/jcorbin/anansi/ansi"
)

func Example_decode_main() {
	switch err := run(); err {
	case nil:
	case io.EOF:
		fmt.Println(err)
	default:
		log.Fatal(err)
	}
}

func run() error {
	const minRead = 128
	var buf bytes.Buffer
	for {
		// read more input...
		buf.Grow(minRead)
		p := buf.Bytes()
		p = p[len(p):cap(p)]
		n, err := os.Stdin.Read(p)
		if err != nil {
			panic(err)
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

		// try to decode a rune, maybe read more bytes to complete a
		// partial escape sequence
		switch r, n := utf8.DecodeRune(buf.Bytes()); r {
		case 0x90, 0x9B, 0x9D, 0x9E, 0x9F: // DCS, CSI, OSC, PM, APC
			return nil
		case 0x1B: // ESC
			if p := buf.Bytes(); len(p) == cap(p) {
				return nil
			}
			fallthrough
		default:
			buf.Next(n)
			handleRune(r)
		}
	}
	return nil
}

func handleEscape(e ansi.Escape, a []byte) {
	fmt.Print(e)
	if len(a) > 0 {
		fmt.Printf(" %q", a)
	}
	fmt.Printf("\r\n")
}

func handleRune(r rune) {
	switch {
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
