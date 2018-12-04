package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var (
	rawMode   = flag.Bool("raw", false, "enable terminal raw mode")
	mouseMode = flag.Bool("mouse", false, "enable terminal mouse reporting")
	altMode   = flag.Bool("alt", false, "enable alternate screen usage")
	stripMode = flag.Bool("strip", false, "strip escape sequences when running non-interactively")
)

func main() {
	flag.Parse()
	anansi.MustRun(run(anansi.NewTerm(os.Stdin, os.Stdout)))
}

func run(term *anansi.Term) error {
	if !term.IsTerminal() {
		return runBatch(term.Input.File, term.Output.File)
	}

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

	if err := term.SetEcho(!*rawMode); err != nil {
		return err
	}
	if err := term.SetRaw(*rawMode); err != nil {
		return err
	}

	return term.RunWith(runInteractive)
}

func runBatch(in, out *os.File) (err error) {
	var bufw = bufio.NewWriter(out)
	defer func() {
		if ferr := bufw.Flush(); err == nil {
			err = ferr
		}
	}()

	const readSize = 4096
	var buf bytes.Buffer
	for err == nil {
		buf.Grow(readSize)
		err = readMore(&buf, in)
		if perr := processBatch(&buf, bufw); err == nil {
			err = perr
		}
	}
	if buf.Len() > 0 {
		log.Printf("undecoded trailer content: %q", buf.Bytes())
	}
	return err
}

func processBatch(buf *bytes.Buffer, w io.Writer) (err error) {
	writeRune := func(r rune) (size int, err error) {
		var b [4]byte
		n := utf8.EncodeRune(b[:], r)
		return w.Write(b[:n])
	}
	if rw := w.(interface {
		WriteRune(r rune) (size int, err error)
	}); rw != nil {
		writeRune = rw.WriteRune
	}

	for err == nil && buf.Len() > 0 {
		e, a, n := ansi.DecodeEscape(buf.Bytes())
		if n > 0 {
			buf.Next(n)
		}
		if e == 0 {
			r, n := utf8.DecodeRune(buf.Bytes())
			switch r {
			case 0x90, 0x9B, 0x9D, 0x9E, 0x9F: // DCS, CSI, OSC, PM, APC
				return
			case 0x1B: // ESC
				if p := buf.Bytes(); len(p) == cap(p) {
					return
				}
			}
			buf.Next(n)
			_, err = writeRune(r)
		} else if !*stripMode {
			if e == ansi.SGR {
				attr, _, decErr := ansi.DecodeSGR(a)
				if decErr == nil {
					_, err = fmt.Fprintf(w, "[ansi:SGR %v]", attr)
				} else {
					_, err = fmt.Fprintf(w, "[ansi:SGR ERR:%v %q]", decErr, a)
				}
			} else if len(a) > 0 {
				_, err = fmt.Fprintf(w, "[ansi:%v %q]", e, a)
			} else {
				_, err = fmt.Fprintf(w, "[ansi:%v]", e)
			}
		}
	}
	return err
}

func readMore(buf *bytes.Buffer, r io.Reader) error {
	b := buf.Bytes()
	b = b[len(b):cap(b)]
	n, err := r.Read(b)
	b = b[:n]
	buf.Write(b)
	return err
}

func runInteractive(term *anansi.Term) error {
	var (
		stop      = anansi.Notify(syscall.SIGTERM, syscall.SIGHUP)
		interrupt = anansi.Notify(syscall.SIGINT)
		resize    = anansi.Notify(syscall.SIGWINCH)
	)

	type readData struct {
		e   ansi.Escape
		a   []byte
		err error
	}

	input := make(chan readData)
	go func() {
		defer close(input)
		for {
			_, err := term.ReadMore()
			for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
				if a != nil {
					a = append([]byte(nil), a...)
				}
				input <- readData{e: e, a: a}
			}
			if err != nil {
				input <- readData{err: err}
				return
			}
		}
	}()

	stop.Open()
	interrupt.Open()
	resize.Open()

	defer stop.Close()
	defer interrupt.Close()
	defer resize.Close()

	for {
		select {
		case sig := <-stop.C:
			return anansi.SigErr(sig)

		case <-interrupt.C:
			// synthesize and handle a ^C escape
			if err := handle(term, 0x03, nil); err != nil {
				return err
			}

		case <-resize.C:
			if sz, err := term.Size(); err != nil {
				fmt.Printf("[ resize err:%v ]", err)
			} else {
				fmt.Printf("[ resize size:%v ]", sz)
			}

		case in := <-input:
			if in.err != nil {
				return in.err
			}
			if err := handle(term, in.e, in.a); err != nil {
				return err
			}
		}
	}
}

var prior ansi.Escape

func handle(term *anansi.Term, e ansi.Escape, a []byte) error {
	if _, err := fmt.Printf("%U %v", e, e); err != nil {
		return err
	}

	if len(a) > 0 {
		if _, err := fmt.Printf(" %q", a); err != nil {
			return err
		}
	}

	switch e {

	// print detail for mouse reporting
	case ansi.CSI('M'), ansi.CSI('m'):
		btn, pt, decErr := ansi.DecodeXtermExtendedMouse(e, a)
		if decErr != nil {
			if _, err := fmt.Printf(" mouse-err:%v", decErr); err != nil {
				return err
			}
		} else if _, err := fmt.Printf(" mouse-%v@%v", btn, pt); err != nil {
			return err
		}

	// ^C to quit
	case 0x03:
		if prior == 0x03 {
			return errors.New("goodbye")
		} else if _, err := fmt.Printf(" \x1b[91m<press Ctrl-C again to quit>\x1b[0m"); err != nil {
			return err
		}

	// ^L to clear
	case 0x0c:
		if prior == 0x0c {
			// 2 ED CUP
			if _, err := fmt.Printf("\x1b[2J\x1b[H"); err != nil {
				return err
			}
		} else if _, err := fmt.Printf(" \x1b[93m<press Ctrl-L again to quit>\x1b[0m"); err != nil {
			return err
		}

	// ^Z to suspend
	case 0x1a:
		if prior == 0x1a {
			if err := term.Suspend(); err != nil {
				return err
			}
		} else if _, err := fmt.Printf(" \x1b[92m<press Ctrl-Z again to suspend>\x1b[0m"); err != nil {
			return err
		}

	}

	prior = e

	_, err := fmt.Printf("\r\n")
	return err
}
