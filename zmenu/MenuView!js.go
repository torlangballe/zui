//go:build !js

package zmenu

import (
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
)

func NewView(name string, items zdict.Items, value any) *MenuView {
	return &MenuView{}
}

func (v *MenuView) SetSelectedHandler(handler func()) {}
func (v *MenuView) Empty()                            {}
func (v *MenuView) AddSeparator()                     {}

func (v *MenuView) SelectWithValue(value any) bool                          { return true }
func (v *MenuView) SetFont(font *zgeo.Font)                                 {}
func (v *MenuView) AddItem(name string, value any)                          {}
func (v *MenuView) RemoveItemByValue(value any)                             {}
func (v *MenuView) ChangeNameForValue(name string, value any)               {}
func (v *MenuView) UpdateItems(items zdict.Items, value any, isAction bool) {}

func menuViewGetHackedFontForSize(font *zgeo.Font) *zgeo.Font {
	return font
}
