//go:build zui

package zcheckbox

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
)

func init() {
	if zdevice.CurrentWasmBrowser == zdevice.Chrome {
		checkboxSize = zgeo.SizeD(14, 14)
	}
}

func NewWithStore(defaultVal bool, storeKey string) *CheckBox {
	val := defaultVal
	if storeKey != "" {
		v, got := zkeyvalue.DefaultStore.GetBool(storeKey, defaultVal)
		// zlog.Info("NewWithStore:", val, v, got, storeKey, defaultVal)
		if got {
			val = v
		}
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
	// v.JSStyle().Set("margin-left", "4px")
	v.SetCanTabFocus(false)
	v.View = v
	v.SetValue(on)
	return v
}

func (v *CheckBox) SetRect(rect zgeo.Rect) {
	rect.Pos.Y -= 1 // 3
	v.NativeView.SetRect(rect)
}

func (v *CheckBox) SetValueHandler(id string, handler func(edited bool)) {
	v.changed.Add(id, handler)
	if v.changed.Count() == 1 {
		v.JSSet("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
			if v.storeKey != "" {
				// zlog.Info("StoreCheck:", v.On(), v.storeKey)
				zkeyvalue.DefaultStore.SetBool(v.On(), v.storeKey, true)
			}
			v.changed.CallAll(true)
			return nil
		}))
	}
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
	v.JSSet("indeterminate", b.IsUnknown())
	v.JSSet("checked", b.Bool())
}

func (v *CheckBox) SetInteractive(interactive bool) {
	v.NativeView.SetInteractive(interactive)
	if interactive {
		v.JSSet("inert", nil)
		return
	}
	v.JSSet("inert", "true")
}
