// +build zui

package zui

import (
	"time"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

//  Created by Tor Langballe on /4/12/15.

type ListView struct {
	ScrollView
	//	Scrolling bool
	spacing float64

	GetRowCount          func() int
	GetRowHeight         func(i int) float64
	CreateRow            func(rowSize zgeo.Size, i int) View
	RowUpdater           func(i int, edited bool)
	HandleRowSelected    func(i int, selected, fromPress bool)
	HandleScrolledToRows func(y float64, first, last int)

	GetRowColor      func(i int) zgeo.Color
	HighlightColor   zgeo.Color
	highlightedIndex int
	MinRows          int // MinRows is the minimum number of rows used to calculate size of list

	selectionIndexes     map[int]bool
	PressSelectable      bool
	PressUnselectable    bool
	MultiSelect          bool
	HoverHighlight       bool
	ExposeSetsRowBGColor bool
	PreCreateRows        int
	SelectedColor        zgeo.Color
	HasUniformHight      bool

	// topPos float64
	scrollTimer      *ztimer.Timer
	scrollToYOnTimer float64
	stack            *CustomView
	rows             map[int]View
	rowHeights       map[int]float64
}

type ListViewIDGetter interface {
	GetID(index int) string
}

var ListViewDefaultSelectedColor = StyleGrayF(0.6, 0.5)
var ListViewDefaultBGColor = StyleGrayF(0.9, 0.45)

func ListViewNew(name string, selection map[int]bool) *ListView {
	v := &ListView{}
	v.Init(v, name, selection)
	return v
	//        allowsSelection = true // selectable
}

func (v *ListView) Init(view View, name string, selection map[int]bool) {
	v.ScrollView.Init(view, name)
	v.rows = map[int]View{}
	v.rowHeights = map[int]float64{}
	v.scrollTimer = ztimer.TimerNew()
	v.GetRowColor = func(i int) zgeo.Color {
		return zgeo.ColorWhite
	}
	v.SelectedColor = ListViewDefaultSelectedColor()
	if selection != nil {
		v.selectionIndexes = selection
	} else {
		v.selectionIndexes = map[int]bool{}
	}
	v.highlightedIndex = -1
	v.SetBGColor(ListViewDefaultBGColor())
	v.SetScrollHandler(func(pos zgeo.Pos, infinityDir int) {
		// zlog.Info("ScrollTo:", pos.Y)
		// v.topPos = pos.Y
		v.scrollToYOnTimer = pos.Y
		if !v.scrollTimer.IsRunning() {
			v.scrollTimer.StartIn(0.2, func() {
				first, last := v.layoutRows()
				if v.HandleScrolledToRows != nil {
					v.HandleScrolledToRows(v.scrollToYOnTimer, first, last)
				}
			})
		}
	})
}

func (v *ListView) SelectionIndex() int {
	for i := range v.selectionIndexes {
		return i
	}
	return -1
}

func (v *ListView) IsRowSelected(index int) bool {
	return v.selectionIndexes[index]
}

func (v *ListView) IsRowHighlighted(index int) bool {
	return v.highlightedIndex == index
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
		h += getRowHeight(v, i, total)
		// zlog.Info("ListView.CalculatedSize2:", v.ObjectName(), i, h)
	}
	s.H = h
	s.Maximize(v.MinSize())
	// zlog.Info("ListView.CalculatedSize:", v.ObjectName(), s, count, v.MinRows)

	return s
}

func (v *ListView) SetSpacing(spacing float64) *ListView {
	v.spacing = spacing
	return v
}

func (v *ListView) Spacing() float64 {
	return v.spacing
}

func (v *ListView) SetRect(rect zgeo.Rect) {
	//  zlog.Info("ListView SetRect:", v.ObjectName(), rect)
	v.ScrollView.SetRect(rect)
	if v.stack == nil {
		v.stack = CustomViewNew("listview.stack")
		v.SetCanFocus(false)
		v.AddChild(v.stack, -1)
		if v.highlightedIndex == -1 && v.HighlightColor.Valid {
			v.highlightedIndex = 0
		}
	}
	count := v.GetRowCount()
	var pos zgeo.Pos
	h := 0.0
	for i := 0; i < count; i++ {
		h += getRowHeight(v, i, rect.Size)
		if i != 0 {
			h += v.spacing
		}
	}
	// zlog.Info("ListView SetRect:", v.ObjectName(), h)

	w := rect.Size.W
	r := zgeo.Rect{pos, zgeo.Size{w, h}}
	v.stack.SetRect(r)
	// zlog.Info("List set rect: stack", rect, r)
	v.layoutRows()
}

