//go:build zui
// +build zui

package ztext

import "github.com/torlangballe/zutil/zgeo"

type LayoutOwner interface {
	//Font() *Font
	SetText(text string)
	//Text() string
	// SetTextAlignment(a zgeo.Alignment) View
	SetFont(font *zgeo.Font)
	//TextAlignment() zgeo.Alignment
}
