// The zgridlist package defines GridListView, a view for displaying cells in multiple rows and columns.
// It knows nothing about the content of its cells.
// If CellHeightFunc is set, each cells hight is based on that, otherwise based on the first cell created.
// See also the slicegridview package which expands the grid list to use slices, make tables with headers and automatic sql tables.
// Also TableView, which use a GridListView and zfield package to create rows based on a struct.
// Cells can be hovered over, pressed and selected. Multiple selections are possible.
// If HierarchyLevelFunc is set, it will insert a branch toggle (BrangeToggleView) widgets if level returned > 1 and leaf is false,
// but only calls LayoutCells when this toggles are changed.
// Hierarchy toggle states in OpenBranches are stored in zkeyvalue.DefaultStore based on the grids storeName.

//go:build zui

package zgridlist

import (
	"math"
	"strconv"
	"strings"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zscrollview"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmap"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
)

type GridListView struct {
	zscrollview.ScrollView
	Spacing                   zgeo.Size
	Selectable                bool
	MultiSelectable           bool
	MakeFullSize              bool
	RecreateCells             bool // RecreateCells forces creation of new cells on next layout
	MaxColumns                int
	MinColumns                int
	MinRowsForFullSize        int
	MaxRowsForFullSize        int
	BorderColor               zgeo.Color
	CellColor                 zgeo.Color
	CellColorFunc             func(id string) zgeo.Color // Always set
	MultiplyColorAlternate    float32
	PressedColor              zgeo.Color
	SelectColor               zgeo.Color
	HoverColor                zgeo.Color
	BranchToggleType          zwidgets.BranchToggleType
	OpenBranches              map[string]bool
	CurrentHoverID            string
	FocusWidth                float64
	HorizontalFirst           bool // HorizontalFirst means 1 2 3 on first row, not down first
	RestoreOffsetOnNextLayout bool
	UpdateOnceOnSetRect       bool
	DeselectOnEscape          bool
	DisabledCells             map[string]bool

	CellCountFunc              func() int
	IDAtIndexFunc              func(i int) string
	CreateCellFunc             func(grid *GridListView, id string) zview.View
	UpdateCellFunc             func(grid *GridListView, id string)
	UpdateCellSelectionFunc    func(grid *GridListView, id string)
	CellHeightFunc             func(id string) float64 // only needed to have variable-height
	HandleSelectionChangedFunc func()
	HandleRowPressedFunc       func(id string) bool // this only gets called for non-selectable grids. Return if it eats press
	HandleRowDoubleTappedFunc  func(id string)
	HandleHoverOverFunc        func(id string) // this gets "" id when hovering out of a cell
	HandleKeyFunc              func(km zkeyboard.KeyMod, down bool) bool
	HierarchyLevelFunc         func(id string) (level int, leaf bool)

	children    map[string]zview.View
	selectedIDs map[string]bool
	DirtyIDs    map[string]bool // if not nil, only call update on these or newly created rows

	ignoreMouseEvent bool
	cellsView        *zcustom.CustomView
	margin           zgeo.Rect
	Columns          int
	Rows             int
	VisibleRows      int
	layoutDirty      bool
	cachedChildSize  zgeo.Size
	selectedIndex    int
	pressStartIndex  int
	pressEndIndex    int
}

var (
	DefaultCellColor    = zgeo.ColorNewGray(0.95, 1)
	DefaultBorderColor  = zgeo.ColorDarkGray
	DefaultPressedColor = zstyle.DefaultFGColor().WithOpacity(0.6)
	DefaultSelectColor  = zstyle.Col(zgeo.ColorNew(0.6, 0.9, 0.8, 1), zgeo.ColorNew(0.2, 0.5, 0.4, 1))
	DefaultHoverColor   = DefaultPressedColor
)

func NewView(storeName string) *GridListView {
	v := &GridListView{}
	v.Init(v, storeName)
	return v
}

func (v *GridListView) Init(view zview.View, storeName string) {
	v.ScrollView.Init(view, storeName)
	v.SetMinSize(zgeo.SizeBoth(100))
	v.MinColumns = 1
	v.FocusWidth = 3
	v.children = map[string]zview.View{}
	v.selectedIndex = -1
	v.pressStartIndex = -1
	v.selectedIDs = map[string]bool{}
	v.SetObjectName(storeName)
	v.cellsView = zcustom.NewView("GridListView.cells")
	v.AddChild(v.cellsView, nil)
	v.cellsView.SetPressUpDownMovedHandler(v.handleUpDownMovedHandler)
	v.cellsView.SetPointerEnterHandler(true, v.handleHover)
	v.CellColor = DefaultCellColor
	v.BorderColor = DefaultBorderColor
	v.SelectColor = DefaultSelectColor
	v.PressedColor = DefaultPressedColor
	v.HoverColor = DefaultHoverColor
	v.OpenBranches = map[string]bool{}
	v.BranchToggleType = zwidgets.BranchToggleTriangle
	v.Spacing = zgeo.SizeD(14, 6)
	v.MultiplyColorAlternate = 0.95
	v.DeselectOnEscape = true
	v.DisabledCells = map[string]bool{}

	v.SetCanTabFocus(true)
	v.SetKeyHandler(v.handleKeyPressed)
	v.loadOpenBranches()
	v.SetScrollHandler(func(pos zgeo.Pos, infinityDir int, delta float64) {
		v.LayoutCells(false)
	})
	v.CellColorFunc = func(id string) zgeo.Color {
		return v.CellColor
	}
	v.IDAtIndexFunc = strconv.Itoa
}

