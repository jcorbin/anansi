package platform

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jcorbin/anansi/ansi"
)

const (
	hudTimeFmt = "15:04:05.000000"
)

// HUD implements a toggle-able debug overlay.
type HUD struct {
	HUDState

	logs      *LogView
	profilers profilers

	detailWidth int
	sep         string

	bla image.Rectangle

	uicur UIID
}

// UIID identifies a user interface component
type UIID int

// HUDState contains serializable HUD state.
type HUDState struct {
	// last mouse for status display
	Mouse Mouse

	// structural state
	Visible    bool
	TimeDetail bool
	ProfDetail bool
	FPSControl bool
	FPSDetail  bool

	// theme
	SelectAttr ansi.SGRAttr
	ButtonAttr ansi.SGRAttr

	// user interaction state
	Active UIID
	EdLin  EditLine
}

var profileSpinner = [4]rune{'◴', '◷', '◶', '◵'}

func (hud *HUD) nextID() UIID {
	hud.uicur++
	return hud.uicur
}

// Update the HUD (only if visible).
func (hud *HUD) Update(ctx *Context, client Client) error {
	hud.uicur = 0

	if m, have := ctx.Input.LastMouse(false); have {
		hud.Mouse = m
	}

	// Ctrl-K toggles hud
	if n := ctx.Input.CountRune('\x0b'); n%2 == 1 {
		hud.Visible = !hud.Visible
	}

	// TODO should we elide all input from the client when paused?
	// TODO use (screen layers?) so that we can overlay the client's output,
	// but still have full (mouse) input priority.
	err := client.Update(ctx)

	if !hud.Visible {
		hud.Active = 0
		return err
	}

	// TODO why is input broken?

	ctx.Output.WriteESC(ansi.DECSC)

	ctx.Output.To(image.Pt(ctx.Output.Size.X+1, 1))

	box := hud.rightSegment(ctx, ctx.Time.Format(hudTimeFmt))
	if box != hud.bla {
		hud.bla = box
	}
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		hud.TimeDetail = !hud.TimeDetail
	}

	box = hud.rightSegment(ctx, fmt.Sprintf("FPS:%.0f/s", ctx.FPS()))
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		hud.FPSDetail = !hud.FPSDetail
	}

	hud.updateProfilers(ctx.Platform)
	if !hud.profilers.isActive() {
		box = hud.rightSegment(ctx, "⊙")
	} else {
		spinnerI := int(time.Duration(ctx.Platform.Time.Nanosecond()) / time.Millisecond / 250)
		box = hud.rightSegment(ctx, string(profileSpinner[spinnerI]))
	}
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		hud.ProfDetail = !hud.ProfDetail
	}

	hud.rightSegment(ctx, fmt.Sprintf("%vx%v", ctx.Output.Size.X, ctx.Output.Size.Y))
	hud.rightSegment(ctx, fmt.Sprintf("W:% 5v", ctx.Platform.output.Flushed))
	hud.rightSegment(ctx, hud.Mouse.String())

	// TODO better placed in footer? overlay?
	// if ctx.Platform.recording != nil {
	// 	var recBox image.Rectangle
	// 	recBox = hud.rightSegment(ctx, "RECORDING")
	// 	if n := ctx.CountPressesIn(recBox, 1); n%2 == 1 {
	// 		err = errOr(err, ctx.toggleRecRep())
	// 	}
	// }

	if hud.TimeDetail || hud.ProfDetail || hud.FPSDetail {
		var derr error
		ctx.Output.To(image.Pt(ctx.Output.Size.X+1, 2))
		derr = hud.drawTimingDetail(ctx)

		if hud.TimeDetail || hud.FPSDetail {
			hud.drawTelemetry(ctx)
		}

		hud.drawProfDetail(ctx)
		err = errOr(err, derr)
		hud.drawFPSDetail(ctx)
	}

	err = errOr(err, hud.logs.Update(ctx))

	ctx.Output.WriteESC(ansi.DECRC)

	return err
}

