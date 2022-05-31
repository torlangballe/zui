//go:build zui

package zgridlist

import (
	"math"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmap"
)

type GridListView struct {
	zui.ScrollView
	Spacing         zgeo.Size
	Selectable      bool
	MultiSelectable bool
	MaxColumns      int
	MinColumns      int
	//	CellIDs         []string

	BorderColor       zgeo.Color
	CellColor         zgeo.Color
	MultiplyAlternate float32
	PressedColor      zgeo.Color
	SelectColor       zgeo.Color
	HoverColor        zgeo.Color

	CellCount              func() int
	IDAtIndex              func(i int) string
	CreateCell             func(id string) zui.View
	UpdateCell             func(grid *GridListView, id string)
	UpdateSelection        func(id string)
	CellHeight             func(id string) float64 // only need to have variable-height
	HandleSelectionChanged func()
	HandkeKey              func(key zkeyboard.Key, mod zkeyboard.Modifier) bool
	children               map[string]zui.View
	selectedIDs            map[string]bool
	pressedIDs             map[string]bool
	selectingFromIndex     int
	appending              bool
	ignoreMouseEvent       bool
	grid                   *zui.CustomView
	margin                 zgeo.Rect
	columns                int
	rows                   int
	layoutDirty            bool
}

func New(name string) *GridListView {
	v := &GridListView{}
	v.Init(v, name)
	return v
}

func (v *GridListView) Init(view zui.View, name string) {
	v.ScrollView.Init(view, name)
	v.SetMinSize(zgeo.SizeBoth(100))
	v.MinColumns = 1
	v.children = map[string]zui.View{}
	v.selectedIDs = map[string]bool{}
	v.pressedIDs = map[string]bool{}
	v.SetObjectName("GridListView (Scroller)")
	v.grid = zui.CustomViewNew("GridListView.grid")
	v.AddChild(v.grid, -1)
	v.grid.SetPressUpDownMovedHandler(v.handleUpDownMovedHandler)

	v.PressedColor = zui.StyleDefaultFGColor().WithOpacity(0.6)
	v.HoverColor = zui.StyleDefaultFGColor().WithOpacity(0.6)
	v.SelectColor = zui.StyleDefaultFGColor().WithOpacity(0.3)
	v.SetScrollHandler(func(pos zgeo.Pos, infinityDir int) {
		// zlog.Info("Scroll:", pos)
		v.LayoutCells(false)
	})
}

func (v *GridListView) IndexOfID(id string) int {
	for i := 0; i < v.CellCount(); i++ {
		if v.IDAtIndex(i) == id {
			return i
		}
	}
	return -1
}

func (v *GridListView) IsSelected(id string) bool {
	return v.selectedIDs[id]
}

func (v *GridListView) SelectedIDs() []string {
	ids := make([]string, len(v.selectedIDs))
	i := 0
	for sid := range v.selectedIDs {
		ids[i] = sid
		i++
	}
	return ids
}

func (v *GridListView) SelectCell(id string, animateScroll bool) {
	v.SelectCells([]string{id}, animateScroll)
}

func (v *GridListView) SelectCells(ids []string, animateScroll bool) {
	v.selectedIDs = map[string]bool{}
	for _, id := range ids {
		v.selectedIDs[id] = true
	}
	v.layoutDirty = true
	for i := 0; i < v.CellCount(); i++ {
		sid := v.IDAtIndex(i)
		if v.selectedIDs[sid] {
			v.ScrollToCell(sid, animateScroll)
			break
		}
	}
	if v.layoutDirty { // this is to avoid layout if scroll did it
		v.LayoutCells(true)
	}
}

func (v *GridListView) IsPressed(id string) bool {
	return v.pressedIDs[id]
}

func (v *GridListView) CellView(id string) zui.View {
	return v.children[id]
}

func (v *GridListView) updateCellBackgrounds() {
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if !visible {
			return true
		}
		child := v.children[cid]
		// zlog.Assert(child != nil, x, y, cid)
		if child != nil {
			v.updateCellBackground(cid, x, y, child)
		}
		return true
	})
}

func (v *GridListView) updateCellBackground(cid string, x, y int, child zui.View) {
	if child == nil {
		child = v.children[cid]
		if child == nil {
			return
		}
	}
	col := v.CellColor
	if v.pressedIDs[cid] && v.PressedColor.Valid {
		col = v.PressedColor
	} else if v.selectedIDs[cid] && v.SelectColor.Valid {
		col = v.SelectColor
	}
	if col.Valid {
		if v.MultiplyAlternate != 0 && x%2 != y%2 {
			col = col.MultipliedBrightness(v.MultiplyAlternate)
		}
		child := v.children[cid]
		if child != nil {
			child.SetBGColor(col)
		}
	}
	if v.BorderColor.Valid {
		zui.ViewGetNative(child).SetStroke(1, v.BorderColor)
	}
	if v.UpdateSelection != nil {
		v.UpdateSelection(cid)
	}
}

