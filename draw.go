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
func DrawBitmap(dst Grid, src *Bitmap, styles ...Style) {
	style := Styles(styles...)
	for gp, bp := dst.Rect.Min, src.Rect.Min; bp.Y < src.Rect.Max.Y; bp.Y += 4 {
		gp.X, bp.X = dst.Rect.Min.X, src.Rect.Min.X
		for gp.X < dst.Rect.Max.X && bp.X < src.Rect.Max.X {
			if i, ok := dst.CellOffset(gp); ok {
				r, a := src.Rune(bp), dst.Attr[i]
				sp := ansi.PtFromImage(bp)
				pr, pa := dst.Rune[i], dst.Attr[i]
				if r, a = style.Style(sp, pr, r, pa, a); r != 0 {
					dst.Rune[i], dst.Attr[i] = r, a
				}
			}
			gp.X++
			bp.X += 2
		}
		gp.Y++
	}
}
