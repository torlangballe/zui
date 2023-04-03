//go:build !js && zui

package zcheckbox

import "github.com/torlangballe/zutil/zbool"

func (c *CheckBox) Value() zbool.BoolInd {
	return zbool.False
}

func New(on zbool.BoolInd) *CheckBox {
	c := &CheckBox{}
	return c
}

func (c *CheckBox) SetValue(b zbool.BoolInd)                  {}
func (s *CheckBox) SetValueHandler(handler func())            {}
func (v *CheckBox) Press()                                    {}
func NewWithStore(storeKey string, defaultVal bool) *CheckBox { return nil }
