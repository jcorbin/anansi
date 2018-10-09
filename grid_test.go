package anansi_test

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
	anansitest "github.com/jcorbin/anansi/test"
)

func TestGrid_CopyIntoAt(t *testing.T) {
	over := Grid{
		Size: image.Pt(5, 2),
		Rune: []rune{
			'h', 'e', 'l', 'l', 'o',
			'w', 'o', 'r', 'l', 'd',
		},
		Attr: make([]ansi.SGRAttr, 10),
	}

	under := Grid{
		Size: image.Pt(10, 6),
		Rune: make([]rune, 60),
		Attr: make([]ansi.SGRAttr, 60),
	}

	over.CopyIntoAt(&under, image.Pt(3, 3))

	lines := anansitest.GridLines(under, '.')
	assert.Equal(t, []string{
		"..........",
		"..........",
		"..hello...",
		"..world...",
		"..........",
		"..........",
	}, lines)
}
