//go:build zui

package zradio

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zkeyvalue"
)

func init() {
	// if zdevice.CurrentWasmBrowser == zdevice.Chrome {
	// 	radioButtonSizeSize = zgeo.SizeD(14, 14)
	// }
}

func NewWithStore(defaultVal bool, id, group, storeKey string) *RadioButton {
	val := defaultVal
	if storeKey != "" {
		v, got := zkeyvalue.DefaultStore.GetBool(storeKey, defaultVal)
		// zlog.Info("NewWithStore:", val, v, got, storeKey, defaultVal)
		if got {
			val = v
		}
	}
	v := NewButton(val, id, group)
	v.storeKey = storeKey
	return v
}

func NewButton(on bool, id, group string) *RadioButton {
	v := &RadioButton{}
	// zlog.Info("NewButton:", on, id, group)
	v.Element = zdom.DocumentJS.Call("createElement", "input")
	v.JSSet("style", "position:absolute")
	v.JSSet("type", "radio")
	v.SetObjectName(id)
	v.JSSet("value", id)
	v.JSSet("name", group)
	if on {
		v.JSSet("checked", true)
	}
	v.SetCanTabFocus(false)
	v.View = v
	v.SetValue(on)
	return v
}

// func (v *RadioButton) SetRect(rect zgeo.Rect) {
// 	rect.Pos.Y -= 3
// 	v.NativeView.SetRect(rect)
// }

func (v *RadioButton) SetValueHandler(id string, handler func(edited bool)) {
	v.changed.Add(id, handler)
	if v.changed.Count() == 1 {
		v.JSSet("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
			if v.storeKey != "" {
				// zlog.Info("StoreCheck:", v.On(), v.storeKey)
				zkeyvalue.DefaultStore.SetBool(v.Value(), v.storeKey, true)
			}
			v.changed.CallAll(true)
			return nil
		}))
	}
}

func (v *RadioButton) Value() bool {
	b := v.JSGet("checked").Bool()
	return b
}

func (v *RadioButton) SetValue(b bool) {
	v.JSSet("checked", b)
}

func (v *RadioButton) SetInteractive(interactive bool) {
	v.NativeView.SetInteractive(interactive)
	if interactive {
		v.JSSet("inert", nil)
		return
	}
	v.JSSet("inert", "true")
}
