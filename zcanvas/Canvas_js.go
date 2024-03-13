package zcanvas

import (
	"strings"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zcache"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"

	"fmt"
	"image"
	"syscall/js"
)

// interesting: https://github.com/markfarnan/go-canvas

func init() {
	zimage.DrawInCanvasFunc = RenderToImage
}

type canvasNative struct {
	element js.Value
	context js.Value
}

var idCount int

func New() *Canvas {
	c := Canvas{}
	c.element = zdom.DocumentJS.Call("createElement", "canvas")
	c.element.Set("style", "position:absolute;pointer-events:none") // pointer-events:none makes canvas not be target when pressed
	c.context = c.element.Call("getContext", "2d")
	c.element.Set("id", fmt.Sprintf("canvas-%d", idCount))
	idCount++

	// c.context.Set("imageSmoothingEnabled", true)
	// c.context.Set("imageSmoothingQuality", "high")
	return &c
}

func (c *Canvas) JSContext() js.Value {
	return c.context
}

func (c *Canvas) JSElement() js.Value {
	return c.element
}

func (c *Canvas) SetSize(size zgeo.Size) {
	c.element.Set("width", size.W) // scale?
	c.element.Set("height", size.H)
	c.size = size
}

func (c *Canvas) Element() js.Value {
	return c.element
}

func setElementRect(e js.Value, rect zgeo.Rect) {
	style := e.Get("style")
	style.Set("left", fmt.Sprintf("%fpx", rect.Pos.X))
	style.Set("top", fmt.Sprintf("%fpx", rect.Pos.Y))
	style.Set("width", fmt.Sprintf("%fpx", rect.Size.W))
	style.Set("height", fmt.Sprintf("%fpx", rect.Size.H))
}

func (c *Canvas) SetRect(rect zgeo.Rect) {
	setElementRect(c.element, rect)
}

func (c *Canvas) setColor(color zgeo.Color, stroke bool) {
	// var vcolor = color
	str := color.Hex()
	name := "fillStyle"
	if stroke {
		name = "strokeStyle"
	}
	c.context.Set(name, str)
}

func (c *Canvas) SetColor(color zgeo.Color) {
	c.setColor(color, false)
	c.setColor(color, true)
}

// func (c *Canvas) SetTile(image *zimage.Image, hor, vert bool) {
// 	rep := "repeat"
// 	if !hor {
// 		rep += "-y"
// 	} else if !vert {
// 		rep += "-x"
// 	}
// 	pattern := c.context.Call("createPattern", image.ImageJS, rep)
// 	c.context.Set("fillStyle", pattern)
// }

func (c *Canvas) FillPath(path *zgeo.Path) {
	c.setPath(path)
	c.context.Call("fill")
}

func (c *Canvas) FillPathEO(path *zgeo.Path) {
	c.setPath(path)
	c.context.Call("fill", "evenodd")
}

func (c *Canvas) SetFont(font *zgeo.Font, matrix *zgeo.Matrix) error {
	str := zdom.GetFontStyle(font)
	// zlog.Info("canvas set font:", str)
	c.context.Set("font", str)
	return nil
}

func (c *Canvas) SetMatrix(matrix zgeo.Matrix) {
	c.currentMatrix = matrix
	c.setMatrix(matrix)
}

func (c *Canvas) Transform(matrix zgeo.Matrix) {
	c.currentMatrix = c.currentMatrix.Multiplied(matrix)
}

func (c *Canvas) ClipPath(path *zgeo.Path, eofill bool) {
	c.setPath(path)
	c.context.Call("clip")
}

func (c *Canvas) GetClipRect() zgeo.Rect {
	return zgeo.Rect{}
	//        return SetRect(context.boundingBoxOfClipPath)
}

func (c *Canvas) StrokePath(path *zgeo.Path, width float64, ltype zgeo.PathLineType) {
	c.setPath(path)
	c.setLineType(ltype)
	c.setLineWidth(width)

	var array []interface{}
	for _, d := range path.Dashes {
		array = append(array, d)
	}
	c.context.Call("setLineDash", array)
	c.context.Call("stroke")
}

func (c *Canvas) DrawPath(path *zgeo.Path, strokeColor zgeo.Color, width float64, ltype zgeo.PathLineType, eofill bool) {
	c.setPath(path)
	c.context.Call("fill")
	c.PushState()
	c.setColor(strokeColor, true)
	c.setLineType(ltype)
	c.context.Call("stroke")
	c.PopState()
}

// TODO: Use zcache?
type scaledImage struct {
	path string
	size zgeo.Size
}

var scaledImageMap = zcache.NewExpiringMap[scaledImage, *zimage.Image](60 * 10)

