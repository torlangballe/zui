package zcanvas

import (
	"math"
	"sync"

	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /21/10/15.
// Check out: https://github.com/tdewolff/canvas
type Canvas struct {
	size zgeo.Size
	canvasNative
	currentMatrix    zgeo.Matrix // is currentTransform...
	DownsampleImages bool
}

func (c *Canvas) Size() zgeo.Size {
	return c.size
}

func (c *Canvas) DrawImageAt(image *zimage.Image, pos zgeo.Pos, useDownsampleCache bool, opacity float32) {
	if image == nil {
		return
	}
	s := image.Size()
	sr := zgeo.Rect{Size: s}
	dr := sr
	dr.Pos = pos
	c.DrawImage(image, useDownsampleCache, dr, opacity, sr)
}

func (c *Canvas) DrawImage(image *zimage.Image, useDownsampleCache bool, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) bool {
	if image == nil {
		return true
	}
	// if strings.Contains(image.Path, "auth") {
	// zlog.Info("Canvas.DrawImage:", image.Size(), destRect, image.Path)
	// }
	if sourceRect.IsNull() {
		sourceRect = zgeo.Rect{Size: image.Size()}
	}
	if image.CapInsets().IsNull() {
		if sourceRect.IsNull() {
			sourceRect = zgeo.Rect{Size: image.Size()}
		}
		return c.drawPlainImage(image, useDownsampleCache, destRect, opacity, sourceRect)
	}
	c.drawInsetImage(image, image.CapInsets(), destRect, opacity)
	return true
	// zlog.Info("Canvas.DrawImage Done", image.Path)
}

func (c *Canvas) drawInsetRow(image *zimage.Image, inset, dest zgeo.Rect, sy, sh, dy, dh float64, opacity float32) {
	size := image.Size()
	ds := dest.Size
	zlog.ErrorIf(ds.W < -inset.Size.W, ds.W, -inset.Size.W, image.Path)
	zlog.ErrorIf(ds.H < -inset.Size.H, ds.H, -inset.Size.H, image.Size(), image.Path, inset, "ds:", ds, "dest:", dest)

	useDownsampleCache := false
	insetMid := size.Minus(inset.Size.Negative())
	c.drawPlainImage(image, useDownsampleCache, zgeo.RectFromXYWH(0, dy, inset.Pos.X, dh), opacity, zgeo.RectFromXYWH(0, sy, inset.Pos.X, sh))
	midMaxX := math.Floor(dest.Max().X + inset.Max().X) // inset.Max is negative
	// zlog.Info("drawInsetRow:", size)
	c.drawPlainImage(image, useDownsampleCache, zgeo.RectFromXYWH(inset.Pos.X, dy, math.Ceil(midMaxX-inset.Pos.X), dh), opacity, zgeo.RectFromXYWH(inset.Pos.X, sy, insetMid.W, sh))
	c.drawPlainImage(image, useDownsampleCache, zgeo.RectFromXYWH(midMaxX, dy, -inset.Max().X, dh), opacity, zgeo.RectFromXYWH(size.W+inset.Max().X, sy, -inset.Max().X, sh))
}

func (c *Canvas) drawInsetImage(image *zimage.Image, inset, dest zgeo.Rect, opacity float32) {
	size := image.Size()
	insetMid := size.Minus(inset.Size.Negative())
	diff := dest.Size.Minus(size).Plus(insetMid)
	// zlog.Info("drawInsetImage:", dest, size, insetMid, diff)
	c.drawInsetRow(image, inset, dest, 0, inset.Pos.Y, dest.Min().Y, inset.Pos.Y, opacity)
	c.drawInsetRow(image, inset, dest, inset.Pos.Y, insetMid.H, dest.Min().Y+inset.Pos.Y, diff.H, opacity)
	c.drawInsetRow(image, inset, dest, size.H+inset.Max().Y, -inset.Max().Y, dest.Max().Y+inset.Max().Y, -inset.Max().Y, opacity)
}

func (c *Canvas) FillRect(rect zgeo.Rect) {
	path := zgeo.PathNewRect(rect, zgeo.Size{})
	c.FillPath(path)
}

func (c *Canvas) Fill() {
	rect := zgeo.Rect{Size: c.size}
	path := zgeo.PathNewRect(rect, zgeo.Size{})
	c.FillPath(path)
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
type measurement struct {
	Font zgeo.Font
	Text string
}

var measuredTexts = map[measurement]zgeo.Size{}
var measureTextMutex sync.Mutex
var measureCanvas *Canvas

func GetTextSize(text string, font *zgeo.Font) zgeo.Size {
	measureTextMutex.Lock()
	m := measurement{Font: *font, Text: text}
	s, got := measuredTexts[m]
	measureTextMutex.Unlock()
	if got {
		// zlog.Info("canvas get Text size, using cache:", text)
		return s
	}
	// zlog.Info("canvas measure text")
	if measureCanvas == nil {
		measureCanvas = New()
		measureCanvas.SetSize(zgeo.Size{800, 100})
	}
	s = measureCanvas.MeasureText(text, font)
	measureTextMutex.Lock()
	// fmt.Println("canvas measure lock time 2:", time.Since(start), len(measureTextCanvases))
	measuredTexts[m] = s
	measureTextMutex.Unlock()
	return s
}

func (c *Canvas) StrokeHorizontal(x1, x2, y float64, width float64, ltype zgeo.PathLineType) {
	path := zgeo.PathNew()
	path.MoveTo(zgeo.Pos{x1, y})
	path.LineTo(zgeo.Pos{x2, y})
	c.StrokePath(path, width, ltype)
}

func (c *Canvas) StrokeVertical(x, y1, y2 float64, width float64, ltype zgeo.PathLineType) {
	path := zgeo.PathNew()
	path.MoveTo(zgeo.Pos{x, y1})
	path.LineTo(zgeo.Pos{x, y2})
	c.StrokePath(path, width, ltype)
}

func (c *Canvas) DrawRectGradientVertical(rect zgeo.Rect, col1, col2 zgeo.Color) {
	colors := []zgeo.Color{col1, col2}
	path := zgeo.PathNewRect(rect, zgeo.Size{})
	c.DrawGradient(path, colors, rect.Min(), rect.BottomLeft(), nil)
}

func MergeImages(box zgeo.Size, images []*zimage.ImageGetter, done func(img *zimage.Image)) {
	zimage.GetImages(images, func(all bool) {
		if !all {
			zlog.Error(nil, "Not all images got")
			return
		}
		if box.IsNull() {
			for _, ig := range images {
				box.Maximize(ig.Image.Size())
			}
		}
		canvas := New()
		canvas.SetSize(box)
		for _, ig := range images {
			r := zgeo.Rect{Size: box}.Align(ig.Image.Size(), ig.Alignment, ig.Margin)
			canvas.DrawImageAt(ig.Image, r.Pos, false, ig.Opacity)
		}
		canvas.ZImage(false, done)
	})
}
