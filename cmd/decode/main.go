package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	rawMode   = flag.Bool("raw", false, "enable terminal raw mode")
	mouseMode = flag.Bool("mouse", false, "enable terminal mouse reporting")
	altMode   = flag.Bool("alt", false, "enable alternate screen usage")
)

func main() {
	flag.Parse()

	term := anansi.NewTerm(os.Stdout)

	if *mouseMode {
		term.AddMode(
			ansi.ModeMouseSgrExt,
			ansi.ModeMouseBtnEvent,
			ansi.ModeMouseAnyEvent,
		)
	}

	if *altMode {
		term.AddMode(
			ansi.ModeAlternateScreen,
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
		// read more input…
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

		// …and process it
		if err := process(term, &buf); err != nil {
			return err
		}
	}
}

func process(term *anansi.Term, buf *bytes.Buffer) error {
	for buf.Len() > 0 {
		// Try to decode an escape sequence…
		e, a, n := ansi.DecodeEscape(buf.Bytes())
		if n > 0 {
			buf.Next(n)
		}

		// …fallback to decoding a rune otherwise…
		if e == 0 {
			r, n := utf8.DecodeRune(buf.Bytes())
			switch r {
			case 0x90, 0x9D, 0x9E, 0x9F: // DCS, OSC, PM, APC
				return nil // …need more bytes to complete a partial string.

			case 0x9B: // CSI
				return nil // …need more bytes to complete a partial control sequence.

			case 0x1B: // ESC
				if p := buf.Bytes(); len(p) == cap(p) {
					return nil // …need more bytes to determine if an escape sequence can be decoded.
				}
				// …pass as literal ESC…
			}

			// …consume and handle the rune.
			buf.Next(n)
			e = ansi.Escape(r)
		}

		handle(term, e, a)
	}
	return nil
}

var prior ansi.Escape

func handle(term *anansi.Term, e ansi.Escape, a []byte) {
	fmt.Printf("%U %v", e, e)

	if len(a) > 0 {
		fmt.Printf(" %q", a)
	}

	// print detail for mouse reporting
	if e == ansi.CSI('M') || e == ansi.CSI('m') {
		btn, pt, err := ansi.DecodeXtermExtendedMouse(e, a)
		if err != nil {
			fmt.Printf(" mouse-err:%v", err)
		} else {
			fmt.Printf(" mouse-%v@%v", btn, pt)
		}
	}

	switch e {

	// ^C to quit
	case 0x03:
		if prior == 0x03 {
			panic("goodbye")
		} else {
			fmt.Printf(" \x1b[91m<press Ctrl-C again to quit>\x1b[0m")
		}

	// ^L to clear
	case 0x0c:
		if prior == 0x0c {
			fmt.Printf("\x1b[2J\x1b[H") // 2 ED CUP
		} else {
			fmt.Printf(" \x1b[93m<press Ctrl-L again to quit>\x1b[0m")
		}

	// ^Z to suspend
	case 0x1a:
		if prior == 0x1a {
			if err := term.RunWithout(suspend); err != nil {
				panic(err)
			}
		} else {
			fmt.Printf(" \x1b[92m<press Ctrl-Z again to suspend>\x1b[0m")
		}

	}

	prior = e

	fmt.Printf("\r\n")
}

func suspend(_ *anansi.Term) error {
	cont := make(chan os.Signal)
	signal.Notify(cont, syscall.SIGCONT)
	log.Printf("suspending")
	if err := syscall.Kill(0, syscall.SIGTSTP); err != nil {
		return err
	}
	<-cont
	log.Printf("resumed")
	return nil
}
