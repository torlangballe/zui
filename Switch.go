package zgo

type Switch struct {
	NativeView
	valueChanged func(view View)
}

func (s *Switch) GetCalculatedSize(total Size) Size {
	return Size{20, 20}
}
