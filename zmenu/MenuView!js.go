//go:build !js

package zmenu

import (
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
)

func NewView(name string, items zdict.Items, value interface{}) *MenuView {
	return &MenuView{}
}

func (v *MenuView) SetSelectedHandler(handler func())                       {}
func (v *MenuView) Empty()                                                  {}
func (v *MenuView) AddSeparator()                                           {}
func (v *MenuView) SetValues(items zdict.NamedValues)                       {}
func (v *MenuView) SetAndSelect(items zdict.NamedValues, value interface{}) {}
func (v *MenuView) SelectWithValue(value interface{}) *MenuView             { return v }
func (v *MenuView) SetFont(font *zgeo.Font)                                 {}
func (v *MenuView) AddItem(name string, value interface{})                  {}
func (v *MenuView) RemoveItemByValue(value interface{})                     {}
func (v *MenuView) ChangeNameForValue(name string, value interface{})       {}

func menuViewGetHackedFontForSize(font *zgeo.Font) *zgeo.Font {
	return font
}
