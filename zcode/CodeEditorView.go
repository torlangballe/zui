//go:build zui

package zcode

import (
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zutil/zgeo"
)

type CodeEditorView struct {
	ztext.TextView
}

func NewEditorView(text string, cols, rows int) *CodeEditorView {
	v := &CodeEditorView{}
	v.TextView.Init(v, text, ztext.Style{}, cols, rows)
	v.SetBGColor(zgeo.ColorNewGray(0.3, 1))
	v.SetColor(zgeo.ColorNewGray(0.8, 1))
	v.SetObjectName("editor2")
	v.SetMargin(zgeo.RectFromXY2(10, 10, -10, -10))
	font := zgeo.FontNew("Lucida Console, Monaco, monospace", 14, zgeo.FontStyleNormal)
	v.SetFont(font)
	return v
}
