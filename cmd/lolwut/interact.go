package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/anansi/anui"
)

type schotterDemoUI struct {
	*schotterDemo
	lastDraw time.Time
	rps      float64
}

func runInteractive() {
	anansi.MustRun(func() error {
		var ui schotterDemoUI
		ui.schotterDemo = &sd
		ui.squareSide = 20 // TODO push down, pre-compute based on initial width and squaresPerRow

		term, err := anansi.OpenTerm()
		if err != nil {
			return err
		}

		term.AddMode(
			ansi.ModeMouseSgrExt,
			ansi.ModeMouseBtnEvent,
		)

		return anui.RunTermLayer(term, &drawTimeStats{
			Layer: &ui,
			// log: true,
			overlay: func(sc anansi.Screen) anansi.Screen {
				return sc.SubAt(sc.Rect.Min.Add(image.Pt(0, sc.Rect.Dy()-1)))
			},
			elapsed: elapsedStats{
				e: make([]time.Duration, 15),
			},
		},
			anui.DefaultOptions,
			// anui.WithSyncDrawRate(60),
			anui.WithAsyncDrawRate(60),
		)
	}())
}

func (sd *schotterDemoUI) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	switch e {

	// mouse-wheel zooms
	case ansi.CSI('M'), ansi.CSI('m'):
		m, err := ansi.ParseMouseEvent(e, a)
		if err != nil {
			return false, err
		}
		switch m.State.ButtonID() {
		case 4: // wheel-up
			sd.squareSide--
			if sd.squareSide < 1 {
				sd.squareSide = 1
			}
		case 5: // wheel-down
			sd.squareSide++
		}

	// speed controls
	case '-':
		sd.rps /= 2
	case '+':
		sd.rps *= 2

	}
	return false, nil
}

func (sd *schotterDemoUI) NeedsDraw() time.Duration {
	return 1
}

func (sd *schotterDemoUI) Draw(sc anansi.Screen, now time.Time) anansi.Screen {
	// compute animation time elapsed
	var elapsed time.Duration
	if !sd.lastDraw.IsZero() {
		elapsed = now.Sub(sd.lastDraw)
	}
	sd.lastDraw = now

	// animate rotation angle
	if sd.rps == 0 {
		sd.rps = 1.0
	}
	var angleRate = 2 * math.Pi * sd.rps / float64(time.Second)
	sd.angleOffset += math.Mod(float64(elapsed)*angleRate, 2*math.Pi)

	// compute canvas size, resize if needed
	screenSize := sc.Bounds().Size()
	canvasSize := sd.canvas.Rect.Size()
	sd.padding = 0
	if screenSize.X > 2 {
		sd.padding = 2
	}
	canvasSize.X = screenSize.X * 2
	canvasSize.Y = screenSize.Y * 4
	roundUp := sd.squareSide - 1
	sd.squaresPerRow = ((screenSize.X-sd.padding)*2 + roundUp) / sd.squareSide
	sd.squaresPerCol = ((screenSize.Y-sd.padding)*4 + roundUp) / sd.squareSide
	if canvasSize != sd.canvas.Rect.Size() {
		sd.canvas.Resize(canvasSize)
	}

	// clear canvas and redraw
	for i := range sd.canvas.Bit {
		sd.canvas.Bit[i] = false
	}
	sd.draw()

	// draw bitmap canvas into terminal grid
	anansi.DrawBitmap(sc.Grid, sd.canvas)

	return sc
}

type drawTimeStats struct {
	anui.Layer

	overlay func(sc anansi.Screen) anansi.Screen
	log     bool

	i int
	t []time.Time

	elapsed elapsedStats

	elBuf, outBuf bytes.Buffer
}

func (dts *drawTimeStats) collect(now time.Time) {
	if len(dts.t) == 0 {
		dts.t = make([]time.Time, 4*60)
	}
	if len(dts.t) > 0 {
		dts.t[dts.i] = now
		dts.i = (dts.i + 1) % len(dts.t)
	}
}

func (dts *drawTimeStats) count(now time.Time) (n int, over time.Duration) {
	if len(dts.t) == 0 {
		return 0, 0
	}
	start := now
	for i := (dts.i - 1 + len(dts.t)) % len(dts.t); i != dts.i; i = (i - 1 + len(dts.t)) % len(dts.t) {
		t := dts.t[i]
		if t.IsZero() {
			break
		}
		if !t.Equal(now) && !t.Before(now) {
			break
		}
		start = t
		n++
	}
	return n, now.Sub(start)
}

