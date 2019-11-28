package zgo

import "fmt"

//  Created by Tor Langballe on /4/12/15.

type ListView struct {
	ScrollView
	//ListViewNative
	//	Scrolling bool
	spacing float64

	GetRowCount          func() int
	GetRowHeight         func(i int) float64
	CreateRow            func(rowSize Size, i int) View
	HandleRowSelected    func(i int)
	HandleScrolledToRows func(y float64, first, last int)

	RowColors []Color

	selectionIndex int
	Selectable     bool
	selectedColor  Color

	topPos float64
	stack  *CustomView
	rows   map[int]View
}

func ListViewNew(name string) *ListView {
	v := &ListView{}
	v.init(v, name)
	v.RowColors = []Color{ColorWhite}
	return v
	//        allowsSelection = true // selectable
}

func (v *ListView) drawIfExposed() {
	for _, view := range v.rows {
		et, got := view.(ExposableType)
		if got {
			et.drawIfExposed()
		}
	}
}

func (v *ListView) Spacing(spacing float64) *ListView {
	v.spacing = spacing
	return v
}

func (v *ListView) GetSpacing() float64 {
	return v.spacing
}

func (v *ListView) SelectedColor(col Color) *ListView {
	v.selectedColor = col
	return v
}

func (v *ListView) init(view View, name string) {
	v.ScrollView.init(view, name)
	v.Selectable = true
	v.selectionIndex = -1
	v.rows = map[int]View{}
	v.HandleScroll = func(pos Pos) {
		v.topPos = pos.Y
		first, last := v.layoutRows()
		if v.HandleScrolledToRows != nil {
			v.HandleScrolledToRows(pos.Y, first, last)
		}
	}
}

func (v *ListView) Rect(rect Rect) View {
	v.ScrollView.Rect(rect)
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
	r := Rect{pos, Size{w, h}}
	v.stack.Rect(r)
	v.layoutRows()
	return v
}

func (v *ListView) layoutRows() (first, last int) {
	count := v.GetRowCount()
	ls := v.GetLocalRect().Size
	oldRows := map[int]View{}
	y := 0.0
	for k, v := range v.rows {
		oldRows[k] = v
	}
	first = -1
	//	fmt.Println("\nlayout rows")
	for i := 0; i < count; i++ {
		var s Size
		s.H = v.GetRowHeight(i)
		s.W = ls.W + v.Margin.Size.W
		r := Rect{Pos{0, y}, s}
		if r.Max().Y >= v.topPos && r.Min().Y <= v.topPos+ls.H {
			if first == -1 {
				first = i
			}
			last = i
			row := v.rows[i]
			if row != nil {
				if row.GetRect() != r {
					row.Rect(r)
				}
				delete(oldRows, i)
			} else {
				row = v.CreateRow(s, i)
				v.stack.AddChild(row, -1)
				v.rows[i] = row
				v.setRowBGColor(i)
				row.Rect(r)
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

func (v *ListView) UpdateVisibleRows(animate bool) {
	//  reloadRows(at indexPathsForVisibleRows ?? [], withanimate ? UIListView.RowAnimation.automatic  UIListView.RowAnimation.none)
}

func (v *ListView) ScrollToMakeRowVisible(row int, animate bool) {
}

func (v *ListView) UpdateRow(row int) {
}

func (v *ListView) ReloadData() {
	v.layoutRows()
}

func (v *ListView) MoveRow(fromIndex int, toIndex int) {
}

func (v *ListView) GetRowViewFromIndex(i int) *View {
	return nil
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
		row.BGColor(col)
	}
}

func (v *ListView) Select(i int) {
	old := v.selectionIndex
	v.selectionIndex = i
	if old != -1 {
		v.setRowBGColor(old)
	}
	fmt.Println("List Select:", i)
	v.setRowBGColor(i)
}

func (v *ListView) DeleteChildRow(i int, transition PresentViewTransition) { // call this after removing data
}

func (v *ListView) IsFocused() bool {
	return false
}
