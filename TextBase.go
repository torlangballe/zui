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

func TextLayoutOwnerCalculateSize(o TextLayoutOwner) Size {
	return TextLayoutCalculateSize(o.GetTextAlignment(), o.GetFont(), o.GetText(), o.GetMaxLines(), o.GetMaxWidth())
}

func TextLayoutCalculateSize(alignment Alignment, font *Font, text string, maxLines int, maxWidth float64) Size {
	var t TextInfo
	t.Alignment = alignment
	t.Text = text
	noWidth := false
	if maxWidth != 0 {
		noWidth = true
		t.Rect = Rect{Size: Size{maxWidth, 99999}}
	}
	t.Font = font
	if maxLines != 0 {
		t.MaxLines = maxLines
	}
	t.Wrap = TextInfoWrapWord
	rect := t.GetBounds(noWidth)
	// fmt.Println("TextLayoutCalculateSize:", rect, text, font.Size, font.Name)

	rect.Size.W += 4
	return rect.Size
}