func getRowHeight(v *ListView, i int, total zgeo.Size) float64 {
	if v.GetRowHeight != nil {
		return v.GetRowHeight(i)
	}
	if v.HasUniformHight {
		for _, h := range v.rowHeights {
			return h
		}
	}
	h := v.rowHeights[i]
	if h != 0 {
		return h
	}
	row := v.makeRow(total, i)
	h = row.CalculatedSize(total).H
	//	fmt.Printf("CreateRowH: %p %f\n", row, h)
	v.rowHeights[i] = h
	return h
}

// var ListLayoutStart time.Time

func (v *ListView) layoutRows() (first, last int) {
	// zlog.Info("layout rows", v.ObjectName())
	// ListLayoutStart = time.Now()
	count := v.GetRowCount()
	ls := v.LocalRect().Size
	oldRows := map[int]View{}
	y := 0.0
	for k, v := range v.rows {
		oldRows[k] = v
	}
	first = -1
	// zlog.Info("\nlayout rows", len(oldRows), count)
	// start := time.Now()
	for i := 0; i < count; i++ {
		var s zgeo.Size
		s.H = getRowHeight(v, i, v.Rect().Size)
		s.W = ls.W
		r := zgeo.Rect{zgeo.Pos{0, y}, s}
		if r.Min().Y > v.YOffset+ls.H {
			break
		}
		if r.Max().Y >= v.YOffset {
			// zlog.Info("actually layout row:", i)
			if first == -1 {
				// zlog.Info("layout rows", v.ObjectName(), i, y)
				first = i
			}
			last = i
			row := v.rows[i]
			if row != nil {
				nv := ViewGetNative(row)
				calc := !nv.Presented || (row.Rect() != r)
				ct, _ := row.(ContainerType)
				if ct != nil {
					includeCollapsed := false
					ContainerTypeRangeChildren(ct, true, includeCollapsed, func(view View) bool {
						nv := ViewGetNative(view)
						if !nv.Presented {
							calc = true
							return false
						}
						return true
					})
				}
				// fmt.Printf("ListLay: %s %p %v %p\n", nv.Hierarchy(), nv, calc)
				if nv.Parent() == nil {
					calc = true
					v.stack.AddChild(row, -1)
				}
				//! v.UpdateRow(i, false)
				PresentViewCallReady(row, false) // do we need this?
				if calc {
					row.SetRect(r)
				}
				//!	v.refreshRow(i)
				delete(oldRows, i)
			} else {
				// tart := time.Now()
				row = v.makeRow(s, i)
				v.stack.AddChild(row, -1)
				PresentViewCallReady(row, false)
				row.SetRect(r)
				v.refreshRow(i)
			}
		}
		y += s.H + v.spacing
	}
	if v.PreCreateRows > 0 {
		go v.preCreateRows(first, last)
	}
	for i, view := range oldRows {
		// zlog.Info("LV DelRow:", view.ObjectName(), view.Rect())
		v.stack.RemoveChild(view)
		delete(v.rows, i)
	}
	// et, _ := v.View.(ExposableType)
	// if et != nil {
	// 	et.drawIfExposed()
	// }
	// zlog.Info("ListLayout done:", time.Since(ListLayoutStart))
	return
}

func (v *ListView) preCreateRows(before, after int) {
	ls := v.LocalRect().Size
	for i := after; i < v.GetRowCount(); i++ {
		if i > after+v.PreCreateRows {
			break
		}
		_, got := v.rows[i]
		if !got {
			var s zgeo.Size
			s.H = getRowHeight(v, i, v.Rect().Size)
			s.W = ls.W
			v.makeRow(s, i)
			// PresentViewCallReady(row, false)
		}
	}
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
		if ViewGetNative(view).Parent() != nil {
			v.stack.RemoveChild(view)
		}
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
		// zlog.Info("ReloadRow:", i)
	}
}

func (v *ListView) refreshRow(index int) {
	row, _ := v.rows[index]
	if row != nil {
		et, got := row.(ExposableType)
		if got {
			et.Expose()
		}
	}
	if !v.ExposeSetsRowBGColor {
		v.UpdateRowBGColor(index)
	}
}

