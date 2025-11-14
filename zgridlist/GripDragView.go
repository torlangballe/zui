//go:build zui

package zgridlist

import (
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type GripDragHandler interface {
	HandleGripDrag(offset float64, id string, down zbool.BoolInd) bool
}

type GripDragView struct {
	zwidgets.GripView
	id      string
	handler GripDragHandler
}

func NewGripDragView(id string, handler GripDragHandler) *GripDragView {
	v := &GripDragView{}
	v.GripView.Init()
	v.GripView.SetMinSize(zgeo.SizeD(40, 14))
	v.id = id
	v.handler = handler
	v.SetPressUpDownMovedHandler(v.handleUpDownMovedHandler)
	return v
}

func (v *GripDragView) handleUpDownMovedHandler(pos zgeo.Pos, down zbool.BoolInd) bool {
	return v.handler.HandleGripDrag(pos.Y, v.id, down)
}
