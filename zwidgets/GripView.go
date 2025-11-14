//go:build zui

package zwidgets

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zutil/zgeo"
)

type GripView struct {
	zcontainer.StackView
}

func (v *GripView) Init() {
	v.StackView.Init(v, true, "grip")
	v.SetBGColor(zgeo.ColorClear)
	v.SetStrokeSide(1, zgeo.ColorDarkGray, zgeo.BottomRight, true)
	v.SetStrokeSide(1, zgeo.ColorNewGray(0.65, 1), zgeo.TopLeft, true)

	v.SetTilePath("images/zcore/grippy.png", zgeo.SizeBoth(65))
	v.SetCorner(5)
	v.SetCursor(zcursor.Grab)
}

func GripViewNew() *GripView {
	v := &GripView{}
	v.Init()
	return v
}
