package zgo

import (
	"syscall/js"
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

func (c *Canvas) SetRect(rect Rect) {
	setElementRect(c.element, rect)
}

func (c *Canvas) SetColor(color Color, opacity float32) {
	var vcolor = color
	if opacity != -1 {
		vcolor = vcolor.OpacityChanged(opacity)
	}
	str := makeRGBAString(vcolor)
	c.context.Set("fillStyle", str)
}

func (c *Canvas) FillPath(path *Path) {
	c.setPath(path)
	c.context.Call("fill")
}

func (c *Canvas) FillPathEO(path *Path) {
}

func (c *Canvas) SetFont(font *Font, matrix *Matrix) {
	str := getFontStyle(font)
	c.context.Set("font", str)
	//    state.font = afontCreateTransformed(amatrix)
}

func (c *Canvas) SetMatrix(matrix Matrix) {
	c.currentMatrix = matrix
	c.setMatrix(matrix)
}

func (c *Canvas) Transform(matrix Matrix) {
	c.currentMatrix = c.currentMatrix.Multiplied(matrix)
}

func (c *Canvas) ClipPath(path *Path, exclude bool, eofill bool) {
	c.setPath(path)
	//      context.clip(using(eofill ? .evenOdd  .winding))
}

func (c *Canvas) GetClipRect() Rect {
	return Rect{}
	//        return Rect(context.boundingBoxOfClipPath)
}

func (c *Canvas) StrokePath(path *Path, width float64, ltype PathLineType) {
	c.setPath(path)
	c.setLineType(ltype)
	// context.setLineWidth(CGFloat(width))
	// context.strokePath()
}

func (c *Canvas) DrawPath(path *Path, strokeColor Color, width float64, ltype PathLineType, eofill bool) {
	c.setPath(path)
	//        context.setStrokeColor(strokeColor.color.cgColor)

	c.setLineType(ltype)
	//        context.setLineWidth(CGFloat(width))

	//        context.drawPath(using eofill ? CGPathDrawingMode.eoFillStroke  CGPathDrawingMode.fillStroke)
}

func (c *Canvas) drawPlainImage(image *Image, destRect Rect, opacity float32, blendMode CanvasBlendMode, sourceRect Rect) {
	sr := sourceRect.TimesD(float64(image.scale))
	// fmt.Println("drawPlainImage:", destRect, sourceRect, sr)
	c.context.Call("drawImage", image.imageJS, sr.Pos.X, sr.Pos.Y, sr.Size.W, sr.Size.H, destRect.Pos.X, destRect.Pos.Y, destRect.Size.W, destRect.Size.H)
}

func (c *Canvas) PushState() {
	//      context.saveGState()
}

func (c *Canvas) PopState() {
	//      context.restoreGState()
}

func (c *Canvas) ClearRect(rect Rect) {
	if rect.IsNull() {
		rect.Size.W = c.Element().Get("width").Float()
		rect.Size.H = c.Element().Get("height").Float()
	}
	c.context.Call("clearRect", rect.Pos.X, rect.Pos.Y, rect.Size.W, rect.Size.H)
}

func (c *Canvas) SetDropShadow(deltaSize Size, blur float32, color Color) {
	// moffset := delta.GetCGSize()    //Mac    moffset.height *= -1
	//context.setShadow(offset moffset, blur CGFloat(blur), color color.color.cgColor)
}

func (c *Canvas) SetDropShadowOff(opacity float64) {
	//        context.setShadow(offset CGSize.zero, blur 0, color nil)
	if opacity != 1 {
		//            context.setAlpha(CGFloat(opacity))
	}
}

func (c *Canvas) createGradient(colors []Color, locations []float32) *int { // returns int for now...
	return nil
}

func (c *Canvas) DrawGradient(path *Path, colors []Color, pos1 Pos, pos2 Pos, locations []float32) {
	c.PushState()
	if path != nil {
		c.ClipPath(path, false, false)
	}
	gradient := c.createGradient(colors, locations)
	if gradient != nil {
		//            context.drawLinearGradient(gradient, start pos1.GetCGPoint(), end pos2.GetCGPoint(), options CGGradientDrawingOptions(rawValueCGGradientDrawingOptions.drawsBeforeStartLocation.rawValue | CGGradientDrawingOptions.drawsBeforeStartLocation.rawValue))
		c.PopState()
	}
}

func (c *Canvas) DrawRadialGradient(path *Path, colors []Color, center Pos, radius float64, endCenter *Pos, startRadius float64, locations []float32) {
	c.PushState()
	if path != nil {
		//            self.ClipPath(path!)
	}
	gradient := c.createGradient(colors, locations)
	if gradient != nil {
		//            let c = UIGraphicsGetCurrentContext()
		//            context.drawRadialGradient(gradient, startCentercenter.GetCGPoint(), startRadiusCGFloat(startRadius), endCenter(endCenter == nil ? center  endCenter!).GetCGPoint(), endRadiusCGFloat(radius), options CGGradientDrawingOptions())
	}
	c.PopState()
}

func (c *Canvas) setPath(path *Path) {
	c.context.Call("beginPath")
	path.ForEachPart(func(part node) {
		switch part.Type {
		case PathMove:
			// fmt.Println("moveTo", part.Points[0].X, part.Points[0].Y)
			c.context.Call("moveTo", part.Points[0].X, part.Points[0].Y)
		case PathLine:
			// fmt.Println("lineTo", part.Points[0].X, part.Points[0].Y)
			c.context.Call("lineTo", part.Points[0].X, part.Points[0].Y)
		case PathClose:
			// fmt.Println("pathClose")
			c.context.Call("closePath")
		case PathQuadCurve:
			c.context.Call("quadraticCurveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y)
			// fmt.Println("quadCurve")
			break
		case PathCurve:
			c.context.Call("bezierCurveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y, part.Points[2].X, part.Points[2].Y)
			// fmt.Println("curveTo")
			break
		}
	})
}

func (c *Canvas) setMatrix(m Matrix) {
	c.currentMatrix = m
}

func (c *Canvas) setLineType(ltype PathLineType) {
}

func (c *Canvas) DrawTextInPos(pos Pos, text string, attributes Dictionary) {
	c.context.Call("fillText", text, pos.X, pos.Y)
}

var measureTextCanvas *Canvas

func canvasGetTextSize(text string, font *Font) Size {
	// https://stackoverflow.com/questions/118241/calculate-text-width-with-javascript

	var s Size
	if measureTextCanvas == nil {
		measureTextCanvas = CanvasNew()
	}
	measureTextCanvas.SetFont(font, nil)
	var metrics = measureTextCanvas.context.Call("measureText", text)

	s.W = metrics.Get("width").Float()
	//	s.H = metrics.Get("height").Float()

	s.W -= 3 // seems to wrap otherwise, maybe it's rounded down to int somewhere
	s.H = font.LineHeight()
	return s
}
