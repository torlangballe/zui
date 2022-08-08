//go:build zui

package zcolor

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

// user input type color to make in js

type ColorView struct {
	zview.NativeView
	valueChanged func(view zview.View)
}

func (v *ColorView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{30, 20}
}
