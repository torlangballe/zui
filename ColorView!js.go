// +build !js,zui

package zui

import "github.com/torlangballe/zutil/zgeo"

func ColorViewNew(col zgeo.Color) *ColorView { return nil }
func (v *ColorView) SetColor(col zgeo.Color) {}
func (v *ColorView) Color() zgeo.Color {
	return zgeo.Color{}
}
