// +build zui

package zui

import "github.com/torlangballe/zutil/zgeo"

type TextLayoutOwner interface {
	//Font() *Font
	SetText(text string)
	//Text() string
	// SetTextAlignment(a zgeo.Alignment) View
	SetFont(font *zgeo.Font)
	//TextAlignment() zgeo.Alignment
}
