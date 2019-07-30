package zgo

/*
func (v *CustomView) IsUsable() bool {
	return true
}

func (v *CustomView) GetAlpha() float32 {
	return 1
}

func (v *CustomView) Opaque(opaque bool) View {
	return v
}

func (v *CustomView) GetView() *ViewNative {
	return nil
}

func (v *CustomView) GetChild(path string) *ViewNative {
	return nil
}

func (v *CustomView) DumpTree() {
}

func (v *CustomView) RemoveFromParent() {
}

func (v *CustomView) Show(show bool) View {
	return v
}

func (v *CustomView) IsVisible() bool {
	return true
}

func (v *CustomView) IsFocused() bool {
	return true
}

func (v *CustomView) Focus(focus bool) View {
	return v
}

func (v *CustomView) GetParent() View {
	return nil
}

func (v *CustomView) Color(color Color) View {
	return v
}

func (v *CustomView) GetColor() Color {
	return Color{}
}

func (v *CustomView) BGColor(color Color) View {
	return v
}

func (v *CustomView) GetBackgroundColor() Color {
	return Color{}
}

func (v *CustomView) CornerRadius(radius float64) View {
	return v
}

func (v *CustomView) GetCornerRadius() float64 {
	return 0
}

func (v *CustomView) Stroke(width float64, color Color) View {
	return v
}

func (v *CustomView) GetStroke() (width float64, color Color) {
	return 1, Color{}
}

func (v *CustomView) Scale(scale float64) View {
	return v
}

func (v *CustomView) GetScale() float64 {
	return 1
}

func (v *CustomView) Expose(fadeIn float32) {

}

func (v *CustomView) GetCalculatedSize(total Size) Size {
	return Size{}
}

func (v *CustomView) GetRect() Rect {
	return Rect{}
}

func (v *CustomView) Rect(rect Rect) View {
	return v
}

func (v *CustomView) GetLocalRect() Rect {
	return Rect{}
}

func (v *CustomView) LocalRect(rect Rect) View {
	return v
}
*/

func (v *CustomView) Activate(activate bool) { // like being activated/deactivated for first time
}

func (v *CustomView) Rotate(degrees float64) {
	// r := MathDegToRad(degrees)
	//self.transform = CGAffineTransform(rotationAngle CGFloat(r))
}

func zConvertViewSizeThatFitstToSize(view *ViewNative, sizeIn Size) Size {
	//    return Size(view.sizeThatFits(sizeIn.GetCGSize()))
	return Size{}
}

func zViewSetRect(view *ViewNative, rect Rect, layout bool) { // layout only used on android
	//    view.frame = frame.GetCGRect()
}

func zRemoveViewFromSuper(view *ViewNative, detachFromContainer bool) {
	//view.removeFromSuperview()
}
