package zgo

import (
	"math"
)

//  Created by Tor Langballe on /21/10/15.

type Canvas struct {
	canvasNative
	currentMatrix Matrix // is currentTransform...
}

type CanvasBlendMode int

const (
	CanvasBlendModeNormal CanvasBlendMode = iota
	CanvasBlendModeMultiply
	CanvasBlendModeScreen
	CanvasBlendModeOverlay
	CanvasBlendModeDarken
	CanvasBlendModeLighten
	CanvasBlendModeColorDodge
	CanvasBlendModeColorBurn
	CanvasBlendModeSoftLight
	CanvasBlendModeHardLight
	CanvasBlendModeDifference
	CanvasBlendModeExclusion
	CanvasBlendModeHue
	CanvasBlendModeSaturation
	CanvasBlendModeColor
	CanvasBlendModeLuminosity
)

func (c *Canvas) DrawImage(image *Image, destRect Rect, opacity float32, blendMode CanvasBlendMode, sourceRect Rect) {
	//	fmt.Println("Canvas.DrawImage", sourceRect, image.path, destRect, image.Size())
	if image != nil {
		if image.GetCapInsets().IsNull() {
			if sourceRect.IsNull() {
				sourceRect = Rect{Size: image.Size()}
			}
			c.drawPlainImage(image, destRect, opacity, blendMode, sourceRect)
		} else {
			c.drawInsetImage(image, image.GetCapInsets(), destRect, opacity, blendMode)
		}
	}
}

func (c *Canvas) drawInsetRow(image *Image, inset, dest Rect, sy, sh, dy, dh float64, opacity float32, blendMode CanvasBlendMode) {
	size := image.Size()
	// diff := dest.Size.Minus(size)
	insetMid := size.Minus(inset.Size.Negative())
	c.drawPlainImage(image, RectFromXYWH(0, dy, inset.Pos.X, dh), opacity, blendMode, RectFromXYWH(0, sy, inset.Pos.X, sh))
	midMaxX := math.Floor(dest.Max().X + inset.Max().X) // inset.Max is negative
	// fmt.Println("drawInsetRow:", RectFromXYWH(inset.Pos.X, dy, math.Ceil(midMaxX-inset.Pos.X), dh))
	c.drawPlainImage(image, RectFromXYWH(inset.Pos.X, dy, math.Ceil(midMaxX-inset.Pos.X), dh), opacity, blendMode, RectFromXYWH(inset.Pos.X, sy, insetMid.W, sh))
	c.drawPlainImage(image, RectFromXYWH(midMaxX, dy, -inset.Max().X, dh), opacity, blendMode, RectFromXYWH(size.W+inset.Max().X, sy, -inset.Max().X, sh))
}

func (c *Canvas) drawInsetImage(image *Image, inset, dest Rect, opacity float32, blendMode CanvasBlendMode) {
	size := image.Size()
	insetMid := size.Minus(inset.Size.Negative())
	diff := dest.Size.Minus(size).Plus(insetMid)
	// fmt.Println("drawInsetImage:", dest, size, insetMid, diff)
	c.drawInsetRow(image, inset, dest, 0, inset.Pos.Y, dest.Min().Y, inset.Pos.Y, opacity, blendMode)
	c.drawInsetRow(image, inset, dest, inset.Pos.Y, insetMid.H, dest.Min().Y+inset.Pos.Y, diff.H, opacity, blendMode)
	c.drawInsetRow(image, inset, dest, size.H+inset.Max().Y, -inset.Max().Y, dest.Max().Y+inset.Max().Y, -inset.Max().Y, opacity, blendMode)
}