func (hud *HUD) apply(p *Platform) error {
	hud.detailWidth = 16
	hud.sep = strings.Repeat("-", hud.detailWidth)
	hud.logs = NewLogView(&Logs)
	hud.profilers = profilers{"All", make([]profiler, 0, 2*(2+len(p.pprofProfiles)))}

	hud.SelectAttr = ansi.SGRCube38.FG()
	hud.ButtonAttr = ansi.SGRCube42.FG()

	return nil
}

func (hud *HUD) updateProfilers(p *Platform) {
	hud.profilers.ps = hud.profilers.ps[:0]
	hud.profilers.ps = append(hud.profilers.ps,
		&p.cpuProfile,
		&p.traceProfile,
	)
	for i := range p.pprofProfiles {
		hud.profilers.ps = append(hud.profilers.ps, &p.pprofProfiles[i])
	}
}

func (hud *HUD) calcKVBox(ctx *Context, key string) (kbox, vbox image.Rectangle) {
	n := utf8.RuneCountInString(key)
	kbox.Min = image.Pt(ctx.Output.Size.X+1-hud.detailWidth-n-1, ctx.Output.Y)
	vbox.Min = image.Pt(ctx.Output.Size.X+1-hud.detailWidth, ctx.Output.Y)
	kbox.Max = image.Pt(kbox.Min.X+n, ctx.Output.Y+1)
	vbox.Max = image.Pt(vbox.Min.X+hud.detailWidth, ctx.Output.Y+1)
	return kbox, vbox
}

func (hud *HUD) detailHeader(ctx *Context, title string) (box image.Rectangle) {
	kbox, vbox := hud.calcKVBox(ctx, title)
	box.Min = kbox.Min
	box.Max = vbox.Max
	ctx.Output.To(kbox.Min)
	ctx.Output.WriteString(title)
	ctx.Output.WriteRune(' ')
	ctx.Output.To(vbox.Min)
	ctx.Output.WriteString(hud.sep)
	ctx.Output.To(box.Max)
	return box
}

func (hud *HUD) detailRow(ctx *Context, key string, val string) (box image.Rectangle) {
	kbox, vbox := hud.calcKVBox(ctx, key)
	box.Min = kbox.Min
	box.Max = vbox.Max
	n := utf8.RuneCountInString(key) + 2 + hud.detailWidth
	ctx.Output.To(image.Pt(ctx.Output.X-n+1, ctx.Output.Y))
	ctx.Output.WriteString(key)
	ctx.Output.WriteRune(' ')
	for i := utf8.RuneCountInString(val); i < hud.detailWidth; i++ {
		ctx.Output.WriteRune(' ')
	}
	ctx.Output.WriteString(val)
	ctx.Output.To(box.Max)
	return box
}

func (hud *HUD) rightSegment(ctx *Context, s string) (box image.Rectangle) {
	box.Min = ctx.Output.Point
	if box.Min.X < ctx.Output.Size.X {
		box.Min.X--
	}
	box.Max = image.Pt(box.Min.X, box.Min.Y+1)
	box.Min.X -= utf8.RuneCountInString(s)
	ctx.Output.To(box.Min)
	ctx.Output.WriteString(s)
	ctx.Output.To(box.Min)
	return box
}

