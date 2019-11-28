// +build !js

package zgo

func MenuViewNew(vals Dictionary, value interface{}) *MenuView {
	return &MenuView{}
}

func (v *MenuView) ChangedHandler(handler func(key string, val interface{})) {}

func (v *MenuView) NameAndValue() (string, interface{}) {
	return "", nil
}
