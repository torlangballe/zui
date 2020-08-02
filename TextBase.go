package zui

type TextLayoutOwner interface {
	//Font() *Font
	SetText(text string) View
	//Text() string
	// SetTextAlignment(a zgeo.Alignment) View
	SetFont(font *Font) View
	//TextAlignment() zgeo.Alignment
}
