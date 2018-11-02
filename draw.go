package anansi

import "github.com/jcorbin/anansi/ansi"

// DrawGrid copies the source grid's cells into the destination grid, applying
// any optional styles to each cell.
//
// The default is an opaque copy: each source cell simply overwrites each
// corresponding destination cell.
//
// A (partially) transparent draw may be done by providing one or more style
// options.
//
// Use sub-grids to copy to/from specific regions; see Grid.SubRect.
func DrawGrid(dst, src Grid, styles ...Style) {
	style := Styles(styles...)
	if style == NoopStyle {
		copyGrid(dst, src)
		return
	}
	for dp, sp, di, si := copySetup(dst, src); sp.Y < src.Rect.Max.Y && dp.Y < dst.Rect.Max.Y; {
		sii, dii := si, di
		sp.X = src.Rect.Min.X
		dp.X = dst.Rect.Min.X
		for sp.X < src.Rect.Max.X && dp.X < dst.Rect.Max.X {
			dr, da := dst.Rune[dii], dst.Attr[dii]
			sr, sa := src.Rune[sii], src.Attr[sii]
			if sr, sa = style.Style(dp, dr, sr, da, sa); sr != 0 {
				dst.Rune[dii], dst.Attr[dii] = sr, sa
			}
			sii++
			dii++
			sp.X++
		}
		si += src.Stride
		di += dst.Stride
		dp.Y++
	}
}

func copyGrid(dst, src Grid) {
	stride := src.Rect.Dx()
	if dstride := dst.Rect.Dx(); stride > dstride {
		stride = dstride
	}
	for dp, sp, di, si := copySetup(dst, src); sp.Y < src.Rect.Max.Y && dp.Y < dst.Rect.Max.Y; {
		copy(dst.Rune[di:di+stride], src.Rune[si:si+stride])
		sp.Y++
		dp.Y++
		si += src.Stride
		di += dst.Stride
	}
	for dp, sp, di, si := copySetup(dst, src); sp.Y < src.Rect.Max.Y && dp.Y < dst.Rect.Max.Y; {
		copy(dst.Attr[di:di+stride], src.Attr[si:si+stride])
		sp.Y++
		dp.Y++
		si += src.Stride
		di += dst.Stride
	}
}

func copySetup(dst, src Grid) (dp, sp ansi.Point, di, si int) {
	dp, sp = dst.Rect.Min, src.Rect.Min
	di, _ = dst.CellOffset(dp)
	si, _ = src.CellOffset(sp)
	return dp, sp, di, si
}

// DrawBitmap draw's a bitmap's braille runes into the destination grid.
//
// Optional rendering styles may be passed to control the graphical rendition
// and transparency of the braille runes. The styles are passed any prior grid
// attributes for each target cell.
//
// One particularly useful style to use is ElideStyle(0x2800), which will map
// any empty braille runes to the zero rune, causing only non-empty braille
// runes to be drawn.
//
// Use sub-grids to target specific regions; see Grid.SubRect.
func DrawBitmap(dst Grid, src Bitmap, styles ...Style) {
	style := Styles(styles...)

	// Only likely to fail for empty (sub-)Grid
	di, withinDst := dst.CellOffset(dst.Rect.Min)
	if !withinDst {
		return
	}

	// Only likely to fail for empty (sub-)Bitmap
	si, withinSrc := src.index(src.Rect.Min)
	if !withinSrc {
		return
	}

	// Loop counter maxes used by all 3 phases below
	ddx, sdx := dst.Rect.Dx(), src.Rect.Dx()
	if sdx > ddx*2 {
		sdx = ddx * 2
	} else if tdx := (sdx + 1) / 2; ddx > tdx {
		ddx = tdx
	}
	ddy, sdy := dst.Rect.Dy(), src.Rect.Dy()
	if sdy > ddy*4 {
		sdy = ddy * 4
	} else if tdy := (sdy + 3) / 4; ddy > tdy {
		ddy = tdy
	}

	oddSDX := sdx%2 == 1
	if oddSDX {
		sdx--
	}

	// Clear grid by filling with empty braille runes (U+2800)
	for ky := 0; ky < ddy; ky++ {
		for kx := 0; kx < ddx; kx++ {
			dst.Rune[di] = 0x2800
			di++ // next grid column
		}
		di -= ddx        // return to start of grid row
		di += dst.Stride // next grid row
	}
	di -= dst.Stride * ddy // return to top of grid

	// Render runes into the grid; essentially a pivoted copy of Bitmap.Rune
	bits := []struct{ left, right rune }{
		{0x0001, 0x0008}, // bitmap first row
		{0x0002, 0x0010}, // bitmap second row
		{0x0004, 0x0020}, // bitmap third row
		{0x0040, 0x0080}, // bitmap fourth row
	}

	// for each grid row
	for ky := 0; ky < ddy; ky++ {
		// for each bitmap row that maps to it
		for _, bits := range bits {
			// for each pair of bitmap columns...
			for kx := 0; kx < sdx; {

				if src.Bit[si] {
					dst.Rune[di] |= bits.left
				}
				si++ // next bitmap column
				kx++

				if src.Bit[si] {
					dst.Rune[di] |= bits.right
				}
				si++ // next bitmap column
				kx++

				di++ // next grid column
			}
			// ...may have a final odd bitmap column
			if oddSDX {
				if src.Bit[si] {
					dst.Rune[di] |= bits.left
				}
				// si++ NOTE would just need to be immediately decremented back
			}
			si -= sdx // return to start of bitmap row
			di -= ddx // return to start of grid row

			si += src.Stride // next bitmap row
		}

		di += dst.Stride // next grid row
	}
	si -= src.Stride * sdy // return to top of bitmap
	di -= dst.Stride * ddy // return to top of grid

	// Apply any styles
	for gp, ky := dst.Rect.Min, 0; ky < ddy; ky++ {
		gp.X = dst.Rect.Min.X
		for kx := 0; kx < ddx; kx++ {
			pr, pa := dst.Rune[di], dst.Attr[di]
			if r, a := style.Style(gp, pr, pr, pa, pa); r != 0 {
				dst.Rune[di], dst.Attr[di] = r, a
			}
			gp.X++
			di++ // next grid column
		}
		gp.Y++
		di -= ddx        // return to start of grid row
		di += dst.Stride // next grid row
	}
}
