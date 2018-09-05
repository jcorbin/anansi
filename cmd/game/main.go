package main

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	"github.com/jcorbin/anansi/x/platform"
)

var errInt = errors.New("interrupt")

func main() {
	platform.MustRun(os.Stdout, Run, platform.FrameRate(60))
}

// Run the game under an active terminal platform.
func Run(p *platform.Platform) error {
	for {
		var g game

		log.Printf("running")

		if err := p.Run(&g); platform.IsReplayDone(err) {
			continue // loop replay
		} else if err == io.EOF || err == errInt {
			return nil
		} else if err != nil {
			log.Printf("exiting due to %v", err)
			return err
		}
	}
}

type game struct {
	Grid anansi.Grid

	Off float32Point
}

type float32Point struct {
	X, Y float32
}

func fpt(x, y float32) float32Point { return float32Point{x, y} }
func (pt float32Point) Add(other float32Point) float32Point {
	pt.X += other.X
	pt.Y += other.Y
	return pt
}

func (g *game) Update(ctx *platform.Context) (err error) {
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

	g.Grid.Resize(ctx.Output.Size)

	// rate := fpt(
	// 	float32(time.Second/60)/float32(time.Millisecond),
	// 	float32(time.Second/30)/float32(time.Millisecond),
	// )

	// dt := float32(ctx.Time.Sub(ctx.LastTime)) / float32(time.Millisecond)
	// g.Off = g.Off.Add(fpt(dt/rate.X, dt/rate.Y))
	// log.Printf("dt:%v off:%v", dt, g.Off)

	dx, dy := int(g.Off.X), int(g.Off.Y)

	for i, y := 0, 1; y <= ctx.Output.Size.Y; y++ {
		for x := 0; x < ctx.Output.Size.X; x++ {
			a := ansi.RGB(0, uint8((x+dx)*8), uint8((y+dy)*16)).BG()
			g.Grid.Attr[i] = a
			g.Grid.Rune[i] = 0
			i++
		}
	}

	// TODO Grid method
	// model := ansi.ColorModelID
	// at := image.Pt(1, 1)
	def := ' '
	x, y := 1, 1
	for i, r := range g.Grid.Rune {
		if r == 0 {
			r = def
		}
		ctx.Output.Cell(ansi.Pt(x, y)).Set(r, g.Grid.Attr[i])
		if x++; x > g.Grid.Size.X {
			x = 1
			y++
		}
	}

	return err
}
