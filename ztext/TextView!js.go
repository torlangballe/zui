//go:build !js && zui

package ztext

import (
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

func (tv *TextView) Init(view zview.View, text string, style Style, rows, cols int) {
	tv.View = view
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	tv.SetFont(f)
}

func (v *TextView) SetTextAlignment(a zgeo.Alignment) {
	v.alignment = a
}

// func (v *TextView) IsReadOnly(is bool) *TextView {
// 	return v
// }

// func (v *TextView) IsPassword(is bool) *TextView {
// 	return v
// }

func (v *TextView) SetChangedHandler(handler func())                                            {}
func (v *TextView) SetKeyHandler(handler func(key zkeyboard.Key, mods zkeyboard.Modifier) bool) {}
func (v *TextView) ScrollToBottom()                                                             {}
func (v *TextView) SetIsStatic(s bool)                                                          {}
func (v *TextView) Select(from, to int)                                                         {}
func (v *TextView) SetPlaceholder(str string)                                                   {}
func (v *TextView) SetMargin(m zgeo.Rect)                                                       {}
