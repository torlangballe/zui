//go:build zui

package zcheckbox

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type CheckBox struct {
	zview.NativeView
	valueChanged func()
	storeKey     string
}

func (c *CheckBox) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{20, 20}
}

func (c *CheckBox) On() bool {
	return c.Value() == zbool.True
}

func (c *CheckBox) SetOn(on bool) {
	c.SetValue(zbool.ToBoolInd(on))
}
