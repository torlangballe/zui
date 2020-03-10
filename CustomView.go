package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

type CustomView struct {
	NativeView
	canvas        *Canvas
	minSize       zgeo.Size
	pressed       func()
	valueChanged  func(view View)
	draw          func(rect zgeo.Rect, canvas *Canvas, view View)
	exposed       bool
	timers        []*ztimer.Timer
	color         zgeo.Color
	IsHighlighted bool
	exposeTimer   ztimer.Timer
	isSetup       bool
}

func CustomViewNew(name string) *CustomView {
	c := &CustomView{}
	c.init(c, name)
	return c
}

func (v *CustomView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.MinSize()
}

func (v *CustomView) Expose() {
	v.exposeInSecs(0.1) //0.01)
}

func (v *CustomView) exposeInSecs(secs float64) {
	if v.exposed || presentViewPresenting {
		return
	}
	// fmt.Println("exposeInSecs", v.ObjectName())
	v.exposed = true
	v.exposeTimer.StartIn(secs, true, func() {
		io, got := v.View.(ImageOwner)
		// fmt.Println("exposeInSecs draw", v.ObjectName(), got)
		if got {
			image := io.GetImage()
			if image != nil {
				// fmt.Println("CV exposeInSecs draw image", v.ObjectName(), image.loading)
			}
			if image != nil && image.loading {
				v.exposeInSecs(0.1)
				return
			}
		}
		et, _ := v.View.(ExposableType)
		if et != nil {
			// fmt.Println("CV exposeInSecs draw", v.ObjectName())
			et.drawIfExposed()
		}
	})
}

func (v *CustomView) SetColor(c zgeo.Color) View {
	v.NativeView.SetColor(c)
	v.color = c

	return v
}

func (v *CustomView) Color() zgeo.Color {
	return v.color
}

func (v *CustomView) SetMinSize(s zgeo.Size) *CustomView {
	v.minSize = s
	return v
}

func (v *CustomView) MinSize() zgeo.Size {
	return v.minSize
}

func (v *CustomView) SetValueHandler(handler func(view View)) {
	v.valueChanged = handler
}

func (v *CustomView) SetDrawHandler(handler func(rect zgeo.Rect, canvas *Canvas, view View)) {
	v.draw = func(rect zgeo.Rect, canvas *Canvas, view View) {
		if handler != nil {
			handler(rect, canvas, view)
		}
	}
}

func (v *CustomView) GetPosFromMe(pos zgeo.Pos, inView View) zgeo.Pos {
	return zgeo.Pos{}
}

func (v *CustomView) GetPosToMe(pos zgeo.Pos, inView View) zgeo.Pos {
	return zgeo.Pos{}
}

func (v *CustomView) GetViewsRectInMyCoordinates(view View) zgeo.Rect {
	return zgeo.Rect{}
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

func zConvertViewSizeThatFitstToSize(view *NativeView, sizeIn zgeo.Size) zgeo.Size {
	//    return Size(view.sizeThatFits(sizeIn.GetCGSize()))
	return zgeo.Size{}
}

func (v *CustomView) getStateColor(col zgeo.Color) zgeo.Color {
	if v.IsHighlighted {
		g := col.GetGrayScale()
		if g < 0.5 {
			col = col.Mix(zgeo.ColorWhite, 0.5)
		} else {
			col = col.Mix(zgeo.ColorBlack, 0.5)
		}
	}
	if !v.Usable() {
		col = col.OpacityChanged(0.3)
	}
	return col
}