func (v *GridListView) SetDirtyRow(id string) {
	if v.DirtyIDs == nil {
		v.DirtyIDs = map[string]bool{}
	}
	v.DirtyIDs[id] = true
}

func (v *GridListView) ClearDirtyRow(id string) {
	if v.DirtyIDs == nil {
		return
	}
	delete(v.DirtyIDs, id)
}

func (v *GridListView) IndexOfID(id string) int {
	count := v.CellCountFunc()
	for i := 0; i < count; i++ {
		if v.IDAtIndexFunc(i) == id {
			return i
		}
	}
	return -1
}

func (v *GridListView) AllIDs() []string {
	count := v.CellCountFunc()
	ids := make([]string, count, count)
	for i := 0; i < count; i++ {
		ids[i] = v.IDAtIndexFunc(i)
	}
	return ids
}

func (v *GridListView) IsSelected(id string) bool {
	return v.selectedIDs[id]
}

func (v *GridListView) IsHoverCell(id string) bool {
	return id == v.CurrentHoverID
}

func (v *GridListView) makeOpenBranchesKey() string {
	return "zgridlist.Branches." + v.ObjectName()
}

func (v *GridListView) loadOpenBranches() {
	str, _ := zkeyvalue.DefaultStore.GetString(v.makeOpenBranchesKey())
	v.OpenBranches = map[string]bool{}
	if str == "" {
		return
	}
	for _, id := range strings.Split(str, ",") {
		v.OpenBranches[id] = true
	}
}

func (v *GridListView) saveOpenBranches() {
	ids := zmap.KeysAsStrings(v.OpenBranches)
	str := strings.Join(ids, ",")
	zkeyvalue.DefaultStore.SetString(str, v.makeOpenBranchesKey(), true)
}

func (v *GridListView) SelectedIDsOrHoverID() []string {
	if len(v.selectedIDs) == 0 { // if no selected, return hover-id if any
		if v.CurrentHoverID != "" {
			return []string{v.CurrentHoverID}
		}
		return nil
	}
	return v.SelectedIDs() // else return all selected, if any
}

func (v *GridListView) SelectedIDsOrHoverIDOrAll() []string {
	ids := v.SelectedIDsOrHoverID()
	if len(ids) == 0 {
		ids = v.AllIDs()
	}
	return ids
}

func (v *GridListView) SelectNextRow(up bool) (newID string) {
	index := 0
	ids := v.SelectedIDsOrHoverID()
	if len(ids) != 0 {
		index = v.IndexOfID(ids[0])
	}
	if up && index == 0 || !up && index == v.CellCountFunc()-1 {
		return ""
	}
	if up {
		index--
	} else {
		index++
	}
	id := v.IDAtIndexFunc(index)
	v.SelectCell(id, false, false)
	v.SetHoverID(id)
	return id
}

func (v *GridListView) SelectedIDInt64() int64 {
	str := v.SelectedID()
	if str == "" {
		return 0
	}
	n, _ := strconv.ParseInt(str, 10, 64)
	return n
}

// SelectedID returns "" if nothing selected, or a random id if more than one selected.
func (v *GridListView) SelectedID() string {
	if len(v.selectedIDs) == 0 {
		return ""
	}
	return zmap.GetAnyKeyAsString(v.selectedIDs)
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

func (v *GridListView) SelectCell(id string, scrollToSelected bool, animate bool) (scrolledTo bool) {
	return v.SelectCells([]string{id}, scrollToSelected, animate)
}

func (v *GridListView) SelectCells(ids []string, scrollToSelected bool, animate bool) (scrolledTo bool) {
	changedIDs := zslice.Exclusion(v.SelectedIDs(), ids)
	v.selectedIDs = map[string]bool{}
	for _, id := range ids {
		v.selectedIDs[id] = true
	}
	v.layoutDirty = true
	if scrollToSelected {
		for i := 0; i < v.CellCountFunc(); i++ {
			sid := v.IDAtIndexFunc(i)
			if v.selectedIDs[sid] {
				scrolledTo = true
				v.ScrollToCell(sid, animate)
				break
			}
		}
	}
	if v.layoutDirty { // this is to avoid layout if scroll did it
		v.LayoutCells(true)
	} else {
		v.updateCellBackgrounds(changedIDs) // this calls UpdateCellSelectionFunc also
		for _, id := range changedIDs {
			if v.children[id] == nil {
				continue
			}
			if v.UpdateCellFunc != nil {
				v.UpdateCellFunc(v, id)
			}
		}
	}
	if v.HandleSelectionChangedFunc != nil {
		v.HandleSelectionChangedFunc()
	}
	return scrolledTo
}

func (v *GridListView) UnselectAll(callHandlers bool) {
	if callHandlers {
		v.SelectCells([]string{}, false, false)
		return
	}
	v.selectedIDs = map[string]bool{}
}

func (v *GridListView) SelectAll(callHandlers bool) {
	var all []string
	v.selectedIDs = map[string]bool{}
	for i := 0; i < v.CellCountFunc(); i++ {
		id := v.IDAtIndexFunc(i)
		v.selectedIDs[id] = true
		all = append(all, id)
	}
	if callHandlers {
		v.SelectCells(all, false, false)
		return
	}
}

// func (v *GridListView) IsPressed(id string) bool {
// 	return v.pressedIDs[id]
// }

func (v *GridListView) CellView(id string) zview.View {
	return v.children[id]
}

func (v *GridListView) updateCellBackgrounds(ids []string) {
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if !visible {
			return true
		}
		if ids == nil || zstr.StringsContain(ids, cid) {
			child := v.children[cid]
			if child != nil {
				v.updateCellBackground(cid, x, y, child)
			}
		}
		return true
	})
}

