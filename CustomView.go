package zgo

import "github.com/torlangballe/zutil/ztimer"

type CustomView struct {
	NativeView
	canvas        *Canvas
	minSize       Size
	pressed       func()
	valueChanged  func(view View)
	draw          func(rect Rect, canvas *Canvas, view View)
	exposed       bool
	timers        []*ztimer.Timer
	color         Color
	IsHighlighted bool
	exposeTimer   ztimer.Timer
	isSetup       bool
}

func CustomViewNew(name string) *CustomView {
	c := &CustomView{}
	c.init(c, name)
	return c
}

func (v *CustomView) GetCalculatedSize(total Size) Size {
	return v.GetMinSize()
}

func (v *CustomView) Expose() {
	v.exposeInSecs(0.01)
}

func (v *CustomView) exposeInSecs(secs float64) {
	if v.exposed || presentViewPresenting {
		return
	}
	// fmt.Println("exposeInSecs", v.GetObjectName())
	v.exposed = true
	v.exposeTimer.Stop()
	v.exposeTimer.Set(secs, true, func() {
		io, got := v.View.(ImageOwner)
		if got {
			image := io.GetImage()
			if image != nil && image.loading {
				v.exposeInSecs(0.1)
				return
			}
		}
		et, _ := v.View.(ExposableType)
		if et != nil {
			// fmt.Println("CV exposeInSecs draw", v.GetObjectName())
			et.drawIfExposed()
		}
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
	v.makeCanvas()
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

func zConvertViewSizeThatFitstToSize(view *NativeView, sizeIn Size) Size {
	//    return Size(view.sizeThatFits(sizeIn.GetCGSize()))
	return Size{}
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
