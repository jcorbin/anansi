package main

import (
	"errors"
	"math"
	"os"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/x/platform"
)

type schotterDemoUI struct {
	*schotterDemo
}

func runInteractive() {
	platform.MustRun(os.Stdin, os.Stdout, func(p *platform.Platform) error {
		var ui schotterDemoUI
		ui.schotterDemo = &sd
		ui.squareSide = 20 // TODO push down, pre-compute based on initial width and squaresPerRow

		return p.Run(&ui)
	}, platform.FrameRate(60))
}

func (sd *schotterDemoUI) Update(ctx *platform.Context) (err error) {
	// Ctrl-C interrupts
	if ctx.Input.HasTerminal('\x03') {
		// ... AFTER any other available input has been processed
		err = errors.New("interrupt")
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

	zoomed := false
	if n := ctx.Input.TotalScrollIn(ctx.Output.Bounds()); n != 0 {
		sd.squareSide += n
		if sd.squareSide < 1 {
			sd.squareSide = 1
		}
		zoomed = true
	}

	canvasSize := sd.canvas.Rect.Size()

	if screenSize := ctx.Output.Bounds().Size(); screenSize.X != canvasSize.X/2 || zoomed {
		sd.padding = 0
		if screenSize.X > 2 {
			sd.padding = 2
		}

		canvasSize.X = screenSize.X * 2
		canvasSize.Y = screenSize.Y * 4

		roundUp := sd.squareSide - 1
		sd.squaresPerRow = ((screenSize.X-sd.padding)*2 + roundUp) / sd.squareSide
		sd.squaresPerCol = ((screenSize.Y-sd.padding)*4 + roundUp) / sd.squareSide

		// TODO resize if != nil
		sd.canvas.Resize(canvasSize)
	}

	for i := range sd.canvas.Bit {
		sd.canvas.Bit[i] = false
	}
	sd.draw()

	ctx.Output.Clear()
	anansi.DrawBitmap(ctx.Output.Grid, sd.canvas)

	sd.angleOffset += 0.01
	if sd.angleOffset > math.Pi {
		sd.angleOffset -= math.Pi
	}

	return err
}
