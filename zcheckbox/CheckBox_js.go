//go:build zui

package zcheckbox

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
)

func NewWithStore(storeKey string, defaultVal bool) *CheckBox {
	val, got := zkeyvalue.DefaultStore.GetBool(storeKey, false)
	// zlog.Info("NewWithStore:", val, got, storeKey, defaultVal)
	if !got {
		val = defaultVal
	}
	v := New(zbool.FromBool(val))
	v.storeKey = storeKey
	return v
}

func New(on zbool.BoolInd) *CheckBox {
	v := &CheckBox{}
	v.Element = zdom.DocumentJS.Call("createElement", "input")
	v.JSSet("style", "position:absolute")
	v.JSSet("type", "checkbox")
	v.JSStyle().Set("margin-top", "4px")
	v.SetCanFocus(true)
	v.View = v
	v.SetValue(on)
	return v
}

func (v *CheckBox) SetRect(rect zgeo.Rect) {
	rect.Pos.Y -= 4
	v.NativeView.SetRect(rect)
}

func (v *CheckBox) SetValueHandler(handler func()) {
	v.valueChanged = handler
	v.JSSet("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.storeKey != "" {
			// zlog.Info("StoreCheck:", v.On(), v.storeKey)
			zkeyvalue.DefaultStore.SetBool(v.On(), v.storeKey, true)
		}
		if v.valueChanged != nil {
			v.valueChanged()
		}
		return nil
	}))
}

func (v *CheckBox) Value() zbool.BoolInd {
	i := v.JSGet("indeterminate").Bool()
	if i {
		return zbool.Unknown
	}
	b := v.JSGet("checked").Bool()
	return zbool.ToBoolInd(b)
}

func (v *CheckBox) SetValue(b zbool.BoolInd) {
	if b.IsUnknown() {
		v.JSSet("indeterminate", true)
	} else {
		v.JSSet("checked", b.Bool())
	}
}

func (v *CheckBox) Press() {
	v.JSCall("click")
}
