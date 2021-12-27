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
	v.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
	v.exposeTimer = ztimer.TimerNew()
	v.exposed = true
	// v.style().Set("overflow", "hidden") // this clips the canvas, otherwise it is on top of corners etc
}

func (v *CustomView) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.setjs("className", "widget")
	v.setjs("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		zlog.Assert(len(args) > 0)
		event := args[0]
		// zlog.Info("Pressed", reflect.ValueOf(v).Type(), v.ObjectName(), event.Get("target").Get("id").String())
		if !event.Get("target").Equal(v.Element) {
			return nil
		}
		v.PressedPos.X = event.Get("offsetX").Float()
		v.PressedPos.Y = event.Get("offsetY").Float()
		_, KeyboardModifiersAtPress = getKeyAndModsFromEvent(event)
		(&v.LongPresser).HandleOnClick(v)
		event.Call("stopPropagation")
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
	v.canvas.SetSize(s) // scale?
	// zlog.Info("setCanvasSize:", v.ObjectName(), scale)
	v.canvas.context.Call("scale", scale, scale) // this must be AFTER setElementRect, doesn't do anything!
	setElementRect(v.canvas.element, zgeo.Rect{Size: size})
}

func (v *CustomView) ReadyToShow(beforeWindow bool) {
	if beforeWindow {
		return
	}
	v.SetHandleExposed(func(intersectsViewport bool) {
		if intersectsViewport && v.exposed {
			go v.drawSelf()
		}
		v.visible = intersectsViewport
	})

}
func (v *CustomView) SetRect(rect zgeo.Rect) {
	r := rect.ExpandedToInt()
	s := zgeo.Size{}
	if v.HasSize() {
		s = v.Rect().Size
	}
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
			v.exposeTimer.Stop()
			v.makeCanvas()
			if !v.OpaqueDraw {
				v.canvas.Clear()
			}
			v.draw(r, v.canvas, v.View)
		}
		v.drawing = false
	}
	v.exposed = false
}

func (v *CustomView) Expose() {
	if v.visible {
		v.exposeTimer.StartIn(0.1, func() {
			go v.drawSelf()
		})
		v.exposed = false
	} else {
		v.exposed = true
	}
}
