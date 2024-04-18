//go:build zui

package zwidgets

import (
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type BranchToggleType int

type BranchToggleListener interface {
	HandleBranchToggleChanged(id string, open bool)
}

const (
	BranchToggleNone BranchToggleType = iota
	BranchToggleTriangle
)

type BranchToggleView struct {
	zimageview.ImageView
	openImage   string
	closedImage string
	isOpen      bool
}

func BranchToggleViewNew(btype BranchToggleType, id string, open bool) *BranchToggleView {
	v := &BranchToggleView{}
	v.Init(v, btype, id, open)
	return v
}

func (v *BranchToggleView) Init(view zview.View, btype BranchToggleType, id string, open bool) {
	// zlog.Info("BranchToggleView Init:", id, open)
	switch btype {
	case BranchToggleTriangle:
		v.openImage = "images/triangle-down-gray.png"
		v.closedImage = "images/triangle-right-gray.png"
	}
	v.ImageView.Init(v, true, nil, "", zgeo.SizeD(16, 16))
	v.SetObjectName(id)
	v.SetOpen(open, false)
	v.SetPressedHandler(func() {
		v.SetOpen(!v.isOpen, true)
	})
}

func (v *BranchToggleView) SetOpen(open bool, tellParents bool) {
	v.isOpen = open
	v.updateToggle()
	if tellParents {
		for _, p := range v.AllParents() {
			hl, _ := p.View.(BranchToggleListener)
			if hl != nil {
				hl.HandleBranchToggleChanged(v.ObjectName(), v.isOpen)
				return
			}
		}
	}
}

func (v *BranchToggleView) IsOpen() bool {
	return v.isOpen
}

func (v *BranchToggleView) updateToggle() {
	str := v.closedImage
	if v.isOpen {
		str = v.openImage
	}
	v.SetImage(nil, str, nil)
}