func (hud *HUD) drawTimingDetail(ctx *Context) (err error) {
	if !hud.TimeDetail {
		return nil
	}
	hud.detailRow(ctx, "last", ctx.Time.Format(hudTimeFmt))
	hud.detailRow(ctx, "∂t", fmt.Sprintf("%.1fms", float64(ctx.Time.Sub(ctx.LastTime))/float64(time.Millisecond)))

	// Play/Pause control
	kbox, box := hud.calcKVBox(ctx, "paused")
	ctx.Output.To(kbox.Min)
	ctx.Output.WriteString("paused ")
	ctx.Output.To(box.Min)
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		ctx.Paused = !ctx.Paused
		ctx.Time = ctx.Platform.Time
	}
	if ctx.Paused {
		hud.drawButton(ctx, box, "Resume")
	} else {
		hud.drawButton(ctx, box, "Pause")
	}

	// Record/Replay control
	kbox, box = hud.calcKVBox(ctx, "replaying") // NOTE same size as "recording"
	ctx.Output.To(kbox.Min)

	if ctx.replay != nil {
		ctx.Output.WriteString("replaying ")
		ctx.Output.To(box.Min)
		repFmt := "%v/%v"
		if !ctx.replay.pause.IsZero() {
			repFmt = "paused %v/%v"
		}
		fmt.Fprintf(ctx.Output, repFmt,
			len(ctx.replay.input)-len(ctx.replay.cur), len(ctx.replay.input))
	} else {
		ctx.Output.WriteString("recording ")
		ctx.Output.To(box.Min)
		name := "Start"
		if ctx.recording != nil {
			name = ctx.recording.Name()
		}
		hud.drawButton(ctx, box, name)
		if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
			err = ctx.toggleRecRep()
		}
	}

	ctx.Output.To(box.Max)
	return err
}

func (hud *HUD) drawProfDetail(ctx *Context) {
	if !hud.ProfDetail {
		return
	}
	box := hud.detailHeader(ctx, "# Profiling Control:")
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		hud.ProfDetail = !hud.ProfDetail
		// TODO regrettable that we write a frame with lame-duck control header
		return
	}
	hudProfileControl{hud, hud.profilers}.update(ctx)
	for _, prof := range hud.profilers.ps {
		hudProfileControl{hud, prof}.update(ctx)
	}
	hud.drawPProfSelector(ctx)
	return
}

func (hud *HUD) drawPProfSelector(ctx *Context) {
	kbox, box := hud.calcKVBox(ctx, "PProf")
	ctx.Output.To(kbox.Min)
	ctx.Output.WriteString("PProf ")
	ctx.Output.To(box.Min)

	id := hud.nextID()
	selecting := false
	if hud.Active != id {
		if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
			hud.Active = id
		}
	}
	if hud.Active == id {
		selecting = true

		// collect list of inactive pprof profiles
		var pprofs []*pprof.Profile
		for _, prof := range pprof.Profiles() {
			have := false
			for i := range ctx.Platform.pprofProfiles {
				if ctx.Platform.pprofProfiles[i].profile == prof {
					have = true
					break
				}
			}
			if !have {
				pprofs = append(pprofs, prof)
			}
		}
		if len(pprofs) == 0 {
			return
		}

		// draw selector
		withAttr(ctx, hud.SelectAttr, ansi.SGRAttrClear, func(ctx *Context) {
			for _, prof := range pprofs {
				withOverAttr(ctx, hud.Mouse,
					image.Rect(box.Min.X, ctx.Output.Y, box.Max.X, ctx.Output.Y+1),
					hud.ButtonAttr, hud.SelectAttr,
					func(ctx *Context, _ bool) {
						ctx.Output.WriteString(prof.Name())
						ctx.Output.WriteString(strings.Repeat(" ", ctx.Output.Size.X-ctx.Output.X))
					})
				ctx.Output.To(image.Pt(box.Min.X, ctx.Output.Y+1))
				box.Max.Y = ctx.Output.Y
			}
		})

		// process selector click(s)
		for eid, kind := range ctx.Input.Type {
			if kind == EventMouse {
				m := ctx.Input.Mouse(eid)
				sid, pressed := m.State.IsPress()
				if pressed && sid == 1 {
					if m.Point.In(box) {
						rel := m.Point.Sub(box.Min)
						ctx.Platform.pprofProfiles = append(ctx.Platform.pprofProfiles, pprofProfileContext{
							profile: pprofs[rel.Y],
							debug:   1,
						})
						ctx.Input.Type[eid] = EventNone
					}
					hud.Active = 0
				}
			}
		}
	}

	if !selecting {
		hud.drawButton(ctx, box, "Add")
	}
	ctx.Output.To(box.Max)
}

