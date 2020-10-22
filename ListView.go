package zui

import (
	"github.com/torlangballe/zutil/zgeo"
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

	RowColors  []zgeo.Color
	HoverColor zgeo.Color
	MinRows    int

	selectionIndex int
	Selectable     bool
	SelectedColor  zgeo.Color

	// topPos float64
	stack *CustomView
	rows  map[int]View
}

type ListViewIDGetter interface {
	GetID(index int) string
}

var DefaultSelectedColor = zgeo.ColorNew(0.4, 0.4, 0.8, 1)

func ListViewNew(name string) *ListView {
	v := &ListView{}
	v.init(v, name)
	v.RowColors = []zgeo.Color{zgeo.ColorWhite}
	v.SelectedColor = DefaultSelectedColor
	return v
	//        allowsSelection = true // selectable
}

func (v *ListView) init(view View, name string) {
	v.ScrollView.init(view, name)
	v.Selectable = true
	v.selectionIndex = -1
	v.rows = map[int]View{}
	v.SetScrollHandler(func(pos zgeo.Pos, infinityDir int) {
		// zlog.Info("ScrollTo:", pos.Y)
		// v.topPos = pos.Y
		first, last := v.layoutRows(-1)
		if v.HandleScrolledToRows != nil {
			v.HandleScrolledToRows(pos.Y, first, last)
		}
	})
}

func (v *ListView) SelectionIndex() int {
	return v.selectionIndex
}

func (v *ListView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.ScrollView.CalculatedSize(total)
	h := 0.0
	count := v.GetRowCount()
	for i := 0; i < count; i++ {
		if v.MinRows != 0 && i >= v.MinRows {
			break
			// something happening here, no size calculated in table when nested in something
		}
		h += v.GetRowHeight(i)
		// zlog.Info("ListView.CalculatedSize2:", v.ObjectName(), i, h)
	}
	s.H = h
	// zlog.Info("ListView.CalculatedSize:", v.ObjectName(), s, count, v.MinRows)

	return s
}

