package anui

import (
	"image"
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// ViewClient is run under a ViewLayer to Render() some viewport area within
// its Bounds().
type ViewClient interface {
	// Bounds returns the bounding box of the client space.
	Bounds() image.Rectangle

	// Render a viewport from client space into the given grid.
	Render(g anansi.Grid, viewport image.Rectangle)
}

// ViewLayer manages a movable viewport within a renderable view client,
// centered around a focus point.
type ViewLayer struct {
	Client    ViewClient
	needsDraw time.Duration
	focus     image.Point
	offset    image.Point
}

// HandleInput processes arrow keys to move the view focus point.
func (view *ViewLayer) HandleInput(e ansi.Escape, a []byte) (handled bool, err error) {
	switch e {

	// arrow keys to move view
	case ansi.CUB, ansi.CUF, ansi.CUU, ansi.CUD:
		if d, ok := ansi.DecodeCursorCardinal(e, a); ok {
			p := view.focus.Add(d)
			if bounds := view.Client.Bounds(); !bounds.Empty() {
				if p.X < bounds.Min.X {
					p.X = bounds.Min.X
				}
				if p.Y < bounds.Min.Y {
					p.Y = bounds.Min.Y
				}
				if p.X >= bounds.Max.X {
					p.X = bounds.Max.X - 1
				}
				if p.Y >= bounds.Max.Y {
					p.Y = bounds.Max.Y - 1
				}
			}
			if view.focus != p {
				view.focus = p
				view.needsDraw = time.Millisecond
			}
		}
		return true, nil

	}
	return false, nil
}

// NeedsDraw returns non-zero if the layer needs to be drawn.
func (view *ViewLayer) NeedsDraw() time.Duration {
	return view.needsDraw
}

// Draw a screen-sized viewport of the view client's content the focus point
// into center of the screen.
func (view *ViewLayer) Draw(screen anansi.Screen, now time.Time) {
	screenSize := screen.Bounds().Size()
	screenLoMid := screenSize.Div(2)
	screenHiMid := screenSize.Add(image.Pt(1, 1)).Div(2)
	view.offset = screenLoMid.Sub(view.focus)
	view.needsDraw = 0
	worldView := image.Rectangle{
		view.focus.Sub(screenLoMid),
		view.focus.Add(screenHiMid),
	}
	bnd := view.Client.Bounds()
	if !bnd.Empty() {
		worldView = worldView.Intersect(bnd)
	}
	subGrid := screen.Grid.
		SubAt(ansi.PtFromImage(worldView.Min.Add(view.offset))).
		SubSize(worldView.Size())
	view.Client.Render(subGrid, worldView)
}

// Offset returns the current view offset (as of the last draw).
func (view *ViewLayer) Offset() image.Point {
	return view.offset
}

// SetFocus sets the view center point used to determine offset when Draw-ing.
func (view *ViewLayer) SetFocus(p image.Point) {
	view.focus = p
	view.needsDraw = time.Millisecond
}
