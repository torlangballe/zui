package zlabel

import (
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

	label.pressWithModifierToClipboard = -1
	label.SetColor(zstyle.DefaultFGColor())
	// zlog.Info("LABCOL:", zstyle.DefaultFGColor)
	//	label.SetColor(zgeo.ColorRed)
	label.wrap = ztextinfo.WrapNone
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
	setPadding(label)
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
	if v.Text() == text {
		return
	}
	v.text = text
	v.NativeView.SetText(text)
}

func setPadding(v *Label) {
	pad := v.margin
	pad.Pos.Y += v.Font().Size / 8

	v.SetNativePadding(pad)
}

func (v *Label) SetBGColor(c zgeo.Color) {
	// x := 0.0   this padding stuff messes with layout in table
	// if c.Valid && c.Opacity() != 0 {
	// 	x = 4
	// }
	// v.padding.SetMinX(x)
	// v.padding.SetMaxX(-x)
	// setPadding(v)
	v.NativeView.SetBGColor(c)
}

func (v *Label) SetFont(font *zgeo.Font) {
	v.NativeView.SetFont(font)
	setPadding(v)
}

func (v *Label) SetWrap(wrap ztextinfo.WrapType) {
	zlog.Assert(wrap == ztextinfo.WrapTailTruncate)
	style := v.JSStyle()
	v.wrap = wrap
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
	v.NativeView.SetRect(r)
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
	setPadding(v)
}
