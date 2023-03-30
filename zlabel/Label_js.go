package zlabel

import (
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"syscall/js"
)

func New(text string) *Label {
	v := &Label{}
	v.Element = zdom.DocumentJS.Call("createElement", "label")
	// zlog.Info("Label New:", v.textInfo.SplitItems, text)
	style := v.JSStyle()
	style.Set("position", "absolute")
	style.Set("textAlign", "left")
	style.Set("display", "block")
	// style.Set("verticalAlign", "middle")
	style.Set("whiteSpace", "preWrap")
	// style.Set("white-space", "pre-wrap")
	//	style.Set("overflow", "hidden")
	//	style.Set("textOverflow", "clip")
	// style.Set("wordWrap", "break-word")
	//	white-space: pre-wrap for multi-lines
	//	style.Set("padding-top", "3px")

	v.SetColor(zstyle.DefaultFGColor())
	v.View = v
	// zlog.Info("LABCOL:", zstyle.DefaultFGColor)
	//	v.SetColor(zgeo.ColorRed)
	v.SetObjectName(text)
	v.SetMaxLines(1)
	textNode := zdom.DocumentJS.Call("createTextNode", text)
	v.JSCall("appendChild", textNode)
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	v.SetFont(f)
	return v
}

func (v *Label) SetWrap(wrap ztextinfo.WrapType) {
	zlog.Assert(wrap == ztextinfo.WrapTailTruncate)
	style := v.JSStyle()
	// style.Set("textOverflow", "ellipsis")
	style.Set("display", "inline-block")
	style.Set("overflow", "hidden")
	style.Set("whiteSpace", "nowrap")
}

func (v *Label) SetMaxLines(max int) {
	// zlog.Info("Label.SetMaxLines:", max, v.ObjectName())
	v.maxLines = max
	style := v.JSStyle()
	if max == 1 {
		// style.Set("overflow", "hidden")
		// style.Set("display", "inline-block")
		// zlog.Info("Label.SetMaxLines here!")
		// style.Set("text-overflow", "ellipsis")
		style.Set("white-space", "nowrap")

	} else {
		// zlog.Info("Label.SetMaxLines here2!")
		style.Set("text-overflow", "initial")
		style.Set("whiteSpace", "pre-wrap")
	}
	//	style.Set("textOverflow", "clip")
	//	white-space: pre-wrap for multi-lines
}

func (v *Label) SetRect(r zgeo.Rect) {
	r.Add(v.margin) // we need to inset margin, as padding (which margin is set as) is outside of this rect.
	r.Pos.Y -= 1
	v.NativeView.SetRect(r)
}

func (v *Label) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.JSSet("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnClick(v)
		return nil
		return nil
	}))
	v.JSSet("class", "widget")
}

func (v *Label) SetLongPressedHandler(handler func()) {
	// zlog.Info("Label.SetLongPressedHandler:", v.ObjectName())
	v.longPressed = handler
	v.JSSet("className", "widget")
	v.JSSet("onmousedown", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnMouseDown(v)
		return nil
	}))
	v.JSSet("onmouseup", js.FuncOf(func(js.Value, []js.Value) interface{} {
		// fmt.Println("MOUSEUP")
		(&v.LongPresser).HandleOnMouseUp(v)
		return nil
	}))
}

func (v *Label) SetTextAlignment(a zgeo.Alignment) {
	str := "left"
	if a&zgeo.Right != 0 {
		str = "right"
	} else if a&zgeo.HorCenter != 0 {
		str = "center"
	}
	v.JSStyle().Set("textAlign", str)
	str = "middle"
	if a&zgeo.Top != 0 {
		str = "top"
	} else if a&zgeo.Bottom != 0 {
		str = "bottom"
	}
	v.JSStyle().Set("verticalAlign", str)
}

func (v *Label) SetMargin(m zgeo.Rect) {
	v.margin = m
	v.NativeView.SetMargin(m)
}
