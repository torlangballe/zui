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

func (v *CustomView) SetLongPressedHandler(handler func()) {
}

func (v *CustomView) PressedHandler() func() {
	return nil
}

func (v *CustomView) LongPressedHandler() func() {
	return nil
}
