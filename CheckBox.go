package zgo

type CheckBox struct {
	NativeView
	valueChanged func(view View)
}

func (s *CheckBox) GetCalculatedSize(total Size) Size {
	return Size{20, 20}
}
