package zgo

import (
	"syscall/js"
)

func LabelNew(text string) *Label {
	label := &Label{}
	label.Element = DocumentJS.Call("createElement", "label")
	label.set("style", "position:absolute")
	label.View = label
	textNode := DocumentJS.Call("createTextNode", text)
	label.call("appendChild", textNode)
	f := FontNice(18, FontStyleNormal)
	label.Font(f)
	return label
}

func (v *Label) PressedHandler(handler func(pos Pos)) {
	v.pressed = handler
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.pressed != nil {
			v.pressed(Pos{})
		}
		return nil
	}))
}
