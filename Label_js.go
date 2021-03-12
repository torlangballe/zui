package zui

import (
	"fmt"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

func LabelNew(text string) *Label {
	v := &Label{}
	v.Element = DocumentJS.Call("createElement", "label")
	// zlog.Info("Label New:", v.TextInfo.SplitItems, text)
	style := v.style()
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

	v.View = v
	v.SetObjectName(text)
	v.SetMaxLines(1)
	textNode := DocumentJS.Call("createTextNode", text)
	v.call("appendChild", textNode)
	f := FontNice(FontDefaultSize, FontStyleNormal)
	v.SetFont(f)
	return v
}

func (v *Label) SetMaxLines(max int) View {
	// zlog.Info("Label.SetMaxLines:", max, v.ObjectName())
	v.maxLines = max
	style := v.style()
	if max == 1 {
		// style.Set("overflow", "hidden")
		// style.Set("display", "inline-block")
		// zlog.Info("Label.SetMaxLines here!")
		style.Set("text-overflow", "ellipsis")
		style.Set("white-space", "nowrap")

	} else {
		// zlog.Info("Label.SetMaxLines here2!")
		style.Set("text-overflow", "initial")
		style.Set("white-space", "normal")
	}
	//	style.Set("textOverflow", "clip")
	//	white-space: pre-wrap for multi-lines

	return v
}

func (v *Label) SetRect(r zgeo.Rect) {
	//	zlog.Info("Label SetRect:", v.ObjectName(), r)
	//	r.Pos.Y -= 6
	v.NativeView.SetRect(r)
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
