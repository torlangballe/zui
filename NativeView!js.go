// +build !js

package zgo

type NativeView struct {
	View      View
	presented bool
}

func (v *NativeView) GetView() *NativeView {
	return v
}

func (v *NativeView) Rect(rect Rect) View {
	return v
}

func (v *NativeView) GetRect() Rect {
	return Rect{}
}

func (v *NativeView) GetLocalRect() Rect {
	return Rect{}
}

func (v *NativeView) LocalRect(rect Rect) {
}

func (v *NativeView) GetCalculatedSize(total Size) Size {
	return Size{10, 10}
}

func (v *NativeView) ObjectName(name string) View {
	return v
}

func (v *NativeView) GetObjectName() string {
	return ""
}

func (v *NativeView) Color(c Color) View {
	return v
}

func (v *NativeView) Alpha(alpha float32) View {
	return v
}

func (v *NativeView) GetAlpha() float32 {
	return 1
}

func (v *NativeView) BGColor(c Color) View {
	return v
}

func (v *NativeView) CornerRadius(radius float64) View {
	return v
}

func (v *NativeView) Stroke(width float64, c Color) View {
	return v
}

func (v *NativeView) Scale(scale float64) View {
	return v
}

func (v *NativeView) GetScale() float64 {
	return 1
}

func (v *NativeView) Show(show bool) View {
	return v
}

func (v *NativeView) IsShown() bool {
	return true
}

func (v *NativeView) Usable(usable bool) View {
	return v
}

func (v *NativeView) IsUsable() bool {
	return true
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
	return v
}

func (v *NativeView) GetFont() *Font {
	return nil
}

func (v *NativeView) Text(text string) View {
	return v
}

func (v *NativeView) GetText() string {
	return ""
}

func (v *NativeView) AddChild(child View, index int) {
}

func (v *NativeView) RemoveChild(child View) {
}

func (v *NativeView) SetDropShadow(deltaSize Size, blur float32, color Color) {
}
