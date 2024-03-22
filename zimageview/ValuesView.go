//go:build zui

package zimageview

import (
	"fmt"
	"strings"

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/ztimer"
)

// A ValuesView is a set of string-values and paths to images for each value.
// Pressing it chooses the next value/image in variants.
type ValuesView struct {
	ImageView
	ValueChangedHandlerFunc func()
	variants                []variant
	currentValue            string
	storeKey                string
}

type variant struct {
	value string
	path  string
}

func (v *ValuesView) Init(view zview.View, fitSize zgeo.Size, key string) {
	v.storeKey = key
	v.ImageView.Init(view, true, nil, "", fitSize)
	v.SetPressedHandler(v.pressed)
	v.SetObjectName("ValuesView")
	ztimer.StartIn(0.1, func() {
		if !v.Rect().Size.IsNull() { // Don't do if not laid out
			v.update() // a bit of a hack to use a timer, but v.ReadyToShow() seems to mess up exposing. TODO: Fix
		}
	})
}

func NewValuesView(fitSize zgeo.Size, key string) *ValuesView {
	v := &ValuesView{}
	v.Init(v, fitSize, key)
	return v
}

func (v *ValuesView) AddVariant(value, path string) {
	v.variants = append(v.variants, variant{value: value, path: path})
}

func (v *ValuesView) SetValue(value string) {
	v.currentValue = value
	for _, a := range v.variants {
		if a.value == value {
			v.SetImage(nil, a.path, nil)
			if v.ValueChangedHandlerFunc != nil {
				v.ValueChangedHandlerFunc()
			}
			if v.storeKey != "" {
				zkeyvalue.DefaultStore.SetString(value, v.storeKey, true)
			}
			break
		}
	}
}

func (v *ValuesView) SetBoolValue(value bool) {
	s := zbool.ToString(value)
	v.SetValue(s)
}

func (v *ValuesView) Value() string {
	return v.currentValue
}

func (v *ValuesView) BoolValue() bool {
	return zbool.FromString(v.currentValue, false)
}

func (v *ValuesView) pressed() {
	if len(v.variants) == 0 {
		return
	}
	var set int
	for i, a := range v.variants {
		if a.value == v.currentValue {
			set = i + 1
			if set >= len(v.variants) {
				set = 0
			}
			break
		}
	}
	v.SetValue(v.variants[set].value)
}

// SetAsToggle sets an image for true and false, with path as prefix/format:
// path="x/" strue="1.png" sfalse="0.png" -> true: x/1.png false: x/0.png
// path = x/%s.png" strue="1" sfalse="0" gives same result.
func (v *ValuesView) SetAsToggle(path, ptrue, pfalse string, initialValue bool) {
	if strings.Contains(path, `%s`) {
		ptrue = fmt.Sprintf(path, ptrue)
		pfalse = fmt.Sprintf(path, pfalse)
	} else {
		ptrue = path + ptrue
		pfalse = path + pfalse
	}
	v.AddVariant(zbool.ToString(true), ptrue)
	v.AddVariant(zbool.ToString(false), pfalse)
	v.SetBoolValue(initialValue)
}

func (v *ValuesView) update() {
	if len(v.variants) == 0 {
		return
	}
	if v.currentValue == "" {
		var val string
		if v.storeKey != "" {
			val, _ = zkeyvalue.DefaultStore.GetString(v.storeKey)
		}
		if val == "" {
			val = v.variants[0].value
		}
		v.SetValue(val)
	}
}
