// +build !js

package zui

import "github.com/torlangballe/zutil/zdict"

func MenuViewNew(name string, items zdict.NamedValues, value interface{}, isStatic bool) *MenuView {
	return &MenuView{}
}

func (v *MenuView) IDAndValue() (id string, value interface{}) {
	return "", nil
}

func (v *MenuView) SetWithID(id string) *MenuView {
	return v
}

func (v *MenuView) ChangedHandler(handler func(id, name string, value interface{})) {}
func (v *MenuView) UpdateValues(items zdict.NamedValues)                            {}
func (v *MenuView) Empty()                                                          {}
func (v *MenuView) AddSeparator()                                                   {}
func (v *MenuView) AddAction(id, name string)                                       {}
func (v *MenuView) updateVals(items zdict.NamedValues, value interface{})           {}
func (v *MenuView) SetValues(items zdict.NamedValues, value interface{})            {}

func menuViewGetHackedFontForSize(font *Font) *Font {
	return font
}

func (v *MenuView) SetFont(font *Font) View {
	return v
}
