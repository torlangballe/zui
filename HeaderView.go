package zui

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
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
	Sortable  bool
}
type HeaderView struct {
	StackView
	SortSmallerFirst   map[string]bool
	PressedHandler     func(id string)
	LongPressedHandler func(id string)
	SortingPressed     func()
}

func HeaderViewNew(id string) *HeaderView {
	v := &HeaderView{}
	v.StackView.init(v, "header")
	v.Vertical = false
	v.SetObjectName(id)
	v.SetSpacing(6)
	v.SortSmallerFirst = map[string]bool{}
	return v
}

func updateTriangle(triangle *Label, small bool) {
	str := "▽"
	if small {
		str = "△"
	}
	triangle.SetText(str)
}

func (v *HeaderView) makeKey(id string) string {
	return "HeaderView/" + v.ObjectName() + "/" + id + "/SmallFirst"
}

func (v *HeaderView) handleButtonPressed(button *Button, h Header) {
	if h.Sortable {
		small, _ := v.SortSmallerFirst[h.ID]
		small = !small
		triangle := (*button.FindViewWithName("sort", false)).(*Label)
		updateTriangle(triangle, small)
		v.SortSmallerFirst[h.ID] = small
		key := v.makeKey(button.ObjectName())
		DefaultLocalKeyValueStore.SetBool(small, key, true)
		if v.SortingPressed != nil {
			v.SortingPressed()
		}
	}
	if v.PressedHandler != nil {
		v.PressedHandler(h.ID)
	}
}

func (v *HeaderView) Populate(headers []Header) {
	for _, h := range headers {
		cell := ContainerViewCell{}
		cell.Alignment = h.Align
		id := h.ID // nned to get actual ID here, not just h.ID (h is pointer)

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
		button.SetObjectName(id)
		// if !h.ImageSize.IsNull() {
		// 	cell.MaxSize = h.ImageSize.Plus(zgeo.Size{8, 8})
		// }
		if h.Tip != "" {
			button.SetToolTip(h.Tip)
		}
		cell.View = button

		button.SetPressedHandler(func() {
			v.handleButtonPressed(button, h)
		})
		button.SetLongPressedHandler(func() {
			if v.LongPressedHandler != nil {
				v.LongPressedHandler(button.ObjectName())
			}
		})
		if h.Sortable {
			triangle := LabelNew("△▽")
			triangle.SetObjectName("sort")
			button.Add(zgeo.RightCenter, triangle, zgeo.Size{4, 0})
			key := v.makeKey(h.ID)
			small, _ := DefaultLocalKeyValueStore.BoolForKey(key, true)
			v.SortSmallerFirst[h.ID] = small
			updateTriangle(triangle, small)
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
