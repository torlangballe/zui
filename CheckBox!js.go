// +build !js

package zui

func (c *CheckBox) GetValue() BoolInd {
	return BoolFalse
}

func CheckBoxNew(on BoolInd) *CheckBox {
	c := &CheckBox{}
	return c
}

func (c *CheckBox) ValueHandler(handler func(view View)) {}

func (c *CheckBox) Value(b BoolInd) *CheckBox {
	return c
}
