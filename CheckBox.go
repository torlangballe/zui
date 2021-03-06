// +build zui

package zui

import "github.com/torlangballe/zutil/zgeo"

type CheckBox struct {
	NativeView
	valueChanged func(view View)
}

func (s *CheckBox) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{20, 20}
}

func (c *CheckBox) Labelize(title string) (*Label, *StackView) {
	label := LabelNew(title)
	label.SetObjectName("$checkBoxLabel:[" + title + "]")
	stack := StackViewHor("$labledCheckBoxStack.[" + title + "]")
	stack.SetSpacing(0)
	stack.Add(c, zgeo.Left|zgeo.VertCenter, zgeo.Size{0, -4})
	stack.Add(label, zgeo.Left|zgeo.VertCenter, zgeo.Size{6, 0})

	return label, stack
}