func (v *ListView) drawIfExposed() {
	// zlog.Info("LV:drawIfExposed")
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

func (v *ListView) SetRect(rect zgeo.Rect) View {
	// zlog.Info("ListView SetRect:", v.ObjectName(), rect)
	v.ScrollView.SetRect(rect)
	if v.stack == nil {
		v.stack = CustomViewNew("listview.stack")
		v.AddChild(v.stack, -1)
	}
	count := v.GetRowCount()
	var pos zgeo.Pos
	h := 0.0
	for i := 0; i < count; i++ {
		h += v.GetRowHeight(i)
		if i != 0 {
			h += v.spacing
		}
	}
	w := rect.Size.W
	r := zgeo.Rect{pos, zgeo.Size{w, h}}
	v.stack.SetRect(r)
	// zlog.Info("List set rect: stack", rect, r)
	v.layoutRows(-1)
	return v
}

func (v *ListView) layoutRows(onlyIndex int) (first, last int) {
	// zlog.Info("layout rows", v.ObjectName(), zlog.GetCallingStackString())
	count := v.GetRowCount()
	ls := v.LocalRect().Size
	oldRows := map[int]View{}
	y := 0.0
	for k, v := range v.rows {
		if onlyIndex == -1 || k == onlyIndex {
			oldRows[k] = v
		}
	}
	first = -1
	// zlog.Info("\nlayout rows", len(oldRows), count)
	// start := time.Now()
	for i := 0; i < count; i++ {
		var s zgeo.Size
		s.H = v.GetRowHeight(i)
		s.W = ls.W
		r := zgeo.Rect{zgeo.Pos{0, y}, s}
		// zlog.Info("layout row:", i, v.YOffset)
		if (onlyIndex == -1 || i == onlyIndex) && r.Max().Y >= v.YOffset && r.Min().Y <= v.YOffset+ls.H {
			// zlog.Info("actually layout row:", i)
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
				// tart := time.Now()
				row = v.CreateRow(s, i)
				if v.HoverColor.Valid {
					index := i
					ViewGetNative(row).SetPointerEnterHandler(func(inside bool) {
						if v.GetRowCount() > 1 {
							if inside {
								row.SetBGColor(v.HoverColor)
							} else {
								v.setRowBGColor(index)
							}
						}
					})
				}
				// zlog.Info("LV Create Row:", time.Since(start))
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
	et, _ := v.View.(ExposableType)
	if et != nil {
		et.drawIfExposed()
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
	first, last := v.GetFirstLastVisibleRowIndexes()
	// zlog.Info("ScrollToMakeRowVisible:", row, first, last)
	if row <= first || row >= last {
		rect := v.GetRectOfRow(row)
		if row <= first {
			v.ScrollView.SetContentOffset(rect.Pos.Y, animate)
		} else {
			h := v.LocalRect().Size.H
			v.ScrollView.SetContentOffset(rect.Pos.Y-h+rect.Size.H, animate)
		}
	}
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
		col := v.SelectedColor
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
	ztimer.RepeatNow(0.1, func() bool {
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

func findViewInRow(row, find View) bool {
	ct, _ := row.(ContainerType)
	found := false
	if ct != nil {
		ContainerTypeRangeChildren(ct, true, func(view View) bool {
			if view == find {
				found = true
				return false
			}
			return true
		})
	}
	return found
}

func (v *ListView) GetFocusedParts() (element, row View, index int) {
	focused := v.GetFocusedView()
	if focused == nil {
		return nil, nil, -1
	}
	for index, row := range v.rows {
		if findViewInRow(row, focused) {
			// zlog.Info("GetFocusedParts.find:", index, row.ObjectName())
			return focused, row, index
		}
	}
	return nil, nil, -1
}

func (v *ListView) UpdateWithOldNewSlice(oldSlice, newSlice ListViewIDGetter) {
	// zlog.Info("update:", v.selectionIndex)
	i := 0
	reload := false
	var oldSelectionIndex = v.selectionIndex
	var selectionID, focusedID, focusedObjectName string
	selectionSet := true
	different := false
	if v.selectionIndex != -1 {
		selectionID = oldSlice.GetID(v.selectionIndex)
		selectionSet = false
	}
	focusedView, _, focusedIndex := v.GetFocusedParts()
	if focusedIndex != -1 {
		focusedID = oldSlice.GetID(focusedIndex)
		focusedObjectName = focusedView.ObjectName()
	}
	for j := 0; ; j++ {
		sid := oldSlice.GetID(j)
		if sid == "" {
			break
		}
		// zlog.Info("OLDID:", sid)
	}

	// zlog.Info("UpdateWithOldNewSlice focus is:", focusedView != nil, focusedIndex, focusedID)
	// zlog.Info("Sel:", selectionID, i, v.selectionIndex)
	for {
		oid := oldSlice.GetID(i)
		nid := newSlice.GetID(i)
		// zlog.Info("new id", nid)
		if nid != "" && focusedID == nid {
			focusedIndex = i
		} else if nid != "" && selectionID == nid {
			// zlog.Info("Found new selection:", i, nid)
			v.selectionIndex = i
			if different && focusedIndex == -1 {
				break
			}
			selectionSet = true
		}
		if nid != oid {
			different = true
			reload = true
			if selectionSet && focusedIndex == -1 {
				break
			}
		}
		if oid == "" || nid == "" {
			break
		}
		i++
	}
	// zlog.Info("UpdateWithOldNewSlice", reload, v.selectionIndex, oldSelectionIndex)
	if focusedIndex != -1 {
		v.ScrollToMakeRowVisible(focusedIndex, false)
	} else if oldSelectionIndex != v.selectionIndex {
		v.ScrollToMakeRowVisible(v.selectionIndex, false)
	}
	if reload {
		v.ReloadData()
	} else {
		v.UpdateVisibleRows()
	}
	if focusedIndex != -1 {
		// zlog.Info("UpdateWithOldNewSlice focus to", reload, focusedIndex, focusedObjectName)
		// for n, v := range v.rows {
		// 	zlog.Info("UpdateWithOldNewSlice rows", n, v.ObjectName())
		// }
		row, _ := v.rows[focusedIndex]
		if row != nil {
			ct := row.(ContainerType)
			newFocused := ContainerTypeFindViewWithName(ct, focusedObjectName, true)
			if newFocused != nil {
				newFocused.Focus(true)
			}
		}
	}
}

func (v *ListView) GetFirstLastVisibleRowIndexes() (first int, last int) {
	first = -1
	last = -1
	// if !v.Presented {
	// 	return
	// }
	y := 0.0
	count := v.GetRowCount()
	ls := v.LocalRect().Size
	for i := 0; i < count; i++ {
		e := y + v.GetRowHeight(i)
		if e >= v.YOffset && y <= v.YOffset+ls.H {
			if first == -1 {
				first = i
			}
			last = i
		}
		if y > v.YOffset+ls.H {
			break
		}
		y = e + v.spacing
	}
	return
}

func (v *ListView) GetRectOfRow(row int) zgeo.Rect {
	if !v.Presented {
		return zgeo.Rect{}
	}
	y := 0.0
	count := v.GetRowCount()
	ls := v.LocalRect().Size
	for i := 0; i < count; i++ {
		var s zgeo.Size
		s.H = v.GetRowHeight(i)
		if row == i {
			s.W = ls.W
			r := zgeo.Rect{zgeo.Pos{0, y}, s}
			return r
		}
		y += s.H + v.spacing
	}
	return zgeo.Rect{}
}

func (v *ListView) UpdateVisibleRows() {
	first, last := v.GetFirstLastVisibleRowIndexes()
	for i := first; i <= last; i++ {
		edited := false
		v.UpdateRow(i, edited)
	}
}

func (v *ListView) GetChildren() []View {
	var views []View
	for _, v := range v.rows {
		views = append(views, v)
	}
	return views
}

func (v *ListView) ArrangeChildren(onlyChild *View) {
	// zlog.Info("ListView ArrangeChildren:", len(v.rows))
	for _, v := range v.rows {
		ct, _ := v.(ContainerType)
		if ct != nil {
			ct.ArrangeChildren(nil)
		}
	}
}

func (v *ListView) ReplaceChild(child, with View) {

}