func (c *Canvas) drawCachedScaledImage(image *zimage.Image, useDownsampleCache bool, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) bool {
	proportional := false
	ds := destRect.Size.Ceil()
	si := scaledImage{image.Path, ds}
	sourceRect = zgeo.Rect{Size: ds}
	var newImage *zimage.Image
	if useDownsampleCache {
		newImage, _ = scaledImageMap.Get(si)
	}
	// if strings.Contains(image.Path, "plus-circled-darkgray.png") {
	// zlog.Info("drawCachedScaledImage:", image.Size(), destRect, image.Path, destRect, newImage != nil)
	// }
	if newImage != nil {
		image = newImage
		c.rawDrawPlainImage(image, destRect, opacity, sourceRect)
		return true
	}
	scale := zscreen.GetMain().Scale
	ds.MultiplyD(scale)
	image.ShrunkInto(ds, proportional, func(shrunkImage *zimage.Image) {
		if shrunkImage == nil {
			return
		}
		shrunkImage.Scale = int(scale)
		if useDownsampleCache {
			if strings.HasPrefix(image.Path, "images/") {
				// zlog.Info("SetForever:", image.Path)
				scaledImageMap.SetForever(si, shrunkImage)
			} else {
				scaledImageMap.Set(si, shrunkImage)
			}
		}
		c.rawDrawPlainImage(shrunkImage, destRect, opacity, sourceRect)
		if !useDownsampleCache {
			shrunkImage.Release()
		}
	})
	return false
}

func (c *Canvas) drawPlainImage(image *zimage.Image, useDownsampleCache bool, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) bool {
	if destRect.Size.H < 0 {
		zlog.Info("drawPlainImage BAD!:", image.Size(), destRect, sourceRect, c)
		return true
	}
	ss := sourceRect.Size
	ds := destRect.Size
	drawnNow := true
	// if strings.Contains(image.Path, "sort-triangle-down") {
	// 	zlog.Info("drawPlainImage:", image.Size(), destRect, image.Path, c.DownsampleImages, ss.Area() < 1000000, ss == image.Size(), sourceRect.Pos.IsNull(), ds.W/ss.W < 0.95, ds.H/ss.H < 0.95)
	// }
	if image.Path != "" && c.DownsampleImages && ss.Area() < 1000000 && ss == image.Size() && sourceRect.Pos.IsNull() && (ds.W/ss.W < 0.95 || ds.H/ss.H < 0.95) {
		drawnNow = c.drawCachedScaledImage(image, useDownsampleCache, destRect, opacity, sourceRect)
		// zlog.Info("drawPlainImage draw-cache:", image.Size(), destRect, image.Path)
		if drawnNow {
			return true
		}
		// if it returns false, it wasn't in cache, so we draw unscaled below
	}
	// zlog.Info("drawPlain:", image.Path, destRect, sourceRect, opacity)
	c.rawDrawPlainImage(image, destRect, opacity, sourceRect)
	return drawnNow
}

func (c *Canvas) rawDrawPlainImage(image *zimage.Image, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) {
	sr := sourceRect.TimesD(float64(image.Scale))

	// zlog.Info("drawRaw:", image.Path, image.Scale, sr, destRect)
	oldAlpha := c.context.Get("globalAlpha").Float()
	if opacity != 1 {
		c.context.Set("globalAlpha", opacity)
	}
	// if strings.Contains(image.Path, "sort-triangle-down") {
	// 	c.SetColor(zgeo.ColorRandom())
	// 	c.FillRect(destRect)
	// 	zlog.Info("rawDrawPlainImage:", image.ImageJS, sr.Pos.X, sr.Pos.Y, sr.Size.W, sr.Size.H, destRect.Pos.X, destRect.Pos.Y, destRect.Size.W, destRect.Size.H)
	// }
	c.context.Call("drawImage", image.ImageJS, sr.Pos.X, sr.Pos.Y, sr.Size.W, sr.Size.H, destRect.Pos.X, destRect.Pos.Y, destRect.Size.W, destRect.Size.H)
	if opacity != 1 {
		c.context.Set("globalAlpha", oldAlpha)
	}
}

func (c *Canvas) PushState() {
	c.context.Call("save")
}

func (c *Canvas) PopState() {
	c.context.Call("restore")
}

func (c *Canvas) Clear() {
	var rect zgeo.Rect
	rect.Size.W = c.Element().Get("width").Float()
	rect.Size.H = c.Element().Get("height").Float()
	c.context.Call("clearRect", rect.Pos.X, rect.Pos.Y, rect.Size.W, rect.Size.H)
}

func (c *Canvas) SetDropShadow(d zstyle.DropShadow) {
	c.context.Set("shadowOffsetX", d.Delta.W)
	c.context.Set("shadowOffsetY", d.Delta.H)
	c.context.Set("shadowBlur", d.Blur)
	c.context.Set("shadowColor", d.Color.Hex())
}

func (c *Canvas) ClearDropShadow() {
	c.SetDropShadow(zstyle.DropShadowClear)
}

func (c *Canvas) DrawGradient(path *zgeo.Path, colors []zgeo.Color, pos1 zgeo.Pos, pos2 zgeo.Pos, locations []float64) {
	// make this a color type instead?? Maybe no point, as it has fixed start/end pos for gradient
	c.PushState()
	c.ClipPath(path, false)
	gradient := c.context.Call("createLinearGradient", pos1.X, pos1.Y, pos2.X, pos2.Y)
	if len(locations) == 0 {
		locations = canvasCreateGradientLocations(len(colors))
	}
	for i, c := range colors {
		gradient.Call("addColorStop", locations[i], c.Hex())
	}
	c.context.Set("fillStyle", gradient)
	c.FillPath(path)
	c.PopState()
}

