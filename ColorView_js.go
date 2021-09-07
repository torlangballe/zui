package zui

import "github.com/torlangballe/zutil/zgeo"

func ColorViewNew(col zgeo.Color) *ColorView {
	v := &ColorView{}
	v.MakeJSElement(v, "color")
	v.SetColor(col)
	return v
}

func (v *ColorView) SetColor(col zgeo.Color) {
	v.setjs("value", makeRGBAString(col))
}

func (v *ColorView) Color() zgeo.Color {
	str := v.getjs("value").String()
	return zgeo.ColorFromString(str)
}
