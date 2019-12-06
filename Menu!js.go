// +build !js

package zgo

import "github.com/torlangballe/zutil/zdict"

func MenuViewNew(vals zdict.Dict, value interface{}) *MenuView {
	return &MenuView{}
}

func (v *MenuView) ChangedHandler(handler func(key string, val interface{})) {}

func (v *MenuView) NameAndValue() (string, interface{}) {
	return "", nil
}