func (v *GridListView) updateCellBackground(cid string, x, y int, child zview.View) {
	if child == nil {
		child = v.children[cid]
		if child == nil {
			return
		}
	}
	col := v.CellColorFunc(cid)
	index := v.IndexOfID(cid)
	min, max := zint.MinMax(v.pressStartIndex, v.pressEndIndex)
	if min != -1 && index >= min && index <= max {
		// zlog.Info("updateCellBackground4", min, max, "xy:", x, y, "inx:", index)
		col = v.PressedColor
	}
	if v.selectedIDs[cid] && v.SelectColor.Valid {
		col = v.SelectColor
	}
	if col.Valid {
		if v.MultiplyColorAlternate != 0 && x%2 != y%2 {
			col = col.MultipliedBrightness(v.MultiplyColorAlternate)
		}
		if v.CurrentHoverID == cid && v.HoverColor.Valid {
			col = col.Mixed(v.HoverColor, 0.5)
		}
		child := v.children[cid]
		if child != nil {
			child.SetBGColor(col)
		}
	}
	if v.BorderColor.Valid {
		child.Native().SetStrokeSide(1, v.BorderColor, zgeo.BottomRight, true) // we set if for non also, in case it moved
	}
	if v.UpdateCellSelectionFunc != nil {
		v.UpdateCellSelectionFunc(v, cid)
	}
}

func (v *GridListView) setPressed(index int) {
	id := v.IDAtIndexFunc(index)
	if !v.MultiSelectable {
		v.selectedIDs[id] = true
		v.SetHoverID(id)
		return
	}
	v.pressEndIndex = index
}

func (v *GridListView) isInsideInteractiveChildCell(id string, pos zgeo.Pos) (interactive bool, eventHandled bool) {
	var contained bool
	view := v.CellView(id)
	pos.Subtract(view.Rect().Pos)
	zcontainer.ViewRangeChildren(view, false, false, func(view zview.View) bool {
		r := view.Rect()
		if r.Contains(pos) {
			contained = true
			_, isLabel := view.(*zlabel.Label)
			if isLabel {
				eventHandled = true
				return false
			}
			cv, _ := view.(*zcustom.CustomView)
			if cv != nil && cv.PressedHandler() != nil {
				eventHandled = true
				return false
			}
			nv := view.Native()
			if nv.HasPressedDownHandler() {
				eventHandled = true
				return false
			}
			interactive = true
			return false
		}
		return true
	})
	if !contained {
		eventHandled = true
	}
	return
}

func (v *GridListView) handleHover(pos zgeo.Pos, inside zbool.BoolInd) {
	id, insideCell := v.FindCellForPos(pos)
	if inside.IsFalse() || !insideCell {
		id = ""
	}
	if id == v.CurrentHoverID {
		return
	}
	if v.DisabledCells[id] {
		return
	}
	if v.CellHeightFunc != nil && v.children[id] != nil {
		h := v.CellHeightFunc(id)
		if h <= 4 { // hack to avoid hovering over small separator-type cells
			return
		}
	}
	v.SetHoverID(id)
}

func (v *GridListView) SetHoverID(id string) {
	var ids []string
	// c, r := v.FindColRowForID(id)
	// zlog.Info("SetHoverID:", id, c, r)
	if v.CurrentHoverID == id {
		return
	}
	if v.CurrentHoverID != "" && v.children[v.CurrentHoverID] != nil {
		ids = []string{v.CurrentHoverID}
		if v.UpdateCellFunc != nil {
			v.UpdateCellFunc(v, v.CurrentHoverID)
		}
	}
	v.CurrentHoverID = id
	if v.CurrentHoverID != "" && v.children[v.CurrentHoverID] != nil {
		ids = append(ids, v.CurrentHoverID)
		//!!! if v.UpdateCellFunc != nil {
		// 	v.UpdateCellFunc(v, v.CurrentHoverID)
		// }
	}
	if v.HoverColor.Valid {
		v.updateCellBackgrounds(ids)
	}
	if v.HandleHoverOverFunc != nil {
		v.HandleHoverOverFunc(id)
	}
}

