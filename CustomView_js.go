package zgo

import (
	"syscall/js"
)

func (v *CustomView) PressedHandler(handler func(pos Pos)) {
	v.pressed = handler
	v.native.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
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

func (c *CustomView) init(view View, name string) {
	c.exposed = true
	vbh := ViewBaseHandler{}
	c.canvas = CanvasNew()
	v := ViewNative(c.canvas.element)
	vbh.native = &v
	vbh.view = view
	c.ViewBaseHandler = vbh // this must be set after vbh is set up
	view.ObjectName(name)
	v.set("disabled", false)
}

func (v *CustomView) drawIfExposed() {
	if v.exposed && v.draw != nil {
		v.exposeTimer.Stop()
		s := v.native.GetLocalRect().Size
		r := Rect{Size: s}
		v.draw(r, v.canvas, v.view)
		v.exposed = false
	}
}

func zViewSetRect(view View, rect Rect, layout bool) { // layout only used on android
	cv, _ := view.(CustomViewProtocol)
	if cv != nil {
		canvas := cv.getCanvas()
		dpr := WindowJS.Get("devicePixelRatio").Float()
		canvas.element.Set("width", rect.Size.W*dpr)
		canvas.element.Set("height", rect.Size.H*dpr)
		canvas.context.Call("scale", dpr, dpr)
	}
	view.GetView().Rect(rect)
}
