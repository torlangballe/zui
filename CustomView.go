//go:build zui
// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
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
	valueChanged     func()
	// pointerEnclosed func(inside bool)
	draw    func(rect zgeo.Rect, canvas *Canvas, view View)
	exposed bool
	visible bool
	drawing bool
	// color         zgeo.Color
	exposeTimer   *ztimer.Timer
	isSetup       bool
	isHighlighted bool
	PressedPos    zgeo.Pos
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

func (v *CustomView) PressedHandler() func() {
	return v.pressed
}

func (v *CustomView) LongPressedHandler() func() {
	return v.longPressed
}

func (v *CustomView) SetColor(c zgeo.Color) {
	v.NativeView.SetColor(c)
	//	v.color = c
	v.Expose()
}

// func (v *CustomView) Color() zgeo.Color {
// 	return v.color
// }

func (v *CustomView) SetMinSize(s zgeo.Size) {
	v.minSize = s
}

func (v *CustomView) MinSize() zgeo.Size {
	return v.minSize
}

func (v *CustomView) SetValueHandler(handler func()) {
	v.valueChanged = handler
}

func (v *CustomView) SetDrawHandler(handler func(rect zgeo.Rect, canvas *Canvas, view View)) {
	v.draw = handler
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
