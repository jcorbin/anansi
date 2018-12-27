package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

func main() {
	anansi.MustRun(run())
}

func run() error {
	colors, err := readColors(os.Stdin)
	if err != nil {
		return err
	}
	for _, c := range colors {
		r, g, b := c.RGB()
		oc := ansi.Palette8.Convert(c)
		fmt.Printf(
			"#%02x%02x%02x\t%v\t%q\t%v\t%q\n",
			r, g, b,
			c,
			c.FG().ControlString(),
			oc,
			oc.FG().ControlString(),
		)
	}

	return nil
}

var hexLinePattern = regexp.MustCompile(`^#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$`)

func readColors(r io.Reader) (colors []ansi.SGRColor, _ error) {
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
