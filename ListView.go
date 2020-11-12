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
	HandleRowSelected    func(i int, selected bool)
	HandleScrolledToRows func(y float64, first, last int)

	RowColors  []zgeo.Color
	HoverColor zgeo.Color
	MinRows    int // MinRows is the minimum number of rows used to calculate size of list

	selectionIndex    int
	PressSelectable   bool
	PressUnselectable bool
	SelectedColor     zgeo.Color

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
	v.Init(v, name)
	return v
	//        allowsSelection = true // selectable
}

func (v *ListView) Init(view View, name string) {
	v.ScrollView.Init(view, name)
	v.selectionIndex = -1
	v.rows = map[int]View{}
	v.RowColors = []zgeo.Color{zgeo.ColorWhite}
	v.SelectedColor = DefaultSelectedColor
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

func (v *ListView) IsRowSelected(index int) bool {
	return index == v.selectionIndex
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
	s.Maximize(v.MinSize())
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
	// zlog.Info("ListView SetRect:", v.ObjectName(), h)

	w := rect.Size.W
	r := zgeo.Rect{pos, zgeo.Size{w, h}}
	v.stack.SetRect(r)
	// zlog.Info("List set rect: stack", rect, r)
	v.layoutRows(-1)
	return v
}

func (v *ListView) layoutRows(onlyIndex int) (first, last int) {
	// zlog.Info("layout rows", v.ObjectName())
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
				row = v.makeRow(s, i)
				v.stack.AddChild(row, -1)
				row.SetRect(r)
				// zlog.Info("LV Create Row:", i, r)
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
	v.SetRect(v.Rect()) // this will cause layoutrows and resizing of v.stack
}

func (v *ListView) ReloadRow(i int) {
	row, _ := v.rows[i]
	if row != nil {
		size := row.Rect().Size
		newRow := v.makeRow(size, i)
		ViewGetNative(v.stack).ReplaceChild(row, newRow)
		zlog.Info("ReloadRow:", i)
	}
}

func (v *ListView) makeRow(rowSize zgeo.Size, index int) View {
	row := v.CreateRow(rowSize, index)
	v.rows[index] = row
	v.setRowBGColor(index)
	if v.HoverColor.Valid {
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
	if v.PressSelectable {
		p, _ := row.(Pressable)
		if p != nil {
			old := p.PressedHandler()
			p.SetPressedHandler(func() {
				zlog.Info("Pressed", index, v.selectionIndex)
				if index == v.selectionIndex {
					if v.PressUnselectable {
						v.Unselect()
					}
				} else {
					v.Select(index)
				}
				if old != nil {
					old()
				}
			})
		}
	}
	return row
}

// func (v *ListView) MoveRow(fromIndex int, toIndex int) {
// }

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
		if v.selectionIndex != i {
			if len(v.RowColors) == 0 {
				col = zgeo.ColorWhite
			} else {
				col = v.RowColors[i%len(v.RowColors)]
			}
		}
		// zlog.Info("setRowBGColor", i, col)
		row.SetBGColor(col)
	}
}

func (v *ListView) Select(i int) {
	zlog.Info("SELECT:", i)
	v.ScrollToMakeRowVisible(i, false) // scroll first, so unselect doesn't update row that might not be visible anyway
	v.Unselect()
	v.selectionIndex = i
	v.HandleRowSelected(i, true)
	v.setRowBGColor(i) // this must be after v.HandleRowSelected, as it might make a new row, also for old above
}

func (v *ListView) Unselect() {
	old := v.selectionIndex
	v.selectionIndex = -1
	if old != -1 {
		zlog.Info("Unselect:", old)
		v.HandleRowSelected(old, false)
		v.setRowBGColor(old)
	}
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
	// zlog.Info("UpdateWithOldNewSlice:", v.selectionIndex, zlog.GetCallingStackString())
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
	} else if v.selectionIndex != -1 && oldSelectionIndex != v.selectionIndex {
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
