package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

// TODO: store pressed/logpressed js function, and release when adding new one

func (v *CustomView) Init(view View, name string) {
	v.MakeJSElement(view, "div")
	v.SetObjectName(name)
	v.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	// v.style().Set("overflow", "hidden") // this clips the canvas, otherwise it is on top of corners etc
}

func (v *CustomView) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.setjs("className", "widget")
	v.setjs("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		zlog.Assert(len(args) > 0)
		// target := args[0].Get("target") // is it not event?
		// zlog.Info("Pressed:", v.ObjectName(), target.Get("id"), this.Equal(target), v.canvas != nil)
		// if v.canvas != nil {
		// 	zlog.Info("Eq:", v.canvas.element.Equal(target))
		// }
		// zlog.Info("Pressed", v.ObjectName(), this.Equal(target), v.canvas != nil && v.canvas.element.Equal(target))
		//		if this.Equal(target) || (v.canvas != nil && v.canvas.element.Equal(target)) {
		(&v.LongPresser).HandleOnClick(v)
		event := args[0]
		event.Call("stopPropagation")
		//		}
		// }
		return nil
	}))
}

func (v *CustomView) SetLongPressedHandler(handler func()) {
	// zlog.Info("SetLongPressedHandler:", v.ObjectName())
	v.longPressed = handler
	v.setjs("className", "widget")
	v.setjs("onmousedown", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnMouseDown(v)
		return nil
	}))
	v.setjs("onmouseup", js.FuncOf(func(js.Value, []js.Value) interface{} {
		// fmt.Println("MOUSEUP")
		(&v.LongPresser).HandleOnMouseUp(v)
		return nil
	}))
}

func (v *CustomView) setCanvasSize(size zgeo.Size, scale float64) {
	s := size.TimesD(scale)
	// if v.ObjectName() == "provider" {
	// zlog.Info("setCanvasSize:", v.ObjectName(), s, zlog.GetCallingStackString())
	// }
	v.canvas.SetSize(s) // scale?
	// zlog.Info("setCanvasSize:", v.ObjectName(), scale)
	v.canvas.context.Call("scale", scale, scale) // this must be AFTER setElementRect, doesn't do anything!
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
		s := v.LocalRect().Size
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
		v.canvas.element.Set("id", v.ObjectName()+".canvas")
		v.call("appendChild", v.canvas.element)
		s := v.LocalRect().Size
		scale := ScreenMain().Scale
		v.setCanvasSize(s, scale)
	}
	// set z index!!
}

func (v *CustomView) drawIfExposed() {
	if !presentViewPresenting && v.draw != nil && v.Parent() != nil { //&& v.exposed
		// zlog.Info("CV drawIfExposed", v.ObjectName(), presentViewPresenting, v.exposed, v.draw, v.Parent() != nil)
		r := v.LocalRect()
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
			v.draw(r, v.canvas, v.View)
			// if !v.Usable() {
			// 	v.canvas.PopState()
			// }
			v.exposed = false
			//		println("CV drawIfExposed end: " + v.ObjectName() + " " + time.Since(start).String())
		}
	}
}
