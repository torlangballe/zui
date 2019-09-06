package zgo

// type ViewBaseHandler struct {
// 	view    View
// 	parent  View
// 	native  *NativeView
// }

// func (l *ViewBaseHandler) GetView() *NativeView {
// 	//	fmt.Println("ViewBaseHandler GetView:", l)
// 	return l.native
// }

// func (l *ViewBaseHandler) SetNative(n *NativeView) {
// 	l.native = n
// }

// func (v *ViewBaseHandler) Parent() View {
// 	return v.parent
// }

type TextLayoutOwner interface {
	Font(font *Font) View
	GetFont() *Font
	Text(text string) View
	GetText() string
	TextAlignment(a Alignment) View
	GetTextAlignment() Alignment
	GetMinWidth() float64
	GetMaxWidth() float64
	GetMaxLines() int
	MinWidth(min float64) View
	MaxWidth(max float64) View
	MaxLines(max int) View
}

func CalculateSize(o TextLayoutOwner, total Size) Size {
	var t TextInfo
	t.Alignment = o.GetTextAlignment()
	t.Text = o.GetText()
	noWidth := false
	if o.GetMaxWidth() != 0 {
		noWidth = true
		t.Rect = Rect{Size: Size{o.GetMaxWidth(), 99999}}
	}
	t.Font = o.GetFont()
	if o.GetMaxLines() != 0 {
		t.MaxLines = o.GetMaxLines()
	}
	t.Wrap = TextInfoWrapWord
	rect := t.GetBounds(noWidth)

	rect.Size.W += 4
	return rect.Size
}