func (v *GridListView) handleUpDownMovedHandler(pos zgeo.Pos, down zbool.BoolInd) bool {
	// zlog.Info("handleUpDownMovedHandler", down)
	eventHandled := true
	var index int
	id, inside := v.FindCellForPos(pos)
	if v.DisabledCells[id] {
		return false
	}
	if id != "" {
		if v.HandleRowPressedFunc != nil {
			if inside && down.IsTrue() {
				// zlog.Info("HandleRowPressedFunc:", id, inside)
				handled := v.HandleRowPressedFunc(id)
				if !v.Selectable && !v.MultiSelectable {
					return handled
				}
			}
		}
		index = v.IndexOfID(id)
		if index == -1 {
			zlog.Info("No index for id:", id)
			return false
		}
	}
	if !v.Selectable && !v.MultiSelectable {
		return false
	}
	var selectionChanged bool
	switch down {
	case zbool.True:
		v.Focus(true)
		// zlog.Info("handleUpDownMovedHandler down", inside, id)
		if !inside || id == "" {
			return false
		}
		// var interactive bool
		// interactive, eventHandled = v.isInsideInteractiveChildCell(id, pos)
		// if interactive {
		// 	zlog.Info("NOT INTERACTIVE gridlistpress!!??!?!")
		// 	return false
		// }
		if zkeyboard.ModifiersAtPress&zkeyboard.ModifierMenu != 0 {
			v.ignoreMouseEvent = true
			if v.selectedIDs[id] {
				delete(v.selectedIDs, id)
			} else {
				if len(v.selectedIDs) == 0 || v.MultiSelectable {
					v.selectedIDs[id] = true
				}
			}
			selectionChanged = true
			v.SetHoverID("")
			break
		}
		if zkeyboard.ModifiersAtPress&zkeyboard.ModifierShift != 0 && len(v.selectedIDs) != 0 {
			v.shiftAppendSelection(index)
			break
		}
		clear := (len(v.selectedIDs) == 1 && v.selectedIDs[id])
		v.selectedIndex = -1
		v.selectedIDs = map[string]bool{}
		if clear {
			v.SetHoverID("")
			return false
		}
		v.pressStartIndex = index
		v.pressEndIndex = v.pressStartIndex
	case zbool.Unknown:
		// zlog.Info("updateCellBackground?", id, v.ignoreMouseEvent, v.MultiSelectable)
		if id == "" {
			return false
		}
		if !v.ignoreMouseEvent && id != "" && v.MultiSelectable {
			// zlog.Info("Move", id)
			// zlog.Info("updateCellBackground", index)
			v.pressEndIndex = index
			// v.dragged = true
		}
	case zbool.False:
		if v.ignoreMouseEvent {
			v.ignoreMouseEvent = false
			break
		}
		selectionChanged = true
		v.selectedIndex = index
		min, max := zint.MinMax(v.pressStartIndex, v.pressEndIndex)
		if min != -1 {
			for i := min; i <= max; i++ {
				v.selectedIDs[v.IDAtIndexFunc(i)] = true
			}
		}
		v.pressStartIndex = -1
		v.pressEndIndex = -1
		// id := v.IDAtIndexFunc(index)
		// if id != "" {
		// v.SetHoverID(id)
		// }
		v.SetHoverID("")
	}
	v.updateCellBackgrounds(nil)
	v.ExposeIn(0.0001)
	if down.IsFalse() && selectionChanged && v.HandleSelectionChangedFunc != nil {
		v.HandleSelectionChangedFunc()
	}
	return eventHandled
}

