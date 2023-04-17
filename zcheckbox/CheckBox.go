//go:build zui

package zcheckbox

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type CheckBox struct {
	zview.NativeView
	valueChanged func()
	storeKey     string
}

func (c *CheckBox) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{24, 24}
}

func (c *CheckBox) On() bool {
	return c.Value() == zbool.True
}

func (c *CheckBox) SetOn(on bool) {
	c.SetValue(zbool.ToBoolInd(on))
}

func Labelize(c *CheckBox, title string) (*zlabel.Label, *zcontainer.StackView) {
	label := zlabel.New(title)
	label.SetObjectName("$checkBoxLabel:[" + title + "]")
	label.SetPressedHandler(func() {
		c.Press()
	})
	stack := zcontainer.StackViewHor("$labledCheckBoxStack.[" + title + "]")
	stack.SetSpacing(0)
	stack.Add(c, zgeo.Left|zgeo.VertCenter, zgeo.Size{0, -4})
	stack.Add(label, zgeo.Left|zgeo.VertCenter, zgeo.Size{6, 0})

	return label, stack
}

func NewWithLabel(def bool, title, storeKey string) (check *CheckBox, label *zlabel.Label, stack *zcontainer.StackView) {
	check = NewWithStore(def, storeKey)
	label, stack = Labelize(check, title)
	return
}
