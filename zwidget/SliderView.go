//go:build zui

package zwidget

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

// user input type range to make in js

type SliderView struct {
	zview.NativeView
	valueChanged func(view zview.View)
}

func (s *SliderView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{100, 20}
}
