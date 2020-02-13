package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

//  Created by Tor Langballe on /4/12/15.

type ListView struct {
	ScrollView
	//ListViewNative
	//	Scrolling bool
	spacing float64

	GetRowCount          func() int
	GetRowHeight         func(i int) float64
	CreateRow            func(rowSize zgeo.Size, i int) View
	RowUpdater           func(i int, edited bool)
	HandleRowSelected    func(i int)
	HandleScrolledToRows func(y float64, first, last int)

	RowColors []zgeo.Color
	MinRows   int

	selectionIndex int
	Selectable     bool
	selectedColor  zgeo.Color

	topPos float64
	stack  *CustomView
	rows   map[int]View
}

func ListViewNew(name string) *ListView {
	v := &ListView{}
	v.init(v, name)
	v.RowColors = []zgeo.Color{zgeo.ColorWhite}
	v.selectedColor = zgeo.ColorNew(0.6, 0.6, 0.8, 1)
	return v
	//        allowsSelection = true // selectable
}

func (v *ListView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.ScrollView.CalculatedSize(total)
	h := 0.0
	count := v.GetRowCount()
	for i := 0; i < v.MinRows && i < count; i++ {
		h += v.GetRowHeight(i)
	}
	s.H = h

	return s
}

func (v *ListView) drawIfExposed() {
	for _, view := range v.rows {
		et, got := view.(ExposableType)
		if got {
			et.drawIfExposed()
		}
	}
}

func (v *ListView) SetSpacing(spacing float64) *ListView {
	v.spacing = spacing
	return v
}

func (v *ListView) Spacing() float64 {
	return v.spacing
}

func (v *ListView) SelectedColor(col zgeo.Color) *ListView {
	v.selectedColor = col
	return v
}

func (v *ListView) init(view View, name string) {
	v.ScrollView.init(view, name)
	v.Selectable = true
	v.selectionIndex = -1
	v.rows = map[int]View{}
	v.HandleScroll = func(pos zgeo.Pos) {
		v.topPos = pos.Y
		first, last := v.layoutRows(-1)
		if v.HandleScrolledToRows != nil {
			v.HandleScrolledToRows(pos.Y, first, last)
		}
	}
}

func (v *ListView) SetRect(rect zgeo.Rect) View {
	v.ScrollView.SetRect(rect)
	if v.stack == nil {
		v.stack = CustomViewNew("listview.stack")
		v.AddChild(v.stack, -1)
	}
	count := v.GetRowCount()
	pos := v.Margin.Min()
	h := 0.0
	for i := 0; i < count; i++ {
		h += v.GetRowHeight(i)
		if i != 0 {
			h += v.spacing
		}
	}
	w := rect.Size.W + v.Margin.Size.W
	r := zgeo.Rect{pos, zgeo.Size{w, h}}
	v.stack.SetRect(r)
	v.layoutRows(-1)
	return v
}

func (v *ListView) layoutRows(onlyIndex int) (first, last int) {
	count := v.GetRowCount()
	ls := v.GetLocalRect().Size
	oldRows := map[int]View{}
	y := 0.0
	for k, v := range v.rows {
		if onlyIndex == -1 || k == onlyIndex {
			oldRows[k] = v
		}
	}
	first = -1
	// fmt.Println("\nlayout rows", len(oldRows), count)
	for i := 0; i < count; i++ {
		var s zgeo.Size
		s.H = v.GetRowHeight(i)
		s.W = ls.W + v.Margin.Size.W
		r := zgeo.Rect{zgeo.Pos{0, y}, s}
		if (onlyIndex == -1 || i == onlyIndex) && r.Max().Y >= v.topPos && r.Min().Y <= v.topPos+ls.H {
			if first == -1 {
				first = i
			}
			last = i
			row := v.rows[i]
			if row != nil {
				if row.Rect() != r {
					row.SetRect(r)
				}
				delete(oldRows, i)
			} else {
				row = v.CreateRow(s, i)
				v.stack.AddChild(row, -1)
				v.rows[i] = row
				v.setRowBGColor(i)
				row.SetRect(r)
			}
		}
		y += s.H + v.spacing
	}
	for i, view := range oldRows {
		v.stack.RemoveChild(view)
		delete(v.rows, i)
	}
	return
}

func (v *ListView) ExposeRows() {
	// for i in indexPathsForVisibleRows ?? [] {
	//     if let c = self.cellForRow(at i) {
	//         exposeAll(c.contentView)
	//     }
	// }
}

func (v *ListView) ScrollToMakeRowVisible(row int, animate bool) {
}

func (v *ListView) ReloadData() {
	for _, view := range v.rows {
		v.stack.RemoveChild(view)
	}
	v.rows = map[int]View{}
	v.layoutRows(-1)
}

func (v *ListView) MoveRow(fromIndex int, toIndex int) {
}

func (v *ListView) GetVisibleRowViewFromIndex(i int) View {
	return v.rows[i]
}

func (v *ListView) GetIndexFromRowView(view View) *int {
	return nil
}

func ListViewGetIndexFromRowView(view View) int {
	return -1
}

func (v *ListView) setRowBGColor(i int) {
	row := v.rows[i]
	if row != nil {
		col := v.selectedColor
		if !v.Selectable || v.selectionIndex == -1 || v.selectionIndex != i {
			col = v.RowColors[i%len(v.RowColors)]
		}
		row.SetBGColor(col)
	}
}

func (v *ListView) Select(i int) {
	v.ScrollToMakeRowVisible(i, false)
	old := v.selectionIndex
	v.selectionIndex = i
	if old != -1 {
		v.setRowBGColor(old)
	}
	v.setRowBGColor(i)
}

func (v *ListView) FlashSelect(i int) {
	count := 0
	v.Select(i)
	ztimer.Repeat(0.1, true, true, func() bool {
		if count%2 == 0 {
			v.Select(i)
		} else {
			v.Select(-1)
		}
		count++
		return (count < 8)
	})
}

func (v *ListView) UpdateRow(i int, edited bool) {
	if v.RowUpdater != nil {
		v.RowUpdater(i, edited)
	}
}

func (v *ListView) DeleteChildRow(i int, transition PresentViewTransition) { // call this after removing data
}

func (v *ListView) IsFocused() bool {
	return false
}

type ListViewIDGetter interface {
	GetID(index int) string
}

func (v *ListView) UpdateWithOldNewSlice(oldSlice, newSlice ListViewIDGetter) {
	i := 0
	reload := false
	for {
		oid := oldSlice.GetID(i)
		nid := newSlice.GetID(i)
		if nid != oid {
			reload = true
			break
		}
		if oid == "" || nid == "" {
			break
		}
		i++
	}
	// zlog.Info("UpdateWithOldNewSlice", reload)
	if reload {
		v.ReloadData()
	} else {
		v.UpdateVisibleRows()
	}
}

func (v *ListView) UpdateVisibleRows() {
	zlog.Info("ListView UpdateVisibleRows")
	first := -1
	last := -1
	y := 0.0
	count := v.GetRowCount()
	ls := v.GetLocalRect().Size
	for i := 0; i < count; i++ {
		e := y + v.GetRowHeight(i)
		if e >= v.topPos && y <= v.topPos+ls.H {
			if first == -1 {
				first = i
			}
			last = i
		}
		if y > v.topPos+ls.H {
			break
		}
		y = e + v.spacing
	}
	for i := first; i <= last; i++ {
		edited := false
		v.UpdateRow(i, edited)
	}
}