func (c *Canvas) DrawRadialGradient(path *zgeo.Path, colors []zgeo.Color, center zgeo.Pos, radius float64, endCenter *zgeo.Pos, startRadius float64, locations []float32) {
	c.PushState()
	if path != nil {
		c.ClipPath(path, false)
	}
	c.PopState()
}

func (c *Canvas) setPath(path *zgeo.Path) {
	// zlog.Info("\n\nsetPath")
	c.context.Call("beginPath")
	path.ForEachPart(func(part zgeo.PathNode) {
		switch part.Type {
		case zgeo.PathMove:
			// zlog.Info("moveTo", part.Points[0].X, part.Points[0].Y)
			c.context.Call("moveTo", part.Points[0].X, part.Points[0].Y)
		case zgeo.PathLine:
			// zlog.Info("lineTo", part.Points[0].X, part.Points[0].Y)
			c.context.Call("lineTo", part.Points[0].X, part.Points[0].Y)
		case zgeo.PathClose:
			// zlog.Info("pathClose")
			c.context.Call("closePath")
		case zgeo.PathQuadCurve:
			c.context.Call("quadraticCurveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y)
			// zlog.Info("quadCurve")
			break
		case zgeo.PathCurve:
			c.context.Call("bezierCurveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y, part.Points[2].X, part.Points[2].Y)
			// zlog.Info("curveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y, part.Points[2].X, part.Points[2].Y)
			break
		}
	})
}

func (c *Canvas) setMatrix(m zgeo.Matrix) {
	c.currentMatrix = m
}

func (c *Canvas) setLineType(ltype zgeo.PathLineType) {
}

func (c *Canvas) setLineWidth(width float64) {
	c.context.Set("lineWidth", width)
}

// DrawTextInPos fills or strokes *text* at bottom-left position *pos*.
// Strokes if strokeWidth is non zero, and changes the canvases current *lineWidth*.
func (c *Canvas) DrawTextInPos(pos zgeo.Pos, text string, strokeWidth float64) {
	name := "fillText"
	if strokeWidth != 0 {
		c.context.Set("lineWidth", strokeWidth)
		name = "strokeText"
	}
	c.context.Call(name, text, pos.X, pos.Y)
}

func (c *Canvas) MeasureText(text string, font *zgeo.Font) zgeo.Size {
	var s zgeo.Size
	c.SetFont(font, nil)
	var metrics = c.context.Call("measureText", text)
	s.W = metrics.Get("width").Float()
	//	zlog.Info("c measure:", text)
	s.H = font.LineHeight() * 1.1
	return s
}

func (c *Canvas) GoImage(cut zgeo.Rect) image.Image {
	if cut.IsNull() {
		cut = zgeo.Rect{Size: c.Size()}
	}
	idata := c.context.Call("getImageData", cut.Pos.X, cut.Pos.X, cut.Size.W, cut.Size.H)
	clamped := idata.Get("data")
	ilen := clamped.Length()
	// zlog.Info("getImageData:", ilen)
	buf := make([]byte, ilen, ilen)
	js.CopyBytesToGo(buf, clamped)
	// zlog.Info("canvas.Image:", cut, ilen, n)
	newImage := image.NewNRGBA(zgeo.Rect{Size: cut.Size}.GoRect())
	newImage.Pix = buf
	return newImage
}

func (c *Canvas) SetGoImage(img image.Image, pos zgeo.Pos) {
	irgba := zimage.GoImageToGoRGBA(img)
	rgba := irgba.(*image.RGBA)
	bytes := rgba.Pix
	array := js.Global().Get("Uint8ClampedArray").New(len(bytes))
	js.CopyBytesToJS(array, bytes)
	s := zimage.GoImageZSize(img)
	iDataType := js.Global().Get("ImageData")
	idata := iDataType.New(array, int(s.W), int(s.H))
	c.context.Call("putImageData", idata, pos.X, pos.Y)
}

func (c *Canvas) ZImage(ensureCopy bool, got func(img *zimage.Image)) {
	// gi := c.GoImage(cut)
	// if gi == nil {
	// 	return nil
	// }
	// return ImageFromGo(gi)
	surl := c.element.Call("toDataURL").String()
	zimage.FromPath(surl, false, got)
}

func CanvasFromGoImage(i image.Image) *Canvas {
	canvas := New()
	canvas.SetSize(zimage.GoImageZSize(i))
	canvas.SetGoImage(i, zgeo.Pos{0, 0})
	return canvas
}

func RenderToImage(size zgeo.Size, draw func(canvasContext js.Value)) image.Image {
	canvas := New()
	// canvas.element.Set("id", "render-canvas")
	canvas.SetSize(size)
	draw(canvas.context)
	return canvas.GoImage(zgeo.Rect{})
}
