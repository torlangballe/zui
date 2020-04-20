package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

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
	SetFont(font *Font) View
	Font() *Font
	SetText(text string) View
	Text() string
	SetTextAlignment(a zgeo.Alignment) View
	TextAlignment() zgeo.Alignment
	MinWidth() float64
	MaxWidth() float64
	MaxLines() int
	SetMinWidth(min float64) View
	SetMaxWidth(max float64) View
	SetMaxLines(max int) View
}

func TextLayoutOwnerCalculateSize(o TextLayoutOwner) zgeo.Size {
	return TextLayoutCalculateSize(o.TextAlignment(), o.Font(), o.Text(), o.MaxLines(), o.MaxWidth())
}

func TextLayoutCalculateSize(alignment zgeo.Alignment, font *Font, text string, maxLines int, maxWidth float64) zgeo.Size {
	var t TextInfo
	t.Alignment = alignment
	t.Text = text
	noWidth := false
	if maxWidth != 0 {
		noWidth = true
		t.Rect = zgeo.Rect{Size: zgeo.Size{maxWidth, 99999}}
	}
	t.Font = font
	if maxLines != 0 {
		t.MaxLines = maxLines
	}
	t.Wrap = TextInfoWrapWord
	size := t.GetBounds(noWidth)
	// fmt.Println("TextLayoutCalculateSize:", maxWidth, rect, text, font.Size, font.Name)

	size.W += 4
	return size
}
