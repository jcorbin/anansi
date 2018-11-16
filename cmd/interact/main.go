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
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/anansi/x/platform"
)

var errInt = errors.New("interrupt")

func main() {
	platform.MustRun(os.Stdout, Run, platform.FrameRate(60))
}

// Run the demo under an active terminal platform.
func Run(p *platform.Platform) error {
	cmd := flag.Args()
	for {
		in := inspect{}
		in.setCmd(cmd)
		if err := p.Run(&in); platform.IsReplayDone(err) {
			continue // loop replay
		} else if err == io.EOF || err == errInt {
			return nil
		} else if err != nil {
			log.Printf("exiting due to %v", err)
			return err
		}
	}
}

type inspect struct {
	cmd  []string
	argi []int
	arg  []string
	val  []string

	edid int
	ed   platform.EditLine

	cmdOutput anansi.Screen
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

func (in *inspect) Update(ctx *platform.Context) (err error) {
	// Ctrl-C interrupts
	if ctx.Input.HasTerminal('\x03') {
		// ... AFTER any other available input has been processed
		err = errInt
		// ... NOTE err != nil will prevent wasting any time flushing the final
		//          lame-duck frame
	}

	// Ctrl-Z suspends
	if ctx.Input.CountRune('\x1a') > 0 {
		defer func() {
			if err == nil {
				err = ctx.Suspend()
			} // else NOTE don't bother suspending, e.g. if Ctrl-C was also present
		}()
	}

	ctx.Output.Clear()
	p := ansi.Pt(1, 1)
	if ctx.HUD.Visible {
		p.Y++
	}
	ctx.Output.To(p)

	j := 0
	for i, arg := range in.cmd {
		if i > 0 {
			ctx.Output.WriteRune(' ')
		}
		var attr ansi.SGRAttr
		var val *string
		if j < len(in.argi) && in.argi[j] == i {
			attr = ansi.SGRCyan.FG()
			val = &in.val[j]
			j++
		} else if i == 0 {
			attr = ansi.SGRGreen.FG()
		}
		if attr != 0 {
			ctx.Output.WriteSGR(attr)
		}
		var r ansi.Rectangle
		r.Min = ctx.Output.Point
		ctx.Output.WriteString(arg)
		r.Max = ctx.Output.Point
		r.Max.Y++
		if attr != 0 {
			ctx.Output.WriteSGR(ansi.SGRAttrClear)
		}
		if val != nil && *val != "" {
			ctx.Output.WriteRune('=')
			ctx.Output.WriteSGR(ansi.SGRBrightBlue.FG())
			ctx.Output.WriteString(*val)
			ctx.Output.WriteSGR(ansi.SGRAttrClear)
		}

		var enter bool
		edid := i + 1
		if ctx.Input.CountPressesIn(r, 1)%2 == 1 {
			if in.edid == edid {
				in.edid = 0
			} else {
				enter = true
				in.edid = edid
			}
		}

		if in.edid == edid {
			func() {
				exit := false
				defer func() {
					if exit {
						in.ed.Reset()
						in.edid = 0
						in.runCmd()
					}
				}()
				if val == nil {
					return
				}

				if enter {
					in.ed.Reset()
					in.ed.Buf = append(in.ed.Buf, *val...)
				}

				done := false
				defer func() {
					if done {
						if len(in.ed.Buf) > 0 {
							*val = string(in.ed.Buf)
						}
						exit = true
					}
				}()

				in.ed.Box = r // TODO expand more space
				if in.ed.Update(ctx); in.ed.Active() {
					// if len(in.ed.Buf) == 0 && name != "" { }
				} else if in.ed.Done() {
					done = true
				} else {
					exit = true
				}
			}()
		}

	}
	p.Y++

	sz := ctx.Output.Bounds().Size()
	sz.Y -= p.Y
	if !in.cmdOutput.Bounds().Size().Eq(sz) {
		log.Printf("resize cmd out %v", sz)
		in.cmdOutput.Resize(sz)
		in.runCmd()
	}

	// TODO scroll w/in cmdOutput
	anansi.DrawGrid(ctx.Output.Grid.SubAt(p), in.cmdOutput.Grid)

	return err
}
