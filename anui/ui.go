package anui

import (
	"time"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

// BannerLayer provides a header banner overlay.
type BannerLayer struct {
	banner    []byte
	needsDraw time.Duration
}

var _ Layer = (*BannerLayer)(nil)

// Say sets the message string.
func (ban *BannerLayer) Say(mess string) {
	ban.banner = []byte(mess)
	ban.needsDraw = time.Millisecond
}

// HandleInput is a no-op.
func (ban *BannerLayer) HandleInput(e ansi.Escape, a []byte) (bool, error) {
	return false, nil
}

// Draw the banner overlay.
func (ban *BannerLayer) Draw(sc anansi.Screen, now time.Time) anansi.Screen {
	ban.needsDraw = 0
	at := sc.Grid.Rect.Min
	bannerWidth := MeasureText(ban.banner).Dx()
	screenWidth := sc.Bounds().Dx()
	at.X += screenWidth/2 - bannerWidth/2
	bsc := sc.SubAt(at)
	anansi.Process(&bsc, ban.banner) // TODO pre-process into an internal Grid for composition
	return sc
}

// NeedsDraw returns non-zero if the layer needs to be drawn (if the mesage
// has changed since last draw).
func (ban *BannerLayer) NeedsDraw() time.Duration {
	return ban.needsDraw
}
