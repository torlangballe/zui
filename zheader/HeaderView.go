// HeaderView is a stack of buttons that represent a header.
// These are created using the Populate method with a slice of Header structs.
// if the Header struct's SortSmallFirst is not undefined, it handles pressing the header button
// and switching sorting small first/last, and setting SortOrder to a list of what to sort first.
// The FitToRowStack method makes the header buttons the same size as the items in a stack.

//go:build zui

package zheader

import (
	"fmt"
	"math"
	"sort"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
)

type Header struct {
	FieldName      string
	Title          string
	Align          zgeo.Alignment
	Justify        zgeo.Alignment
	Height         float64
	ImagePath      string
	MinWidth       float64
	MaxWidth       float64
	ImageSize      zgeo.Size
	Tip            string
	SortSmallFirst zbool.BoolInd
	SortPriority   int
}

type HeaderView struct {
	zcontainer.StackView
	SortOrder             []zfields.SortInfo
	HeaderPressedFunc     func(id string)
	HeaderLongPressedFunc func(id string)
	SortingPressedFunc    func()
}

func NewView(storeName string) *HeaderView {
	v := &HeaderView{}
	v.StackView.Init(v, false, storeName)
	v.SetSpacing(0)
	return v
}

func (v *HeaderView) updateTriangle(triangle *zimageview.ImageView, id string) {
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

	str = fmt.Sprintf("images/zcore/sort-triangle-%s.png", str)
	triangle.SetImage(nil, str, nil)
}

func makeKey(name string) string {
	return "zheader.HeaderView/SortOrder/" + name
}

func getUserAdjustedSortOrder(tableName string) (order []zfields.SortInfo) {
	key := makeKey(tableName)
	zkeyvalue.DefaultStore.GetObject(key, &order)
	return
}

func SetUserAdjustedSortOrder(tableName string, order []zfields.SortInfo) {
	key := makeKey(tableName)
	zkeyvalue.DefaultStore.SetObject(order, key, true)
}

func (v *HeaderView) findSortInfo(sortOrderID string) int {
	for i := range v.SortOrder {
		if v.SortOrder[i].FieldName == sortOrderID {
			return i
		}
	}
	return -1
}

func (v *HeaderView) handleButtonPressed(button *zshape.ImageButtonView, h Header) {
	if h.SortSmallFirst != zbool.Unknown {
		si := v.findSortInfo(h.FieldName)
		sorting := v.SortOrder[si]
		sorting.SmallFirst = !sorting.SmallFirst
		zslice.RemoveAt(&v.SortOrder, si)
		v.SortOrder = append([]zfields.SortInfo{sorting}, v.SortOrder...)
		for _, c := range v.GetChildren(false) {
			ib, _ := c.(*zshape.ImageButtonView)
			if ib != nil {
				tri, _ := ib.FindViewWithName("sort", false)
				if tri != nil {
					tri.Native().SetAlpha(0.5)
				}
			}
		}
		sort, _ := button.FindViewWithName("sort", false)
		triangle := sort.(*zimageview.ImageView)
		v.updateTriangle(triangle, h.FieldName)
		SetUserAdjustedSortOrder(v.ObjectName(), v.SortOrder)
		if v.SortingPressedFunc != nil {
			v.SortingPressedFunc()
		}
	}
	if v.HeaderPressedFunc != nil {
		v.HeaderPressedFunc(h.FieldName)
	}
}

func (v *HeaderView) Populate(headers []Header) {
	type newSort struct {
		fieldName string
		small     bool
		pri       int
	}
	v.SortOrder = getUserAdjustedSortOrder(v.ObjectName())
	zslice.RemoveIf(&v.SortOrder, func(i int) bool { // let's remove any incorrect id's from user stored sort order, in case we changed field names
		for _, h := range headers {
			if v.SortOrder[i].FieldName == h.FieldName {
				return false
			}
		}
		return true
	})
	var newSorts []newSort
	for _, h := range headers {
		if h.SortSmallFirst != zbool.Unknown && v.findSortInfo(h.FieldName) == -1 {
			newSorts = append(newSorts, newSort{fieldName: h.FieldName, small: h.SortSmallFirst.Bool(), pri: h.SortPriority})
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
		v.SortOrder = append(v.SortOrder, zfields.SortInfo{n.fieldName, n.small})
	}
	SetUserAdjustedSortOrder(v.ObjectName(), v.SortOrder)
	// for _, s := range v.SortOrder {
	// 	zlog.Info("SO:", s.FieldName)
	// }
	for _, h := range headers {
		// zlog.Info("POPULATE:", h.FieldName, h.Title)
		cell := zcontainer.Cell{}
		cell.Alignment = h.Align
		header := h
		s := zgeo.Size{h.MinWidth, 26}
		button := zshape.ImageButtonViewNew(h.Title, "gray-header", s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
		// zlog.Info("HEADER:", h.Title, h.Justify)
		j := h.Justify
		if j == zgeo.AlignmentNone {
			j = zgeo.Left
		}
		button.SetTextAlignment(j)
		if h.ImagePath != "" {
			iv := zimageview.New(nil, h.ImagePath, h.ImageSize)
			iv.SetMinSize(h.ImageSize)
			iv.SetObjectName(h.FieldName + ".image")
			button.Add(iv, zgeo.Center, zgeo.Size{})
		}
		button.SetTextColor(zgeo.ColorWhite)
		button.SetObjectName(h.FieldName)
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
			if v.HeaderLongPressedFunc != nil {
				v.HeaderLongPressedFunc(button.ObjectName())
			}
		})
		if h.SortSmallFirst != zbool.Unknown {
			triangle := zimageview.New(nil, "", zgeo.Size{6, 5})
			triangle.SetObjectName("sort")
			//			triangle.Show(false)
			button.Add(triangle, zgeo.TopRight, zgeo.Size{2, 3})
			v.updateTriangle(triangle, h.FieldName)
		}
		zfloat.Maximize(&h.MinWidth, button.CalculatedSize(zgeo.Size{}).W)
		if h.MaxWidth != 0 {
			zfloat.Maximize(&cell.MaxSize.W, math.Max(h.MaxWidth, h.MinWidth))
		}
		v.AddCell(cell, -1)
	}
}

func (v *HeaderView) FitToRowStack(stack *zcontainer.StackView, gap float64, rightStackMargin, headerIncX float64) {
	var cells []zcontainer.Cell
	for _, c := range stack.Cells {
		if !c.Collapsed && !c.Free {
			cells = append(cells, c)
		}
	}
	var hviews []zview.View
	for _, c := range v.Cells {
		if !c.Collapsed && !c.Free {
			hviews = append(hviews, c.View)
		}
	}
	zlog.Assert(len(cells) == len(hviews), len(cells), len(hviews), stack.Hierarchy())
	x := 0.0
	w := stack.Rect().Size.W
	// zlog.Info("HeaderFit", v.ObjectName())
	for i := range cells {
		var e float64
		if i < len(cells)-1 {
			e = cells[i+1].View.Rect().Pos.X
			e -= gap
		} else {
			e = w + 2 + rightStackMargin
			e = v.LocalRect().Max().X
		}
		hv := hviews[i]
		hr := hv.Rect()
		hr.Pos.X = x
		if i != 0 {
			hr.Pos.X -= headerIncX
		}
		hr.SetMaxX(e - headerIncX)
		x = e
		hr.Size.H--
		hv.SetRect(hr)
		// zlog.Info("Header View rect item:", stack.ObjectName(), hv.ObjectName(), hv.Rect())
	}
}
