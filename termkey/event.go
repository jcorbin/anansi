package termkey

import (
	"image"
)

// Event represents a keyboard or mouse event.
type Event struct {
	Mod Modifier // one of Mod* constants or 0
	Key Key      // one of Key* constants, invalid if 'Ch' is not 0
	Ch  rune     // a unicode character

	// TODO terminal.Point
	image.Point // if Key is one of Mouse*
}