func (v *GridListView) setPressed(index int) {
	// zlog.Info("setPressed", index, v.selectingFromIndex)
	if !v.MultiSelectable {
		v.pressedIDs[v.IDAtIndex(index)] = true
		return
	}
	if v.appending {
		for id := range v.selectedIDs {
			v.pressedIDs[id] = true
		}
	}
	if v.selectingFromIndex != -1 {
		v.pressedIDs = map[string]bool{}
		min, max := zint.MinMax(index, v.selectingFromIndex)
		for i := min; i <= max; i++ {
			v.pressedIDs[v.IDAtIndex(i)] = true
		}
	}
}

func (v *GridListView) handleUpDownMovedHandler(pos zgeo.Pos, down zbool.BoolInd) {
	if !v.Selectable && !v.MultiSelectable {
		return
	}
	var index int
	id, inside := v.FindCellForPos((pos))
	if id != "" {
		index = v.IndexOfID(id)
		if index == -1 {
			zlog.Info("No index for id:", id)
			return
		}
	}
	// zlog.Info("updown:", id, index, pos)
	switch down {
	case zbool.True:
		if !inside || id == "" {
			return
		}
		if zkeyboard.ModifiersAtPress&zkeyboard.ModifierCommand != 0 {
			v.ignoreMouseEvent = true
			if v.selectedIDs[id] {
				delete(v.selectedIDs, id)
			} else {
				v.selectedIDs[id] = true
			}
			break
		}
		v.pressedIDs[id] = true
		v.appending = (v.MultiSelectable && (zkeyboard.ModifiersAtPress&zkeyboard.ModifierShift != 0))
		if !v.appending || v.selectingFromIndex == -1 {
			v.selectingFromIndex = index
		}
		v.setPressed(index)
	case zbool.Unknown:
		if id == "" {
			return
		}
		if !v.ignoreMouseEvent && id != "" && v.MultiSelectable {
			// zlog.Info("Move", id)
			v.setPressed(index)
			// v.dragged = true
		}
	case zbool.False:
		if v.ignoreMouseEvent {
			v.ignoreMouseEvent = false
			break
		}
		if len(v.selectedIDs) > 0 && len(v.pressedIDs) == 1 {
			pid := zmap.GetAnyKeyAsString(v.pressedIDs)
			if v.selectedIDs[pid] {
				v.selectedIDs = map[string]bool{}
			} else {
				v.selectedIDs = v.pressedIDs
			}
		} else {
			v.selectedIDs = v.pressedIDs
		}
		v.selectingFromIndex = -1
		v.pressedIDs = map[string]bool{}
	}
	v.updateCellBackgrounds()
	v.ExposeIn(0.0001)
	if v.HandleSelectionChanged != nil {
		v.HandleSelectionChanged()
	}
}

func (v *GridListView) CellRects(cellID string) (fouter, finner zgeo.Rect) {
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if cid == cellID {
			finner = inner
			fouter = outer
			return false
		}
		return true
	})
	return
}

func (v *GridListView) FindCellForColRow(col, row int) (id string) {
	v.ForEachCell(func(cellID string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if col == x && row == y {
			id = cellID
			return false
		}
		return true
	})
	return
}

func (v *GridListView) FindCellForPos(pos zgeo.Pos) (id string, inside bool) {
	var before bool
	v.ForEachCell(func(cellID string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if pos.Y < outer.Pos.Y {
			before = true
		}
		if outer.Contains(pos) {
			// zlog.Info("box:", cellID, outer, pos)
			id = cellID
			inside = true
			return false
		}
		return true
	})
	if id == "" && v.CellCount() > 0 {
		if before {
			return v.IDAtIndex(0), false
		}
		return v.IDAtIndex(v.CellCount() - 1), false
	}
	return
}

func (v *GridListView) CalculateColumnsAndRows(childWidth, totalWidth float64) (nx, ny int) {
	nx = zgeo.MaxCellsInSize(v.Spacing.W, v.margin.Size.W, childWidth, totalWidth)
	if v.MaxColumns != 0 {
		zint.Minimize(&nx, v.MaxColumns)
	}
	zint.Maximize(&nx, v.MinColumns)
	ny = (v.CellCount() + nx - 1) / nx
	return
}

func (v *GridListView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.MinSize()
	if v.CellCount() == 0 {
		return v.MinSize()
	}
	childSize := v.getAChildSize(total)
	mx := math.Max(1, float64(v.MinColumns))
	w := childSize.W*mx + v.Spacing.W*(mx-1) - v.margin.Size.W
	zfloat.Maximize(&s.W, w)
	// zlog.Info("GridListView CS:", s, mx, w, childSize)
	return s
}

