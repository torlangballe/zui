package zgo

func (v *ViewNative) Rect(rect Rect) {
}

func (v *ViewNative) GetRect() Rect {
	return Rect{}
}

func (v *ViewNative) GetLocalRect() Rect {
	return Rect{}
}

func (v *ViewNative) LocalRect(rect Rect) {
}

func (v *ViewBaseHandler) ObjectName(name string) View {
	return v.view
}

func (v *ViewBaseHandler) GetObjectName() string {
	return ""
}

func (v *ViewBaseHandler) Color(c Color) View {
	return v.view
}

func (v *ViewBaseHandler) Rect(rect Rect) View {
	return v.view
}

func (v *ViewBaseHandler) Alpha(alpha float32) View {
	return v.view
}

func (v *ViewBaseHandler) GetAlpha() float32 {
	return 1
}

func (v *ViewBaseHandler) BGColor(c Color) View {
	return v.view
}

func (v *ViewBaseHandler) CornerRadius(radius float64) View {
	return v.view
}

func (v *ViewBaseHandler) Stroke(width float64, c Color) View {
	return v.view
}

func (v *ViewBaseHandler) Scale(scale float64) View {
	return v.view
}

func (v *ViewBaseHandler) GetScale() float64 {
	return 1
}

func (v *ViewBaseHandler) Show(show bool) View {
	return v.view
}

func (v *ViewBaseHandler) IsShown() bool {
	return true
}

func (v *ViewBaseHandler) Usable(usable bool) View {
	return v.view
}

func (v *ViewBaseHandler) IsUsable() bool {
	return true
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
}

func (v *TextBaseHandler) Font(font *Font) View {
	return v.view
}

func (v *TextBaseHandler) GetFont() *Font {
	return &Font{}
}

func (v *TextBaseHandler) Text(text string) View {
	return v.view
}

func (v *TextBaseHandler) GetText() string {
	return ""
}
func (v *TextBaseHandler) TextAlignment(a Alignment) View {
	return v.view
}

func (v *TextBaseHandler) GetTextAlignment() Alignment {
	return AlignmentLeft
}

// ??? do we need this?
// func (v *ViewNative) GetChild(path string) View {
// 	return nil
// }
