package zgo

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

func LabelNew(text string) *Label {
	label := &Label{}
	label.Element = DocumentJS.Call("createElement", "label")
	label.set("style", "position:absolute")
	label.View = label
	textNode := DocumentJS.Call("createTextNode", text)
	label.call("appendChild", textNode)
	f := FontNice(FontDefaultSize, FontStyleNormal)
	label.Font(f)
	return label
}

func (v *Label) PressedHandler(handler func()) {
	v.pressed = handler
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.pressed != nil {
			v.pressed()
		}
		return nil
	}))
}

func (l *Label) TextAlignment(a zgeo.Alignment) View {
	l.alignment = a
	str := "left"
	if a&zgeo.AlignmentRight != 0 {
		str = "right"
	} else if a&zgeo.AlignmentHorCenter != 0 {
		str = "center"
	}
	l.style().Set("textAlign", str)
	return l
}