func (dts *drawTimeStats) Draw(sc anansi.Screen, now time.Time) anansi.Screen {
	sc = dts.Layer.Draw(sc, now)

	dts.collect(now)

	if dts.elapsed.collect(now) {
		q1 := dts.elapsed.quantile(0.25)
		q2 := dts.elapsed.quantile(0.5)
		q3 := dts.elapsed.quantile(0.75)
		iqr := q3 - q1
		hi := q2 + iqr + iqr/2
		normRank := dts.elapsed.rank(hi)

		if dts.log {
			log.Printf(
				"elapsed q1:%v q2:%v q3:%v normRank:%v/%v",
				q1, q2, q3, normRank, len(dts.elapsed.edb),
			)
		}
		if dts.overlay != nil {
			dts.elBuf.Reset()
			_, _ = fmt.Fprintf(&dts.elBuf,
				"∂t:[|%.1fms %.1fms %.1fms| iqr:%.1fms",
				float64(q1)/float64(time.Millisecond),
				float64(q2)/float64(time.Millisecond),
				float64(q3)/float64(time.Millisecond),
				float64(iqr)/float64(time.Millisecond),
			)

			if out := len(dts.elapsed.edb) - normRank; out > 0 {
				_, _ = fmt.Fprintf(&dts.elBuf, " out:%v/%v", out, len(dts.elapsed.edb))
			}

			_ = dts.elBuf.WriteByte(']')
		}
	}

	if dts.overlay != nil {
		dts.outBuf.Reset()
		if dts.elBuf.Len() > 0 {
			_, _ = dts.outBuf.Write(dts.elBuf.Bytes())
		}
	}

	if n, over := dts.count(now); n > 0 && over > 0 {
		fps := float64(n) / (float64(over) / float64(time.Second))
		if dts.log {
			log.Printf("draw timing rate:%v/s over:%v", fps, over)
		}
		if dts.overlay != nil {
			if dts.outBuf.Len() > 0 {
				_ = dts.outBuf.WriteByte(' ')
			}
			_, _ = dts.outBuf.WriteString("FPS:")
			_, _ = dts.outBuf.WriteString(strconv.FormatFloat(fps, 'f', 0, 64))
		}
	}

	if dts.overlay != nil && dts.outBuf.Len() > 0 {
		sc := dts.overlay(sc)
		sc = sc.SubAt(sc.Rect.Min.Add(image.Pt(sc.Rect.Dx()-dts.outBuf.Len(), 0)))
		anansi.Process(&sc, dts.outBuf.Bytes())
	}

	return sc
}

type elapsedStats struct {
	lastDraw time.Time
	e        []time.Duration
	edb      []time.Duration
}

func (es *elapsedStats) collect(now time.Time) (updated bool) {
	if es.e != nil {
		if !es.lastDraw.IsZero() {
			elapsed := now.Sub(es.lastDraw)
			es.e = append(es.e, elapsed)
			if len(es.e) == cap(es.e) {
				sort.Slice(es.e, func(i, j int) bool {
					return es.e[i] < es.e[j]
				})
				es.edb = mergeSortedDurations(es.edb, es.e)
				es.e = es.e[:0]
				updated = true
			}
		}
	}

	es.lastDraw = now

	return updated
}

func (es *elapsedStats) quantile(q float64) time.Duration {
	n := float64(len(es.edb)) * q
	i := int(math.Floor(n))
	if j := int(math.Ceil(n)); i != j {
		return es.edb[i]/2 + es.edb[j]/2
	}
	return es.edb[i]
}

func (es *elapsedStats) rank(val time.Duration) int {
	return sort.Search(len(es.edb), func(i int) bool {
		return val < es.edb[i]
	})
}

func mergeSortedDurations(a, b []time.Duration) []time.Duration {
	edb := make([]time.Duration, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] <= b[j] {
			edb = append(edb, a[i])
			i++
		} else {
			edb = append(edb, b[j])
			j++
		}
	}
	for ; i < len(a); i++ {
		edb = append(edb, a[i])
	}
	for ; j < len(b); j++ {
		edb = append(edb, b[j])
	}
	return edb
}
