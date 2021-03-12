package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

func CheckBoxNew(on zbool.BoolInd) *CheckBox {
	s := &CheckBox{}
	s.Element = DocumentJS.Call("createElement", "input")
	s.setjs("style", "position:absolute")
	s.setjs("type", "checkbox")
	s.SetCanFocus(true)
	s.View = s
	s.SetValue(on)
	return s
}

func (v *CheckBox) SetRect(rect zgeo.Rect) {
	rect.Pos.Y -= 4
	v.NativeView.SetRect(rect)
}

func (s *CheckBox) SetValueHandler(handler func(view View)) {
	s.valueChanged = handler
	s.setjs("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if s.valueChanged != nil {
			s.valueChanged(s)
		}
		return nil
	}))
}

func (s *CheckBox) Value() zbool.BoolInd {
	i := s.getjs("indeterminate").Bool()
	if i {
		return zbool.Unknown
	}
	b := s.getjs("checked").Bool()
	return zbool.ToBoolInd(b)
}

func (s *CheckBox) SetValue(b zbool.BoolInd) *CheckBox {
	if b.IsUndetermined() {
		s.setjs("indeterminate", true)
	} else {
		s.setjs("checked", b.BoolValue())
	}
	return s
}
