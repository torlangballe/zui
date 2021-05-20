// +build zui,!js

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

func ButtonNew(text string) *Button {
	return nil
}

func (v *Button) MakeEnterDefault() {
}

func (v *Button) SetPressedHandler(handler func()) {
}

func (v *Button) SetLongPressedHandler(handler func()) {
}

func (v *Button) SetMargin(m zgeo.Rect) {
	v.margin = m
}
