package zui

import "github.com/torlangballe/zutil/zgeo"

// user input type range to make in js

type SliderView struct {
	NativeView
	valueChanged func(view View)
}

func (s *SliderView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{100, 20}
}

