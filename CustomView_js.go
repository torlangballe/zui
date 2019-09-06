package zgo

import (
	"fmt"
	"syscall/js"
)

func (v *CustomView) PressedHandler(handler func(pos Pos)) {
	v.pressed = handler
	v.set("className", "widget")
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.pressed != nil {
			v.pressed(Pos{})
		}
		return nil
	}))
}

func CustomViewNew(name string) *CustomView {
	c := &CustomView{}
	c.init(c, name)
	return c
}

// func (v *CustomView) Init(view View) {
// 	v.Element = DocumentJS.Call("createElement", "div")
// 	v.Element.Set("style", "position:absolute")
// 	v.View = view
// 	fmt.Printf("CustomView init %p\n", v)
// }

func (v *CustomView) init(view View, name string) {
	v.Element = DocumentJS.Call("createElement", "div")
	v.Element.Set("style", "position:absolute")
	v.exposed = true
	v.View = view
	v.ObjectName(name)
}

func (v *CustomView) drawIfExposed() {
	if v.exposed && v.draw != nil {
		fmt.Printf("drawIfExposed %p %s\n", v, v.GetObjectName())
		v.exposeTimer.Stop()
		s := v.GetLocalRect().Size
		r := Rect{Size: s}
		v.draw(r, v.canvas, v.View)
		v.exposed = false
		fmt.Println("drawIfExposed2", v.GetObjectName())
	}
}

func (v *CustomView) Rect(rect Rect) View {
	if v.canvas != nil {
		dpr := WindowJS.Get("devicePixelRatio").Float()
		v.canvas.element.Set("width", rect.Size.W*dpr)
		v.canvas.element.Set("height", rect.Size.H*dpr)
		v.canvas.context.Call("scale", dpr, dpr)
	}
	setElementRect(v.Element, rect)
	return v
}
