// +build zui

package zui

type TextLayoutOwner interface {
	//Font() *Font
	SetText(text string)
	//Text() string
	// SetTextAlignment(a zgeo.Alignment) View
	SetFont(font *Font)
	//TextAlignment() zgeo.Alignment
}
