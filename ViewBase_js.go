package zgo

import (
	"fmt"
	"syscall/js"
)

// func (n *ViewNative) Style() css { // for now...
// 	return css(n.get("style"))
// }

func (v *ViewNative) RemoveFromParent() {

}

func (v *ViewNative) Parent() {

}

func (v *ViewNative) GetRect() Rect {
	var pos Pos
	pos.X = v.get("offsetLeft").Float()
	pos.Y = v.get("offsetTop").Float()
	size := v.GetLocalRect().Size
	return Rect{pos, size}
}

func setElementRect(e js.Value, rect Rect) {
	style := e.Get("style")
	style.Set("left", fmt.Sprintf("%fpx", rect.Pos.X))
	//	jv.GetView().style()).Set("left", fmt.Sprintf("%fpx", rect.Pos.X))
	style.Set("top", fmt.Sprintf("%fpx", rect.Pos.Y))
	style.Set("width", fmt.Sprintf("%fpx", rect.Size.W))
	style.Set("height", fmt.Sprintf("%fpx", rect.Size.H))
}

func (v *ViewNative) Rect(rect Rect) {
	setElementRect(js.Value(*v), rect)
}

func (v *ViewNative) GetLocalRect() Rect {
	style := v.style()
	w := parseCoord(style.Get("width"))
	h := parseCoord(style.Get("height"))
	return RectMake(0, 0, w, h)
}

func (v *ViewNative) LocalRect(rect Rect) {

}

func (v *ViewBaseHandler) GetObjectName() string {
	return v.view.GetView().get("id").String()
}

func makeRGBAString(c Color) string {
	rgba := c.GetRGBA()
	return fmt.Sprintf("rgba(%d,%d,%d,%g)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), rgba.A)
}

func (v *ViewBaseHandler) Color(c Color) View {
	v.view.GetView().style().Set("color", makeRGBAString(c))
	return v.view
}

func (n *ViewNative) style() js.Value {
	return n.get("style")
}

func (v *ViewNative) setUsable(usable bool) {
	v.set("disabled", !usable)
}

func (v *ViewNative) isUsable() bool {
	return !v.get("disabled").Bool()
}

func (v *ViewBaseHandler) Alpha(alpha float32) View {
	v.view.GetView().style().Set("alpha", alpha)
	return v.view
}

func (v *ViewBaseHandler) GetAlpha() float32 {
	return float32(v.view.GetView().style().Get("alpha").Float())
}

func (v *ViewBaseHandler) ObjectName(name string) View {
	v.view.GetView().set("id", name)
	return v.view
}

func (v *ViewBaseHandler) BGColor(c Color) View {
	v.view.GetView().style().Set("background", makeRGBAString(c))
	return v.view
}

func (v *ViewBaseHandler) CornerRadius(radius float64) View {
	style := v.view.GetView().style()
	s := fmt.Sprintf("%dpx", int(radius))
	style.Set("-moz-border-radius", s)
	style.Set("-webkit-border-radius", s)
	style.Set("border-radius", s)
	return v.view
}

func (v *ViewBaseHandler) Stroke(width float64, c Color) View {
	style := v.view.GetView().style()
	style.Set("border-color", makeRGBAString(c))
	style.Set("border", "solid 1px transparent")
	return v.view
}

func (v *ViewBaseHandler) Scale(scale float64) View {
	return v.view
}

func (v *ViewBaseHandler) GetScale() float64 {
	return 1
}

func (v *ViewBaseHandler) Show(show bool) View {
	str := "hidden"
	if show {
		str = "visible"
	}
	v.view.GetView().style().Set("visibility", str)
	return v.view
}

func (v *ViewBaseHandler) IsShown() bool {
	return v.view.GetView().style().Get("visibility").String() == "visible"
}

func (v *ViewBaseHandler) Usable(usable bool) View {
	v.view.GetView().setUsable(usable)
	return v.view
}

func (v *ViewBaseHandler) IsUsable() bool {
	return v.view.GetView().isUsable()
}

func (v *ViewBaseHandler) IsFocused() bool {
	return true
}

func (v *ViewBaseHandler) Focus(focus bool) View {
	return v.view
}

func (v *ViewBaseHandler) Opaque(opaque bool) View {
	return v.view
}

func (v *ViewBaseHandler) GetChild(path string) *ViewNative {
	return nil
}

func (v *ViewBaseHandler) DumpTree() {
}

func (v *ViewBaseHandler) RemoveFromParent() {
	if v.parent != nil {
		v.parent.GetView().call("removeChild", js.Value(*v.view.GetView()))
	}
}

// func (v *ViewBaseHandler) Parent() View {
// 	return v.parent
// }

// func (v *ViewBaseHandler) GetLocalRect() Rect {
// 	w := v.view.GetView().get("offsetHeight").Float()
// 	h := v.view.GetView().get("offsetHeight").Float()
// 	return Rect{Size: Size{w, h}}
// }

// func (v *ViewBaseHandler) LocalRect(rect Rect) View {
// 	w := v.view.GetView().get("width").Float()
// 	h := v.view.GetView().get("height").Float()
// 	style := v.view.GetView().style()
// 	style.Set("width", fmt.Sprintf("%fpx", w))
// 	style.Set("height", fmt.Sprintf("%fpx", h))
// 	return v.view
// }

func (v *TextBaseHandler) Font(font *Font) View {
	style := v.view.GetView().style()
	style.Set("font-family", font.Name)
	style.Set("font-size", fmt.Sprintf("%gpx", font.Size))
	return v.view
}

func (v *TextBaseHandler) GetFont() *Font {
	style := v.view.GetView().style()
	name := style.Get("font-family").String()
	size := parseCoord(style.Get("font-size"))
	return FontNew(name, size, FontNormal)
}

func (v *TextBaseHandler) Text(text string) View {
	v.view.GetView().set("innerHTML", text)
	return v.view
}

func (v *TextBaseHandler) GetText() string {
	text := v.view.GetView().get("innerHTML").String()
	return text
}
func (v *TextBaseHandler) TextAlignment(a Alignment) View {
	return v.view
}

func (v *TextBaseHandler) GetTextAlignment() Alignment {
	return AlignmentLeft
}
