//go:build zui

package zcheckbox

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type CheckBox struct {
	zview.NativeView
	changed  zview.ValueHandlers
	storeKey string
}

var checkboxSize = zgeo.SizeBoth(20)

func (c *CheckBox) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	return checkboxSize, checkboxSize
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
	label.SetPressedHandler("$press.label", zkeyboard.ModifierNone, func() {
		c.ClickAll()
	})
	stack := zcontainer.StackViewHor("$labledCheckBoxStack.[" + title + "]")
	stack.SetMargin(zgeo.RectFromXY2(0, 3, 0, -3))
	stack.SetSpacing(0)
	stack.Add(c, zgeo.Left|zgeo.VertCenter, zgeo.SizeD(0, -4))
	stack.Add(label, zgeo.Left|zgeo.VertCenter, zgeo.SizeD(6, 0))
	return label, stack
}

func NewWithLabel(def bool, title, storeKey string) (check *CheckBox, label *zlabel.Label, stack *zcontainer.StackView) {
	check = NewWithStore(def, storeKey)
	label, stack = Labelize(check, title)
	return
}

func (v *CheckBox) Toggle() {
	v.SetOn(!v.On())
}

func MakeNextToView(view zview.View) (*CheckBox, *zcontainer.StackView) {
	stack := zcontainer.StackViewHor(view.ObjectName() + ".check-stack")
	stack.SetSpacing(1)
	check := New(zbool.False)
	stack.Add(check, zgeo.CenterLeft)
	stack.Add(view, zgeo.CenterLeft|zgeo.HorExpand)
	check.SetValueHandler("$stacked-handler", func(edited bool) {
		view.SetUsable(check.On())
	})
	return check, stack
}
