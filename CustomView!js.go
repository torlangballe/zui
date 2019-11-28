// +build !js

package zgo

func (v *CustomView) drawIfExposed() {
}

func (c *CustomView) init(view View, name string) {
	c.ObjectName(name)
}

func (v *CustomView) makeCanvas() {
}

func (v *CustomView) PressedHandler(handler func()) {
}
