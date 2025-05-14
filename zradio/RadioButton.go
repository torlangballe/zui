//go:build zui

package zradio

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type RadioButton struct {
	zview.NativeView
	changed  zview.ValueHandlers
	storeKey string
}

var radioButtonSize = zgeo.SizeBoth(20)

func (c *RadioButton) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	return radioButtonSize, radioButtonSize
}

func (v *RadioButton) Toggle() {
	v.SetValue(!v.Value())
}
