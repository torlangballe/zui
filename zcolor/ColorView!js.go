//go:build !js && zui

package zcolor

import "github.com/torlangballe/zutil/zgeo"

func NewView(col zgeo.Color) *ColorView      { return nil }
func (v *ColorView) SetColor(col zgeo.Color) {}
func (v *ColorView) Color() zgeo.Color {
	return zgeo.Color{}
}
