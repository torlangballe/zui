package zgo

import (
	"fmt"
	"syscall/js"
)

type canvasNative struct {
	element js.Value
	context js.Value
}

func CanvasNew() *Canvas {
	c := Canvas{}
	c.element = DocumentJS.Call("createElement", "canvas")
	c.element.Set("style", "position:absolute")
	c.context = c.element.Call("getContext", "2d")
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
	// fmt.Println("drawPlainImage:", destRect, sourceRect)
	sr := sourceRect.TimesD(float64(image.scale))
	c.context.Call("drawImage", image.imageJS, sr.Pos.X, sr.Pos.Y, sr.Size.W, sr.Size.H, destRect.Pos.X, destRect.Pos.Y, destRect.Size.W, destRect.Size.H)
}

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
	//	diff := dest.Size.Minus(size)
	insetMid := size.Minus(inset.Size.Negative())
	//	fmt.Println("drawInsetRow:", "sy:", sy, "sh:", sh, "dy:", dy, "dh:", dh)
	c.drawPlainImage(image, RectFromXYWH(0, dy, inset.Pos.X, dh), opacity, blendMode, RectFromXYWH(0, sy, inset.Pos.X, sh))
	midMaxX := dest.Max().X + inset.Max().X // inset.Max is negative
	c.drawPlainImage(image, RectFromXYWH(inset.Pos.X, dy, midMaxX-inset.Pos.X, dh), opacity, blendMode, RectFromXYWH(inset.Pos.X, sy, insetMid.W, sh))
	c.drawPlainImage(image, RectFromXYWH(midMaxX, dy, -inset.Max().X, dh), opacity, blendMode, RectFromXYWH(size.W+inset.Max().X, sy, -inset.Max().X, sh))
}

func (c *Canvas) drawInsetImage(image *Image, inset, dest Rect, opacity float32, blendMode CanvasBlendMode) {
	size := image.Size()
	insetMid := size.Minus(inset.Size.Negative())
	diff := dest.Size.Minus(size).Plus(insetMid)
	//	fmt.Println("drawInsetImage:", dest, size, insetMid, diff)
	c.drawInsetRow(image, inset, dest, 0, inset.Pos.Y, dest.Min().Y, inset.Pos.Y, opacity, blendMode)
	c.drawInsetRow(image, inset, dest, inset.Pos.Y, insetMid.H, dest.Min().Y+inset.Pos.Y, diff.H, opacity, blendMode)
	c.drawInsetRow(image, inset, dest, size.H+inset.Max().Y, -inset.Max().Y, dest.Max().Y+inset.Max().Y, -inset.Max().Y, opacity, blendMode)
}

func (c *Canvas) PushState() {
	//      context.saveGState()
}

func (c *Canvas) PopState() {
	//      context.restoreGState()
}

func (c *Canvas) ClearRect(rect Rect) {
	//      context.clear(rect.GetCGRect())
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

var measureDiv *js.Value

func canvasGetTextSize(text string, font *Font) Size {
	// https://stackoverflow.com/questions/118241/calculate-text-width-with-javascript
	var s Size
	if measureDiv == nil {
		e := DocumentJS.Call("createElement", "div")
		e.Set("hidden", "true")
		DocumentElementJS.Call("appendChild", e)
		measureDiv = &e
	}
	style := measureDiv.Get("style")

	style.Set("fontSize", fmt.Sprintf("%dpx", int(font.Size)))
	style.Set("position", "absolute")
	style.Set("left", "-1000")
	style.Set("top", "-1000")
	measureDiv.Set("innerHTML", text)

	s.W = measureDiv.Get("clientWidth").Float()
	s.H = measureDiv.Get("clientHeight").Float()

	s.W += 2 // seems to wrap otherwise, maybe it's rounded down to int somewhere
	return s
}