func (v *GridListView) CalculatedGridSize(total zgeo.Size) zgeo.Size {
	if v.CellCount() == 0 {
		return v.MinSize()
	}
	childSize := v.getAChildSize(total)
	nx, ny := v.CalculateColumnsAndRows(childSize.W, total.W)
	s := v.margin.Size.Negative()
	x := float64(nx)
	y := float64(ny)
	s.W += childSize.W*x + v.Spacing.W*(x-1)
	s.H += childSize.H*y + v.Spacing.H*(y-1)
	// zlog.Info(childSize, "CalculatedGridSize:", nx, ny, s, childSize.H*x, v.Spacing.H*(y-1))
	return s
}

func (v *GridListView) RemoveCell(id string) bool {
	child := v.children[id]
	if child != nil {
		v.grid.RemoveChild(child)
		delete(v.children, id)
		return true
	}
	return false
}

func (v *GridListView) getAChildSize(total zgeo.Size) zgeo.Size {
	cid := v.IDAtIndex(0)
	child := v.CreateCell(cid)
	s := child.CalculatedSize(total)
	return s
}

func (v *GridListView) GetChildren(collapsed bool) (children []zui.View) {
	for _, c := range v.children {
		children = append(children, c)
	}
	return
}

func (v *GridListView) SetRect(rect zgeo.Rect) {
	v.ScrollView.SetRect(rect)
	ct := v.View.(zui.ContainerType) // in case we are a stack or something inheriting from GridListView
	ct.ArrangeChildren()
}

func (v *GridListView) makeOrGetChild(id string) zui.View {
	child := v.children[id]
	if child != nil {
		return child
	}
	child = v.CreateCell(id)
	v.children[id] = child
	v.grid.AddChild(child, -1)
	zui.PresentViewCallReady(child, false)
	if v.UpdateCell != nil {
		// zlog.Info("UpdateNewCell", id, len(v.children))
		v.UpdateCell(v, id)
	}
	child.(zui.ExposableType).Expose()
	return child
}

func (v *GridListView) ForEachCell(got func(cellID string, outer, inner zgeo.Rect, x, y int, visible bool) bool) {
	if v.CellCount() == 0 {
		return
	}
	rect := v.LocalRect()
	pos := rect.Pos
	childSize := rect.Size

	if v.CellHeight == nil {
		childSize = v.getAChildSize(rect.Size)
	} else {
		zlog.Assert(v.MaxColumns <= 1)
	}
	// zlog.Info("4Each:", rect, childSize)
	v.columns, v.rows = v.CalculateColumnsAndRows(childSize.W, rect.Size.W)
	var x, y int
	for i := 0; i < v.CellCount(); i++ {
		cellID := v.IDAtIndex(i)
		if v.CellHeight != nil {
			childSize.H = v.CellHeight(cellID)
		}
		r := zgeo.Rect{Pos: pos, Size: childSize}
		lastx := (x == v.columns-1)
		lasty := (y == v.rows-1)
		var minx, maxx, miny, maxy float64
		if x == 0 {
			minx = v.margin.Min().X
		} else {
			minx = v.Spacing.W / 2
		}
		if lastx {
			maxx = -(rect.Max().X - r.Max().X)
		} else {
			maxx -= v.Spacing.W / 2
		}
		if y == 0 {
			miny = v.margin.Min().Y
		} else {
			miny += v.Spacing.H / 2
		}
		if lasty {
			maxy = v.margin.Max().Y
		} else {
			maxy -= v.Spacing.H / 2
		}
		marg := zgeo.RectFromXY2(minx, miny, maxx, maxy)
		r.Size.Add(marg.Size.Negative())
		cr := r.Plus(marg)
		visible := (r.Max().Y >= v.YOffset && r.Min().Y <= v.YOffset+v.LocalRect().Size.H)
		r2 := r
		if v.BorderColor.Valid {
			r2.Size.Subtract(zgeo.Size{1, 1})
		}
		// zlog.Info(cellID, "Vis:", marg, cellID, visible, r, v.YOffset, v.YOffset+v.LocalRect().Size.H)
		if !got(cellID, r2, cr, x, y, visible) {
			break
		}
		if lastx {
			pos.X = rect.Pos.X
			x = 0
			y++
			pos.Y += r.Size.H
		} else {
			pos.X += r.Size.W
			x++
		}
	}
}

func (v *GridListView) ArrangeChildren() {
	// zlog.Info("glist.ArrangeChildren")
	v.ScrollView.ArrangeChildren()
	s := v.CalculatedGridSize(v.LocalRect().Size)
	r := zgeo.Rect{Size: s}
	v.grid.SetRect(r)
	v.LayoutCells(false)
}

