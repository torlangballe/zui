package zgo

import "syscall/js"

func LabelNew(text string) *Label {
	label := &Label{}
	vbh := ViewBaseHandler{}
	tbh := TextBaseHandler{}
	e := DocumentJS.Call("createElement", "label")
	e.Set("style", "position:absolute")
	v := ViewNative(e)
	vbh.native = &v
	vbh.view = label
	tbh.view = label
	label.TextBaseHandler = tbh // this must be set after tbh is set up
	label.ViewBaseHandler = vbh // this must be set after vbh is set up
	textNode := DocumentJS.Call("createTextNode", text)
	e.Call("appendChild", textNode)
	f := FontNice(18, FontNormal)
	label.Font(f)
	return label
}

func (v *Label) PressedHandler(handler func(pos Pos)) {
	v.pressed = handler
	v.native.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.pressed != nil {
			v.pressed(Pos{})
		}
		return nil
	}))
}