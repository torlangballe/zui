package zui

import "github.com/torlangballe/zutil/zgeo"

// user input type color to make in js

type ColorView struct {
	NativeView
	valueChanged func(view View)
}

func (s *ColorView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{30, 20}
}