func (v *GridListView) shiftAppendSelection(clickIndex int) {
	min := -1
	max := -1
	for id, _ := range v.selectedIDs {
		index := v.IndexOfID(id)
		if min == -1 || min > index {
			min = index
		}
		if max == -1 || max < index {
			max = index
		}
	}
	s, e := zint.MinMax(clickIndex, min)
	if clickIndex > min {
		s, e = zint.MinMax(max, clickIndex)
	}
	for i := s; i <= e; i++ {
		v.selectedIDs[v.IDAtIndexFunc(i)] = true
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

func (v *GridListView) FindColRowForID(id string) (col, row int) {
	col = -1
	row = -1
	v.ForEachCell(func(cellID string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if cellID == id {
			row = y
			col = x
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
			id = cellID
			inside = true
			return false
		}
		return true
	})
	if id == "" && v.CellCountFunc() > 0 {
		if before {
			return v.IDAtIndexFunc(0), false
		}
		return v.IDAtIndexFunc(v.CellCountFunc() - 1), false
	}
	return
}

func (v *GridListView) CalculateColumnsAndRows(childWidth, totalWidth float64) (nx, ny int) {
	nx = zgeo.MaxCellsInLength(v.Spacing.W, childWidth, totalWidth-v.margin.Size.W)
	// zlog.Info("CalculateColumnsAndRows:", v.ObjectName(), nx, v.MaxColumns)
	if v.MaxColumns != 0 {
		zint.Minimize(&nx, v.MaxColumns)
	}
	zint.Maximize(&nx, v.MinColumns)
	ny = (v.CellCountFunc() + nx - 1) / nx
	// zlog.Info("CalculateColumnsAndRows:", childWidth, totalWidth, nx, ny)
	return
}

func (v *GridListView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	focusMarg := zgeo.SizeD(6, 6)
	v.cachedChildSize = zgeo.SizeNull
	s = v.MinSize()
	// zlog.Info("GLV CalculatedSize:", v.Hierarchy(), total, s, v.MakeFullSize)
	if v.CellCountFunc() == 0 {
		return s.Plus(focusMarg), s.Plus(focusMarg)
	}
	childSize := v.getAChildSize(total)
	maxCols := math.Max(1, float64(v.MinColumns))
	w := childSize.W*maxCols + v.Spacing.W*(maxCols-1) - v.margin.Size.W
	zfloat.Maximize(&s.W, w)
	max = v.CalculatedGridSize(total)
	if v.MakeFullSize {
		s = max
		// zlog.Info("CalcGridSizeAllRows", total, v.ObjectName(), s)
		// if v.CellHeightFunc != nil {
		// 	for y := 0; y < int(max) && y < v.CellCountFunc(); y++ {
		// 		s.H += CellHeightFunc()
		// 	}
		// } else {
		// 	my := (float64(v.CellCountFunc()) + max - 1) / max
		// 	s.H = childSize.H*my + v.Spacing.H*(my-1) - v.margin.Size.H
		// }
	}
	max.W = 0
	max.H += focusMarg.H
	s = s.Plus(focusMarg)
	// zlog.Info("GLV CalculatedSize2:", total, childSize.H, v.ObjectName(), s, v.MinSize(), v.CellCountFunc(), max)
	return s, max
}

func (v *GridListView) CalculatedGridSize(total zgeo.Size) zgeo.Size {
	if v.CellCountFunc() == 0 {
		return v.MinSize()
	}
	totalMinusBar := total.Minus(zgeo.SizeD(v.BarSize(), 0))
	childSize := v.getAChildSize(totalMinusBar)
	nx, ny := v.CalculateColumnsAndRows(childSize.W, totalMinusBar.W)
	// zlog.Info("CalculatedSize cs:", total, v.BarSize(), totalMinusBar, childSize, nx, ny)
	// zlog.Info("CalculatedGridSize childsize:", childSize, totalMinusBar.W, nx, ny)
	zint.Maximize(&ny, v.MinRowsForFullSize)
	if v.MaxRowsForFullSize != 0 {
		zint.Minimize(&ny, v.MaxRowsForFullSize)
	}
	s := v.margin.Size.Negative()
	x := float64(nx)
	y := float64(ny)
	// zlog.Info("GLV CalculatedGridSize:", total, childSize, nx, ny, s, v.CellHeightFunc != nil)
	if v.CellHeightFunc != nil {
		for i := 0; i < ny; i++ {
			s.H += v.CellHeightFunc(v.IDAtIndexFunc(i))
		}
		s.H += v.Spacing.H * y //(y - 1)
	} else {
		s.H += childSize.H*y + v.Spacing.H*y //(y-1)
	}
	s.W += childSize.W*x + v.Spacing.W*(x-1)
	s.H++
	return s
}

func (v *GridListView) RemoveCell(id string) bool {
	child := v.children[id]
	if child != nil {
		v.cellsView.RemoveChild(child, true)
		delete(v.children, id)
		return true
	}
	return false
}

func (v *GridListView) getAChildSize(total zgeo.Size) zgeo.Size {
	if !v.cachedChildSize.IsNull() {
		return v.cachedChildSize
	}
	cid := v.IDAtIndexFunc(0)
	// zlog.Info("getAChildSize:CreateCellFunc", v.CreateCellFunc != nil)
	child := v.CreateCellFunc(v, cid)
	s, _ := child.CalculatedSize(total)
	if v.CellHeightFunc != nil {
		zfloat.Maximize(&s.H, v.CellHeightFunc(cid))
	}
	child.Native().PerformAddRemoveFuncs(false)
	zfloat.Maximize(&s.W, v.MinSize().W)
	v.cachedChildSize = s
	return s
}

func (v *GridListView) GetChildren(collapsed bool) (children []zview.View) {
	for _, c := range v.children {
		children = append(children, c)
	}
	return
}

func (v *GridListView) insertBranchToggle(id string, child zview.View) {
	level, leaf := v.HierarchyLevelFunc(id)
	co, _ := child.(zcontainer.CellsOwner)
	aa, _ := child.(zcontainer.AdvancedAdder)
	if level > 0 {
		w := float64(level-2) * 14
		if v.BranchToggleType != zwidgets.BranchToggleNone {
			if !leaf {
				bt := zwidgets.BranchToggleViewNew(v.BranchToggleType, id, v.OpenBranches[id])
				aa.AddAdvanced(bt, zgeo.CenterLeft, zgeo.RectNull, zgeo.SizeNull, 0, false)
			}
			w += 24
		}
		cells := co.GetCells()
		(*cells)[0].Margin.Pos.X += w
		// zlog.Info("insertBranch:", (*cells)[0].View.ObjectName(), (*cells)[0].Margin.W)
		// (*cells)[0].MinSize.W -= w
		// (*cells)[0].MaxSize.W -= w
	}
}

func (v *GridListView) addDebugIndexLabel(child zview.View, id string) {
	a, _ := child.(zcontainer.AdvancedAdder)
	if a == nil {
		return
	}
	i := v.IndexOfID(id)
	label := zlabel.New(strconv.Itoa(i))
	label.SetObjectName("debug-index")
	label.SetZIndex(22111)
	label.SetBGColor(zgeo.ColorOrange)

	a.AddAdvanced(label, zgeo.TopRight, zgeo.RectNull, zgeo.SizeNull, -1, true)
}

func (v *GridListView) makeOrGetChild(id string) (zview.View, bool) {
	child := v.children[id]
	// zlog.Info("makeOrGetChild:", id, child != nil)
	if child != nil {
		return child, true
	}
	if v.DirtyIDs != nil { // set this before v.CreateCellFunc below, so it can clear it.
		v.DirtyIDs[id] = true
	}
	child = v.CreateCellFunc(v, id)
	// v.addDebugIndexLabel(child, id)
	if v.HierarchyLevelFunc != nil {
		v.insertBranchToggle(id, child)
	}
	if v.HandleRowDoubleTappedFunc != nil {
		child.Native().SetDoublePressedHandler(func() {
			v.HandleRowDoubleTappedFunc(id)
		})
	}
	v.children[id] = child
	// child.Show(false)
	v.cellsView.AddChild(child, nil)
	child.Native().SetJSStyle("userSelect", "none")
	zpresent.CallReady(child, true)
	zpresent.CallReady(child, false)
	// e, _ := child.(zview.ExposableType)
	// if e != nil {
	// 	e.Expose()
	// }
	return child, false
}

func (v *GridListView) innerRect() zgeo.Rect {
	return v.LocalRect().ExpandedD(-v.FocusWidth)
}

func (v *GridListView) xForStartOfNthCell(n int) float64 {
	r := v.LocalRect()
	r.Size.W -= v.BarSize()
	// r.Add(v.margin)
	//	r.Size.W -= v.Spacing.W * 2
	x := r.Pos.X + float64(n)*r.Size.W/float64(v.Columns)
	// zlog.Info("xForStartOfNthCell", r.Size.W, n, r.Pos.X, r, x, v.Spacing.W)
	return x
}

func (v *GridListView) ForEachCell(got func(cellID string, outer, inner zgeo.Rect, x, y int, visible bool) bool) {
	// zlog.Info("ForEachCell", v.margin)
	if v.CellCountFunc() == 0 {
		return
	}
	lrect := v.LocalRect()
	lrect.Size.W -= v.BarSize()
	// zlog.Info("ForEachCell", rect)
	pos := lrect.Pos.Floor()
	width := lrect.Size.W + v.margin.Size.W - v.Spacing.W // -spacing.W is for space/2 on either side
	localSize := v.LocalRect().Size
	childSize := v.getAChildSize(localSize.Ceil())
	zlog.Assert(v.MaxColumns <= 1)
	v.Columns, v.Rows = v.CalculateColumnsAndRows(childSize.W, lrect.Size.W)
	width -= (float64(v.Columns) - 1) * v.Spacing.W
	var x, y int
	v.VisibleRows = 0
	for i := 0; i < v.CellCountFunc(); i++ {
		cellID := v.IDAtIndexFunc(i)
		if v.CellHeightFunc != nil {
			childSize.H = v.CellHeightFunc(cellID)
		}
		var s zgeo.Size
		x1 := v.xForStartOfNthCell(x)
		x2 := v.xForStartOfNthCell(x + 1)
		s.W = math.Ceil(x2 - x1)
		s.H = childSize.H + v.Spacing.H
		pos.X = x1
		r := zgeo.Rect{Pos: pos, Size: s}
		lastx := (x == v.Columns-1)
		lasty := (y == v.Rows-1)
		var minx, maxx, miny, maxy float64
		minx = v.Spacing.W / 2
		if x == 0 {
			minx += v.margin.Min().X
		} else {
			r.SetMinX(r.Pos.X)
		}
		maxx = -v.Spacing.W / 2
		if lastx {
			maxx += v.margin.Max().X
		}
		miny = v.Spacing.H / 2
		if y == 0 {
			miny += v.margin.Min().Y
		}
		maxy = -v.Spacing.H / 2
		if lasty {
			maxy += v.margin.Max().Y + 1
		}
		marg := zgeo.RectFromXY2(minx, miny, maxx, maxy).ExpandedToInt()
		// r.Size.Add(marg.Size.Negative()) // expand r's size to include margins
		r = r.ExpandedToInt()
		cr := r.Plus(marg)
		visible := (r.Max().Y >= v.YOffset && r.Min().Y <= v.YOffset+v.innerRect().Size.H)
		r2 := r
		if v.BorderColor.Valid {
			r2.Size.W--
		}
		if !got(cellID, r2, cr, x, y, visible) {
			break
		}
		if v.HorizontalFirst {
			if lastx {
				x = 0
				y++
				pos.Y += r.Size.H
				if visible {
					v.VisibleRows++
				}
			} else {
				//				pos.X += r.Size.W - 12
				x++
			}
			continue
		} else {
			pos.Y++
		}
		if lasty {
			pos.Y = lrect.Pos.Y
			y = 0
			x++
			//			pos.X += r.Size.W - 1
		} else {
			if visible {
				v.VisibleRows++
			}
			pos.Y += r.Size.H - 1
			y++
		}
	}
}

func (v *GridListView) SetRect(rect zgeo.Rect) {
	if rect.Size.W == 0 {
		// zlog.Info("GLV:SetRect rect.W==0:", v.Hierarchy(), rect)
		return
	}
	r := rect.ExpandedD(-v.FocusWidth)
	s := v.CalculatedGridSize(r.Size)
	w := r.Size.W - v.BarSize()
	cs := zgeo.SizeD(w, s.H)
	v.ScrollView.SetRectWithChildSize(r, cs)
	v.LayoutCells(v.UpdateOnceOnSetRect)
	v.UpdateOnceOnSetRect = false
}

func (v *GridListView) ArrangeChildren() {
	v.SetRect(v.Rect())
}

func (v *GridListView) ReplaceChild(child, with zview.View) {
	for id, c := range v.children {
		if c == child {
			v.children[id] = with
			v.Native().ReplaceChild(child, with)
			break
		}
	}
}

func (v *GridListView) LayoutCells(updateCells bool) {
	var oy float64
	var topID string
	if v.RestoreOffsetOnNextLayout { // || v.RestoreTopSelectedRowOnNextLayout {
		var top float64
		for id, child := range v.children {
			y := child.Rect().Pos.Y
			if topID == "" || y < top && y >= 0 {
				top = y
				topID = id
				// zlog.Info("GridListView.LayoutCells with RestoreOffsetOnNextLayout set top", topID, y)
			}
		}
		if topID != "" {
			v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
				if cid == topID { // || v.RestoreTopSelectedRowOnNextLayout && v.selectedIDs[cid] {
					oy = outer.Pos.Y
					v.ShowScrollBars(false, false)
					// zlog.Info("GridListView.LayoutCells: set offset to old id:", y, outer.Pos.Y, len(v.children), v.CellCountFunc())
					return false
				}
				return true
			})
		}
	}
	if v.DirtyIDs == nil {
		v.DirtyIDs = map[string]bool{}
	}
	var selected []string
	oldSelected := v.SelectedIDs()
	var hoverOK bool
	v.layoutDirty = false
	placed := map[string]bool{}
	if v.RecreateCells {
		for cid, view := range v.children {
			v.cellsView.RemoveChild(view, true)
			delete(v.children, cid)
		}
		v.children = map[string]zview.View{}
		v.RecreateCells = false
	}
	v.updateBorder()
	var updateCount int
	// zlog.Info("LayoutCells", v.ObjectName(), updateCells, v.CellCountFunc(), v.Rect())
	// start := time.Now()
	var count int
	// zlog.Info("GridListView.LayoutCells", v.ObjectName(), v.CellCountFunc(), v.RestoreOffsetOnNextLayout, v.RestoreTopSelectedRowOnNextLayout, v.cellsView.Rect())
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if zstr.StringsContain(oldSelected, cid) {
			selected = append(selected, cid)
		}
		if visible {
			count++
			// prof := zlog.NewProfile(0.001, "GV.Layout", v.ObjectName(), cid)
			if v.CurrentHoverID == cid {
				hoverOK = true
			}
			child, _ := v.makeOrGetChild(cid)
			//TODO: exit when !visible after visible
			// prof.Log("After Make")
			ms, _ := child.(zview.Marginalizer)
			if ms != nil {
				marg := inner.Minus(outer)
				// zlog.Info("Grid SetMarg on child", cid, marg, v.margin)
				ms.SetMargin(marg)
			}
			dirty := (v.DirtyIDs != nil && v.DirtyIDs[cid])
			if dirty || !child.Native().HasSize() || child.Rect() != outer {
				// if count == 1 {
				// zlog.Info("GridListView.LayoutCells first visible", v.ObjectName(), outer, cid)
				// }
				child.SetRect(outer)
			}
			// prof.Log("After Set Rect")
			v.updateCellBackground(cid, x, y, child)
			if v.UpdateCellFunc != nil && (updateCells || dirty) {
				updateCount++
				// zlog.Info("GridLayout cell", x, y)
				v.UpdateCellFunc(v, cid)
			}
			placed[cid] = true
			// prof.End("End")
		}
		return true
	})
	// zlog.Info("Layed:", len(v.DirtyIDs), updateCount, v.Hierarchy(), time.Since(start))
	v.DirtyIDs = nil
	for cid, view := range v.children {
		if !placed[cid] {
			// delete(v.selectedIDs, cid)
			v.cellsView.RemoveChild(view, true)
			delete(v.children, cid)
		}
	}
	if !hoverOK {
		v.SetHoverID("")
	}
	// if !zstr.SlicesAreEqual(oldSelected, selected) {
	// 	v.SelectCells(selected, false, false)
	// }
	v.RestoreOffsetOnNextLayout = false
	if topID != "" {
		// zlog.Info("GridListView.LayoutCells: set offset to old set:", oy)
		// v.SetScrollHandler(old)
		v.SetContentOffset(oy, false)
		v.ShowScrollBars(false, true)
	}
	// zlog.Info("GLV.Layout:", v.ObjectName(), time.Since(start), count)
}