func (hud *HUD) drawFPSDetail(ctx *Context) {
	if hud.FPSDetail {
		hud.drawFrameTiming(ctx)
		hud.drawStallsDetail(ctx)
	}
}

func (hud *HUD) drawTelemetry(ctx *Context) {
	hud.detailHeader(ctx, "# Telemetry:")
	hud.drawFileEditRow(ctx, &ctx.Telemetry.coll, "Go Log")
	hud.drawToggleRow(ctx, "Log Ticks", &ctx.LogTicks)
	hud.drawToggleRow(ctx, "Log Timing", &ctx.LogTiming)
	hud.drawToggleRow(ctx, "Log Stalls", &ctx.Telemetry.LogStallData)
}

func (hud *HUD) drawFrameTiming(ctx *Context) {
	box := hud.detailHeader(ctx, "# Frame Control:")
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		hud.FPSControl = !hud.FPSControl
	}
	if hud.FPSControl {
		hud.detailRow(ctx, "last ∂t", fmt.Sprintf("%.1fms", float64(ctx.LastTick.LastDelta)/float64(time.Millisecond)))
		hud.detailRow(ctx, "goal", fmt.Sprintf("%.1fms", float64(ctx.LastTick.GoalDelta)/float64(time.Millisecond)))
		hud.detailRow(ctx, "set", fmt.Sprintf("%.1fms", float64(ctx.LastTick.SetDelta)/float64(time.Millisecond)))
		hud.detailRow(ctx, "adj", fmt.Sprintf("%.1fms", float64(ctx.LastTick.Adjust)/float64(time.Millisecond)))
		hud.detailRow(ctx, "err", fmt.Sprintf("%.1fms", float64(ctx.LastTick.Err)/float64(time.Millisecond)))
		hud.detailRow(ctx, "∫", fmt.Sprintf("%.1fms", float64(ctx.LastTick.CumErr)/float64(time.Millisecond)))
		hud.detailRow(ctx, "∂", fmt.Sprintf("%.1fms", float64(ctx.LastTick.DeltaDiff)/float64(time.Millisecond)))
	}

	box = hud.detailHeader(ctx, "# Frame Timing:")
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		ctx.SetTimingEnabled(!ctx.Telemetry.TimingEnabled)
	}
	if ctx.Telemetry.TimingEnabled {
		stats := ctx.Timing.Stats
		goal := ctx.LastTick.GoalDelta
		hud.detailRow(ctx, "estimate", fmt.Sprintf("%.0f", ctx.FPSEstimate.Value))
		hud.detailRow(ctx, "actual", fmt.Sprintf("%.0f", stats.FPS))
		hud.detailRow(ctx, "∂t min", fmtMSDiff(stats.Min, goal))
		hud.detailRow(ctx, "q1", fmtMSDiff(stats.Q1, goal))
		hud.detailRow(ctx, "q2", fmtMSDiff(stats.Q2, goal))
		hud.detailRow(ctx, "q3", fmtMSDiff(stats.Q3, goal))
		hud.detailRow(ctx, "max", fmtMSDiff(stats.Max, goal))
	}
}

func fmtMSDiff(td, from time.Duration) string {
	e := float64(td)/float64(from) - 1.0
	return fmt.Sprintf("% +.1f%% %.1fms",
		100.0*e,
		float64(td)/float64(time.Millisecond),
	)
}

func (hud *HUD) drawStallsDetail(ctx *Context) {
	box := hud.detailHeader(ctx, "# Output Stalls:")
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		ctx.SetStallTracking(!ctx.Telemetry.StallDataEnabled)
	}
	if ctx.Telemetry.StallDataEnabled {
		stats := ctx.Telemetry.Stalls.Stats
		hud.detailRow(ctx, "Stalls", strconv.Itoa(stats.N))
		hud.detailRow(ctx, "min t", stats.Min.String())
		hud.detailRow(ctx, "max t", stats.Max.String())
		hud.detailRow(ctx, "∑ t", stats.Sum.String())
		hud.detailRow(ctx, "% t", fmt.Sprintf("%.2f%%", 100.0*stats.Pct))
	}
}

