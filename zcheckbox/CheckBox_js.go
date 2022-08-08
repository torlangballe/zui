//go:build zui

package zcheckbox

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

func New(on zbool.BoolInd) *CheckBox {
	c := &CheckBox{}
	c.Element = zdom.DocumentJS.Call("createElement", "input")
	c.JSSet("style", "position:absolute")
	c.JSSet("type", "checkbox")
	c.JSStyle().Set("margin-top", "4px")
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
	c.JSSet("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if c.valueChanged != nil {
			c.valueChanged()
		}
		return nil
	}))
}

func (c *CheckBox) Value() zbool.BoolInd {
	i := c.JSGet("indeterminate").Bool()
	if i {
		return zbool.Unknown
	}
	b := c.JSGet("checked").Bool()
	return zbool.ToBoolInd(b)
}

func (c *CheckBox) SetValue(b zbool.BoolInd) {
	if b.IsUnknown() {
		c.JSSet("indeterminate", true)
	} else {
		c.JSSet("checked", b.Bool())
	}
}
