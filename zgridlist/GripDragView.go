//go:build zui

package zgridlist

import (
	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

type GripDragHandler interface {
	HandleGripDrag(offset float64, id string, down zbool.BoolInd)
}

type GripDragView struct {
	zimageview.ImageView
	id      string
	handler GripDragHandler
}

func NewGripDragView(id string, handler GripDragHandler) *GripDragView {
	v := &GripDragView{}
	v.Init(v, true, nil, "images/zcore/grip-hor.png", zgeo.SizeD(20, 14))
	v.SetCursor(zcursor.Grab)
	v.id = id
	v.handler = handler
	return v
}

func (v *GripDragView) handleUpDownMovedHandler(pos zgeo.Pos, down zbool.BoolInd) {
	v.handler.HandleGripDrag(pos.Y, v.id, down)
}
