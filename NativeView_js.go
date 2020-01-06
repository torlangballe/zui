package zui

import (
	"fmt"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

type NativeView struct {
	Element   js.Value
	View      View
	presented bool
}

func (v *NativeView) Parent() *NativeView {
	e := v.get("parentElement")
	if e.Type() == js.TypeUndefined || e.Type() == js.TypeNull {
		return nil
	}
	//	fmt.Println("ParentElement:", v.GetObjectName(), e)
	n := &NativeView{}
	n.Element = e
	n.View = v
	return n
}

func (v *NativeView) GetNative() *NativeView {
	return v
}

func (v *NativeView) SetRect(rect zgeo.Rect) View {
	// fmt.Println("NV Rect", v.GetObjectName())
	setElementRect(v.Element, rect)
	return v
}

func (v *NativeView) Rect() zgeo.Rect {
	var pos zgeo.Pos
	style := v.style()
	// pos.X = v.Element.Get("offsetLeft").Float()
	// pos.Y = v.Element.Get("offsetTop").Float()
	pos.X = parseElementCoord(style.Get("left"))
	pos.Y = parseElementCoord(style.Get("top"))
	size := v.GetLocalRect().Size
	return zgeo.Rect{pos, size}
}

func setElementRect(e js.Value, rect zgeo.Rect) {
	style := e.Get("style")
	style.Set("left", fmt.Sprintf("%fpx", rect.Pos.X))
	style.Set("top", fmt.Sprintf("%fpx", rect.Pos.Y))
	style.Set("width", fmt.Sprintf("%fpx", rect.Size.W))
	style.Set("height", fmt.Sprintf("%fpx", rect.Size.H))
}

func (v *NativeView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{10, 10}
}

func (v *NativeView) GetLocalRect() zgeo.Rect {
	var w, h float64
	style := v.style()
	sw := style.Get("width")
	sh := style.Get("height")
	if sw.String() != "" {
		h = parseElementCoord(sh)
		w = parseElementCoord(sw)
	} else {
		println("parse empty Coord: " + v.GetObjectName())
		panic("parse empty Coord")
	}

	return zgeo.RectMake(0, 0, w, h)
}

func (v *NativeView) LocalRect(rect zgeo.Rect) {

}

func (v *NativeView) GetObjectName() string {
	return v.get("id").String()
}

func makeRGBAString(c zgeo.Color) string {
	rgba := c.GetRGBA()
	return fmt.Sprintf("rgba(%d,%d,%d,%g)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), rgba.A)
}

func (v *NativeView) Color(c zgeo.Color) View {
	v.style().Set("color", makeRGBAString(c))
	return v
}

func (n *NativeView) style() js.Value {
	return n.get("style")
}

func (v *NativeView) setUsable(usable bool) {
	v.set("disabled", !usable)
}

func (v *NativeView) isUsable() bool {
	dis := v.Element.Get("disabled")
	if dis.Type() == js.TypeUndefined {
		return true
	}
	return dis.Bool()
}

func (v *NativeView) Alpha(alpha float32) View {
	v.style().Set("alpha", alpha)
	return v
}

func (v *NativeView) GetAlpha() float32 {
	return float32(v.style().Get("alpha").Float())
}

func (v *NativeView) ObjectName(name string) View {
	v.set("id", name)
	return v
}

func (v *NativeView) SetBGColor(c zgeo.Color) View {
	v.style().Set("background", makeRGBAString(c))
	return v
}

func (v *NativeView) CornerRadius(radius float64) View {
	style := v.style()
	s := fmt.Sprintf("%dpx", int(radius))
	style.Set("-moz-border-radius", s)
	style.Set("-webkit-border-radius", s)
	style.Set("border-radius", s)
	return v
}

func (v *NativeView) Stroke(width float64, c zgeo.Color) View {
	style := v.style()
	style.Set("border-color", makeRGBAString(c))
	style.Set("border", "solid 1px transparent")
	return v
}

func (v *NativeView) Scale(scale float64) View {
	return v
}

func (v *NativeView) GetScale() float64 {
	return 1
}

func (v *NativeView) Show(show bool) View {
	str := "hidden"
	if show {
		str = "visible"
	}
	v.style().Set("visibility", str)
	return v
}

func (v *NativeView) IsShown() bool {
	return v.style().Get("visibility").String() == "visible"
}

func (v *NativeView) Usable(usable bool) View {
	v.setUsable(usable)
	return v
}

func (v *NativeView) IsUsable() bool {
	return v.isUsable()
}

func (v *NativeView) IsFocused() bool {
	return true
}

func (v *NativeView) Focus(focus bool) View {
	return v
}

func (v *NativeView) CanFocus(can bool) View {
	v.set("tabindex", "0")
	return v
}

func (v *NativeView) Opaque(opaque bool) View {
	return v
}

func (v *NativeView) GetChild(path string) *NativeView {
	return nil
}

func (v *NativeView) DumpTree() {
}

func (v *NativeView) RemoveFromParent() {
}

func (v *NativeView) SetFont(font *Font) View {
	cssStyle := v.style()
	cssStyle.Set("font-style", string(font.Style&FontStyleItalic))
	// zlog.Debug("font:", v.GetObjectName(), font.Style&FontStyleItalic, font.Style&FontStyleBold)
	cssStyle.Set("font-weight", (font.Style & FontStyleBold).String())
	cssStyle.Set("font-family", font.Name)
	cssStyle.Set("font-size", fmt.Sprintf("%gpx", font.Size))
	return v
}

func (v *NativeView) Font() *Font {
	cssStyle := v.style()
	fstyle := FontStyleNormal
	name := cssStyle.Get("font-family").String()
	if cssStyle.Get("font-weight").String() == "bold" {
		fstyle |= FontStyleBold
	}
	if cssStyle.Get("font-style").String() == "italic" {
		fstyle |= FontStyleItalic
	}
	ss := cssStyle.Get("font-size")
	size := parseElementCoord(ss)

	return FontNew(name, size, fstyle)
}

func (v *NativeView) SetText(text string) View {
	v.set("innerHTML", text)
	return v
}

func (v *NativeView) GetText() string {
	text := v.get("innerHTML").String()
	return text
}

func (v *NativeView) AddChild(child View, index int) {
	o, _ := child.(NativeViewOwner)
	if o == nil {
		panic("NativeView AddChild child not native")
	}
	n := o.GetNative()
	v.call("appendChild", n.Element)
}

func (v *NativeView) RemoveChild(child View) {
	o, _ := child.(NativeViewOwner)
	if o == nil {
		panic("NativeView AddChild child not native")
	}
	v.call("removeChild", o.GetNative().Element)
}

func (v *NativeView) SetDropShadow(deltaSize zgeo.Size, blur float32, color zgeo.Color) {
	str := fmt.Sprintf("%dpx %dpx %dpx %s", int(deltaSize.W), int(deltaSize.H), int(blur), makeRGBAString(color))
	v.style().Set("boxShadow", str)
}

func (v *NativeView) SetToolTip(str string) {
	v.set("title", str)
}

func (v *NativeView) Child(path string) View {
	return ViewChild(v.View, path)
}
