package zlabel

import (
	"fmt"
	"strings"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func (label *Label) InitAsLink(view zview.View, title, surl string, newWindow bool) {
	label.MakeJSElement(view, "a")
	label.init(title)
	label.SetURL(surl, newWindow)
}

func (label *Label) Init(view zview.View, text string) {
	label.MakeJSElement(view, "label")
	label.init(text)
}

func (label *Label) init(text string) {
	// zlog.Info("Label New:", label.textInfo.SplitItems, text)
	style := label.JSStyle()
	style.Set("textAlign", "left")
	style.Set("lineHeight", "1")
	style.Set("whiteSpace", "preWrap")
	label.pressWithModifierToClipboard = -1
	label.SetColor(zstyle.DefaultFGColor())
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
	setPadding(v)
}

func setPadding(v *Label) {
	pad := v.margin

	pad.Pos.Y++
	if zdevice.CurrentWasmBrowser != zdevice.Safari {
		pad.Pos.Y++
	}
	v.SetNativePadding(pad)
}

func (v *Label) SetBGColor(c zgeo.Color) {
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
		style.Set("overflow", "hidden")
	}
	// v.SetStyleForAllPlatforms("line-clamp", fmt.Sprint(max))
}

func (v *Label) SetRect(r zgeo.Rect) {
	r.Size.H--
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
}

func (v *Label) SetMargin(m zgeo.Rect) {
	v.margin = m
	setPadding(v)
}

func (v *Label) SetDropShadow(shadow ...zstyle.DropShadow) {
	var parts []string
	for _, s := range shadow {
		str := fmt.Sprintf("%dpx %dpx %dpx %s ", int(s.Delta.W), int(s.Delta.H), int(s.Blur), zdom.MakeRGBAString(s.Color))
		parts = append(parts, str)
	}
	v.SetJSStyle("textShadow", strings.Join(parts, ", "))
}

func (v *Label) OutsideDropStroke(delta float64, col zgeo.Color) {
	var drops []zstyle.DropShadow
	for x := -delta; x <= delta; x++ {
		for y := -delta; y <= delta; y++ {
			drop := zstyle.MakeDropShadow(x, y, 0, col)
			drops = append(drops, drop)
		}
	}
	v.SetDropShadow(drops...)
}
