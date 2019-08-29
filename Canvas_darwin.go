package zgo

import "fmt"

type canvasNative struct {
}

func (c *Canvas) SetRect(rect Rect) {
}

func (c *Canvas) SetColor(color Color, opacity float32) {

	var vcolor = color
	if opacity != -1 {
		vcolor = vcolor.OpacityChanged(opacity)
	}
	// context.setStrokeColor(vcolor.color.cgColor)
	// context.setFillColor(vcolor.color.cgColor)
}

func (c *Canvas) FillPath(path *Path) {
	c.setPath(path)
	// context.fillPath(using(eofill ? .evenOdd  .winding))
}

func (c *Canvas) FillPathEO(path *Path) {
	c.setPath(path)
	// context.fillPath(using(eofill ? .evenOdd  .winding))
}

func (c *Canvas) SetFont(font *Font, matrix *Matrix) {
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

func (c *Canvas) DrawImage(image *Image, destRect Rect, opacity float32, blendMode CanvasBlendMode, sourceRect Rect) Rect {
	fmt.Println("Canvas.DrawImage darwin", image)
	//        image.draw(vdestRect.GetCGRect(), blendMode, opacity)
	return destRect
}

func drawInsetRow(canvas *Canvas, image *Image, inset, dest Rect, sy, sh, dy, dh float64, opacity float32, blendMode CanvasBlendMode) {
	size := image.Size()
	diff := dest.Size.Minus(size)
	insetMid := size.Minus(inset.Size.Negative())
	canvas.DrawImage(image, RectFromXYWH(0, sy, inset.Pos.X, sh), opacity, blendMode, RectFromXYWH(0, dy, inset.Pos.X, dh))
	canvas.DrawImage(image, RectFromXYWH(inset.Pos.X, sy, insetMid.W, sh), opacity, blendMode, RectFromXYWH(inset.Pos.X, dy, diff.W, dh))
	canvas.DrawImage(image, RectFromXYWH(size.W+inset.Max().X, sy, -inset.Max().X, sh), opacity, blendMode, RectFromXYWH(dest.Max().X+inset.Max().X, dy, -inset.Max().X, dh))
}

func (c *Canvas) drawInsetImage(canvas *Canvas, image *Image, inset, dest Rect, opacity float32, blendMode CanvasBlendMode) {
	size := image.Size()
	insetMid := size.Minus(inset.Size.Negative())
	diff := dest.Size.Minus(size)
	drawInsetRow(canvas, image, inset, dest, 0, inset.Pos.Y, 0, inset.Pos.Y, opacity, blendMode)
	drawInsetRow(canvas, image, inset, dest, inset.Pos.Y, insetMid.H, inset.Pos.Y, diff.H, opacity, blendMode)
	drawInsetRow(canvas, image, inset, dest, inset.Size.H+inset.Max().Y, -inset.Max().X, size.H+inset.Max().Y, -inset.Max().Y, opacity, blendMode)
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

}

func (c *Canvas) setMatrix(m Matrix) {
	c.currentMatrix = m
}

func (c *Canvas) setLineType(ltype PathLineType) {
}

func (c *Canvas) DrawTextInPos(pos Pos, text string, attributes Dictionary) {
}

func canvasGetTextSize(text string, font *Font) Size {
	return Size{}
}
