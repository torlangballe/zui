package zlabel

import (
	"syscall/js"

	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func (label *Label) InitAsLink(view zview.View, title, surl string) {
	label.MakeJSElement(view, "a")
	label.init(title)
	label.SetURL(surl, false)
}
func (label *Label) Init(view zview.View, text string) {
	label.MakeJSElement(view, "label")
	label.init(text)
}

func (label *Label) init(text string) {
	// zlog.Info("Label New:", label.textInfo.SplitItems, text)
	style := label.JSStyle()
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

	label.SetColor(zstyle.DefaultFGColor())
	// zlog.Info("LABCOL:", zstyle.DefaultFGColor)
	//	label.SetColor(zgeo.ColorRed)
	label.SetObjectName(text)
	label.SetMaxLines(1)
	label.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if down || km.Modifier != 0 || label.PressedHandler() == nil {
			return false
		}
		if km.Key.IsReturnish() {
			label.PressedHandler()()
			return true
		}
		return false
	})
	// textNode := zdom.DocumentJS.Call("createTextNode", "")
	// label.JSCall("appendChild", textNode)
	label.SetText(text)
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	label.SetFont(f)
}

func (label *Label) SetURL(surl string, newWindow bool) {
	label.JSSet("href", surl)
	label.SetUsable(surl != "")
	if newWindow {
		label.JSSet("target", "_blank")
		label.JSSet("rel", "noopener noreferrer")
		return
	}
	label.JSSet("target", "")
	label.JSSet("rel", "")
}

func (v *Label) SetText(text string) {
	v.text = text
	v.NativeView.SetText(text)
}

func setPadding(v *Label) {
	v.SetNativePadding(v.padding)
}

func (v *Label) SetBGColor(c zgeo.Color) {
	x := 0.0
	if c.Valid && c.Opacity() != 0 {
		x = 4
	}
	v.padding.SetMinX(x)
	v.padding.SetMaxX(-x)
	setPadding(v)
	v.NativeView.SetBGColor(c)
}

func (v *Label) SetFont(font *zgeo.Font) {
	v.NativeView.SetFont(font)
	v.padding.SetMinY(font.Size / 8)
	setPadding(v)
}

func (v *Label) SetWrap(wrap ztextinfo.WrapType) {
	zlog.Assert(wrap == ztextinfo.WrapTailTruncate)
	style := v.JSStyle()
	if wrap == ztextinfo.WrapTailTruncate {
		style.Set("textOverflow", "ellipsis")
	}
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
	// zlog.Info("label.SetPressedHandler:", v.Hierarchy())
	v.JSSet("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		zlog.Info("label.Pressed:", v.Hierarchy())
		e := args[0]
		v.SetStateOnDownPress(e)
		(&v.LongPresser).HandleOnClick(v)
		return nil
	}))
	v.JSSet("class", "widget")
}

func (v *Label) SetPressedDownHandler(handler func()) {
	v.pressed = handler
	v.JSSet("onmousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		v.SetStateOnDownPress(e)
		handler()
		e.Call("preventDefault")
		e.Call("stopPropagation")
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
}
