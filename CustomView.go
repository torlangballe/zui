package zgo

type CustomView struct {
	ViewBaseHandler
	MinSize        Size
	pressedHandler ViewPressedProtocol
	valueHandler   ViewValueChangedProtocol
	drawHandler    ViewDrawProtocol
	timers         []*Timer
}

func CustomViewNew() *CustomView {
	c := &CustomView{}
	return c
}

func (v *CustomView) PressedHandler(handler ViewPressedProtocol) {
	v.pressedHandler = handler
}

func (v *CustomView) ValueHandler(handler ViewValueChangedProtocol) {
	v.valueHandler = handler
}

func (v *CustomView) DrawHandler(handler ViewDrawProtocol) {
	v.drawHandler = handler
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

func (v *CustomView) getStateColor(col Color) Color {
	return Color{}
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

func zViewSetRect(view *ViewNative, rect Rect, layout bool) { // layout only used on android
	view.Rect(rect)
}

func zRemoveViewFromSuper(view *ViewNative, detachFromContainer bool) {
	//view.removeFromSuperview()
}
