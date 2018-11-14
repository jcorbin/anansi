package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"unicode/utf8"

	"github.com/jcorbin/anansi/ansi"
)

var readBuf bytes.Buffer
var outBuf *bufio.Writer

func main() {
	outBuf = bufio.NewWriter(os.Stdout)
	if err := readMore(); err != nil {
		log.Fatalln(err)
	}
}

func readMore() (err error) {
	defer func() {
		if ferr := outBuf.Flush(); err == nil {
			err = ferr
		}
	}()
	for err == nil {
		readBuf.Grow(4096)
		b := readBuf.Bytes()
		b = b[len(b):cap(b)]
		var n int
		n, err = os.Stdin.Read(b)
		ateof := err == io.EOF
		if err == nil || ateof {
			readBuf.Write(b[:n])
			if perr := processMore(ateof); err == nil {
				err = perr
			}
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func processMore(ateof bool) (err error) {
	for err == nil && readBuf.Len() > 0 {
		e, a, n := ansi.DecodeEscape(readBuf.Bytes())
		readBuf.Next(n)
		if e == 0 {
			r, n := utf8.DecodeRune(readBuf.Bytes())
			const esc = 0x1b
			if r == esc && readBuf.Len() == 1 && !ateof {
				break // readMore, try to complete escape sequence
			}
			readBuf.Next(n)
			e = ansi.Escape(r)
		}
		err = process(e, a)
	}
	return err
}

func process(e ansi.Escape, a []byte) error {
	if !e.IsEscape() {
		_, err := outBuf.WriteRune(rune(e))
		return err
	}

	var b [4096]byte
	p := b[:0]

	// try to convert legacy SGR colors to 24-bit; TODO palette selection
	if e == ansi.CSI('m') {
		if attr, _, err := ansi.DecodeSGR(a); err == nil {
			needed := false

			// maybe convert foreground
			if fg, hasFG := attr.FG(); hasFG {
				if newc := fg.To24Bit(); newc != fg {
					attr = attr.SansFG() | newc.FG()
					needed = true
				}
			}

			// maybe convert background
			if bg, hasBG := attr.BG(); hasBG {
				if newc := bg.To24Bit(); newc != bg.To24Bit() {
					attr = attr.SansBG() | newc.BG()
					needed = true
				}
			}

			if needed {
				p = attr.AppendTo(p)
			}
		}
	}

	// not SGR, or conversion failed/not necessary
	if len(p) == 0 {
		p = e.AppendWith(p, a...)
	}

	_, err := outBuf.Write(p)
	return err
}
