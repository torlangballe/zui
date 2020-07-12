package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

func (v *CustomView) init(view View, name string) {
	v.Element = DocumentJS.Call("createElement", "div")
	v.Element.Set("style", "position:absolute")
	v.exposed = true
	v.View = view
	v.SetObjectName(name)
	v.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	// v.style().Set("overflow", "hidden") // this clips the canvas, otherwise it is on top of corners etc
}

func (v *CustomView) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.set("className", "widget")
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnClick(v)
		return nil
	}))
}

func (v *CustomView) SetLongPressedHandler(handler func()) {
	// zlog.Info("SetLongPressedHandler:", v.ObjectName())
	v.longPressed = handler
	v.set("className", "widget")
	v.set("onmousedown", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnMouseDown(v)
		return nil
	}))
	v.set("onmouseup", js.FuncOf(func(js.Value, []js.Value) interface{} {
		// fmt.Println("MOUSEUP")
		(&v.LongPresser).HandleOnMouseUp(v)
		return nil
	}))
}

func (v *CustomView) PressedHandler() func() {
	return v.pressed
}

func (v *CustomView) LongPressedHandler() func() {
	return v.longPressed
}

func (v *CustomView) setCanvasSize(size zgeo.Size, scale float64) {
	s := size.TimesD(scale)
	// if v.ObjectName() == "provider" {
	// zlog.Info("setCanvasSize:", v.ObjectName(), s, zlog.GetCallingStackString())
	// }
	v.canvas.SetSize(s) // scale?
	// zlog.Info("setCanvasSize:", v.ObjectName(), scale)
	v.canvas.context.Call("scale", scale, scale) // this must be AFTER setElementRect, doesn't do anything!
	//!!!!!!!! v.canvas.element.Get("style").Set("zIndex", 0)
	setElementRect(v.canvas.element, zgeo.Rect{Size: size})
}

func (v *CustomView) SetRect(rect zgeo.Rect) View {

	s := zgeo.Size{}
	if v.HasSize() {
		s = v.Rect().Size
	}
	v.NativeView.SetRect(rect)
	// r := v.Rect()
	// if v.ObjectName() == "streams" {
	// 	zlog.Info("CV SetRect:", v.ObjectName(), rect, r)
	// }
	// if rect != r {
	if v.canvas != nil && s != rect.Size {
		// zlog.Debug("set", v.ObjectName(), s, rect.Size)
		s := v.GetLocalRect().Size
		scale := ScreenMain().Scale
		v.setCanvasSize(s, scale)
		v.Expose()
	}
	// }
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
		v.canvas = CanvasNew()
		v.call("appendChild", v.canvas.element)
		s := v.GetLocalRect().Size
		scale := ScreenMain().Scale
		v.setCanvasSize(s, scale)
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
