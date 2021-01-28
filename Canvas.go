package zui

import (
	"math"
	"sync"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /21/10/15.
// Check out: https://github.com/tdewolff/canvas
type Canvas struct {
	canvasNative
	currentMatrix zgeo.Matrix // is currentTransform...
	DownsampleImages  bool
}

func (c *Canvas) DrawImage(image *Image, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) {
	if image != nil {
		if image.CapInsets().IsNull() {
			if sourceRect.IsNull() {
				sourceRect = zgeo.Rect{Size: image.Size()}
			}
			c.drawPlainImage(image, destRect, opacity, sourceRect)
		} else {
			// zlog.Info("Canvas.DrawImage", image.Size(), sourceRect, image.Path, destRect, image.SetCapInsets())
			c.drawInsetImage(image, image.CapInsets(), destRect, opacity)
			// zlog.Info("Canvas.DrawImage Done", image.Path)
		}
	}
}

func (c *Canvas) drawInsetRow(image *Image, inset, dest zgeo.Rect, sy, sh, dy, dh float64, opacity float32) {
	size := image.Size()
	ds := dest.Size
	zlog.ErrorIf(ds.W < -inset.Size.W, ds.W, -inset.Size.W, image.Path)
	zlog.ErrorIf(ds.H < -inset.Size.H, ds.H, -inset.Size.H, image.Size(), image.Path, inset, "ds:", ds, "dest:", dest)

	insetMid := size.Minus(inset.Size.Negative())
	c.drawPlainImage(image, zgeo.RectFromXYWH(0, dy, inset.Pos.X, dh), opacity, zgeo.RectFromXYWH(0, sy, inset.Pos.X, sh))
	midMaxX := math.Floor(dest.Max().X + inset.Max().X) // inset.Max is negative
	// zlog.Info("drawInsetRow:", size)
	c.drawPlainImage(image, zgeo.RectFromXYWH(inset.Pos.X, dy, math.Ceil(midMaxX-inset.Pos.X), dh), opacity, zgeo.RectFromXYWH(inset.Pos.X, sy, insetMid.W, sh))
	c.drawPlainImage(image, zgeo.RectFromXYWH(midMaxX, dy, -inset.Max().X, dh), opacity, zgeo.RectFromXYWH(size.W+inset.Max().X, sy, -inset.Max().X, sh))
}

func (c *Canvas) drawInsetImage(image *Image, inset, dest zgeo.Rect, opacity float32) {
	size := image.Size()
	insetMid := size.Minus(inset.Size.Negative())
	diff := dest.Size.Minus(size).Plus(insetMid)
	// zlog.Info("drawInsetImage:", dest, size, insetMid, diff)
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

// measureTextCanvases is a pool of canvases to do text measurements in. Might not actually speed things up in DOM, which maybe is single-thread
var measureTextCanvases = map[*Canvas]bool{}
var measureTextMutex sync.Mutex

func canvasGetTextSize(text string, font *Font) zgeo.Size {
	var canvas *Canvas
	measureTextMutex.Lock() 
	// fmt.Println("canvas measure lock time:", time.Since(start), len(measureTextCanvases))
	for c, used := range measureTextCanvases {
		if !used {
			measureTextCanvases[c] = true
			canvas = c
			break
		}
	}
	measureTextMutex.Unlock()
	if canvas == nil {
		canvas = CanvasNew()
		canvas.SetSize(zgeo.Size{500, 100})
	}
	s := canvas.MeasureText(text, font)
	measureTextMutex.Lock()
	// fmt.Println("canvas measure lock time 2:", time.Since(start), len(measureTextCanvases))
	measureTextCanvases[canvas] = false
	measureTextMutex.Unlock()
	return s
}
