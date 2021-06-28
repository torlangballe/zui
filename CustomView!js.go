// +build !js
// +build zui

package zui

func (v *CustomView) drawIfExposed() {
}

func (c *CustomView) Init(view View, name string) {
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
