package zui

import (
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

type baseCustomView struct {
	cancelClick     bool
	downClickedTime time.Time
	longTimer       *ztimer.Timer
}

func (v *CustomView) init(view View, name string) {
	v.Element = DocumentJS.Call("createElement", "div")
	v.Element.Set("style", "position:absolute")
	v.exposed = true
	v.View = view
	v.SetObjectName(name)
	v.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	v.style().Set("overflow", "hidden") // this clips the canvas, otherwise it is on top of corners etc
}

func (v *CustomView) SetLongPressedHandler(handler func()) {
	v.longPressed = handler
	v.set("className", "widget")
	v.set("onmousedown", js.FuncOf(func(js.Value, []js.Value) interface{} {
		// fmt.Println("MOUSEDOWN")
		v.downClickedTime = time.Now()
		v.longTimer = ztimer.StartIn(0.5, true, func() {
			// fmt.Println("TIMER")
			if v.longPressed != nil && v.Usable() {
				v.longPressed()
			}
			v.longTimer = nil
			v.cancelClick = true
		})
		return nil
	}))
	v.set("onmouseup", js.FuncOf(func(js.Value, []js.Value) interface{} {
		// fmt.Println("MOUSEUP")
		if v.longTimer != nil {
			v.longTimer.Stop()
		}
		return nil
	}))
}

func (v *CustomView) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.set("className", "widget")
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.longTimer != nil {
			v.longTimer.Stop()
		}
		if !v.cancelClick && v.pressed != nil && v.Usable() {
			v.pressed()
		}
		v.cancelClick = false
		return nil
	}))
}

func (v *CustomView) setCanvasSize(size zgeo.Size) {
	v.canvas.SetSize(size) // scale?
	setElementRect(v.canvas.element, zgeo.Rect{Size: size})
}

func (v *CustomView) SetRect(rect zgeo.Rect) View {
	// zlog.Info("CV SetRect:", v.ObjectName(), rect)
	v.NativeView.SetRect(rect)
	r := v.Rect()
	if rect != r {
		if v.canvas != nil {
			zlog.Debug("set", v.ObjectName(), rect)
			s := v.GetLocalRect().Size
			v.setCanvasSize(s)
			v.Expose()
		}
	}
	v.isSetup = true
	return v
}

func (v *CustomView) SetUsable(usable bool) View {
	v.NativeView.SetUsable(usable)
	v.Expose()
	// zlog.Info("CV SetUsable:", v.ObjectName(), usable, v.Usable())
	return v
}

func (v *CustomView) makeCanvas() {
	if v.canvas == nil {
		// zlog.Info("makeCanvas:", v.ObjectName())
		v.canvas = CanvasNew()
		v.call("appendChild", v.canvas.element)
		s := v.GetLocalRect().Size
		v.setCanvasSize(s)
		v.canvas.context.Call("scale", 2, 2) // this must be AFTER setElementRect, doesn't do anything!
		v.canvas.element.Get("style").Set("zIndex", 0)
	}
	// set z index!!
}

func (v *CustomView) drawIfExposed() {
	if !presentViewPresenting && v.draw != nil && v.Parent() != nil { //&& v.exposed
		// zlog.Info("CV drawIfExposed", v.ObjectName(), presentViewPresenting, v.exposed, v.draw, v.Parent() != nil)
		r := v.GetLocalRect()
		if !r.Size.IsNull() { // if r.Size.IsNull(), it hasn't been caclutated yet in first ArrangeChildren
			// println("CV drawIfExposed2:", v.ObjectName())
			v.exposeTimer.Stop()
			v.makeCanvas()
			v.canvas.ClearRect(zgeo.Rect{})
			// if !v.Usable() {
			// 	// zlog.Info("cv: push for disabled")
			// 	v.canvas.PushState()
			// 	v.canvas.context.Set("globalAlpha", 0.4)
			// }
			// zlog.Info("CV drawIfExposed", v.ObjectName(), v.Usable(), v.canvas.context.Get("globalAlpha"))
			v.draw(r, v.canvas, v.View)
			// if !v.Usable() {
			// 	v.canvas.PopState()
			// }
			v.exposed = false
			//		println("CV drawIfExposed end: " + v.ObjectName() + " " + time.Since(start).String())
		}
	}
}
