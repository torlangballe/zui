package zui

import (
	"fmt"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

func LabelNew(text string) *Label {
	label := &Label{}
	label.Element = DocumentJS.Call("createElement", "label")
	style := label.style()
	style.Set("position", "absolute")
	style.Set("textAlign", "left")
	style.Set("display", "block")
	style.Set("whiteSpace", "nowrap")
	style.Set("overflow", "hidden")
	style.Set("textOverflow", "clip")
	//	style.Set("padding-top", "3px")

	label.maxLines = 1
	label.View = label
	label.SetObjectName(text)
	textNode := DocumentJS.Call("createTextNode", text)
	label.call("appendChild", textNode)
	f := FontNice(FontDefaultSize, FontStyleNormal)
	label.SetFont(f)
	return label
}

func (v *Label) SetRect(r zgeo.Rect) View {
	//	fmt.Println("Label SetRect:", v.ObjectName(), r)
	//	r.Pos.Y -= 6
	v.NativeView.SetRect(r)
	return v
}

func (v *Label) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.pressed != nil {
			v.pressed()
		}
		return nil
	}))
}

func (v *Label) SetTextAlignment(a zgeo.Alignment) View {
	v.alignment = a
	str := "left"
	if a&zgeo.Right != 0 {
		str = "right"
	} else if a&zgeo.HorCenter != 0 {
		str = "center"
	}
	v.style().Set("textAlign", str)
	return v
}

func (v *Label) SetMargin(m zgeo.Rect) *Label {
	v.margin = m
	style := v.style()
	fmt.Println("Label SetMarg:", v.ObjectName(), m)
	style.Set("padding-top", fmt.Sprintf("%dpx", int(m.Min().Y)))
	style.Set("padding-left", fmt.Sprintf("%dpx", int(m.Min().X)))
	style.Set("padding-bottom", fmt.Sprintf("%dpx", int(m.Max().Y)))
	style.Set("padding-right", fmt.Sprintf("%dpx", int(m.Max().X)))
	return v
}
