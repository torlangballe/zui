package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

// TODO: store pressed/logpressed js function, and release when adding new one

func (v *CustomView) Init(view View, name string) {
	stype := "div"
	zstr.HasPrefix(name, "#type:", &stype)
	// zlog.Info("CUSTOMVIEW:", stype)
	v.MakeJSElement(view, stype)
	v.SetObjectName(name)
	v.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	v.exposeTimer = ztimer.TimerNew()
	v.exposed = true
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
		event := args[0]
		v.PressedPos.X = event.Get("offsetX").Float()
		v.PressedPos.Y = event.Get("offsetY").Float()
		(&v.LongPresser).HandleOnClick(v)
		// zlog.Info("Pressed", v.ObjectName(), v.PressedPos)

		_, KeyboardModifiersAtPress = getKeyAndModsFromEvent(event)
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

func (v *CustomView) SetRect(rect zgeo.Rect) {

	r := rect.ExpandedToInt()
	s := zgeo.Size{}
	if v.HasSize() {
		s = v.Rect().Size
	}
	v.NativeView.SetRect(r)
	// r := v.Rect()
	// if v.ObjectName() == "streams" {
	// 	zlog.Info("CV SetRect:", v.ObjectName(), rect, r)
	// }
	// if rect != r {
	if v.canvas != nil && s != r.Size {
		// zlog.Debug("set", v.ObjectName(), s, r.Size)
		s := v.LocalRect().Size
		scale := zscreen.MainScale
		v.setCanvasSize(s, scale)
		v.Expose()
	}
	// }
	v.isSetup = true
}

func (v *CustomView) SetUsable(usable bool) {
	v.NativeView.SetUsable(usable)
	v.Expose()
	// zlog.Info("CV SetUsable:", v.ObjectName(), usable, v.Usable())
}

func (v *CustomView) makeCanvas() {
	if v.canvas == nil {
		v.canvas = CanvasNew()
		v.canvas.element.Set("id", v.ObjectName()+".canvas")
		v.canvas.DownsampleImages = v.DownsampleImages
		firstChild := v.getjs("firstChild")
		v.canvas.element.Get("style").Set("zIndex", 1)
		s := v.LocalRect().Size
		scale := zscreen.MainScale
		v.setCanvasSize(s, scale)

		if firstChild.IsUndefined() {
			v.call("appendChild", v.canvas.element)
		} else {
			v.call("insertBefore", v.canvas.element, firstChild)
		}
	}
	// set z index!!
}

func (v *CustomView) drawSelf() {
	// zlog.Info("CustV drawIfExposed", v.ObjectName(), presentViewPresenting, v.exposed, v.draw)     //, zlog.GetCallingStackString())
	if !v.drawing && !presentViewPresenting && v.draw != nil && v.Parent() != nil && v.HasSize() { //&& v.exposed
		v.drawing = true
		r := v.LocalRect()
		// zlog.Info("CV drawIfExposed", v.ObjectName(), presentViewPresenting, v.exposed, v.draw, v.Parent() != nil, !r.Size.IsNull())
		if !r.Size.IsNull() { // if r.Size.IsNull(), it hasn't been caclutated yet in first ArrangeChildren
			// println("CV drawIfExposed2:", v.ObjectName())
			v.exposeTimer.Stop()
			v.makeCanvas()
			if !v.OpaqueDraw {
				v.canvas.Clear()
			}
			// if !v.Usable() {
			// 	// zlog.Info("cv: push for disabled")
			// 	v.canvas.PushState()
			// 	v.canvas.context.Set("globalAlpha", 0.4)
			// }
			v.draw(r, v.canvas, v.View)
			// if !v.Usable() {
			// 	v.canvas.PopState()
			// }
			//!!!			v.exposed = false // we move this to below, is some case where drawing can never happen if presenting or something, and exposed is never cleared
			//		println("CV drawIfExposed end: " + v.ObjectName() + " " + time.Since(start).String())
		}
		v.drawing = false
	} else {
		// zlog.Info("CustV NOT drawIfExposed", v.ObjectName(), presentViewPresenting, v.exposed, v.draw) //, zlog.GetCallingStackString())
	}
	v.exposed = false
}

func (v *CustomView) Expose() {
	// iv, _ := v.View.(*ImageView)
	// if iv != nil {
	// zlog.Info("CustV Expose", v.visible, v.Hierarchy(), v.exposed, v.draw, presentViewPresenting, "hs:", v.HasSize())
	// }
	if v.visible {
		v.exposeTimer.StartIn(0.1, func() {
			go v.drawSelf()
		})
		v.exposed = false
	} else {
		v.exposed = true
	}
}
