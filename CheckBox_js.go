package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

func CheckBoxNew(on zbool.BoolInd) *CheckBox {
	c := &CheckBox{}
	c.Element = DocumentJS.Call("createElement", "input")
	c.setjs("style", "position:absolute")
	c.setjs("type", "checkbox")
	c.style().Set("margin-top", "4px")
	c.SetCanFocus(true)
	c.View = c
	c.SetValue(on)
	return c
}

func (v *CheckBox) SetRect(rect zgeo.Rect) {
	rect.Pos.Y -= 4
	v.NativeView.SetRect(rect)
}

func (c *CheckBox) SetValueHandler(handler func()) {
	c.valueChanged = handler
	c.setjs("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if c.valueChanged != nil {
			c.valueChanged()
		}
		return nil
	}))
}

func (c *CheckBox) Value() zbool.BoolInd {
	i := c.getjs("indeterminate").Bool()
	if i {
		return zbool.Unknown
	}
	b := c.getjs("checked").Bool()
	return zbool.ToBoolInd(b)
}

func (c *CheckBox) SetValue(b zbool.BoolInd) {
	if b.IsUndetermined() {
		c.setjs("indeterminate", true)
	} else {
		c.setjs("checked", b.Bool())
	}
}
