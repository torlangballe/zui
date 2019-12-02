package zgo

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

func (v *CustomView) init(view View, name string) {
	v.Element = DocumentJS.Call("createElement", "div")
	v.Element.Set("style", "position:absolute")
	v.exposed = true
	v.View = view
	v.ObjectName(name)
	v.Font(FontNice(FontDefaultSize, FontStyleNormal))
}

func (v *CustomView) PressedHandler(handler func()) {
	v.pressed = handler
	v.set("className", "widget")
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.pressed != nil {
			v.pressed()
		}
		return nil
	}))
}

func (v *CustomView) Rect(rect zgeo.Rect) View {
	v.NativeView.Rect(rect)
	if !v.isSetup {
		if v.canvas != nil {
			s := v.GetLocalRect().Size
			setElementRect(v.canvas.element, zgeo.Rect{Size: s})
			v.canvas.element.Set("width", s.W*2) // scale?
			v.canvas.element.Set("height", s.H*2)
			v.canvas.context.Call("scale", 2, 2)
		}
	}
	v.isSetup = true
	return v
}

func (v *CustomView) makeCanvas() {
	v.canvas = CanvasNew()
	v.call("appendChild", v.canvas.element)
	// set z index!!
}

func (v *CustomView) drawIfExposed() {
	// fmt.Println("CV drawIfExposed", v.GetObjectName(), presentViewPresenting, v.exposed, v.draw, v.Parent() != nil)
	if !presentViewPresenting && v.draw != nil && v.Parent() != nil { //&& v.exposed
		r := v.GetLocalRect()
		if !r.Size.IsNull() { // if r.Size.IsNull(), it hasn't been caclutated yet in first ArrangeChildren
			// println("CV drawIfExposed2:", v.GetObjectName())
			v.exposeTimer.Stop()
			v.canvas.ClearRect(zgeo.Rect{})
			v.draw(r, v.canvas, v.View)
			v.exposed = false
			//		println("CV drawIfExposed end: " + v.GetObjectName() + " " + time.Since(start).String())
		}
	}
}
