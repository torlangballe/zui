// The zgridlist package defines GridListView, a view for displaying cells in multiple rows and columns.
// It knows nothing about the content of it's cells.
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
	"github.com/torlangballe/zutil/zstr"
)

type GridListView struct {
	zscrollview.ScrollView
	Spacing                zgeo.Size
	Selectable             bool
	MultiSelectable        bool
	MakeFullSize           bool
	MaxColumns             int
	MinColumns             int
	MinRowsForFullSize     int
	MaxRowsForFullSize     int
	BorderColor            zgeo.Color
	CellColor              zgeo.Color
	CellColorFunc          func(id string) zgeo.Color // Always set
	MultiplyColorAlternate float32
	PressedColor           zgeo.Color
	SelectColor            zgeo.Color
	HoverColor             zgeo.Color
	BranchToggleType       zwidgets.BranchToggleType
	OpenBranches           map[string]bool
	CurrentHoverID         string
	FocusWidth             float64
	HorizontalFirst        bool // HorizontalFirst means 1 2 3 on first row, not down first

	CellCountFunc              func() int
	IDAtIndexFunc              func(i int) string
	CreateCellFunc             func(grid *GridListView, id string) zview.View
	UpdateCellFunc             func(grid *GridListView, id string)
	UpdateSelectionFunc        func(grid *GridListView, id string)
	CellHeightFunc             func(id string) float64 // only need to have variable-height
	HandleSelectionChangedFunc func()
	HandleRowPressed           func(id string) // this only gets called for non-selectable grids
	HandleHoverOverFunc        func(id string) // this gets "" id when hovering out of a cell
	HandleKeyFunc              func(km zkeyboard.KeyMod, down bool) bool
	HierarchyLevelFunc         func(id string) (level int, leaf bool)

	children         map[string]zview.View
	selectedIDs      map[string]bool
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
	DefaultSelectColor  = zstyle.DefaultFGColor().WithOpacity(0.3)
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
	v.AddChild(v.cellsView, -1)
	v.cellsView.SetPressUpDownMovedHandler(v.handleUpDownMovedHandler)
	v.cellsView.SetPointerEnterHandler(true, v.handleHover)
	v.CellColor = DefaultCellColor
	v.BorderColor = DefaultBorderColor
	v.SelectColor = DefaultSelectColor
	v.PressedColor = DefaultPressedColor
	v.HoverColor = DefaultHoverColor
	v.OpenBranches = map[string]bool{}
	v.BranchToggleType = zwidgets.BranchToggleTriangle
	v.Spacing = zgeo.Size{14, 6}
	v.MultiplyColorAlternate = 0.95
	v.SetCanTabFocus(true)
	v.SetKeyHandler(v.handleKeyPressed)
	v.SetStroke(1, zgeo.ColorBlack, false)
	v.loadOpenBranches()
	v.SetScrollHandler(func(pos zgeo.Pos, infinityDir int) {
		// zlog.Info("Scroll:", pos)
		v.LayoutCells(false)
	})
	v.CellColorFunc = func(id string) zgeo.Color {
		return v.CellColor
	}
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
	ids := zmap.GetKeysAsStrings(v.OpenBranches)
	str := strings.Join(ids, ",")
	zkeyvalue.DefaultStore.SetString(str, v.makeOpenBranchesKey(), true)
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
	if animateScroll {
		for i := 0; i < v.CellCountFunc(); i++ {
			sid := v.IDAtIndexFunc(i)
			if v.selectedIDs[sid] {
				v.ScrollToCell(sid, animateScroll)
				break
			}
		}
	}
	if v.layoutDirty { // this is to avoid layout if scroll did it
		v.LayoutCells(true)
	}
	if v.HandleSelectionChangedFunc != nil {
		v.HandleSelectionChangedFunc()
	}
}

