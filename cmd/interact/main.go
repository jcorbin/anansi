package main

/*

This is a very rough prototype of an "interact" command:
- given some unix command, allow the user to define any $variables in it
- uses anansi's x/platform layer for convenience, but that's really overkill
  for an application like this (which has no real need for an animation loop at
  any FPS)
- the variable editing UX leaves much to be desired:
  - the editline doesn't properly occlude the prior drawn varible...
  - ... TODO any easy way would be to dynamically define the editline rectangle
  - ... the lack of even basic emacs-style keybinds chaffs
- beyond basic free-form variables, there's an obvious path for a small DSL to
  specify numeric arguments, enumeration arguments, and such
- another direction would be allow free from adding and removing of arguments
- command output handling is nascent, should at least support paging,
  scrolling, etc; may even consider embedding $PAGER here...
- ...speaking of embedding, a more advanced feature would be to support a stdin
  file, and embed $EDITOR
- finally, other adjacent features, like parity with `watch(1)` come easily to mind

*/

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/anansi/anui"
	"github.com/jcorbin/anansi/x/platform"
)

func main() {
	anansi.MustRun(run())
}

func run() error {
	term, err := anansi.OpenTerm()
	if err != nil {
		return err
	}

	term.AddMode(
		ansi.ModeMouseSgrExt,
		ansi.ModeMouseBtnEvent,
	)

	cmd := flag.Args()
	in := inspect{}
	in.setCmd(cmd)

	return anui.RunTermLayer(term, &in)
}

type inspect struct {
	immLayer

	cmd  []string
	argi []int
	arg  []string
	val  []string

	edid int
	ed   editLine

	cmdOutput anansi.ScreenDiffer // XXX VirtualScreen
}

func (in *inspect) setCmd(cmd []string) {
	in.cmd = append(in.cmd[:0], cmd...)
	in.argi = in.argi[:0]
	in.arg = in.arg[:0]
	in.val = in.val[:0]
	for i, arg := range in.cmd {
		if strings.HasPrefix(arg, "\\$") {
			in.cmd[i] = arg[1:]
		} else if strings.HasPrefix(arg, "$") {
			in.argi = append(in.argi, i)
			in.arg = append(in.arg, arg[1:])
			in.val = append(in.val, "")
		}
	}
	in.runCmd()
}

func (in *inspect) haveAllArgVals() bool {
	for _, val := range in.val {
		if val == "" {
			return false
		}
	}
	return true
}

func (in *inspect) runCmd() {
	in.cmdOutput.Clear()
	if in.cmdOutput.Bounds().Empty() {
		return
	}
	in.cmdOutput.To(ansi.Pt(1, 1))
	if !in.haveAllArgVals() {
		in.cmdOutput.WriteString("Define variables to run")
		return
	}

	args := append([]string(nil), in.cmd...)
	for ii, val := range in.val {
		args[in.argi[ii]] = val
	}

	cmd := exec.Command(args[0], args[1:]...)
	// TODO pipe into cmdOutput; pty
	out, err := cmd.Output()

	// TODO CR handling no work
	fmt.Fprintf(&in.cmdOutput, "status: %v\r\n", cmd.ProcessState)
	if err != nil {
		in.cmdOutput.WriteSGR(ansi.SGRRed.FG())
		fmt.Fprintf(&in.cmdOutput, "error: %v\r\n", err)
		in.cmdOutput.WriteSGR(ansi.SGRAttrClear)
	}
	in.cmdOutput.Write(out)

}

func (in *inspect) Update(inp immInput, out anansi.Screen, now time.Time) (needsDraw time.Duration) {
	run := false
	defer func() {
		if run {
			in.runCmd()
		}
	}()

	p := out.Rect.Min
	i, j := 0, 0

	restart := func() {
		p = out.Rect.Min
		i, j = 0, 0
	}

	for ; i < len(in.cmd); i++ {
		arg := in.cmd[i]
		if i > 0 {
			if !out.SetCell(p, ' ', 0) {
				break
			}
		}
		var keyAttr, valAttr ansi.SGRAttr
		var val *string
		if j < len(in.argi) && in.argi[j] == i {
			keyAttr = ansi.SGRCyan.FG()
			valAttr = ansi.SGRBrightBlue.FG()
			val = &in.val[j]
			j++
		} else if i == 0 {
			keyAttr = ansi.SGRGreen.FG()
		}
		argBox := ansi.Rectangle{p, p}

		for _, r := range arg {
			if !out.SetCell(p, r, keyAttr) {
				break
			}
			p.X++
		}
		if val != nil {
			if out.SetCell(p, '=', 0) {
				p.X++
			}
			for _, r := range *val {
				if !out.SetCell(p, r, valAttr) {
					break
				}
				p.X++
			}
		}

		if p.X > argBox.Min.X {
			argBox.Max = ansi.Pt(p.X, p.Y+1)
		}
		if argBox.Empty() {
			break
		}

		if val == nil {
			continue
		}

		func() {
			edid := i + 1

			if m, is := inp.Mouse(); is && m.Point.In(argBox) {
				in.ed.Reset()
				if in.edid == edid {
					in.edid = 0
					run = true
				} else {
					in.edid = edid
					in.ed.Buf = append(in.ed.Buf, *val...)
				}
				if inp.Next() {
					restart()
					return
				}
			}

			if in.edid != edid {
				return
			}

			needsDraw = in.ed.Update(inp, out.SubRect(argBox), now)

			if !in.ed.Active() {
				// TODO draw ghost
				// if len(in.ed.Buf) == 0 && name != "" { }
			} else {
				if in.ed.Done() && len(in.ed.Buf) > 0 {
					*val = string(in.ed.Buf)
				}
				in.edid = 0
				in.ed.Reset()
				run = true
			}
		}()

	}

	p.Y++

	sz := out.Bounds().Size()
	sz.Y -= p.Y
	if !in.cmdOutput.Bounds().Size().Eq(sz) {
		log.Printf("resize cmd out %v", sz)
		in.cmdOutput.Resize(sz)
		in.runCmd()
	}

	// TODO scroll w/in cmdOutput
	anansi.DrawGrid(out.Grid.SubAt(p), in.cmdOutput.Grid)

	return needsDraw
}