func (v *GridListView) ScrollToCell(cellID string, animate bool) {
	outerRect, _ := v.CellRects(cellID)
	if !outerRect.IsNull() {
		v.MakeRectVisible(outerRect.Plus(zgeo.RectFromXY2(0, 0, 0, 2)), animate)
	}
}

func (v *GridListView) moveHover(incX, incY int, mod zkeyboard.Modifier) bool {
	var sx, sy int
	var fid string

	append := (mod == zkeyboard.ModifierShift)
	if v.CurrentHoverID != "" {
		fid = v.CurrentHoverID
		sx, sy = v.FindColRowForID(fid)
	}
	if fid == "" && v.selectedIndex != -1 {
		fid = v.IDAtIndexFunc(v.selectedIndex)
		sx, sy = v.FindColRowForID(fid)
	}
	if fid == "" {
		v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool { // find a selected cell and x,y
			if visible {
				if append && (len(v.selectedIDs) == 0 || v.selectedIDs[cid]) || !append && fid == v.CurrentHoverID {
					fid = cid
					sx = x
					sy = y
					return false
				}
			}
			return true
		})
	}
	if fid != "" {
		if append && len(v.selectedIDs) == 0 {
			v.selectedIDs[fid] = true
			v.CurrentHoverID = ""
			v.updateCellBackground(fid, sx, sy, nil)
			return true
		} else {
			nx := sx + incX
			ny := sy + incY
			if append && nx < 0 && ny > 0 {
				nx = v.Columns - 1
				ny--
			} else if append && nx >= v.Columns && ny < v.Rows-1 {
				nx = 0
				ny++
			} else if nx < 0 || nx >= v.Columns || ny < 0 || ny >= v.Rows {
				return true
			}
			if ny*v.Columns+nx >= v.CellCountFunc() {
				if incY != 1 {
					return true
				}
				nx = 0
			}
			nid := v.FindCellForColRow(nx, ny)
			if append {
				startIndex := v.IndexOfID(fid)
				endIndex := v.IndexOfID(nid)
				min, max := zint.MinMax(startIndex, endIndex)
				for i := min; i <= max; i++ {
					v.selectedIDs[v.IDAtIndexFunc(i)] = true
				}
				if incX == -1 || incY == -1 {
					v.selectedIndex = min
				} else {
					v.selectedIndex = max
				}
				v.CurrentHoverID = ""
				v.updateCellBackgrounds(nil)
			} else {
				v.SetHoverID(nid)
			}
			v.ScrollToCell(nid, false)
		}
	}
	return true
}

