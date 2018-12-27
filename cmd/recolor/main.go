package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

var colorModels = map[string]ansi.ColorModel{
	"id":    ansi.ColorModelID,
	"3":     ansi.Palette3,
	"4":     ansi.Palette4,
	"8":     ansi.Palette8,
	"24":    ansi.ColorModel24,
	"canon": ansi.SGRColorCanon,
}

type colorModelFlag struct {
	name  string
	model ansi.ColorModel
}

func (cmf *colorModelFlag) String() string { return cmf.name }
func (cmf *colorModelFlag) Set(s string) error {
	model := colorModels[s]
	if model == nil {
		names := make([]string, 0, len(colorModels))
		for name := range colorModels {
			names = append(names, name)
		}
		return fmt.Errorf("no such color model, valid choices:%q", names)
	}
	cmf.name = s
	cmf.model = model
	return nil
}

var colorModel = colorModelFlag{"id", ansi.ColorModelID}

func init() {
	flag.Var(&colorModel, "model", "output color model")
}

func main() {
	// TODO default colorModel from env
	flag.Parse()
	anansi.MustRun(run())
}

func run() error {
	model, err := readModel()
	if err != nil {
		return err
	}
	model = ansi.ColorModels(model, colorModel.model)
	return recolor(os.Stdin, os.Stdout, model)
}

func readModel() (ansi.ColorModel, error) {
	name := flag.Arg(0)
	if name == "" {
		return nil, errors.New("missing palette argument")
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	colors, err := readHexColors(f)
	return ansi.Palette(colors), err
}

var hexLinePattern = regexp.MustCompile(`#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})`)

func readHexColors(r io.Reader) (colors []ansi.SGRColor, _ error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		parts := hexLinePattern.FindStringSubmatch(line)
		if len(parts) == 0 {
			return nil, fmt.Errorf("bad line %q, expected %v", line, hexLinePattern)
		}
		r, _ := strconv.ParseInt(parts[1], 16, 64)
		g, _ := strconv.ParseInt(parts[2], 16, 64)
		b, _ := strconv.ParseInt(parts[3], 16, 64)
		colors = append(colors, ansi.RGB(uint8(r), uint8(g), uint8(b)))
	}
	return colors, sc.Err()
}

var linePattern = regexp.MustCompile(`^(\s*#|\s*TERM)|([^\s]+)\s+([^\s]+)`)

func convertAttr(model ansi.ColorModel, attr ansi.SGRAttr) ansi.SGRAttr {
	// TODO cache
	if fg, has := attr.FG(); has {
		attr = attr.SansFG() | model.Convert(fg).FG()
	}
	if bg, has := attr.BG(); has {
		attr = attr.SansBG() | model.Convert(bg).BG()
	}
	return attr
}

func recolor(r io.Reader, w io.Writer, model ansi.ColorModel) error {
	const minReadSize = 1024
	var inbuf, outbuf bytes.Buffer

	for {
		inbuf.Grow(minReadSize)
		p := inbuf.Bytes()
		p = p[len(p):cap(p)]
		n, err := r.Read(p)
		_, _ = inbuf.Write(p[:n])
		outbuf.Grow(inbuf.Len())

		for done := false; !done && inbuf.Len() > 0; {
			e, a, n := ansi.DecodeEscape(inbuf.Bytes())
			if n > 0 {
				inbuf.Next(n)
			}
			if e != 0 {
				b := outbuf.Bytes()
				b = b[len(b):]
				switch e {
				case ansi.SGR:
					if attr, _, err := ansi.DecodeSGR(a); err == nil {
						attr = convertAttr(model, attr)
						b = attr.AppendTo(b)
					} else {
						b = e.AppendWith(b, a...)
					}
				default:
					b = e.AppendWith(b, a...)
				}
				_, _ = outbuf.Write(b)
				continue
			}

			r, n := utf8.DecodeRune(inbuf.Bytes())
			if err == nil {
				switch r {
				case 0x90, 0x9B, 0x9D, 0x9E, 0x9F: // DCS, CSI, OSC, PM, APC
					done = true
					continue
				case 0x1B: // ESC
					if p := inbuf.Bytes(); len(p) == cap(p) {
						done = true
						continue
					}
				}
			}
			inbuf.Next(n)
			_, _ = outbuf.WriteRune(r)
		}

		if outbuf.Len() > 0 {
			if _, werr := outbuf.WriteTo(w); err == nil {
				err = werr
			}
		}

		if err != nil {
			return err
		}
	}
}
