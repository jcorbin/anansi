package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

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
	flag.Parse()
	anansi.MustRun(run())
}

func run() error {
	model, err := readModel()
	if err != nil {
		return err
	}
	model = ansi.ColorModels(model, colorModel.model)
	return remapDircolors(os.Stdin, os.Stdout, model)
}

func readModel() (ansi.ColorModel, error) {
	name := flag.Arg(0)
	if name == "" {
		return nil, nil
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
	if fg, has := attr.FG(); has {
		attr = attr.SansFG() | model.Convert(fg).FG()
	}
	if bg, has := attr.BG(); has {
		attr = attr.SansBG() | model.Convert(bg).BG()
	}
	return attr
}

func remapDircolors(r io.Reader, w io.Writer, model ansi.ColorModel) error {
	var buf bytes.Buffer
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if parts := linePattern.FindSubmatch(sc.Bytes()); len(parts) != 0 && len(parts[1]) == 0 {
			attr, _, err := ansi.DecodeSGR(parts[3])
			if err != nil {
				return fmt.Errorf("invalid attr %q: %v (in %q)", parts[3], err, sc.Bytes())
			}
			if nattr := convertAttr(model, attr); nattr != attr {
				ctl := nattr.ControlString()
				buf.Write(parts[2])
				buf.WriteByte(' ')
				buf.WriteString(ctl[2 : len(ctl)-1])
				// buf.WriteString(" # prior: ")
				// buf.Write(parts[3])
			}
		}
		if buf.Len() == 0 {
			buf.Write(sc.Bytes())
		}
		buf.WriteByte('\n')
		if _, err := buf.WriteTo(w); err != nil {
			return err
		}
	}
	return sc.Err()
}
