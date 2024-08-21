//go:build zui

package ztext

import (
	"github.com/torlangballe/zutil/zgeo"
)

type TextOwner interface {
	//Font() *Font
	Text() string
	SetText(text string)
	//Text() string
	// SetTextAlignment(a zgeo.Alignment) View
	SetFont(font *zgeo.Font)
	//TextAlignment() zgeo.Alignment
}
