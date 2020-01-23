package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func (v *CustomView) init(view View, name string) {
	v.Element = DocumentJS.Call("createElement", "div")
	v.Element.Set("style", "position:absolute")
	v.exposed = true
	v.View = view
	v.SetObjectName(name)
	v.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
}

func (v *CustomView) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.set("className", "widget")
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.pressed != nil {
			v.pressed()
		}
		return nil
	}))
}

func (v *CustomView) setCanvasSize(size zgeo.Size) {
	v.canvas.element.Set("width", size.W*2) // scale?
	v.canvas.element.Set("height", size.H*2)
	setElementRect(v.canvas.element, zgeo.Rect{Size: size})
}

func (v *CustomView) SetRect(rect zgeo.Rect) View {
	v.NativeView.SetRect(rect)
	r := v.Rect()
	if rect != r {
		if v.canvas != nil {
			zlog.Debug("set", v.GetObjectName(), rect)
			s := v.GetLocalRect().Size
			v.setCanvasSize(s)
			v.Expose()
		}
	}
	v.isSetup = true
	return v
}

func (v *CustomView) makeCanvas() {
	if v.canvas == nil {
		// fmt.Println("makeCanvas:", v.GetObjectName())
		v.canvas = CanvasNew()
		v.call("appendChild", v.canvas.element)
		s := v.GetLocalRect().Size
		v.setCanvasSize(s)
		v.canvas.context.Call("scale", 2, 2) // this must be AFTER setElementRect, doesn't do anything!
	}
	// set z index!!
}

func (v *CustomView) drawIfExposed() {
	// fmt.Println("CV drawIfExposed", v.GetObjectName(), presentViewPresenting, v.exposed, v.draw, v.Parent() != nil)
	if !presentViewPresenting && v.draw != nil && v.Parent() != nil { //&& v.exposed
		r := v.GetLocalRect()
		if !r.Size.IsNull() { // if r.Size.IsNull(), it hasn't been caclutated yet in first ArrangeChildren
			// println("CV drawIfExposed2:", v.GetObjectName())
			v.exposeTimer.Stop()
			v.makeCanvas()
			v.canvas.ClearRect(zgeo.Rect{})
			v.draw(r, v.canvas, v.View)
			v.exposed = false
			//		println("CV drawIfExposed end: " + v.GetObjectName() + " " + time.Since(start).String())
		}
	}
}
