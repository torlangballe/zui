package zcustom

import (
	"syscall/js"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

// TODO: store pressed/logpressed js function, and release when adding new one

var IsPresentingFunc func() bool

func (v *CustomView) Init(view zview.View, name string) {
	stype := "div"
	zstr.HasPrefix(name, "#type:", &stype)
	v.MakeJSElement(view, stype)
	v.SetObjectName(name)
	v.SetJSStyle("overflow", "hidden")
	zview.SetUserSelect(&v.NativeView, "auto")
	v.exposeTimer = ztimer.TimerNew()
	v.exposed = true
}

// SetPressedHandler is not in NativeView, as it needs to work with LongPresser struct storing when pressed etc to generate a long pressed action
func (v *CustomView) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.JSSet("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// zlog.Info("Pressed:", v.Hierarchy(), v.Usable())
		if !v.Usable() {
			return nil
		}
		zlog.Assert(len(args) > 0)
		event := args[0]
		event.Call("stopPropagation")
		if !event.Get("target").Equal(v.Element) {
			return nil
		}
		v.SetStateOnDownPress(event)
		(&v.LongPresser).HandleOnClick(v)
		return nil
	}))
}

func (v *CustomView) SetLongPressedHandler(handler func()) {
	v.longPressed = handler
	v.JSSet("onmousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		v.SetStateOnDownPress(args[0])
		(&v.LongPresser).HandleOnMouseDown(v)
		args[0].Call("preventDefault")
		return nil
	}))
	v.JSSet("onmouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		(&v.LongPresser).HandleOnMouseUp(v)
		args[0].Call("preventDefault")
		return nil
	}))
}

func (v *CustomView) setCanvasSize(size zgeo.Size, scale float64) {
	s := size.TimesD(scale)
	v.canvas.SetSize(s)                              // scale?
	v.canvas.JSContext().Call("scale", scale, scale) // this must be AFTER SetElementRectmentRect, doesn't do anything!
	zview.SetElementRect(v.canvas.JSElement(), zgeo.Rect{Size: size})
}

func (v *CustomView) ReadyToShow(beforeWindow bool) {
	if beforeWindow {
		return
	}
	if v.draw != nil {
		// doing SetHandleExpose AFTER window is created important for observer to be on the new window.
		v.SetHandleExposed(func(intersectsViewport bool) {
			if intersectsViewport && v.exposed {
				v.visible = true
				if v.draw != nil {
					v.drawSelf()
				}
			}
			v.visible = intersectsViewport
		})
	}
}

func (v *CustomView) SetRect(rect zgeo.Rect) {
	r := rect.ExpandedToInt()
	s := zgeo.Size{}
	if v.HasSize() {
		s = v.Rect().Size
	}
	// zlog.Info("CV SetRect", r, rect)
	v.NativeView.SetRect(r)
	if v.canvas != nil && s != r.Size {
		s := v.LocalRect().Size
		scale := zscreen.MainScale
		v.setCanvasSize(s, scale)
		v.Expose()
	}
	v.isSetup = true
}

func (v *CustomView) SetUsable(usable bool) {
	v.NativeView.SetUsable(usable)
	v.Expose()
}

func (v *CustomView) MC() {
	v.makeCanvas()
}

func (v *CustomView) makeCanvas() {
	if v.canvas != nil {
		return
	}
	v.canvas = zcanvas.New()
	// v.canvas.JSElement().Set("id", v.ObjectName()+".canvas")
	v.canvas.DownsampleImages = v.DownsampleImages
	firstChild := v.JSGet("firstChild")
	v.canvas.JSElement().Get("style").Set("zIndex", 1)
	s := v.LocalRect().Size
	scale := zscreen.MainScale
	v.setCanvasSize(s, scale)

	if firstChild.IsUndefined() {
		v.JSCall("appendChild", v.canvas.JSElement())
	} else {
		v.JSCall("insertBefore", v.canvas.JSElement(), firstChild)
	}
}

func (v *CustomView) drawSelf() {
	// v.canvas.SetColor(zgeo.ColorRandom())
	// v.canvas.FillRect(v.LocalRect())
	if !v.drawing && !IsPresentingFunc() && v.draw != nil && v.Parent() != nil && v.HasSize() { //&& v.exposed
		v.drawing = true
		r := v.LocalRect()
		if !r.Size.IsNull() { // if r.Size.IsNull(), it hasn't been caclutated yet in first ArrangeChildren
			v.exposeTimer.Stop()
			v.makeCanvas()
			if !v.OpaqueDraw {
				v.canvas.Clear()
			}
			// if v.ObjectName() == "sort" {
			// 	zlog.Info(v.ObjectName(), "CV drawSelf: draw")
			// }
			v.draw(r, v.canvas, v.View)
		}
		v.drawing = false
	}
	v.exposed = false
}

var count int

func (v *CustomView) ExposeIn(secs float64) {
	// if v.ObjectName() == "sort" {
	// 	zlog.Info(v.ObjectName(), "CV Expose:", v.draw, v.visible)
	// }
	if v.draw == nil {
		return
	}
	if v.visible {
		if v.exposeTimer.IsRunning() {
			return
		}
		count++
		v.drawSelf()
		v.exposed = false
	} else {
		v.exposed = true
	}
}

func (v *CustomView) Expose() {
	v.ExposeIn(0.1)
}
