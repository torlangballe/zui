package zgo

type Canvas struct {
	currentMatrix Matrix // is currentTransform...
}

type CanvasBlendMode int

const (
	normal CanvasBlendMode = iota
	multiply
	screen
	overlay
	darken
	lighten
	colorDodge
	colorBurn
	softLight
	hardLight
	difference
	exclusion
	hue
	saturation
	color
	luminosity
)

/*

function getTextWidth(text, font) {
    // re-use canvas object for better performance
    var canvas = getTextWidth.canvas || (getTextWidth.canvas = document.createElement("canvas"))
    var context = canvas.getContext("2d")
    context.font = font
    var metrics = context.measureText(text)
    return metrics.width
}

console.log(getTextWidth("hello there!", "bold 12pt arial"))  // close to 86

*/

//
//  Canvas.swift
//
//  Created by Tor Langballe on /21/10/15.
//  Copyright © 2015 Capsule.fm. All rights reserved.
//

func MatrixForRotatingAroundPoint(point Pos, deg float64) Matrix {
	var transform = MatrixIdentity
	transform = transform.TranslatedByPos(point)
	transform = transform.Rotated(MathDegToRad(deg))
	transform = transform.TranslatedByPos(point.Negative())

	return transform
}

func MatrixForRotationDeg(deg float64) Matrix {
	var transform = MatrixIdentity
	transform = transform.Rotated(MathDegToRad(deg))
	return transform
}

func (c *Canvas) SetColor(color Color, opacity float32) {

	var vcolor = color
	if opacity != -1 {
		vcolor = vcolor.OpacityChanged(opacity)
	}
	// context.setStrokeColor(vcolor.color.cgColor)
	// context.setFillColor(vcolor.color.cgColor)
}

func (c *Canvas) FillPath(path *Path, eofill bool) {
	c.setPath(path)
	// context.fillPath(using(eofill ? .evenOdd  .winding))
}

func (c *Canvas) SetFont(font Font, matrix Matrix) {
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

func (c *Canvas) StrokePath(path *Path, widthfloat64, ltype PathLineType) {
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

func (c *Canvas) DrawImage(image Image, destRect Rect, align Alignment, opacity float32, blendMode CanvasBlendMode, corner float64, margin Size) Rect {
	var vdestRect = destRect
	if align != AlignmentNone {
		vdestRect = vdestRect.Align(Size(image.size), align, margin, Size{0, 0})
	} else {
		vdestRect = vdestRect.Expanded(margin.Negative())
	}
	if corner != 0 {
		c.PushState()
		path := NewRectPath(vdestRect, Size{corner, corner})
		c.ClipPath(path, false, false)
	}
	//        image.draw(vdestRect.GetCGRect(), blendMode, opacity)
	if corner != 0 {
		c.PopState()
	}
	return vdestRect
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