func (hud *HUD) drawButton(ctx *Context, box image.Rectangle, label string) {
	ctx.Output.To(box.Min)
	withAttr(ctx, hud.SelectAttr, ansi.SGRAttrClear, func(ctx *Context) {
		withOverAttr(ctx, hud.Mouse,
			image.Rect(box.Min.X, ctx.Output.Y, box.Max.X, ctx.Output.Y+1),
			hud.ButtonAttr, hud.SelectAttr,
			func(ctx *Context, _ bool) {
				ctx.Output.WriteString("[ ")
				max := box.Dx() - 4
				n := utf8.RuneCountInString(label)
				for i := 0; i < (max-n)/2; i++ {
					ctx.Output.WriteRune(' ')
				}
				writeTruncated(ctx, max, []byte(label))
				for ctx.Output.X < box.Max.X-2 {
					ctx.Output.WriteRune(' ')
				}
				ctx.Output.WriteString(" ]")
			})
	})
	ctx.Output.To(image.Pt(box.Max.X-1, box.Max.Y))
}

func (hud *HUD) underActivation(ctx *Context, box image.Rectangle, f func(*Context, image.Rectangle, UIID, bool)) bool {
	id := hud.nextID()
	enter := hud.Active != id
	if hud.Active != id {
		if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
			hud.Active = id
		}
	}
	// TODO unify press counting
	if hud.Active == id {
		if ctx.Input.AnyPressesOutside(box) {
			hud.Active = 0
		}
	}
	if hud.Active == id {
		f(ctx, box, id, enter)
	}
	return hud.Active == id
}

type fileable interface {
	name() string
	file() *os.File
}

func (hud *HUD) drawToggleRow(ctx *Context, name string, flag *bool) (kbox, box image.Rectangle) {
	kbox, box = hud.calcKVBox(ctx, name)
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		*flag = !*flag
	}
	ctx.Output.To(kbox.Min)
	ctx.Output.WriteString(name)
	ctx.Output.To(box.Min)
	if *flag {
		hud.drawButton(ctx, box, "true")
	} else {
		hud.drawButton(ctx, box, "false")
	}
	return kbox, box
}

func (hud *HUD) drawFileEditRow(ctx *Context, fil fileable, label string) (kbox, box image.Rectangle) {
	name := fil.name()
	kbox, box = hud.calcKVBox(ctx, name)
	ctx.Output.To(kbox.Min)
	ctx.Output.WriteString(name)
	ctx.Output.To(box.Min)
	if !hud.underActivation(ctx, box, func(ctx *Context, box image.Rectangle, id UIID, enter bool) {
		hud.underFileEdit(fil, ctx, box, id, enter)
	}) {
		if f := fil.file(); f != nil {
			label = f.Name()
		} else if label == "" {
			label = "Choose File"
		}
		hud.drawButton(ctx, box, label)
	}
	ctx.Output.To(box.Max)
	return kbox, box
}

func (hud *HUD) underFileEdit(fil fileable, ctx *Context, box image.Rectangle, id UIID, enter bool) {
	exit := true
	defer func() {
		if exit {
			hud.EdLin.Reset()
			hud.Active = 0
		}
	}()

	type creator interface{ create(string) error }
	creat, canCreate := fil.(creator)
	if !canCreate {
		return
	}

	exit = false
	if enter {
		hud.EdLin.Reset()
		if f := fil.file(); f != nil {
			hud.EdLin.Buf = append(hud.EdLin.Buf, f.Name()...)
		}
	}

	name := ""
	type dfltFN interface{ defaultFileName() string }
	if dfn, haveDflt := fil.(dfltFN); haveDflt {
		name = dfn.defaultFileName()
	}

	done := false
	defer func() {
		if done {
			if len(hud.EdLin.Buf) > 0 {
				name = string(hud.EdLin.Buf)
			}
			if name != "" {
				if err := creat.create(name); err != nil {
					log.Printf(
						"failed to create %q for %s: %v",
						name, fil.name(), err)
				}
			}
			exit = true
		}
	}()

	hud.EdLin.Box = box
	if hud.EdLin.Update(ctx); hud.EdLin.Active() {
		if len(hud.EdLin.Buf) == 0 && name != "" {
			if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
				done = true
			} else {
				ctx.Output.To(hud.EdLin.Box.Min)
				ctx.Output.WriteSGR(ansi.SGRGray12.FG())
				ctx.Output.WriteString(name)
				ctx.Output.WriteSGR(ansi.SGRAttrClear)
			}
		}
	} else if hud.EdLin.Done() {
		done = true
	} else {
		exit = true
	}
}