func (v *GridListView) UnselectAll(callHandlers bool) {
	if callHandlers {
		v.SelectCells([]string{}, false)
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
		v.SelectCells(all, false)
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
	index := v.Columns*y + x
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
	if v.UpdateSelectionFunc != nil {
		v.UpdateSelectionFunc(v, cid)
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
	// zlog.Info("handleUpDownMovedHandler")
	eventHandled := true
	var index int
	id, inside := v.FindCellForPos(pos)
	if id != "" {
		if !v.Selectable && !v.MultiSelectable && v.HandleRowPressed != nil {
			v.HandleRowPressed(id)
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
		if zkeyboard.ModifiersAtPress&zkeyboard.ModifierCommand != 0 {
			v.ignoreMouseEvent = true
			if v.selectedIDs[id] {
				delete(v.selectedIDs, id)
			} else {
				v.selectedIDs[id] = true
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
	// zlog.Info("GL pressed with selection:", down, selectionChanged, v.selectedIDs)
	if selectionChanged && v.HandleSelectionChangedFunc != nil {
		// zlog.Info("GL pressed with selection:", down)
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
	nx = zgeo.MaxCellsInSize(v.Spacing.W, v.margin.Size.W, childWidth, totalWidth)
	if v.MaxColumns != 0 {
		zint.Minimize(&nx, v.MaxColumns)
	}
	zint.Maximize(&nx, v.MinColumns)
	ny = (v.CellCountFunc() + nx - 1) / nx
	return
}

func (v *GridListView) CalculatedSize(total zgeo.Size) zgeo.Size {
	marg := zgeo.Size{6, 6}
	// v.cachedChildSize = zgeo.Size{}
	s := v.MinSize()
	if v.CellCountFunc() == 0 {
		return s.Plus(marg)
	}
	if v.MakeFullSize {
		return v.CalculatedGridSize(total).Plus(marg)
	}
	childSize := v.getAChildSize(total)
	mx := math.Max(1, float64(v.MinColumns))
	w := childSize.W*mx + v.Spacing.W*(mx-1) - v.margin.Size.W
	zfloat.Maximize(&s.W, w)
	// zlog.Info("GLV CalculatedSize:", v.Hierarchy(), s, v.MinSize())
	return s.Plus(marg)
}

func (v *GridListView) CalculatedGridSize(total zgeo.Size) zgeo.Size {
	if v.CellCountFunc() == 0 {
		return v.MinSize()
	}
	childSize := v.getAChildSize(total)
	nx, ny := v.CalculateColumnsAndRows(childSize.W, total.W)
	zint.Maximize(&ny, v.MinRowsForFullSize)
	if v.MaxRowsForFullSize != 0 {
		zint.Minimize(&ny, v.MaxRowsForFullSize)
	}
	s := v.margin.Size.Negative()
	x := float64(nx)
	y := float64(ny)
	if v.CellHeightFunc != nil {
		for i := 0; i < ny; i++ {
			s.H += v.CellHeightFunc(v.IDAtIndexFunc(i))
		}
		s.H += v.Spacing.H * y //(y - 1)
	} else {
		s.H += childSize.H*y + v.Spacing.H*y //(y-1)
	}
	// zlog.Info("GLV CalculatedGridSize2:", s, v.Spacing.H)
	s.W += childSize.W*x + v.Spacing.W*(x-1)
	s.W += v.BarSize // scroll bar
	// zlog.Info("GL:CalculatedGridSize:", s)
	return s
}

func (v *GridListView) RemoveCell(id string) bool {
	child := v.children[id]
	if child != nil {
		v.cellsView.RemoveChild(child)
		delete(v.children, id)
		return true
	}
	return false
}

func (v *GridListView) getAChildSize2(total zgeo.Size) zgeo.Size {
	cid := v.IDAtIndexFunc(0)
	child, exists := v.makeOrGetChild(cid)
	if exists {
		return child.Rect().Size
	}
	s := child.CalculatedSize(total)
	if v.CellHeightFunc != nil {
		zfloat.Maximize(&s.H, v.CellHeightFunc(cid))
	}
	zfloat.Maximize(&s.W, v.MinSize().W)
	child.SetRect(zgeo.Rect{Size: s})
	return s
}

func (v *GridListView) getAChildSize(total zgeo.Size) zgeo.Size {
	if !v.cachedChildSize.IsNull() {
		return v.cachedChildSize
	}
	cid := v.IDAtIndexFunc(0)
	child := v.CreateCellFunc(v, cid)
	s := child.CalculatedSize(total)
	if v.CellHeightFunc != nil {
		zfloat.Maximize(&s.H, v.CellHeightFunc(cid))
	}
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

func (v *GridListView) SetRect(rect zgeo.Rect) {
	v.ScrollView.SetRect(rect.ExpandedD(-v.FocusWidth))
	// zlog.Info("GL: SetRect:", rect, v.Rect(), v.cellsView.Rect())
	at := v.View.(zcontainer.Arranger) // in case we are a stack or something inheriting from GridListView
	// zlog.Info("GL: SetRect:", v.View.Native().Hierarchy(), rect)
	at.ArrangeChildren()
	// zlog.Info("GL: SetRect done:", v.View.Native().Hierarchy(), rect)
}

func (v *GridListView) insertBranchToggle(id string, child zview.View) {
	level, leaf := v.HierarchyLevelFunc(id)
	co, _ := child.(zcontainer.CellsOwner)
	aa, _ := child.(zcontainer.AdvancedAdder)
	if level > 0 {
		w := float64(level-1) * 14
		if v.BranchToggleType != zwidgets.BranchToggleNone {
			if !leaf {
				bt := zwidgets.BranchToggleViewNew(v.BranchToggleType, id, v.OpenBranches[id])
				aa.AddAdvanced(bt, zgeo.CenterLeft, zgeo.Size{4 + w, 0}, zgeo.Size{}, -1, true)
			}
			w += 24
		}
		cells := co.GetCells()
		(*cells)[0].Margin.W += w
		(*cells)[0].MinSize.W -= w
		(*cells)[0].MaxSize.W -= w
	}
}

func (v *GridListView) makeOrGetChild(id string) (zview.View, bool) {
	child := v.children[id]
	if child != nil {
		return child, true
	}
	child = v.CreateCellFunc(v, id)
	if v.HierarchyLevelFunc != nil {
		v.insertBranchToggle(id, child)
	}
	v.children[id] = child
	v.cellsView.AddChild(child, -1)
	child.Native().SetJSStyle("userSelect", "none")
	zpresent.CallReady(child, true)
	if v.UpdateCellFunc != nil {
		v.UpdateCellFunc(v, id)
	}
	zpresent.CallReady(child, false)
	child.(zview.ExposableType).Expose()
	return child, false
}

func (v *GridListView) innerRect() zgeo.Rect {
	return v.LocalRect().ExpandedD(-v.FocusWidth)
}

func (v *GridListView) ForEachCell(got func(cellID string, outer, inner zgeo.Rect, x, y int, visible bool) bool) {
	// zlog.Info("ForEachCell", v.margin)
	if v.CellCountFunc() == 0 {
		return
	}
	rect := v.LocalRect().ExpandedToInt()
	rect.Size.W -= v.BarSize
	// zlog.Info("ForEachCell", rect)
	pos := rect.Pos.Floor()
	childSize := rect.Size.Plus(v.margin.Size)
	width := childSize.W

	if v.CellHeightFunc == nil {
		childSize = v.getAChildSize(rect.Size).Ceil()
	} else {
		zlog.Assert(v.MaxColumns <= 1)
	}
	// zlog.Info("ForEachCell", rect, childSize)
	v.Columns, v.Rows = v.CalculateColumnsAndRows(childSize.W, rect.Size.W)
	var x, y int
	v.VisibleRows = 0
	for i := 0; i < v.CellCountFunc(); i++ {
		cellID := v.IDAtIndexFunc(i)
		if v.CellHeightFunc != nil {
			childSize.H = v.CellHeightFunc(cellID)
		}

		r := zgeo.Rect{Pos: pos, Size: childSize}
		endx := math.Ceil(float64(x+1) * width / float64(v.Columns))
		if r.Max().X < endx {
			r.SetMaxX(endx)
		}
		lastx := (x == v.Columns-1)
		lasty := (y == v.Rows-1)
		var minx, maxx, miny, maxy float64
		minx = v.Spacing.W / 2
		if x == 0 {
			minx += v.margin.Min().X
		}
		maxx = -v.Spacing.W / 2
		if lastx {
			maxx = -(rect.Max().X - r.Max().X)
		}
		miny = v.Spacing.H / 2
		if y == 0 {
			miny += v.margin.Min().Y
		}
		maxy = -v.Spacing.H / 2
		if lasty {
			maxy += v.margin.Max().Y
		}
		marg := zgeo.RectFromXY2(minx, miny, maxx, maxy).ExpandedToInt()
		r.Size.Add(marg.Size.Negative())
		cr := r.Plus(marg)
		visible := (r.Max().Y >= v.YOffset && r.Min().Y <= v.YOffset+v.innerRect().Size.H)
		r2 := r
		if v.BorderColor.Valid {
			r2.Size.W -= 1
		}
		if !got(cellID, r2, cr, x, y, visible) {
			break
		}
		if v.HorizontalFirst {
			if lastx {
				pos.X = rect.Pos.X
				x = 0
				y++
				pos.Y += r.Size.H
				if visible {
					v.VisibleRows++
				}
			} else {
				pos.X += r.Size.W - 1
				x++
			}
			continue
		}
		if lasty {
			pos.Y = rect.Pos.Y
			y = 0
			x++
			pos.X += r.Size.W
		} else {
			if visible {
				v.VisibleRows++
			}
			pos.Y += r.Size.H - 1
			y++
		}
	}
}

func (v *GridListView) ArrangeChildren() {
	w := zscrollview.DefaultBarSize
	if v.CellCountFunc() < 20 {
		w = 0
	}
	v.ScrollView.BarSize = w
	v.ScrollView.ArrangeChildren()
	locSize := v.innerRect().Size
	s := v.CalculatedGridSize(locSize)
	r := zgeo.Rect{Size: zgeo.Size{locSize.W, s.H + 1}}
	r.Pos.Y--
	v.cellsView.SetRect(r)
	v.LayoutCells(false)
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
	// zlog.Info("LayoutCells", v.Hierarchy(), v.CellCountFunc(), v.HorizontalFirst) //, zlog.CallingStackString())
	v.layoutDirty = false
	placed := map[string]bool{}
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if visible {
			child, _ := v.makeOrGetChild(cid)
			//TODO: exit when !visible after visible
			ms, _ := child.(zview.Marginalizer)
			if ms != nil {
				marg := inner.Minus(outer)
				// zlog.Info("Grid SetMarg on child", cid, marg, v.margin)
				ms.SetMargin(marg)
			}
			o := outer.Plus(zgeo.RectFromXY2(0, 0, 0, 0))
			child.SetRect(o)
			v.updateCellBackground(cid, x, y, child)
			if updateCells && v.UpdateCellFunc != nil {
				v.UpdateCellFunc(v, cid)
			}
			placed[cid] = true
		}
		return true
	})
	for cid, view := range v.children {
		if !placed[cid] {
			v.cellsView.RemoveChild(view)
			delete(v.children, cid)
		}
	}
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
	// zlog.Info("List keypress other!", v.ObjectName(), key, mod == zkeyboard.ModifierNone)
	if km.Key.IsReturnish() && (v.Selectable || v.MultiSelectable) {
		id := v.CurrentHoverID
		if id == "" && v.selectedIndex != -1 {
			id = v.IDAtIndexFunc(v.selectedIndex)
		}
		if id != "" {
			v.selectedIndex = v.IndexOfID(id)
			if km.Modifier&zkeyboard.ModifierCommand != 0 {
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
			if v.Selectable || v.MultiSelectable {
				v.UnselectAll(true)
			}
		case 'A':
			if v.MultiSelectable {
				v.SelectAll(true)
			}
		}
	}
	if v.HandleKeyFunc != nil {
		// zlog.Info("List keypress2", key, km.Modifier)
		return v.HandleKeyFunc(km, down)
	}
	return false
}

func (v *GridListView) ReadyToShow(beforeWindow bool) {
	// zlog.Info("List ReadyToShow:", v.ObjectName(), v.CreateCellFunc != nil)
	if beforeWindow && v.BorderColor.Valid {
		// v.cellsView.SetStrokeSide(1, v.BorderColor, zgeo.TopLeft, true) // We set top and left
	}
	// if !beforeWindow && (v.Selectable || v.MultiSelectable) {
	// 	zwindow.FromNativeView(&v.NativeView).AddKeypressHandler(v.View, v.handleKeyPressed)
	// }
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
