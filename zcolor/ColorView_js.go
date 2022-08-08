package zcolor

import (
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zgeo"
)

func NewView(col zgeo.Color) *ColorView {
	v := &ColorView{}
	v.MakeJSElement(v, "color")
	v.SetColor(col)
	return v
}

func (v *ColorView) SetColor(col zgeo.Color) {
	v.JSSet("value", zdom.MakeRGBAString(col))
}

func (v *ColorView) Color() zgeo.Color {
	str := v.JSGet("value").String()
	return zgeo.ColorFromString(str)
}
