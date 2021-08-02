// +build zui

package zui

import (
	"fmt"
	"math"
	"sort"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
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
	v.StackView.Init(v, false, id)
	v.SetSpacing(0)
	return v
}

func (v *HeaderView) updateTriangle(triangle *ImageView, id string) {
	i := v.findSortInfo(id)
	sorting := v.SortOrder[i]
	if i == 0 {
		triangle.SetAlpha(1)
	} else {
		triangle.SetAlpha(0.5)
	}
	str := "down"
	if sorting.SmallFirst {
		str = "up"
	}

	str = fmt.Sprintf("images/sort-triangle-%s.png", str)
	triangle.SetImage(nil, str, nil)
}

func makeKey(name string) string {
	return "HeaderView/SortOrder/" + name
}

func getUserAdjustedSortOrder(tableName string) (order []SortInfo) {
	key := makeKey(tableName)
	DefaultLocalKeyValueStore.GetObject(key, &order)
	return
}

func SetUserAdjustedSortOrder(tableName string, order []SortInfo) {
	key := makeKey(tableName)
	DefaultLocalKeyValueStore.SetObject(order, key, true)
}

func (v *HeaderView) findSortInfo(sortOrderID string) int {
	for i := range v.SortOrder {
		if v.SortOrder[i].ID == sortOrderID {
			return i
		}
	}
	return -1
}

func (v *HeaderView) handleButtonPressed(button *ImageButtonView, h Header) {
	if h.SortSmallFirst != zbool.Unknown {
		si := v.findSortInfo(h.ID)
		sorting := v.SortOrder[si]
		sorting.SmallFirst = !sorting.SmallFirst
		zslice.RemoveAt(&v.SortOrder, si)
		v.SortOrder = append([]SortInfo{sorting}, v.SortOrder...)
		for _, c := range v.GetChildren(false) {
			tri, _ := c.(*ImageButtonView).FindViewWithName("sort", false)
			if tri != nil {
				tri.SetAlpha(0.5)
			}
		}
		sort, _ := button.FindViewWithName("sort", false)
		triangle := sort.(*ImageView)
		v.updateTriangle(triangle, h.ID)
		SetUserAdjustedSortOrder(v.ObjectName(), v.SortOrder)
		if v.SortingPressed != nil {
			v.SortingPressed()
		}
	}
	if v.HeaderPressed != nil {
		v.HeaderPressed(h.ID)
	}
}

func (v *HeaderView) Populate(headers []Header) {
	type newSort struct {
		id    string
		small bool
		pri   int
	}
	v.SortOrder = getUserAdjustedSortOrder(v.ObjectName())
	zslice.RemoveIf(&v.SortOrder, func(i int) bool { // let's remove any incorrect id's from user stored sort order, in case we changed field names
		for _, h := range headers {
			if v.SortOrder[i].ID == h.ID {
				return false
			}
		}
		return true
	})
	var newSorts []newSort
	for _, h := range headers {
		// zlog.Info("POPULATE:", h.ID)
		if h.SortSmallFirst != zbool.Unknown && v.findSortInfo(h.ID) == -1 {
			newSorts = append(newSorts, newSort{id: h.ID, small: h.SortSmallFirst.BoolValue(), pri: h.SortPriority})
		}
	}
	sort.Slice(newSorts, func(i, j int) bool {
		pi := newSorts[i]
		pj := newSorts[j]
		if pi.pri == 0 {
			pi.pri = 1000 + i
		}
		if pj.pri == 0 {
			pj.pri = 1000 + j
		}
		return pi.pri < pj.pri
	})
	for _, n := range newSorts {
		v.SortOrder = append(v.SortOrder, SortInfo{n.id, n.small})
	}
	SetUserAdjustedSortOrder(v.ObjectName(), v.SortOrder)
	// for _, s := range v.SortOrder {
	// 	zlog.Info("SO:", s.ID)
	// }
	for _, h := range headers {
		cell := ContainerViewCell{}
		cell.Alignment = h.Align
		header := h
		s := zgeo.Size{h.MinWidth, 24}
		button := ImageButtonViewNew(h.Title, "grayHeader", s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
		if h.ImagePath != "" {
			iv := ImageViewNew(nil, h.ImagePath, h.ImageSize)
			iv.SetMinSize(h.ImageSize)
			iv.SetObjectName(h.ID + ".image")
			button.Add(iv, zgeo.Center, zgeo.Size{})
		}
		button.SetTextColor(zgeo.ColorWhite)
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
			triangle := ImageViewNew(nil, "", zgeo.Size{6, 5})
			triangle.SetObjectName("sort")
			//			triangle.Show(false)
			button.Add(triangle, zgeo.TopRight, zgeo.Size{2, 3})
			v.updateTriangle(triangle, h.ID)
		}
		zfloat.Maximize(&h.MinWidth, button.CalculatedSize(zgeo.Size{}).W)
		if h.MaxWidth != 0 {
			zfloat.Maximize(&cell.MaxSize.W, math.Max(h.MaxWidth, h.MinWidth))
		}
		v.AddCell(cell, -1)
	}
}

func (v *HeaderView) FitToRowStack(stack *StackView, marg float64) {
	children := stack.GetChildren(true)
	x := 0.0
	w := stack.Rect().Size.W
	// zlog.Info("HeaderFit", v.ObjectName(), len(v.cells), len(children))
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
			continue
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
