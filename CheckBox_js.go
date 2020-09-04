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

func (v *CheckBox) SetRect(rect zgeo.Rect) View {
	rect.Pos.Y -= 3
	return v.NativeView.SetRect(rect)
}

func (s *CheckBox) ValueHandler(handler func(view View)) {
	s.valueChanged = handler
	s.setjs("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if s.valueChanged != nil {
			s.valueChanged(s)
		}
		return nil
	}))
}

func (s *CheckBox) Value() zbool.BoolInd {
	b := s.getjs("checked").Bool()
	i := s.getjs("indeterminate").Bool()
	if i {
		return zbool.Unknown
	}
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
