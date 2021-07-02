// +build zui

package zui

import (
	"fmt"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

type CustomView struct {
	NativeView
	LongPresser

	OpaqueDraw       bool // if OpaqueDraw set, drawing does not clear canvas first, assuming total coverage during draw
	DownsampleImages bool
	canvas           *Canvas
	minSize          zgeo.Size
	pressed          func()
	longPressed      func()
	valueChanged     func(view View)
	// pointerEnclosed func(inside bool)
	draw          func(rect zgeo.Rect, canvas *Canvas, view View)
	exposed       bool
	color         zgeo.Color
	exposeTimer   ztimer.Timer
	isSetup       bool
	isHighlighted bool
}

func CustomViewNew(name string) *CustomView {
	c := &CustomView{}
	c.Init(c, name)
	return c
}

func (v *CustomView) Canvas() *Canvas {
	return v.canvas
}

func (v *CustomView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.MinSize()
}

func (v *CustomView) SetHighlighted(h bool) {
	v.isHighlighted = h
	v.Expose()
}

func (v *CustomView) IsHighlighted() bool {
	return v.isHighlighted
}

func (v *CustomView) ExposeInSecs(secs float64) {
	// if v.ObjectName() == "chart-drawer" {
	// }
	// fmt.Printf("pre-exposedInSecs %s %p %v\n", v.Hierarchy(), v, v.Presented)
	if v.exposed || presentViewPresenting {
		return
	}
	// 	zlog.Info("Exposing:", v.Hierarchy(), v.Presented, zlog.GetCallingStackString())
	v.exposed = true
	v.exposeTimer.StartIn(secs, func() {
		// fmt.Printf("exposedInSecs %s %p %v\n", v.Hierarchy(), v, v.Presented)
		io, got := v.View.(ImageOwner)
		// zlog.Info("exposeInSecs draw", v.ObjectName(), got)
		if got {
			image := io.GetImage()
			if image != nil {
				// zlog.Info("CV exposeInSecs draw image", v.ObjectName(), image.loading)
			}
			if image != nil && image.loading {
				zlog.Info("CV exposeInSecs wait for loading", v.ObjectName())
				v.ExposeInSecs(0.1)
				return
			}
		}
		//		nv := ViewGetNative(v)
		// fmt.Printf("CV exposing: %s %p %v par:%p\n", v.Hierarchy(), nv, v.Presented, nv.Parent())
		if !v.Presented {
			nv := ViewGetNative(v)
			fmt.Printf("Not exposing unpresented: %s %p %v\n", v.Hierarchy(), nv, nv.Presented)
			return
		}
		//		if v.style().Get("width").String() == "" { // it's not arranged, could be collapsed in Container
		et, _ := v.View.(ExposableType)
		if et != nil {
			// if v.ObjectName() == "chart-drawer" {
			// 	zlog.Info("CV ExposeInSecs draw", v.ObjectName())
			// }
			et.drawIfExposed()
		}
	})
}

func (v *CustomView) PressedHandler() func() {
	return v.pressed
}

func (v *CustomView) LongPressedHandler() func() {
	return v.longPressed
}

func (v *CustomView) SetColor(c zgeo.Color) {
	v.NativeView.SetColor(c)
	v.color = c
	v.Expose()
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

func (v *CustomView) DrawHandler() func(rect zgeo.Rect, canvas *Canvas, view View) {
	return v.draw
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

// func (v *CustomView) HandleClosing() {
// 	zlog.Info("CV HandleClosing")
// 	for _, st := range v.StopOnClose {
// 		st.Stop()
// 	}
// 	v.StopOnClose = v.StopOnClose[:]
// }

// func (v *CustomView) Activate(activate bool) { // like being activated/deactivated for first time
// }

func zConvertViewSizeThatFitstToSize(view *NativeView, sizeIn zgeo.Size) zgeo.Size {
	//    return Size(view.sizeThatFits(sizeIn.GetCGSize()))
	return zgeo.Size{}
}

func (v *CustomView) getStateColor(col zgeo.Color) zgeo.Color {
	if v.isHighlighted {
		g := col.GrayScale()
		if g < 0.5 {
			col = col.Mixed(zgeo.ColorWhite, 0.5)
		} else {
			col = col.Mixed(zgeo.ColorBlack, 0.5)
		}
	}
	if !v.Usable() {
		col = col.WithOpacity(0.3)
	}
	return col
}

func (v *CustomView) Focus(focus bool) {
	v.NativeView.Focus(focus)
	// zlog.Info("FOCUS:", v.ObjectName(), focus)
	v.Expose()
}