func (v *ListView) makeRow(rowSize zgeo.Size, index int) View {
	row := v.CreateRow(rowSize, index)
	// fmt.Printf("CreateRow: %p %d\n", row, index)
	nv := ViewGetNative(row)
	v.rows[index] = row
	// v.refreshRow(index)
	if v.HoverHighlight && v.HighlightColor.Valid {
		nv.SetPointerEnterHandler(false, func(pos zgeo.Pos, inside zbool.BoolInd) {
			if time.Since(v.ScrolledAt) < time.Second {
				return
			}
			if v.GetRowCount() > 1 {
				old := v.highlightedIndex
				if inside.Bool() {
					v.highlightedIndex = index
				} else {
					v.highlightedIndex = -1
				}
				if old != -1 && old != v.highlightedIndex {
					v.refreshRow(old)
				}
				if inside.Bool() {
					v.refreshRow(index)
				}
			}
		})
	}
	if v.PressSelectable {
		p, _ := row.(Pressable)
		if p != nil {
			old := p.PressedHandler()
			p.SetPressedHandler(func() {
				// zlog.Info("Pressed", index, v.selectionIndexes, old)
				v.doRowPressed(index)
				if old != nil {
					old()
				}
			})
		}
	}
	return row
}

func (v *ListView) doRowPressed(index int) {
	if v.selectionIndexes[index] {
		if v.MultiSelect || v.PressUnselectable {
			v.Unselect(index, true)
		}
	} else {
		v.Select(index, false, true)
	}

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

func (v *ListView) UpdateRowBGColor(i int) {
	row := v.rows[i]
	if row != nil {
		col := v.SelectedColor
		if !v.selectionIndexes[i] || !col.Valid {
			col = v.GetRowColor(i)
			if v.HighlightColor.Valid && i == v.highlightedIndex {
				col = col.Mixed(v.HighlightColor, v.HighlightColor.Opacity())
			}
		}
		// zlog.Info("COL:", col, i, v.selectionIndexes[i], v.HighlightColor)
		// zlog.Info("UpdateRowBGColor", i, col, v.selectionIndexes[i], v.highlightedIndex)
		row.SetBGColor(col)
	}
}

func (v *ListView) Select(i int, scrollTo, fromPress bool) {
	// zlog.Info("SELECT:", i)
	if scrollTo {
		v.ScrollToMakeRowVisible(i, false) // scroll first, so unselect doesn't update row that might not be visible anyway
	}
	if !v.MultiSelect {
		v.UnselectAll()
	}
	v.selectionIndexes[i] = true
	if v.HandleRowSelected != nil {
		v.HandleRowSelected(i, true, fromPress)
	}
	v.refreshRow(i) // this must be after v.HandleRowSelected, as it might make a new row, also for old above
}

func (v *ListView) UnselectAll() {
	if v.HandleRowSelected != nil {
		for index := range v.selectionIndexes {
			v.HandleRowSelected(index, false, false)
		}
	}
	oldIndexes := v.selectionIndexes
	v.selectionIndexes = map[int]bool{}
	for index := range oldIndexes {
		v.refreshRow(index)
	}
}

func (v *ListView) Unselect(index int, fromPress bool) {
	// zlog.Info("Unselect:", index)
	if v.selectionIndexes[index] {
		delete(v.selectionIndexes, index)
		v.HandleRowSelected(index, false, fromPress)
		v.refreshRow(index)
	}
}

func (v *ListView) FlashSelect(i int) {
	count := 0
	v.Select(i, true, false)
	ztimer.RepeatNow(0.1, func() bool {
		if count%2 == 0 {
			v.Select(i, false, false)
		} else {
			v.Unselect(i, false)
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
		includeCollapsed := false
		ContainerTypeRangeChildren(ct, true, includeCollapsed, func(view View) bool {
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
	oldSelectionIndex := v.SelectionIndex()
	oldSelections := v.selectionIndexes
	v.selectionIndexes = map[int]bool{}
	selectIDs := map[int]string{}
	var focusedID, focusedObjectName string
	for i := range oldSelections {
		selectIDs[i] = oldSlice.GetID(i)
		// zlog.Info("UpdateWithOldNewSlice:", v.selectionIndex, "selid:", selectionID)
	}
	focusedView, _, focusedIndex := v.GetFocusedParts()
	if focusedIndex != -1 {
		focusedID = oldSlice.GetID(focusedIndex)
		focusedObjectName = focusedView.ObjectName()
	}
	// zlog.Info("UpdateWithOldNewSlice focus is:", focusedView != nil, focusedIndex, focusedID)
	// defer zlog.Info("UpdateWithOldNewSlice done")
	for {
		oid := oldSlice.GetID(i)
		nid := newSlice.GetID(i)
		// zlog.Info("new id", i, nid, "oid:", oid, "selid:", selectionID)
		if nid != "" && focusedID == nid {
			focusedIndex = i
		} else if nid != "" && selectIDs[i] == nid {
			// zlog.Info("Found new selection:", i, nid)
			v.selectionIndexes[i] = true
			delete(selectIDs, i)
			// if different && focusedIndex == -1 {
			// 	break
			// }
		}
		if nid != oid {
			reload = true
			if len(selectIDs) == 0 && focusedIndex == -1 {
				// zlog.Info("break2")
				break
			}
		}
		if oid == "" && nid == "" {
			break
		}
		if (oid == "" || nid == "") && len(selectIDs) == 0 {
			// zlog.Info("break3")
			break
		}
		i++
	}
	if focusedIndex != -1 {
		v.ScrollToMakeRowVisible(focusedIndex, false)
	} else if !v.MultiSelect && v.SelectionIndex() != -1 && oldSelectionIndex != v.SelectionIndex() {
		v.ScrollToMakeRowVisible(v.SelectionIndex(), false)
	}
	// zlog.Info("UpdateWithOldNewSlice", reload, v.selectionIndex, oldSelectionIndex)
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
			newFocused, _ := ContainerTypeFindViewWithName(ct, focusedObjectName, true)
			if newFocused != nil {
				// zlog.Info("listview focus something")
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
		e := y + getRowHeight(v, i, v.Rect().Size)
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
		s.H = getRowHeight(v, i, v.Rect().Size)
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
	edited := false
	for i := first; i <= last; i++ {
		v.UpdateRow(i, edited)
	}
	// go func() {
	// 	for i, _ := range v.rows {
	// 		if i < first || i > last {
	// 			v.UpdateRow(i, edited)
	// 		}
	// 	}
	// }()
}

func (v *ListView) GetChildren(includeCollapsed bool) []View {
	var views []View
	for _, v := range v.rows {
		views = append(views, v)
	}
	return views
}

func (v *ListView) ArrangeChildren() {
	// zlog.Info("ListView ArrangeChildren:", len(v.rows))
	for _, v := range v.rows {
		ct, _ := v.(ContainerType)
		if ct != nil {
			ct.ArrangeChildren()
		}
	}
}

func (v *ListView) ReplaceChild(child, with View) {
	zlog.Fatal(nil, "not implemented")
}

func (v *ListView) moveHighlight(delta int) {
	old := v.highlightedIndex
	n := zint.Max(zint.Min(v.highlightedIndex+delta, v.GetRowCount()-1), 0)
	if n != v.highlightedIndex {
		// zlog.Info("moveHighlight:", v.highlightedIndex, "->", n)
		v.highlightedIndex = n
		if old != -1 {
			v.refreshRow(old)
		}
		v.refreshRow(n)
		animate := false
		v.ScrollToMakeRowVisible(n, animate)
	}
}

func (v *ListView) ReadyToShow(beforeWindow bool) {
	// zlog.Info("List ReadyToShow:", beforeWindow)
	if !beforeWindow && v.HighlightColor.Valid {
		win := v.GetWindow()
		win.AddKeypressHandler(v.View, func(key KeyboardKey, mod KeyboardModifier) {
			// zlog.Info("List keypress!", v.ObjectName(), key, mod == KeyboardModifierNone)
			switch key {
			case KeyboardKeyTab:
				if v.HighlightColor.Valid && v.highlightedIndex != -1 {
					row := v.rows[v.highlightedIndex]
					ViewGetNative(row).FocusNext(mod != KeyboardModifierShift)
				}
			case KeyboardKeyUpArrow:
				if mod == KeyboardModifierNone {
					v.moveHighlight(-1)
				}
			case KeyboardKeyDownArrow:
				if mod == KeyboardModifierNone {
					v.moveHighlight(1)
				}
			case KeyboardKeyReturn, KeyboardKeyEnter:
				if mod == KeyboardModifierNone {
					if v.PressSelectable && v.HighlightColor.Valid && v.highlightedIndex != -1 {
						v.doRowPressed(v.highlightedIndex)
					}
				}
			}
		})
	}
}