func (v *GridListView) toggleBranch(open bool) {
	if len(v.selectedIDs) == 0 {
		return
	}
	id := v.SelectedIDs()[0]
	cellView := v.CellView(id)
	zcontainer.ViewRangeChildren(cellView, false, false, func(view zview.View) bool {
		bt, _ := view.(*zwidgets.BranchToggleView)
		if bt != nil {
			tellParents := true
			if open != bt.IsOpen() {
				bt.SetOpen(open, tellParents)
			}
			return false
		}
		return true
	})
}

func (v *GridListView) handleKeyPressed(km zkeyboard.KeyMod, down bool) bool {
	if !down {
		return false
	}
	if km.Key.IsReturnish() && (v.Selectable || v.MultiSelectable) {
		id := v.CurrentHoverID
		if id == "" && v.selectedIndex != -1 && v.selectedIndex < v.CellCountFunc() {
			id = v.IDAtIndexFunc(v.selectedIndex)
		}
		if id != "" {
			v.selectedIndex = v.IndexOfID(id)
			if km.Modifier&zkeyboard.ModifierMenu != 0 {
				clear := v.selectedIDs[id]
				if clear {
					delete(v.selectedIDs, id)
				} else {
					v.selectedIDs[id] = true
				}
			} else if km.Modifier == 0 {
				clear := (len(v.selectedIDs) == 1 && v.selectedIDs[id])
				v.selectedIDs = map[string]bool{}
				if !clear {
					v.selectedIDs[id] = true
				}
			}
			v.CurrentHoverID = ""
			v.updateCellBackgrounds(nil)
		}
	}
	if km.Modifier == zkeyboard.ModifierNone || km.Modifier == zkeyboard.ModifierShift {
		switch km.Key {
		case zkeyboard.KeyUpArrow:
			return v.moveHover(0, -1, km.Modifier)
		case zkeyboard.KeyDownArrow:
			return v.moveHover(0, 1, km.Modifier)
		case zkeyboard.KeyLeftArrow:
			if v.HierarchyLevelFunc != nil && v.MaxColumns == 1 {
				v.toggleBranch(false)
				return true
			}
			if v.MaxColumns != 1 {
				return v.moveHover(-1, 0, km.Modifier)
			}
		case zkeyboard.KeyRightArrow:
			if v.HierarchyLevelFunc != nil && v.MaxColumns == 1 {
				v.toggleBranch(true)
				return true
			}
			if v.MaxColumns != 1 {
				return v.moveHover(1, 0, km.Modifier)
			}
		case zkeyboard.KeyEscape:
			if v.DeselectOnEscape && (v.Selectable || v.MultiSelectable) {
				v.UnselectAll(true)
			}
		case 'A':
			if v.MultiSelectable {
				v.SelectAll(true)
			}
		}
	}
	if v.HandleKeyFunc != nil {
		return v.HandleKeyFunc(km, down)
	}
	return false
}

