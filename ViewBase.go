package zgo

import "fmt"

type ViewBaseHandler struct {
	View
	native *ViewNative
	parent *ViewBaseHandler
}

func (l *ViewBaseHandler) GetView() *ViewNative {
	//	fmt.Println("ViewBaseHandler GetView:", l)
	return l.native
}

func (l *ViewBaseHandler) SetNative(n *ViewNative) {
	l.native = n
}

func (v *ViewBaseHandler) Parent() View {
	return v.parent
}

type TextBase interface {
	Font(font *Font) View
	GetFont() *Font
	Text(text string) View
	GetText() string
	TextAlignment(a Alignment) View
	GetTextAlignment() Alignment
}

type TextBaseHandler struct {
	view     View
	MinWidth float64
	MaxWidth float64
	MaxLines int
}

func (v *TextBaseHandler) GetCalculatedSize(total Size) Size {
	var t TextInfo

	fmt.Println("TextBaseHandler GetCalculatedSize")
	t.Alignment = v.GetTextAlignment()
	t.Text = v.GetText()
	noWidth := false
	if v.MaxWidth != 0 {
		noWidth = true
		t.Rect = RectFromSize(Size{v.MaxWidth, 99999})
	}
	t.Font = v.GetFont()
	if v.MaxLines != 0 {
		t.MaxLines = v.MaxLines
	}
	t.Wrap = TextInfoWrapWord
	rect := t.GetBounds(noWidth)

	fmt.Println("CalcSize:", t.Text, t.Font.Size)
	rect.Size.W += 4
	return rect.Size
}
