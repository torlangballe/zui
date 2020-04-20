package zui

import (
	"math"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /21/10/15.
// Check out: https://github.com/tdewolff/canvas
type Canvas struct {
	canvasNative
	currentMatrix zgeo.Matrix // is currentTransform...
}

func (c *Canvas) DrawImage(image *Image, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) {
	// fmt.Println("Canvas.DrawImage", image.Size(), sourceRect, image.path, destRect, image.GetCapInsets())
	if image != nil {
		if image.GetCapInsets().IsNull() {
			if sourceRect.IsNull() {
				sourceRect = zgeo.Rect{Size: image.Size()}
			}
			c.drawPlainImage(image, destRect, opacity, sourceRect)
		} else {
			c.drawInsetImage(image, image.GetCapInsets(), destRect, opacity)
		}
	}
}

func (c *Canvas) drawInsetRow(image *Image, inset, dest zgeo.Rect, sy, sh, dy, dh float64, opacity float32) {
	size := image.Size()
	ds := dest.Size
	zlog.ErrorIf(ds.W < -inset.Size.W, ds.W, -inset.Size.W, image.path)
	zlog.ErrorIf(ds.H < -inset.Size.H, ds.H, -inset.Size.H, image.path, inset)

	insetMid := size.Minus(inset.Size.Negative())
	c.drawPlainImage(image, zgeo.RectFromXYWH(0, dy, inset.Pos.X, dh), opacity, zgeo.RectFromXYWH(0, sy, inset.Pos.X, sh))
	midMaxX := math.Floor(dest.Max().X + inset.Max().X) // inset.Max is negative
	// fmt.Println("drawInsetRow:", size)
	c.drawPlainImage(image, zgeo.RectFromXYWH(inset.Pos.X, dy, math.Ceil(midMaxX-inset.Pos.X), dh), opacity, zgeo.RectFromXYWH(inset.Pos.X, sy, insetMid.W, sh))
	c.drawPlainImage(image, zgeo.RectFromXYWH(midMaxX, dy, -inset.Max().X, dh), opacity, zgeo.RectFromXYWH(size.W+inset.Max().X, sy, -inset.Max().X, sh))
}

func (c *Canvas) drawInsetImage(image *Image, inset, dest zgeo.Rect, opacity float32) {
	size := image.Size()
	insetMid := size.Minus(inset.Size.Negative())
	diff := dest.Size.Minus(size).Plus(insetMid)
	// fmt.Println("drawInsetImage:", dest, size, insetMid, diff)
	c.drawInsetRow(image, inset, dest, 0, inset.Pos.Y, dest.Min().Y, inset.Pos.Y, opacity)
	c.drawInsetRow(image, inset, dest, inset.Pos.Y, insetMid.H, dest.Min().Y+inset.Pos.Y, diff.H, opacity)
	c.drawInsetRow(image, inset, dest, size.H+inset.Max().Y, -inset.Max().Y, dest.Max().Y+inset.Max().Y, -inset.Max().Y, opacity)
}

func canvasCreateGradientLocations(colors int) []float64 {
	locations := make([]float64, colors, colors)
	last := colors - 1
	for i := 0; i <= last; i++ {
		locations[i] = float64(i) / float64(last)
	}
	return locations
}