// TODO move into anui

type editLine struct {
	platform.EditLine // XXX port to anui
}

func (ed *editLine) Update(in immInput, out anansi.Screen, now time.Time) time.Duration

// TODO move into anui

type immUI interface {
	Update(in immInput, out anansi.Screen, now time.Time) time.Duration
}

type immInput interface {
	Any() bool
	Next() bool
	Rune() (r rune)
	Mouse() (m ansi.MouseEvent, is bool)
	Escape() (e ansi.Escape, a []byte)
}

type immLayer struct {
	input     inputBuffer
	needsDraw time.Duration
	ui        immUI
}

func (im *immLayer) Draw(sc anansi.Screen, now time.Time) anansi.Screen {
	im.needsDraw = im.ui.Update(&im.input, sc, now)
	return sc
}

func (im *immLayer) NeedsDraw() time.Duration {
	return im.needsDraw
}

func (im *immLayer) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	return true, im.input.add(e, a)
}

type inputType uint8

const (
	inputNone inputType = iota
	inputRune
	inputEscape
	inputMouse
)

type inputBuffer struct {
	t []inputType
	e []ansi.Escape
	m []ansi.MouseEvent
	r [][2]int
	b bytes.Buffer

	current int
}

func (ib *inputBuffer) Next() bool {
	if ib.current < len(ib.t) {
		ib.current++
	}
	return ib.Any()
}

func (ib *inputBuffer) Any() bool {
	return ib.current < len(ib.t)
}

func (ib *inputBuffer) Rune() (r rune) {
	if ib.t[ib.current] == inputRune {
		r := rune(ib.e[ib.current])
	}
	return r
}

func (ib *inputBuffer) Mouse() (m ansi.MouseEvent, is bool) {
	if ib.t[ib.current] == inputMouse {
		m := ib.m[ib.current]
		is = true
	}
	return m, is
}

func (ib *inputBuffer) Escape() (e ansi.Escape, a []byte) {
	if ib.t[ib.current] == inputEscape {
		e = ib.e[ib.current]
		if r := ib.r[ib.current]; r[1] > r[0] {
			a = ib.b.Bytes()[r[0]:r[1]]
		}
	}
	return e, a
}

func (ib *inputBuffer) add(e ansi.Escape, a []byte) (err error) {
	var (
		t inputType
		m ansi.MouseEvent
		r [2]int
	)

	switch {

	case e == ansi.CSI('M'), e == ansi.CSI('m'):
		t = inputMouse
		m, err = ansi.ParseMouseEvent(e, a)

	case e.IsEscape(), len(a) > 0:
		t = inputEscape
		if len(a) > 0 {
			r[0] = ib.b.Len()
			_, _ = ib.b.Write(a)
			r[1] = ib.b.Len()
		}

	default:
		t = inputRune

	}

	if err != nil || t == inputNone {
		return err
	}

	if len(ib.t) == cap(ib.t) && ib.current > 0 {
		ib.shift(ib.current)
	}

	ib.t = append(ib.t, t)
	ib.e = append(ib.e, e)
	ib.m = append(ib.m, m)
	ib.r = append(ib.r, r)

	return nil
}

func (ib *inputBuffer) shift(n int) {
	m := 0
	if n >= len(ib.t) {
		ib.b.Reset()
	} else {
		m = copy(ib.t, ib.t[n:])
		copy(ib.e, ib.e[n:])
		copy(ib.m, ib.m[n:])
		bn := ib.r[n][0]
		ib.b.Next(bn)
		for i, j := 0, n; j < len(ib.r); i, j = i+1, j+1 {
			ib.r[i] = [2]int{ib.r[j][0] - bn, ib.r[j][1] - bn}
		}
	}
	ib.t = ib.t[:m]
	ib.e = ib.e[:m]
	ib.m = ib.m[:m]
	ib.r = ib.r[:m]
}