// LogView is a scrollable log viewer.
//
// TODO currently anchored to bottom of screen, and hardcoded to 10 lines high.
type LogView struct {
	LogViewState

	logs       *LogSink
	lastByte   int
	eolOffsets []int
}

// LogViewState contains serializable LogView state.
type LogViewState struct {
	ViewLines int
	Expanded  bool
	Line      int
}

// NewLogView creates a new log view attached to the given log buffer.
func NewLogView(logs *LogSink) *LogView {
	return &LogView{logs: logs}
}

// Update the log view, processing input, and drawing.
func (lv *LogView) Update(ctx *Context) error {
	// view calc input handling
	viewLines := lv.ViewLines
	if viewLines == 0 {
		viewLines = 10
		lv.ViewLines = viewLines
	}
	height := viewLines

	topLeft := image.Pt(1, ctx.Output.Size.Y)
	if lv.Expanded {
		height++
		topLeft.Y -= height - 1
	} else {
		height = 1
	}

	// TODO drag resizing
	if n := ctx.Input.CountPressesIn(image.Rectangle{topLeft, image.Pt(ctx.Output.Size.X+1, topLeft.Y+1)}, 1); n%2 == 1 {
		if lv.Expanded = !lv.Expanded; lv.Expanded {
			height = viewLines + 1
			topLeft.Y = ctx.Output.Size.Y - height + 1
		} else {
			height = 1
			topLeft.Y = ctx.Output.Size.Y
		}
	}

	if !lv.Expanded {
		viewLines = 1
	}

	content := lv.scanContent()
	lines := len(lv.eolOffsets) - 1
	if lines <= 0 {
		return nil
	}

	bounds := image.Rectangle{topLeft, ctx.Output.Size}
	bounds.Max = bounds.Max.Add(image.Pt(0, 1)) // TODO ideally utilize the final column too
	if diff := bounds.Dy() - lines + 1; diff > 0 {
		bounds.Min.Y += diff
	}
	if delta := ctx.Input.TotalScrollIn(bounds); delta != 0 {
		lv.scrollBy(lines, delta)
	}
	start, end := lv.viewWindow(lines, viewLines)

	// render
	ctx.Output.To(bounds.Min)
	if height > 1 {
		fmt.Fprintf(ctx.Output, "Logs (%v-%v/%v):", start, end, lines)
	}
	off := lv.eolOffsets[start-1] + 1
	for _, eol := range lv.eolOffsets[start : end+1] {
		ctx.Output.To(image.Pt(1, ctx.Output.Y+1))
		w := bounds.Max.X - ctx.Output.X
		writeTruncated(ctx, w, content[off:eol])
		off = eol + 1
	}
	return nil
}

func (lv *LogView) scanContent() []byte {
	content := bytes.TrimRight(lv.logs.Bytes(), "\n")
	if len(content) > lv.lastByte {
		if len(lv.eolOffsets) == 0 {
			lv.eolOffsets = append(lv.eolOffsets, -1)
		} else {
			lv.eolOffsets = lv.eolOffsets[:len(lv.eolOffsets)-1]
		}
		for off := lv.lastByte; off < len(content); off++ {
			i := bytes.IndexByte(content[off:], '\n')
			if i < 0 {
				break
			}
			off += i
			lv.eolOffsets = append(lv.eolOffsets, off)
		}
		lv.lastByte = len(content)
		lv.eolOffsets = append(lv.eolOffsets, lv.lastByte)
	}
	return content
}

