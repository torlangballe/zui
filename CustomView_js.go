package zgo

import (
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
	r := v.GetLocalRect()
	if v.exposed && v.draw != nil && !r.Size.IsNull() { // if r.Size.IsNull(), it hasn't been caclutated yet in first ArrangeChildren
		//		println("CV drawIfExposed: " + v.GetObjectName())
		canvas := CanvasNew()
		canvas.SetRect(r)
		v.exposeTimer.Stop()
		v.draw(r, canvas, v.View)
		url := canvas.element.Call("toDataURL").String()
		v.exposed = false
		v.style().Set("backgroundImage", "url("+url+")")
		v.style().Set("background-repeat", "no-repeat")
		//		println("CV drawIfExposed end: " + v.GetObjectName() + " " + time.Since(start).String())
	}
}
