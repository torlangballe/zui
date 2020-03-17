package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
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
	c.context.Set("imageSmoothingEnabled", true)
	c.context.Set("imageSmoothingQuality", "high")
	return &c
}

func (c *Canvas) Element() js.Value {
	return c.element
}

func (c *Canvas) SetRect(rect zgeo.Rect) {
	setElementRect(c.element, rect)
}

func (c *Canvas) setColor(color zgeo.Color, opacity float32, stroke bool) {
	var vcolor = color
	if opacity != -1 {
		vcolor = vcolor.OpacityChanged(opacity)
	}
	str := makeRGBAString(vcolor)
	name := "fillStyle"
	if stroke {
		name = "strokeStyle"
	}
	c.context.Set(name, str)
}

func (c *Canvas) SetColor(color zgeo.Color, opacity float32) {
	c.setColor(color, opacity, false)
	c.setColor(color, opacity, true)
}

func (c *Canvas) FillPath(path *zgeo.Path) {
	c.setPath(path)
	c.context.Call("fill")
}

func (c *Canvas) FillPathEO(path *zgeo.Path) {
}

func (c *Canvas) SetFont(font *Font, matrix *zgeo.Matrix) {
	str := getFontStyle(font)
	// fmt.Println("canvas set font:", str)
	c.context.Set("font", str)
	//    state.font = afontCreateTransformed(amatrix)
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
	//      context.clip(using(eofill ? .evenOdd  .winding))
}

func (c *Canvas) GetClipRect() zgeo.Rect {
	return zgeo.Rect{}
	//        return SetRect(context.boundingBoxOfClipPath)
}

func (c *Canvas) StrokePath(path *zgeo.Path, width float64, ltype zgeo.PathLineType) {
	c.setPath(path)
	c.setLineType(ltype)
	c.context.Call("stroke")

	// context.setLineWidth(CGFloat(width))
}

func (c *Canvas) DrawPath(path *zgeo.Path, strokeColor zgeo.Color, width float64, ltype zgeo.PathLineType, eofill bool) {
	c.setPath(path)
	c.context.Call("fill")
	c.PushState()
	c.setColor(strokeColor, 1, true)
	c.setLineType(ltype)
	c.context.Call("stroke")
	c.PopState()

	//        context.setLineWidth(CGFloat(width))

	//        context.drawPath(using eofill ? CGPathDrawingMode.eoFillStroke  CGPathDrawingMode.fillStroke)
}

func (c *Canvas) drawPlainImage(image *Image, destRect zgeo.Rect, opacity float32, blendMode CanvasBlendMode, sourceRect zgeo.Rect) {
	sr := sourceRect.TimesD(float64(image.scale))
	// fmt.Println("drawPlainImage:", destRect, sourceRect, sr, c)
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

func (c *Canvas) DrawGradient(path *zgeo.Path, colors []zgeo.Color, pos1 zgeo.Pos, pos2 zgeo.Pos, locations []float32) {
	// make this a color type instead?? Maybe no point, as it has fixed start/end pos for gradient
	c.PushState()
	c.ClipPath(path, false, false)
	gradient := c.context.Call("createLinearGradient", pos1.X, pos1.Y, pos2.X, pos2.Y)
	if len(locations) == 0 {
		last := len(colors) - 1
		for i := 0; i <= last; i++ {
			locations = append(locations, float32(i)/float32(last))
		}
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
	// fmt.Println("\n\nsetPath")
	c.context.Call("beginPath")
	path.ForEachPart(func(part zgeo.PathNode) {
		switch part.Type {
		case zgeo.PathMove:
			//  fmt.Println("moveTo", part.Points[0].X, part.Points[0].Y)
			c.context.Call("moveTo", part.Points[0].X, part.Points[0].Y)
		case zgeo.PathLine:
			// fmt.Println("lineTo", part.Points[0].X, part.Points[0].Y)
			c.context.Call("lineTo", part.Points[0].X, part.Points[0].Y)
		case zgeo.PathClose:
			// fmt.Println("pathClose")
			c.context.Call("closePath")
		case zgeo.PathQuadCurve:
			c.context.Call("quadraticCurveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y)
			// fmt.Println("quadCurve")
			break
		case zgeo.PathCurve:
			c.context.Call("bezierCurveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y, part.Points[2].X, part.Points[2].Y)
			// fmt.Println("curveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y, part.Points[2].X, part.Points[2].Y)
			break
		}
	})
}

func (c *Canvas) setMatrix(m zgeo.Matrix) {
	c.currentMatrix = m
}

func (c *Canvas) setLineType(ltype zgeo.PathLineType) {
}

func (c *Canvas) DrawTextInPos(pos zgeo.Pos, text string, attributes zdict.Dict) {
	c.context.Call("fillText", text, pos.X, pos.Y)
}

var measureTextCanvas *Canvas

func canvasGetTextSize(text string, font *Font) zgeo.Size {
	// https://stackoverflow.com/questions/118241/calculate-text-width-with-javascript

	var s zgeo.Size
	if measureTextCanvas == nil {
		measureTextCanvas = CanvasNew()
	}
	measureTextCanvas.SetFont(font, nil)
	var metrics = measureTextCanvas.context.Call("measureText", text)

	s.W = metrics.Get("width").Float()
	//	s.H = metrics.Get("height").Float()

	// if text == "QTT Manager" {
	// 	fmt.Println("canvasGetTextSize:", text, font.Size, font.Name, s, s.W)
	// }
	//	s.W -= 3 // seems to wrap otherwise, maybe it's rounded down to int somewhere
	s.H = font.LineHeight() * 0.85

	// fmt.Println("canvasGetTextSize:", text, font, s)
	return s
}
