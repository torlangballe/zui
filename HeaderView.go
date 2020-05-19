package zui

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

type Header struct {
	ID        string
	Title     string
	Align     zgeo.Alignment
	Height    float64
	ImagePath string
	MinWidth  float64
	MaxWidth  float64
	ImageSize zgeo.Size
	Tip       string
}
type HeaderView struct {
	StackView
}

func HeaderViewNew() *HeaderView {
	v := &HeaderView{}
	v.StackView.init(v, "header")
	v.Vertical = false
	v.SetSpacing(6)

	return v
}

func (v *HeaderView) Populate(headers []Header, pressed func(id string)) {
	for _, h := range headers {
		cell := ContainerViewCell{}
		cell.Alignment = h.Align

		s := zgeo.Size{h.MinWidth, 28}
		button := ButtonNew(h.Title, "grayHeader", s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
		if h.ImagePath != "" {
			iv := ImageViewNew(h.ImagePath, h.ImageSize)
			iv.SetMinSize(h.ImageSize)
			iv.SetObjectName(h.ID + ".image")
			button.Add(zgeo.Center, iv, zgeo.Size{})
		}
		button.TextInfo.Color = zgeo.ColorWhite
		button.TextXMargin = 0
		// if !h.ImageSize.IsNull() {
		// 	cell.MaxSize = h.ImageSize.Plus(zgeo.Size{8, 8})
		// }
		if h.Tip != "" {
			button.SetToolTip(h.Tip)
		}
		cell.View = button

		if pressed != nil {
			id := h.ID // nned to get actual ID here, not just f.ID (f is pointer)
			button.SetPressedHandler(func() {
				pressed(id)
			})
		}
		zfloat.Maximize(&h.MinWidth, button.CalculatedSize(zgeo.Size{}).W)
		if h.MaxWidth != 0 {
			zfloat.Maximize(&cell.MaxSize.W, math.Max(h.MaxWidth, h.MinWidth))
		}
		v.AddCell(cell, -1)
	}
}

func (v *HeaderView) FitToRowStack(stack *StackView, marg float64) {
	children := stack.GetChildren()
	x := 0.0
	w := stack.Rect().Size.W
	for i := range children {
		var e float64
		if i < len(children)-1 {
			e = children[i+1].Rect().Pos.X
			e -= marg
		} else {
			e = w
		}
		hv := v.cells[i].View
		hr := hv.Rect()
		hr.Pos.X = x
		hr.SetMaxX(e)
		x = e
		hv.SetRect(hr)
		// zlog.Info("Header View rect item:", stack.ObjectName(), hv.ObjectName(), hv.Rect())
	}
}