func (v *GridListView) ReplaceChild(child, with zui.View) {
	for id, c := range v.children {
		if c == child {
			v.children[id] = with
			zui.ViewGetNative(v).ReplaceChild(child, with)
			break
		}
	}
}

func (v *GridListView) LayoutCells(updateCells bool) {
	v.layoutDirty = false
	placed := map[string]bool{}
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if visible {
			// zlog.Info("Arrange:", cid, updateCells)
			child := v.makeOrGetChild(cid)
			//TODO: exit when !visible after visible
			ms, _ := child.(zui.Marginalizer)
			if ms != nil {
				marg := inner.Minus(outer)
				ms.SetMargin(marg)
			}
			o := outer.Plus(zgeo.RectFromXY2(0, 0, 0, 0))
			child.SetRect(o)
			v.updateCellBackground(cid, x, y, child)
			if updateCells && v.UpdateCell != nil {
				v.UpdateCell(v, cid)
			}
			placed[cid] = true
		}
		return true
	})
	for cid, view := range v.children {
		if !placed[cid] {
			v.grid.RemoveChild(view)
			delete(v.children, cid)
		}
	}
}

func (v *GridListView) ScrollToCell(cellID string, animate bool) {
	outerRect, _ := v.CellRects(cellID)
	if !outerRect.IsNull() {
		v.MakeRectVisible(outerRect.Plus(zgeo.RectFromXY2(0, 0, 0, 1)), animate)
	}
}

/*
func (v *GridListView) SetNewCellIDs(newIDs []string) {
	newChildren := map[string]zui.View{}
	newSelected := map[string]bool{}
	for _, nid := range newIDs {
		child := v.children[nid]
		if child != nil {
			newChildren[nid] = child
			delete(v.children, nid)
		}
		if v.selectedIDs[nid] {
			newSelected[nid] = true
		}
	}
	oldChildren := v.children
	v.children = newChildren
	v.CellIDs = newIDs
	v.selectedIDs = newSelected
	v.LayoutCells(true)
	for _, c := range oldChildren {
		v.grid.RemoveChild(c)
	}
}
*/

func (v *GridListView) moveSelection(incX, incY int, mod zkeyboard.Modifier) bool {
	var sx, sy int
	var fid string

	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if visible {
			if len(v.selectedIDs) == 0 || v.selectedIDs[cid] {
				fid = cid
				sx = x
				sy = y
				return false
			}
		}
		return true
	})
	if fid != "" {
		if len(v.selectedIDs) == 0 {
			v.selectedIDs[fid] = true
			v.updateCellBackground(fid, sx, sy, nil)
			return true
		} else {
			nx := sx + incX
			ny := sy + incY
			// zlog.Info("moveSel:", v.columns, v.rows, fid, "s:", sx, sy, "n:", nx, ny, ny*v.columns+nx >= len(v.CellIDs))
			if nx < 0 || nx >= v.columns || ny < 0 || ny >= v.rows {
				return true
			}
			if ny*v.columns+nx >= v.CellCount() {
				if incY != 1 {
					return true
				}
				nx = 0
			}
			v.selectedIDs = map[string]bool{}
			v.updateCellBackgrounds()
			nid := v.FindCellForColRow(nx, ny)
			// zlog.Info("MoveTo:", nid, len(v.children))
			if nid != "" {
				v.selectedIDs[nid] = true
				v.ScrollToCell(nid, false)
				v.updateCellBackground(nid, nx, ny, nil)
			}
		}
		if v.HandleSelectionChanged != nil {
			v.HandleSelectionChanged()
		}
	}
	return true
}

func (v *GridListView) ReadyToShow(beforeWindow bool) {
	// zlog.Info("List ReadyToShow:", beforeWindow)
	if !beforeWindow && (v.Selectable || v.MultiSelectable) {
		v.GetWindow().AddKeypressHandler(v.View, func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
			// zlog.Info("List keypress!", v.ObjectName(), key, mod == zkeyboard.ModifierNone)
			if mod == zkeyboard.ModifierNone {
				//					v.doRowPressed(v.highlightedIndex)
				switch key {
				case zkeyboard.KeyUpArrow:
					return v.moveSelection(0, -1, mod)
				case zkeyboard.KeyDownArrow:
					return v.moveSelection(0, 1, mod)
				case zkeyboard.KeyLeftArrow:
					return v.moveSelection(-1, 0, mod)
				case zkeyboard.KeyRightArrow:
					return v.moveSelection(1, 0, mod)
				case zkeyboard.KeyReturn, zkeyboard.KeyEnter:
					break // do select
				}
			}
			if v.HandkeKey != nil {
				return v.HandkeKey(key, mod)
			}
			return false
		})
	}
}

func (v *GridListView) SetMargin(m zgeo.Rect) {
	v.margin = m
}

func (v *GridListView) Margin() zgeo.Rect {
	return v.margin
}
