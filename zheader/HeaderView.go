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
	"github.com/torlangballe/zui/zlabel"
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
	FieldName string
	Title     string
	Align     zgeo.Alignment
	Justify   zgeo.Alignment
	// Height         float64
	ImagePath      string
	MinWidth       float64
	MaxWidth       float64
	ImageSize      zgeo.Size
	Tip            string
	SortSmallFirst zbool.BoolInd
	SortPriority   int
	Lockable       bool
}

type HeaderView struct {
	zcontainer.StackView
	SortOrder             []zfields.SortInfo
	HeaderPressedFunc     func(id string)
	HeaderLongPressedFunc func(id string)
	SortingPressedFunc    func()
	LockPressedFunc       func(id string)
	LockedCount           int
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
	zlog.Info("HeaderColumnPressed", h.FieldName)
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
	// zlog.Info("HeaderView.Populate")
	type newSort struct {
		fieldName string
		small     bool
		pri       int
	}
	v.SortOrder = getUserAdjustedSortOrder(v.ObjectName())
	zslice.DeleteFromFunc(&v.SortOrder, func(si zfields.SortInfo) bool { // let's remove any incorrect id's from user stored sort order, in case we changed field names
		for _, h := range headers {
			if si.FieldName == h.FieldName {
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
		v.SortOrder = append(v.SortOrder, zfields.SortInfo{FieldName: n.fieldName, SmallFirst: n.small})
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
		s := zgeo.SizeD(h.MinWidth, 26)
		button := zshape.ImageButtonViewNew(h.Title, "gray-header", s, zshape.DefaultInsets)
		// zlog.Info("HEADER:", h.Title, h.Justify)
		j := h.Justify
		if j == zgeo.AlignmentNone {
			j = zgeo.Left
		}
		button.SetTextAlignment(j)
		if h.ImagePath != "" {
			iv := zimageview.NewWithCachedPath(h.ImagePath, h.ImageSize)
			iv.SetMinSize(h.ImageSize)
			iv.SetObjectName(h.FieldName + ".image")
			button.Add(iv, zgeo.Center, zgeo.SizeNull)
		}
		button.SetTextColor(zgeo.ColorWhite)
		button.SetObjectName(h.FieldName)
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
		if h.Lockable {
			// zlog.Info("POPULATE: Lock", h.FieldName, h.Title)
			lockLabel := zlabel.New("")
			lockLabel.SetObjectName("lockLabel")
			lockLabel.SetTextAlignment(zgeo.Right)
			lockLabel.SetColor(zgeo.ColorYellow)
			lockLabel.Show(false)
			button.Add(lockLabel, zgeo.CenterRight, zgeo.SizeD(-2, 0))

			lock := zimageview.NewWithCachedPath("images/zcore/lock.png", zgeo.SizeD(14, 20))
			lock.SetObjectName("lock")
			lock.Show(false)
			button.Add(lock, zgeo.CenterRight, zgeo.SizeD(2, 0))
			lock.SetPressedHandler(func() {
				if zlog.ErrorIf(v.LockPressedFunc == nil, h.FieldName) {
					return
				}
				v.LockPressedFunc(h.FieldName)
			})
		}
		if h.SortSmallFirst != zbool.Unknown {
			triangle := zimageview.NewWithCachedPath("", zgeo.SizeD(6, 5))
			triangle.SetObjectName("sort")
			//			triangle.Show(false)
			button.Add(triangle, zgeo.TopRight, zgeo.SizeD(2, 3))
			v.updateTriangle(triangle, h.FieldName)
		}
		zfloat.Maximize(&h.MinWidth, button.CalculatedSize(zgeo.SizeNull).W)
		if h.MaxWidth != 0 {
			zfloat.Maximize(&cell.MaxSize.W, math.Max(h.MaxWidth, h.MinWidth))
		}
		v.AddCell(cell, -1)
	}
}

func GetLockViews(headerButton *zshape.ImageButtonView) (*zlabel.Label, *zimageview.ImageView) {
	lock, _ := headerButton.FindViewWithName("lock", false)
	label, _ := headerButton.FindViewWithName("lockLabel", false)
	if lock == nil {
		return nil, nil
	}
	return label.(*zlabel.Label), lock.(*zimageview.ImageView)
}

func ShowLock(headerButton *zshape.ImageButtonView, show bool) {
	label, lock := GetLockViews(headerButton)
	if lock != nil {
		label.Show(show)
		lock.Show(show)
	}
}

func (v *HeaderView) FitToRowStack(stack *zcontainer.StackView) {
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
	xdiff := stack.AbsoluteRect().Pos.X - v.AbsoluteRect().Pos.X

	zlog.Assert(len(cells) == len(hviews), len(cells), len(hviews), stack.Hierarchy())

	hr := v.Rect()
	x := hr.Pos.X
	for i := range cells {
		cr := cells[i].View.Rect()
		o := cr
		o.Pos.X = x
		if i == len(cells)-1 {
			x = hr.Max().X
		} else {
			x = (cr.Max().X+cells[i+1].View.Rect().Min().X)/2 + xdiff
		}
		o.SetMaxX(x)
		o.Pos.Y = 0
		o.Size.H = hr.Size.H
		hviews[i].SetRect(o)
	}
}

func (v *HeaderView) ColumnView(id string) *zshape.ImageButtonView {
	view, _ := v.FindViewWithName(id, false)
	if view != nil {
		return view.(*zshape.ImageButtonView)
	}
	return nil
}

func (v *HeaderView) RightColumn() *zshape.ImageButtonView {
	for i := len(v.Cells) - 1; i >= 0; i-- {
		if v.Cells[i].Free {
			continue
		}
		b, _ := v.Cells[i].View.(*zshape.ImageButtonView)
		if b != nil {
			return b
		}
	}
	return nil
}

func (v *HeaderView) ArrangeChildren() {
	// We purpously don't do anything here, as we want FitToRowStack to arrange us
}
