package zgo

type CustomView struct {
	ViewBaseHandler
	MinSize             Size
	pressedHandler      ViewPressedProtocol
	valueChangedHandler ViewValueChangedProtocol
	drawHandler         ViewDrawProtocol
	timers              []*Timer
}

func CustomViewNew() *CustomView {
	c := &CustomView{}
	return c
}

func (v *CustomView) SetPressedInPosHandler(handler ViewPressedProtocol) {
	v.pressedHandler = handler
}

func (v *CustomView) SetValueChangedHandler(handler ViewValueChangedProtocol) {
	v.valueChangedHandler = handler
}

func (v *CustomView) SetDrawHandler(handler ViewDrawProtocol) {
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
