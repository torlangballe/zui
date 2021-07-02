package zui

import (
	"image"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

// interesting: https://github.com/markfarnan/go-canvas

type canvasNative struct {
	element js.Value
	context js.Value
}

func CanvasNew() *Canvas {
	c := Canvas{}
	c.element = DocumentJS.Call("createElement", "canvas")
	c.element.Set("style", "position:absolute")
	c.context = c.element.Call("getContext", "2d")
	// c.context.Set("imageSmoothingEnabled", true)
	// c.context.Set("imageSmoothingQuality", "high")
	return &c
}

func (c *Canvas) SetSize(size zgeo.Size) {
	c.element.Set("width", size.W) // scale?
	c.element.Set("height", size.H)
	c.size = size
}

func (c *Canvas) Element() js.Value {
	return c.element
}

func (c *Canvas) SetRect(rect zgeo.Rect) {
	setElementRect(c.element, rect)
}

func (c *Canvas) setColor(color zgeo.Color, stroke bool) {
	var vcolor = color
	str := makeRGBAString(vcolor)
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

func (c *Canvas) FillPath(path *zgeo.Path) {
	c.setPath(path)
	c.context.Call("fill")
}

func (c *Canvas) FillPathEO(path *zgeo.Path) {
}

func (c *Canvas) SetFont(font *Font, matrix *zgeo.Matrix) error {
	str := getFontStyle(font)
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

func (c *Canvas) ClipPath(path *zgeo.Path, exclude bool, eofill bool) {
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

var scaledImageMap = map[scaledImage]*Image{}

func (c *Canvas) drawCachedScaledImage(image *Image, synchronous, useDownsampleCache bool, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) bool {
	proportional := false
	ds := destRect.Size.Ceil()
	si := scaledImage{image.Path, ds}
	sourceRect = zgeo.Rect{Size: ds}
	var newImage *Image
	if useDownsampleCache {
		newImage, _ = scaledImageMap[si]
		if newImage == nil {
			synchronous = false
		}
	}
	do := func() {
		if newImage != nil {
			image = newImage
			c.rawDrawPlainImage(image, destRect, opacity, sourceRect)
			return
		}
		if len(scaledImageMap) > 500 {
			scaledImageMap = map[scaledImage]*Image{}
		}
		image.ShrunkInto(ds, proportional, func(image *Image) {
			if useDownsampleCache {
				scaledImageMap[si] = image
			}
			c.rawDrawPlainImage(image, destRect, opacity, sourceRect)
		})
	}
	if synchronous {
		do()
	} else {
		go do()
	}
	return synchronous
}

func (c *Canvas) drawPlainImage(image *Image, synchronous, useDownsampleCache bool, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) {
	// zlog.Info("rawDrawPlainImage1:", image.Size(), image.Path)
	if destRect.Size.H < 0 {
		zlog.Info("drawPlainImage BAD!:", image.loading, image.Size(), destRect, sourceRect, c)
		return
	}
	ss := sourceRect.Size
	ds := destRect.Size
	// zlog.Info("drawPlain:", image.Size(), image.Path, ss, ds, ss.Area() < 1000000, ss == image.size, sourceRect.Pos.IsNull())
	if image.Path != "" && c.DownsampleImages && ss.Area() < 1000000 && ss == image.size && sourceRect.Pos.IsNull() && (ds.W/ss.W < 0.95 || ds.H/ss.H < 0.95) {
		if c.drawCachedScaledImage(image, synchronous, useDownsampleCache, destRect, opacity, sourceRect) { // if it retturns false, it wasn't in cache, so we draw unscaled
		}
		return
	}
	if synchronous {
		// zlog.Info("rawDrawPlainImage:", image.Size(), ss, ds, image.Path)
		c.rawDrawPlainImage(image, destRect, opacity, sourceRect)
	} else {
		go c.rawDrawPlainImage(image, destRect, opacity, sourceRect)
	}
}

func (c *Canvas) rawDrawPlainImage(image *Image, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) {
	sr := sourceRect.TimesD(float64(image.scale))
	if image.Path == "images/plus-circled-white.png" {
		zlog.Info("rawDrawPlain:", image.Size(), image.Path, sr, destRect, zlog.GetCallingStackString())
	}
	c.context.Call("drawImage", image.imageJS, sr.Pos.X, sr.Pos.Y, sr.Size.W, sr.Size.H, destRect.Pos.X, destRect.Pos.Y, destRect.Size.W, destRect.Size.H)
}

func (c *Canvas) PushState() {
	c.context.Call("save")
}

func (c *Canvas) PopState() {
	c.context.Call("restore")
}

func (c *Canvas) ClearRect(rect zgeo.Rect) {
	if rect.IsNull() {
		rect.Size.W = c.Element().Get("width").Float()
		rect.Size.H = c.Element().Get("height").Float()
	}
	c.context.Call("clearRect", rect.Pos.X, rect.Pos.Y, rect.Size.W, rect.Size.H)
}

func (c *Canvas) SetDropShadow(deltaSize zgeo.Size, blur float32, color zgeo.Color) {
	// moffset := delta.GetCGSize()    //Mac    moffset.height *= -1
	//context.setShadow(offset moffset, blur CGFloat(blur), color color.color.cgColor)
}

func (c *Canvas) SetDropShadowOff(opacity float64) {
	//        context.setShadow(offset CGSize.zero, blur 0, color nil)
	if opacity != 1 {
		//            context.setAlpha(CGFloat(opacity))
	}
}

func (c *Canvas) DrawGradient(path *zgeo.Path, colors []zgeo.Color, pos1 zgeo.Pos, pos2 zgeo.Pos, locations []float64) {
	// make this a color type instead?? Maybe no point, as it has fixed start/end pos for gradient
	c.PushState()
	c.ClipPath(path, false, false)
	gradient := c.context.Call("createLinearGradient", pos1.X, pos1.Y, pos2.X, pos2.Y)
	if len(locations) == 0 {
		locations = canvasCreateGradientLocations(len(colors))
	}
	for i, c := range colors {
		gradient.Call("addColorStop", locations[i], makeRGBAString(c))
	}
	c.context.Set("fillStyle", gradient)
	c.FillPath(path)
	c.PopState()
}

func (c *Canvas) DrawRadialGradient(path *zgeo.Path, colors []zgeo.Color, center zgeo.Pos, radius float64, endCenter *zgeo.Pos, startRadius float64, locations []float32) {
	c.PushState()
	if path != nil {
		c.ClipPath(path, false, false)
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

func (c *Canvas) MeasureText(text string, font *Font) zgeo.Size {
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
	irgba := GoImageToGoRGBA(img)
	rgba := irgba.(*image.RGBA)
	bytes := rgba.Pix
	array := js.Global().Get("Uint8ClampedArray").New(len(bytes))
	js.CopyBytesToJS(array, bytes)
	s := GoImageZSize(img)
	iDataType := js.Global().Get("ImageData")
	idata := iDataType.New(array, int(s.W), int(s.H))
	c.context.Call("putImageData", idata, pos.X, pos.Y)
}

func (c *Canvas) ZImage(got func(img *Image)) {
	// gi := c.GoImage(cut)
	// if gi == nil {
	// 	return nil
	// }
	// return ImageFromGo(gi)
	surl := c.element.Call("toDataURL").String()
	ImageFromPath(surl, got)
}
