package main

import (
	"flag"
	"fmt"

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

var colorCtlWidths = map[string]int{
	"":   24, // default
	"3":  12,
	"4":  12,
	"8":  16,
	"id": 16,
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

func main() {
	model := colorModelFlag{"id", ansi.ColorModelID}
	showThemes := false
	flag.Var(&model, "model", "output color model")
	flag.BoolVar(&showThemes, "themes", false, "list color theme palettes too")
	flag.Parse()

	rst := ansi.SGRAttrClear.ControlString()
	fmt.Printf("using %q model (pass -model to change)\n", model.name)

	ctlWidth := colorCtlWidths[model.name]
	if ctlWidth == 0 {
		ctlWidth = colorCtlWidths[""]
	}

	type palEnt struct {
		name, extra string
		off         int
		pal         ansi.Palette
	}

	pals := []palEnt{
		{"Palette3", "", 0, ansi.Palette3},
		{"Palette4", "Palette3 + ...", 8, ansi.Palette4[8:]},
		{"Palette8", "Palette4 + ...", 16, ansi.Palette8[16:]},
	}

	themes := []palEnt{
		{"VGAPalette", "", 0, ansi.Palette(ansi.VGAPalette)},
		{"CMDPalette", "", 0, ansi.Palette(ansi.CMDPalette)},
		{"TermnialAppPalette", "", 0, ansi.Palette(ansi.TermnialAppPalette)},
		{"PuTTYPalette", "", 0, ansi.Palette(ansi.PuTTYPalette)},
		{"MIRCPalette", "", 0, ansi.Palette(ansi.MIRCPalette)},
		{"XTermPalette", "", 0, ansi.Palette(ansi.XTermPalette)},
		{"XPalette", "", 0, ansi.Palette(ansi.XPalette)},
		{"UbuntuPalette", "", 0, ansi.Palette(ansi.UbuntuPalette)},
	}

	if showThemes {
		pals = append(append(pals[:2:2], themes...), pals[2])
	}

	for i, p := range pals {
		if i > 0 {
			fmt.Printf("\n")
		}
		if p.extra != "" {
			fmt.Printf("%s (%s):\n", p.name, p.extra)
		} else {
			fmt.Printf("%s:\n", p.name)
		}
		for i, c := range p.pal {
			c = model.model.Convert(c)
			fg := c.FG().ControlString()
			bg := c.BG().ControlString()
			fmt.Printf(
				"%s% 3d %s[ fg ]%s%s[ bg ]%s fg_ctl:% -*q bg_ctl:% -*q\n",
				rst, p.off+i,
				fg,
				rst,
				bg,
				rst,
				ctlWidth, fg,
				ctlWidth, bg,
			)
		}
	}
}
