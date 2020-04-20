package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zbool"
)

func CheckBoxNew(on zbool.BoolInd) *CheckBox {
	s := &CheckBox{}
	s.Element = DocumentJS.Call("createElement", "input")
	s.set("style", "position:absolute")
	s.set("type", "checkbox")
	s.CanFocus(true)
	s.View = s
	s.SetValue(on)
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

func (s *CheckBox) Value() zbool.BoolInd {
	b := s.get("checked").Bool()
	i := s.get("indeterminate").Bool()
	if i {
		return zbool.BoolUnknown
	}
	return zbool.ToBoolInd(b)
}

func (s *CheckBox) SetValue(b zbool.BoolInd) *CheckBox {
	if b.IsUndetermined() {
		s.set("indeterminate", true)
	} else {
		s.set("checked", b.Value())
	}
	return s
}
