package zgo

import (
	"fmt"
	"syscall/js"
)

type NativeView struct {
	Element js.Value
	View    View
}

func (v *NativeView) Parent() *NativeView {
	return nil
}

func (v *NativeView) GetNative() *NativeView {
	return v
}

func (v *NativeView) GetRect() Rect {
	var pos Pos
	pos.X = v.Element.Get("offsetLeft").Float()
	pos.Y = v.Element.Get("offsetTop").Float()
	size := v.GetLocalRect().Size
	return Rect{pos, size}
}

func setElementRect(e js.Value, rect Rect) {
	style := e.Get("style")
	style.Set("left", fmt.Sprintf("%fpx", rect.Pos.X))
	style.Set("top", fmt.Sprintf("%fpx", rect.Pos.Y))
	style.Set("width", fmt.Sprintf("%fpx", rect.Size.W))
	style.Set("height", fmt.Sprintf("%fpx", rect.Size.H))
}

func (v *NativeView) GetCalculatedSize(total Size) Size {
	return Size{10, 10}
}

func (v *NativeView) Rect(rect Rect) View {
	setElementRect(v.Element, rect)
	return v
}

func (v *NativeView) GetLocalRect() Rect {
	var w, h float64
	style := v.style()
	sw := style.Get("width")
	sh := style.Get("height")
	if sw.String() != "" {
		h = parseCoord(sh)
		w = parseCoord(sw)
	} else {
		println("parse empty Coord: " + v.GetObjectName())
	}

	return RectMake(0, 0, w, h)
}

func (v *NativeView) LocalRect(rect Rect) {

}

func (v *NativeView) GetObjectName() string {
	return v.get("id").String()
}

func makeRGBAString(c Color) string {
	rgba := c.GetRGBA()
	return fmt.Sprintf("rgba(%d,%d,%d,%g)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), rgba.A)
}

func (v *NativeView) Color(c Color) View {
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

func (v *NativeView) BGColor(c Color) View {
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

func (v *NativeView) Stroke(width float64, c Color) View {
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

func (v *NativeView) Font(font *Font) View {
	//	str := getFontStyle(font)
	//	v.set("font", str)
	style := v.style()
	style.Set("font-family", font.Name)
	style.Set("font-size", fmt.Sprintf("%gpx", font.Size))
	return v
}

func (v *NativeView) GetFont() *Font {
	style := v.style()
	name := style.Get("font-family").String()
	ss := style.Get("font-size")
	size := parseCoord(ss)
	return FontNew(name, size, FontStyleNormal)
}

func (v *NativeView) Text(text string) View {
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
}
