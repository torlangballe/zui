package zgo

import "fmt"

type CustomViewProtocol interface {
	drawIfExposed()
	getCanvas() *Canvas
}

type CustomView struct {
	ViewBaseHandler
	minSize       Size
	pressed       func(pos Pos)
	valueChanged  func(view View)
	draw          func(rect Rect, canvas *Canvas, view View)
	exposed       bool
	canvas        *Canvas
	timers        []*Timer
	color         Color
	IsHighlighted bool
	exposeTimer   Timer
}

func (v *CustomView) getCanvas() *Canvas {
	return v.canvas
}

func (v *CustomView) GetCalculatedSize(total Size) Size {
	return v.GetMinSize()
}
func (v *CustomView) Expose() {
	v.exposeInSecs(0.01)
}

func (v *CustomView) exposeInSecs(secs float64) {
	fmt.Println("CV.Expose:", v.GetObjectName())
	v.exposed = true
	v.exposeTimer.Stop()
	v.exposeTimer.Set(secs, true, func() {
		io, got := v.view.(ImageOwner)
		if got {
			image := io.GetImage()
			if image != nil && image.loading {
				v.exposeInSecs(0.1)
				return
			}
		}
		v.drawIfExposed()
	})
}

func (v *CustomView) Color(color Color) View {
	v.color = color
	return v
}

func (v *CustomView) MinSize(s Size) *CustomView {
	v.minSize = s
	return v
}

func (v *CustomView) GetMinSize() Size {
	return v.minSize
}

func (v *CustomView) ValueHandler(handler func(view View)) {
	v.valueChanged = handler
}

func (v *CustomView) DrawHandler(handler func(rect Rect, canvas *Canvas, view View)) {
	v.draw = handler
}

func (v *CustomView) GetPosFromMe(pos Pos, inView View) Pos {
	return Pos{}
}

func (v *CustomView) GetPosToMe(pos Pos, inView View) Pos {
	return Pos{}
}

func (v *CustomView) GetViewsRectInMyCoordinates(view View) Rect {
	return Rect{}
}

func (v *CustomView) HandleClosing() {
	for _, t := range v.timers {
		t.Stop()
	}
	v.timers = v.timers[:]
}

func (v *CustomView) Activate(activate bool) { // like being activated/deactivated for first time
}

func (v *CustomView) Rotate(degrees float64) {
	// r := MathDegToRad(degrees)
	//self.transform = CGAffineTransform(rotationAngle CGFloat(r))
}

func zConvertViewSizeThatFitstToSize(view *ViewNative, sizeIn Size) Size {
	//    return Size(view.sizeThatFits(sizeIn.GetCGSize()))
	return Size{}
}

func zRemoveViewFromSuper(view *ViewNative, detachFromContainer bool) {
	//view.removeFromSuperview()
}

func (v *CustomView) getStateColor(col Color) Color {
	if v.IsHighlighted {
		g := col.GetGrayScale()
		if g < 0.5 {
			col = col.Mix(ColorWhite, 0.5)
		} else {
			col = col.Mix(ColorBlack, 0.5)
		}
	}
	if !v.IsUsable() {
		col = col.OpacityChanged(0.3)
	}
	return col
}
