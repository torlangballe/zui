// +build !js
// +build zui

package zui

import "github.com/torlangballe/zutil/zbool"

func (c *CheckBox) Value() zbool.BoolInd {
	return zbool.False
}

func CheckBoxNew(on zbool.BoolInd) *CheckBox {
	c := &CheckBox{}
	return c
}

func (c *CheckBox) SetValue(b zbool.BoolInd)       {}
func (s *CheckBox) SetValueHandler(handler func()) {}
