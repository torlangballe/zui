//go:build !js && zui

package zcustom

import "github.com/torlangballe/zui"

func (v *CustomView) drawSelf() {
}

func (c *CustomView) Init(view zui.View, name string) {
	c.SetObjectName(name)
}

func (v *CustomView) makeCanvas() {
}

func (v *CustomView) SetPressedHandler(handler func()) {
	v.pressed = handler
}

func (v *CustomView) SetLongPressedHandler(handler func()) {
	v.longPressed = handler
}

func (v *CustomView) Expose() {
}
