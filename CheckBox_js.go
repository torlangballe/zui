package zgo

import "syscall/js"

func CheckBoxNew(on BoolInd) *CheckBox {
	s := &CheckBox{}
	s.Element = DocumentJS.Call("createElement", "input")
	s.set("style", "position:absolute")
	s.set("type", "checkbox")
	s.CanFocus(true)
	s.View = s
	s.Value(on)
	return s
}

func (s *CheckBox) ValueHandler(handler func(view View)) {
	s.valueChanged = handler
	s.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if s.valueChanged != nil {
			s.valueChanged(s)
		}
		return nil
	}))
}

func (s *CheckBox) GetValue() BoolInd {
	b := s.get("checked").Bool()
	i := s.get("indeterminate").Bool()
	if i {
		return BoolUnknown
	}
	return BoolIndFromBool(b)
}

func (s *CheckBox) Value(b BoolInd) *CheckBox {
	if b.IsIndetermed() {
		s.set("indeterminate", true)
	} else {
		s.set("checked", b.Value())
	}
	return s
}
