package zgo

import "github.com/torlangballe/zutil/zgeo"

type CheckBox struct {
	NativeView
	valueChanged func(view View)
}

func (s *CheckBox) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{20, 20}
}