func (lv *LogView) scrollBy(lines, delta int) {
	start := lv.Line
	if start == 0 {
		start = lines
	}
	start += delta
	if start >= lines {
		start = 0
	} else if start < 1 {
		start = 1
	}
	lv.Line = start
}

func (lv *LogView) viewWindow(lines, viewLines int) (int, int) {
	if lv.Line == 0 {
		if start := lines - viewLines + 1; start > 0 {
			return start, lines
		}
		return 1, lines
	}
	start, end := lv.Line, lv.Line
	for end < lines && end-start+1 < viewLines {
		end++
	}
	for start > 1 && end-start+1 < viewLines {
		start--
	}
	lv.Line = start
	return start, end
}

type hudProfileControl struct {
	*HUD
	prof profiler
}

func (hud hudProfileControl) update(ctx *Context) {
	kbox, box := hud.drawFileEditRow(ctx, hud.prof, "")
	hud.drawSpinner(ctx, image.Rect(kbox.Max.X, kbox.Min.Y, box.Min.X, kbox.Max.Y))
	ctx.Output.To(box.Max)
}

func (hud hudProfileControl) drawSpinner(ctx *Context, box image.Rectangle) {
	spinner := '⊙'
	if hud.prof.isActive() {
		spinner = profileSpinner[int(time.Duration(ctx.Platform.Time.Nanosecond())/time.Millisecond/250)]
	}
	ctx.Output.To(box.Min)
	ctx.Output.WriteRune(spinner)
	if n := ctx.Input.CountPressesIn(box, 1); n%2 == 1 {
		if hud.prof.isActive() {
			if err := hud.prof.stop(); err != nil {
				log.Printf("failed to stop %s profiling: %v", hud.prof.name(), err)
			}
		} else if err := hud.prof.start(); err != nil {
			log.Printf("failed to start %s profiling: %v", hud.prof.name(), err)
		}
	}
}

type profiler interface {
	fileable
	isActive() bool
	start() error
	stop() error
}

type profilers struct {
	nom string
	ps  []profiler
}

var nullFile *os.File

func init() {
	f, err := os.Open(os.DevNull)
	if err != nil {
		panic(err.Error())
	}
	nullFile = f
}

func (ps profilers) name() string { return ps.nom }
func (ps profilers) isActive() bool {
	for i := range ps.ps {
		if ps.ps[i].isActive() {
			return true
		}
	}
	return false
}
func (ps profilers) file() *os.File { return nullFile }
func (ps profilers) start() (err error) {
	for i := range ps.ps {
		err = errOr(err, ps.ps[i].start())
	}
	return err
}
func (ps profilers) stop() (err error) {
	for i := range ps.ps {
		err = errOr(err, ps.ps[i].stop())
	}
	return err
}

func withAttr(ctx *Context, e, x ansi.SGRAttr, f func(*Context)) {
	ctx.Output.WriteSGR(e)
	f(ctx)
	ctx.Output.WriteSGR(x) // TODO should be unnecessary with proper cursor state
}

func withOverAttr(
	ctx *Context, m Mouse,
	box image.Rectangle, o, a ansi.SGRAttr,
	f func(*Context, bool),
) {
	over := m.Point.In(box)
	if over {
		ctx.Output.WriteSGR(o)
	}
	f(ctx, over)
	if over {
		ctx.Output.WriteSGR(a) // TODO should be unnecessary with proper cursor state
	}
}

func writeTruncated(ctx *Context, w int, b []byte) {
	// if !ctx.Output.Point.In(bounds) { return }
	n := utf8.RuneCount(b)
	if rem := n - w; rem > 0 {
		for rem++; rem > 0; rem-- {
			_, m := utf8.DecodeLastRune(b)
			b = b[:len(b)-m]
		}
		ctx.Output.Write(b)
		ctx.Output.WriteRune('…')
	} else {
		ctx.Output.Write(b)
	}
}
