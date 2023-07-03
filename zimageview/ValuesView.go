package zimageview

import (
	"fmt"
	"strings"

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type ValuesView struct {
	ImageView
	ValueChangedHandlerFunc func()
	variants                []variant
	currentValue            string
}

type variant struct {
	value string
	path  string
}

func (v *ValuesView) Init(view zview.View, fitSize zgeo.Size) {
	v.ImageView.Init(view, nil, "", fitSize)
	v.SetPressedHandler(v.pressed)
}

func NewValuesView(fitSize zgeo.Size) *ValuesView {
	v := &ValuesView{}
	v.Init(v, fitSize)
	return v
}

func (v *ValuesView) AddVariant(value, path string) {
	v.variants = append(v.variants, variant{value: value, path: path})
}

func (v *ValuesView) AddPathVariant(value string, imagePath string) {
	v.AddVariant(value, imagePath)
}

func (v *ValuesView) SetValue(value string) {
	v.currentValue = value
	for _, a := range v.variants {
		if a.value == value {
			v.SetImage(nil, a.path, nil)
			if v.ValueChangedHandlerFunc != nil {
				v.ValueChangedHandlerFunc()
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
	v.AddPathVariant(zbool.ToString(true), ptrue)
	v.AddPathVariant(zbool.ToString(false), pfalse)
	v.SetBoolValue(initialValue)
}
