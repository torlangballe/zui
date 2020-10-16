// +build !js

package zui

import "github.com/torlangballe/zutil/zgeo"

type baseNativeView struct {
	View View
}

type LongPresser struct{}

func (v *NativeView) GetView() *NativeView {
	return v
}

func (v *NativeView) SetRect(rect zgeo.Rect) View {
	return v
}

func (v *NativeView) Rect() zgeo.Rect {
	return zgeo.Rect{}
}

func (v *NativeView) LocalRect() zgeo.Rect {
	return zgeo.Rect{}
}

func (v *NativeView) SetLocalRect(rect zgeo.Rect) {
}

func (v *NativeView) Parent() *NativeView {
	return nil
}

func (v *NativeView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{10, 10}
}

func (v *NativeView) SetObjectName(name string) View {
	return v
}

func (v *NativeView) ObjectName() string {
	return ""
}

func (v *NativeView) SetColor(c zgeo.Color) View {
	return v
}

func (v *NativeView) Color() zgeo.Color {
	return zgeo.Color{}
}

func (v *NativeView) SetAlpha(alpha float32) View {
	return v
}

func (v *NativeView) Alpha() float32 {
	return 1
}

func (v *NativeView) SetBGColor(c zgeo.Color) View {
	return v
}

func (v *NativeView) SetCorner(radius float64) View {
	return v
}

func (v *NativeView) SetStroke(width float64, c zgeo.Color) View {
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

func (v *NativeView) SetUsable(usable bool) View {
	return v
}

func (v *NativeView) Usable() bool {
	return true
}

func (v *NativeView) IsFocused() bool {
	return true
}

func (v *NativeView) Focus(focus bool) View {
	return v
}

func (v *NativeView) SetCanFocus(can bool) View {
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
	v.StopStoppers()
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

func (v *NativeView) Text() string {
	return ""
}
func (v *NativeView) AddChild(child View, index int)                   {}
func (v *NativeView) RemoveChild(child View)                           {}
func (v *NativeView) SetDropShadow(shadow zgeo.DropShadow)             {}
func (v *NativeView) SetToolTip(str string)                            {}
func (v *NativeView) SetAboveParent(above bool)                        {}
func NativeViewAddToRoot(v View)                                       {}
func (v *NativeView) SetScrollHandler(handler func(pos zgeo.Pos))      {}
func (v *NativeView) setjs(property string, value interface{})         {}
func (v *NativeView) SetPointerEnterHandler(handler func(inside bool)) {}
func (v *NativeView) AllParents() (all []*NativeView) {
	return
}
