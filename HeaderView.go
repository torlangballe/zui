package zui

import (
	"fmt"
	"math"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type Header struct {
	ID             string
	Title          string
	Align          zgeo.Alignment
	Height         float64
	ImagePath      string
	MinWidth       float64
	MaxWidth       float64
	ImageSize      zgeo.Size
	Tip            string
	SortSmallFirst zbool.BoolInd
	SortPriority   int
}

type SortInfo struct {
	ID         string
	SmallFirst bool
}

type HeaderView struct {
	StackView
	SortOrder         []SortInfo
	HeaderPressed     func(id string) // we call it HeaderPressed and HeaderLongPressed to avoid confusion with pressed stuff inherited from stack
	HeaderLongPressed func(id string)
	SortingPressed    func()
}

func HeaderViewNew(id string) *HeaderView {
	v := &HeaderView{}
	v.StackView.init(v, "header")
	v.Vertical = false
	v.SetObjectName(id)
	v.SetSpacing(6)
	return v
}

func updateTriangle(triangle *ImageView, small bool) {
	str := "down"
	if small {
		str = "up"
	}
	str = fmt.Sprintf("images/sort-triangle-%s.png", str)
	triangle.SetImage(nil, str, nil)
}

func (v *HeaderView) makeKey() string {
	return "HeaderView/SortOrder/" + v.ObjectName()
}

func (v *HeaderView) findSortInfo(id string) *SortInfo {
	for i := range v.SortOrder {
		if v.SortOrder[i].ID == id {
			return &v.SortOrder[i]
		}
	}
	return nil
}

func (v *HeaderView) handleButtonPressed(button *Button, h Header) {
	if h.SortSmallFirst != zbool.Unknown {
		sorting := v.findSortInfo(h.ID)
		sorting.SmallFirst = !sorting.SmallFirst
		triangle := (*button.FindViewWithName("sort", false)).(*ImageView)
		updateTriangle(triangle, sorting.SmallFirst)
		key := v.makeKey()
		DefaultLocalKeyValueStore.SetObject(v.SortOrder, key, true)
		if v.SortingPressed != nil {
			v.SortingPressed()
		}
	}
	if v.HeaderPressed != nil {
		v.HeaderPressed(h.ID)
	}
}

func (v *HeaderView) Populate(headers []Header) {
	key := v.makeKey()
	DefaultLocalKeyValueStore.GetObject(key, &v.SortOrder)
	for _, h := range headers {
		if h.SortSmallFirst != zbool.Unknown {
			if v.findSortInfo(h.ID) == nil {
				v.SortOrder = append(v.SortOrder, SortInfo{h.ID, h.SortSmallFirst.BoolValue()})
			}
		}
	}
	for _, h := range headers {
		cell := ContainerViewCell{}
		cell.Alignment = h.Align
		header := h
		s := zgeo.Size{h.MinWidth, 28}
		button := ButtonNew(h.Title, "grayHeader", s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
		if h.ImagePath != "" {
			iv := ImageViewNew(h.ImagePath, h.ImageSize)
			iv.SetMinSize(h.ImageSize)
			iv.SetObjectName(h.ID + ".image")
			button.Add(zgeo.Center, iv, zgeo.Size{})
		}
		button.SetColor(zgeo.ColorWhite)
		button.TextXMargin = 0
		button.SetObjectName(h.ID)
		// if !h.ImageSize.IsNull() {
		// 	cell.MaxSize = h.ImageSize.Plus(zgeo.Size{8, 8})
		// }
		if h.Tip != "" {
			button.SetToolTip(h.Tip)
		}
		cell.View = button

		button.SetPressedHandler(func() {
			v.handleButtonPressed(button, header)
		})
		button.SetLongPressedHandler(func() {
			if v.HeaderLongPressed != nil {
				v.HeaderLongPressed(button.ObjectName())
			}
		})
		if h.SortSmallFirst != zbool.Unknown {
			triangle := ImageViewNew("", zgeo.Size{6, 5})
			triangle.SetObjectName("sort")
			//			triangle.Show(false)
			button.Add(zgeo.TopRight, triangle, zgeo.Size{2, 3})
			sorting := v.findSortInfo(h.ID)
			triangle.SetAlpha(0.5)
			updateTriangle(triangle, sorting.SmallFirst)
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
		if v.cells[i].Collapsed {
			zlog.Info("Header collapsed!", v.ObjectName(), i)
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
