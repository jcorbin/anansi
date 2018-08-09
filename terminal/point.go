package terminal

import "image"

// Point on a terminal is relative to a image.Pt(1, 1) origin not image.ZP.
type Point struct{ image.Point }

// Pt constructs a new terminal point with supplied X and Y values already in
// terminal <1,1>-origin space.
func Pt(x, y int) Point { return Point{image.Pt(x, y)} }

// Nowhere represents an unknown position on the terminal screen; it is the
// zero value of Point so that, like with Visibility, it's easy to default to
// unknown state.
//
// NOTE since, like any other screen space, negative points are out of bounds,
// any point component <1 should be considered invalid/unusable. So while users
// may pass Nowhere around to have a useful pun for the zero value, consumers
// should be more accepting in their logic and avoid using it for comparisons.
var Nowhere = Point{}

// ImageToTermPoint translates a point into terminal <1,1>-origin space.
func ImageToTermPoint(pt image.Point) Point {
	pt.X++
	pt.Y++
	return Point{pt}
}

// TermToImagePoint translates a point out of terminal <1,1>-origin space.
func TermToImagePoint(pt Point) image.Point {
	pt.X--
	pt.Y--
	return pt.Point
}

// TODO use in internal/termkey so that parsed events carry a terminal point
// rather than an image point, avoiding more janky off-by-one math in the
// decode path.
