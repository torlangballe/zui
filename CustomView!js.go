// +build !js

package zui

func (v *CustomView) drawIfExposed() {
}

func (c *CustomView) init(view View, name string) {
	c.SetObjectName(name)
}

func (v *CustomView) makeCanvas() {
}

func (v *CustomView) SetPressedHandler(handler func()) {
}
