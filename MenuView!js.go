// +build !js

package zui

import "github.com/torlangballe/zutil/zdict"

func MenuViewNew(vals zdict.Items, value interface{}) *MenuView {
	return &MenuView{}
}

func (v *MenuView) ChangedHandler(handler func(item zdict.Item)) {}

func (v *MenuView) NameAndValue() zdict.Item {
	return zdict.Item{}
}

func (v *MenuView) UpdateValues(vals zdict.Items) {}
