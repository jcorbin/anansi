package platform

import (
	"image"
)

func NewTest(size image.Point, client Client) *Platform {
	p := Platform{client: client}
	p.screen.Resize(size)
	p.State.LastSize = size
	return &p
}
