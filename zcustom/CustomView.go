//go:build zui

package zcustom

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

type CustomView struct {
	zview.NativeView
	OpaqueDraw       bool // if OpaqueDraw set, drawing does not clear canvas first, assuming total coverage during draw
	DownsampleImages bool
	canvas           *zcanvas.Canvas
	minSize          zgeo.Size
	pressed          func()
	longPressed      func()
	valueChanged     func()
	draw             func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View)
	exposed          bool
	visible          bool
	drawing          bool
	exposeTimer      *ztimer.Timer
	isSetup          bool
	isHighlighted    bool
}

func NewView(name string) *CustomView {
	c := &CustomView{}
	c.Init(c, name)
	return c
}

func (v *CustomView) DebugInfo() string {
	return zstr.Concat(" ", v.Hierarchy(), v.exposed, v.drawing, v.exposed)
}

func (v *CustomView) SetVisible(visible bool) {
	v.visible = visible
}

func (v *CustomView) SetExposed(exp bool) {
	v.exposed = exp
}

func (v *CustomView) Canvas() *zcanvas.Canvas {
	v.makeCanvas()
	return v.canvas
}

func (v *CustomView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	return v.MinSize(), zgeo.Size{}
}

func (v *CustomView) SetHighlighted(h bool) {
	v.isHighlighted = h
	v.Expose()
}

func (v *CustomView) IsHighlighted() bool {
	return v.isHighlighted
}

func (v *CustomView) PressedHandler() func() {
	// zlog.Info("cv.PressedHandler:", v != nil)
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

func (v *CustomView) SetDrawHandler(handler func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View)) {
	v.draw = handler
}

func (v *CustomView) DrawHandler() func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	return v.draw
}

func (v *CustomView) GetPosFromMe(pos zgeo.Pos, inview zview.View) zgeo.Pos {
	return zgeo.Pos{}
}

func (v *CustomView) GetPosToMe(pos zgeo.Pos, inview zview.View) zgeo.Pos {
	return zgeo.Pos{}
}

func (v *CustomView) GetViewsRectInMyCoordinates(view zview.View) zgeo.Rect {
	return zgeo.Rect{}
}

func zConvertViewSizeThatFitstToSize(view *zview.NativeView, sizeIn zgeo.Size) zgeo.Size {
	//    return Size(view.sizeThatFits(sizeIn.GetCGSize()))
	return zgeo.SizeNull
}

func (v *CustomView) GetStateColor(col zgeo.Color) zgeo.Color {
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
