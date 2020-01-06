// +build !js

package zui

import "github.com/torlangballe/zutil/zgeo"

type NativeView struct {
	View      View
	presented bool
}

func (v *NativeView) GetView() *NativeView {
	return v
}

func (v *NativeView) SetRect(rect zgeo.Rect) View {
	return v
}

func (v *NativeView) Rect() zgeo.Rect {
	return zgeo.Rect{}
}

func (v *NativeView) GetLocalRect() zgeo.Rect {
	return zgeo.Rect{}
}

func (v *NativeView) LocalRect(rect zgeo.Rect) {
}

func (v *NativeView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{10, 10}
}

func (v *NativeView) ObjectName(name string) View {
	return v
}

func (v *NativeView) GetObjectName() string {
	return ""
}

func (v *NativeView) Color(c zgeo.Color) View {
	return v
}

func (v *NativeView) Alpha(alpha float32) View {
	return v
}

func (v *NativeView) GetAlpha() float32 {
	return 1
}

func (v *NativeView) SetBGColor(c zgeo.Color) View {
	return v
}

func (v *NativeView) CornerRadius(radius float64) View {
	return v
}

func (v *NativeView) Stroke(width float64, c zgeo.Color) View {
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

func (v *NativeView) SetFont(font *Font) View {
	return v
}

func (v *NativeView) Font() *Font {
	return nil
}

func (v *NativeView) SetText(text string) View {
	return v
}

func (v *NativeView) GetText() string {
	return ""
}
func (v *NativeView) AddChild(child View, index int)                                    {}
func (v *NativeView) RemoveChild(child View)                                            {}
func (v *NativeView) SetDropShadow(deltaSize zgeo.Size, blur float32, color zgeo.Color) {}
func (v *NativeView) SetToolTip(str string)                                             {}