func (v *GridListView) updateBorder() {
	if v.BorderColor.Valid {
		w := 0.0
		if v.CellCountFunc() > 0 {
			w = 1
		}
		// v.cellsView.SetStrokeSide(w, v.BorderColor, zgeo.TopLeft, true) // we set if for non also, in case it moved
		v.cellsView.SetStrokeSide(w, v.BorderColor, zgeo.TopLeft|zgeo.BottomRight, true) // we set if for non also, in case it moved
	}
}
func (v *GridListView) ReadyToShow(beforeWindow bool) {
	v.ScrollView.ReadyToShow(beforeWindow)
	v.updateBorder()
	if beforeWindow && v.BorderColor.Valid {
		v.SetStroke(1, v.BorderColor, true)
	}
}

func (v *GridListView) SetMargin(m zgeo.Rect) {
	v.margin = m
	// zlog.Info("GL SetMarg:", v.Hierarchy(), m)
}

func (v *GridListView) Margin() zgeo.Rect {
	return v.margin
}

func (v *GridListView) HandleBranchToggleChanged(id string, open bool) {
	if open {
		v.OpenBranches[id] = true
	} else {
		delete(v.OpenBranches, id)
	}
	v.LayoutCells(false)
	v.saveOpenBranches()
}

func (v *GridListView) AnyChildView() zview.View {
	if len(v.children) == 0 {
		return nil
	}
	var view zview.View
	zmap.GetAnyValue(&view, v.children)
	return view
}
