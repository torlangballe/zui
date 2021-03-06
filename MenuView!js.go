// +build !js
// +build zui

package zui

import "github.com/torlangballe/zutil/zdict"

func MenuViewNew(name string, items zdict.Items, value interface{}) *MenuView {
	return &MenuView{}
}

// func (v *MenuView) IDAndValue() (id string, value interface{}) {
// 	return "", nil
// }

// func (v *MenuView) SetWithID(id string) *MenuView {
// 	return v
// }

func (v *MenuView) SetSelectedHandler(handler func())                       {}
func (v *MenuView) Empty()                                                  {}
func (v *MenuView) AddSeparator()                                           {}
func (v *MenuView) SetValues(items zdict.NamedValues)                       {}
func (v *MenuView) SetAndSelect(items zdict.NamedValues, value interface{}) {}
func (v *MenuView) SelectWithValue(value interface{}) *MenuView             { return v }
func (v *MenuView) SetFont(font *Font) {}

func menuViewGetHackedFontForSize(font *Font) *Font {
	return font
}

