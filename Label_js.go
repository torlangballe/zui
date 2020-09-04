package zui

import (
	"fmt"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

func LabelNew(text string) *Label {
	label := &Label{}
	label.Element = DocumentJS.Call("createElement", "label")
	// zlog.Info("Label New:", label.TextInfo.SplitItems, text)
	style := label.style()
	style.Set("position", "absolute")
	style.Set("textAlign", "left")
	style.Set("display", "block")
	style.Set("verticalAlign", "middle")
	style.Set("whiteSpace", "pre-wrap")
	//	style.Set("overflow", "hidden")
	//	style.Set("textOverflow", "clip")
	style.Set("wordWrap", "break-word")
	//	white-space: pre-wrap for multi-lines
	//	style.Set("padding-top", "3px")

	label.SetMaxLines(1)
	label.View = label
	label.SetObjectName(text)
	textNode := DocumentJS.Call("createTextNode", text)
	label.call("appendChild", textNode)
	f := FontNice(FontDefaultSize, FontStyleNormal)
	label.SetFont(f)
	return label
}

func (v *Label) SetRect(r zgeo.Rect) View {
	//	zlog.Info("Label SetRect:", v.ObjectName(), r)
	//	r.Pos.Y -= 6
	v.NativeView.SetRect(r)
	return v
}

func (v *Label) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.setjs("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnClick(v)
		return nil
		return nil
	}))
	v.setjs("class", "widget")
}

func (v *Label) SetLongPressedHandler(handler func()) {
	// zlog.Info("Label.SetLongPressedHandler:", v.ObjectName())
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

func (v *Label) SetTextAlignment(a zgeo.Alignment) View {
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
	// zlog.Info("Label SetMarg:", v.ObjectName(), m)
	style.Set("padding-top", fmt.Sprintf("%dpx", int(m.Min().Y)))
	style.Set("padding-left", fmt.Sprintf("%dpx", int(m.Min().X)))
	style.Set("padding-bottom", fmt.Sprintf("%dpx", int(m.Max().Y)))
	style.Set("padding-right", fmt.Sprintf("%dpx", int(m.Max().X)))
	return v
}
