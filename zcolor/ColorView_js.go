package zcolor

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

func New(col zgeo.Color) *ColorView {
	v := &ColorView{}
	v.MakeJSElement(v, "input")
	v.JSSet("type", "color")
	v.SetColor(col)
	v.JSSet("oninput", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		v.SetToolTip(v.Color().HexNoAlpha())
		if v.valueChangedHandlerFunc != nil {
			v.valueChangedHandlerFunc(true)
		}
		return nil
	}))
	return v
}

func (v *ColorView) SetColor(col zgeo.Color) {
	v.JSSet("value", col.HexNoAlpha())
	v.SetToolTip(v.Color().HexNoAlpha())
}

func (v *ColorView) Color() zgeo.Color {
	str := v.JSGet("value").String()
	return zgeo.ColorFromString(str)
}

func (v *ColorView) SetValueHandler(f func(edited bool)) {
	v.valueChangedHandlerFunc = f
}
