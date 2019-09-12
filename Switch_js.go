package zgo

import "syscall/js"

type BoolInd int

const (
	BoolTrue    BoolInd = 1
	BoolFalse           = 0
	BoolUnknown         = -1
)

func BoolIndMake(b bool) BoolInd {
	if b {
		return BoolTrue
	}
	return BoolFalse
}

func (b BoolInd) Value() bool {
	return b == 1
}

func (b BoolInd) IsIndetermed() bool {
	return b == -1
}

func SwitchNew(on BoolInd) *Switch {
	s := &Switch{}
	s.Element = DocumentJS.Call("createElement", "input")
	s.set("style", "position:absolute")
	s.set("type", "checkbox")
	s.View = s
	s.Value(on)
	return s
}

func (s *Switch) ValueHandler(handler func(view View)) {
	s.valueChanged = handler
	s.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if s.valueChanged != nil {
			s.valueChanged(s)
		}
		return nil
	}))
}

func (s *Switch) GetValue() BoolInd {
	b := s.get("checked").Bool()
	i := s.get("indeterminate").Bool()
	if i {
		return BoolUnknown
	}
	return BoolIndMake(b)
}

func (s *Switch) Value(b BoolInd) *Switch {
	if b.IsIndetermed() {
		s.set("indeterminate", true)
	} else {
		s.set("checked", b.Value())
	}
	return s
}
