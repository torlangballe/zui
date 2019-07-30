package zgo

import (
	"fmt"
	"strconv"
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
	return RectFromPosSize(pos, size)
}

func (v *ViewNative) Rect(rect Rect) {
	style := v.style()
	style.Set("left", fmt.Sprintf("%fpt", rect.Pos.X))
	//	jv.GetView().style()).Set("left", fmt.Sprintf("%fpt", rect.Pos.X))
	style.Set("top", fmt.Sprintf("%fpt", rect.Pos.Y))
	style.Set("width", fmt.Sprintf("%fpt", rect.Size.W))
	style.Set("height", fmt.Sprintf("%fpt", rect.Size.H))
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
	return v.GetView().get("id").String()
}

func makeRGBA(c Color) string {
	rgba := c.GetRGBA()
	return fmt.Sprintf("rgba(%d,%d,%d,%d)", int(rgba.R*255), int(rgba.G*255), int(rgba.B*255), int(rgba.A*255))
}

func (v *ViewBaseHandler) Color(c Color) View {
	v.GetView().style().Set("color", makeRGBA(c))
	return v
}

func (n *ViewNative) style() js.Value {
	return n.get("style")
}

func (v *ViewNative) setUsable(usable bool) {
	v.set("disabled", !usable)
}

func (v *ViewNative) isUsable() bool {
	return v.get("disabled").Bool()
}

func (v *ViewBaseHandler) Alpha(alpha float32) View {
	v.GetView().style().Set("alpha", alpha)
	return v
}

func (v *ViewBaseHandler) GetAlpha() float32 {
	return float32(v.GetView().style().Get("alpha").Float())
}

func (v *ViewBaseHandler) ObjectName(name string) ViewSimple {
	v.GetView().set("id", name)
	return v
}

func (v *ViewBaseHandler) BGColor(c Color) View {
	v.GetView().style().Set("background", makeRGBA(c))
	return v
}

func (v *ViewBaseHandler) CornerRadius(radius float64) View {
	style := v.GetView().style()
	s := fmt.Sprintf("%dpt", int(radius))
	style.Set("-moz-border-radius", s)
	style.Set("-webkit-border-radius", s)
	style.Set("border-radius", s)
	return v
}

func (v *ViewBaseHandler) Stroke(width float64, c Color) View {
	style := v.GetView().style()
	style.Set("border-color", makeRGBA(c))
	style.Set("border", "solid 1px transparent")
	return v
}

func (v *ViewBaseHandler) Scale(scale float64) View {
	return v
}

func (v *ViewBaseHandler) GetScale() float64 {
	return 1
}

func (v *ViewBaseHandler) Show(show bool) View {
	str := "hidden"
	if show {
		str = "visible"
	}
	v.GetView().style().Set("visibility", str)
	return v
}

func (v *ViewBaseHandler) IsShown() bool {
	return v.GetView().style().Get("visibility").String() == "visible"
}

func (v *ViewBaseHandler) Usable(usable bool) View {
	v.native.setUsable(usable)
	return v
}

func (v *ViewBaseHandler) IsUsable() bool {
	return v.native.isUsable()
}

func (v *ViewBaseHandler) IsFocused() bool {
	return true
}

func (v *ViewBaseHandler) Focus(focus bool) View {
	return v
}

func (v *ViewBaseHandler) Opaque(opaque bool) View {
	return v
}

func (v *ViewBaseHandler) GetChild(path string) *ViewNative {
	return nil
}

func (v *ViewBaseHandler) DumpTree() {
}

func (v *ViewBaseHandler) RemoveFromParent() {
	if v.parent != nil {
		v.parent.GetView().call("removeChild", js.Value(*v.native))
	}
}

// func (v *ViewBaseHandler) Parent() View {
// 	return v.parent
// }

func (v *ViewBaseHandler) GetCalculatedSize(total Size) Size {
	fmt.Println("ViewBaseHandler GetCalculatedSize")
	return Size{10, 10}
}

func (v *ViewBaseHandler) GetLocalRect() Rect {
	w := v.GetView().get("offsetHeight").Float()
	h := v.GetView().get("offsetHeight").Float()
	return RectFromSize(Size{w, h})
}

func (v *ViewBaseHandler) LocalRect(rect Rect) View {
	w := v.GetView().get("width").Float()
	h := v.GetView().get("height").Float()
	style := v.GetView().style()
	style.Set("width", fmt.Sprintf("%fpt", w))
	style.Set("height", fmt.Sprintf("%fpt", h))
	return v
}

func (v *TextBaseHandler) Font(font *Font) View {
	style := v.view.GetView().style()
	style.Set("font-family", font.Name)
	style.Set("font-size", fmt.Sprintf("%gpx", font.Size))
	return v.view
}

func (v *TextBaseHandler) GetFont() *Font {
	style := v.view.GetView().style()
	name := style.Get("font-family").String()
	ss := style.Get("font-size").String()
	size, _ := strconv.ParseFloat(ss, 32)
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
