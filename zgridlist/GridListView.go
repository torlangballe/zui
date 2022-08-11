// The zgridlist package defines GridListView, a view for displaying cells in multiple rows and columns.
// It knows nothing about the content of it's cells.
// If CellHeightFunc is set, each cells hight is based on that, otherwise based on first view created.
// zgridlist also contains SliceGridView which creates a grid from any slice.
// Also TableView, which use a GridListView and zfield package to create rows based on a struct.
// GridListView is based on functions for getting a count of cells, getting and id for an index and creating a cell's view for a given id.
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
	"github.com/torlangballe/zui/zwidget"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmap"
)

type GridListView struct {
	zscrollview.ScrollView
	Spacing            zgeo.Size
	Selectable         bool
	MultiSelectable    bool
	MakeFullSize       bool
	MaxColumns         int
	MinColumns         int
	MinRowsForFullSize int
	MaxRowsForFullSize int
	BorderColor        zgeo.Color
	CellColor          zgeo.Color
	CellColorFunc      func(id string) zgeo.Color
	MultiplyAlternate  float32
	PressedColor       zgeo.Color
	SelectColor        zgeo.Color
	HoverColor         zgeo.Color
	BranchToggleType   zwidget.BranchToggleType
	OpenBranches       map[string]bool

	CellCountFunc              func() int
	IDAtIndexFunc              func(i int) string
	CreateCellFunc             func(grid *GridListView, id string) zview.View
	UpdateCellFunc             func(grid *GridListView, id string)
	UpdateSelectionFunc        func(grid *GridListView, id string)
	CellHeightFunc             func(id string) float64 // only need to have variable-height
	HandleSelectionChangedFunc func()
	HandleHoverOverFunc        func(id string) // this gets "" id when hovering out of a cell
	HandleKeyFunc              func(key zkeyboard.Key, mod zkeyboard.Modifier) bool
	HierarchyLevelFunc         func(id string) (level int, leaf bool)

	children           map[string]zview.View
	selectedIDs        map[string]bool
	pressedIDs         map[string]bool
	selectingFromIndex int
	appending          bool
	ignoreMouseEvent   bool
	cellsView          *zcustom.CustomView
	margin             zgeo.Rect
	columns            int
	rows               int
	layoutDirty        bool
	CurrentHoverID     string
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
	v.selectingFromIndex = -1
	v.children = map[string]zview.View{}
	v.selectedIDs = map[string]bool{}
	v.pressedIDs = map[string]bool{}
	v.SetObjectName(storeName)
	v.cellsView = zcustom.NewView("GridListView.grid")
	v.AddChild(v.cellsView, -1)
	v.cellsView.SetPressUpDownMovedHandler(v.handleUpDownMovedHandler)
	v.cellsView.SetPointerEnterHandler(true, v.handleHover)
	v.CellColor = DefaultCellColor
	v.BorderColor = DefaultBorderColor
	v.SelectColor = DefaultSelectColor
	v.PressedColor = DefaultPressedColor
	v.HoverColor = DefaultHoverColor
	v.OpenBranches = map[string]bool{}
	v.BranchToggleType = zwidget.BranchToggleTriangle
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
	for i := 0; i < v.CellCountFunc(); i++ {
		if v.IDAtIndexFunc(i) == id {
			return i
		}
	}
	return -1
}

func (v *GridListView) IsSelected(id string) bool {
	return v.selectedIDs[id]
}

func (v *GridListView) IsHoverCell(id string) bool {
	return id == v.CurrentHoverID
}

func (v *GridListView) makeOpenBranchesKey() string {
	return "ZGridListView.branches." + v.ObjectName()
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
	for i := 0; i < v.CellCountFunc(); i++ {
		sid := v.IDAtIndexFunc(i)
		if v.selectedIDs[sid] {
			v.ScrollToCell(sid, animateScroll)
			break
		}
	}
	if v.layoutDirty { // this is to avoid layout if scroll did it
		v.LayoutCells(true)
	}
	if v.HandleSelectionChangedFunc != nil {
		v.HandleSelectionChangedFunc()
	}
}

func (v *GridListView) UnselectAll() {
	v.SelectCells([]string{}, false)
}

func (v *GridListView) IsPressed(id string) bool {
	return v.pressedIDs[id]
}

func (v *GridListView) CellView(id string) zview.View {
	return v.children[id]
}

func (v *GridListView) updateCellBackgrounds() {
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if !visible {
			return true
		}
		child := v.children[cid]
		if child != nil {
			v.updateCellBackground(cid, x, y, child)
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
	col := v.CellColor
	if v.CurrentHoverID == cid && v.HoverColor.Valid {
		col = v.HoverColor
	} else if v.pressedIDs[cid] && v.PressedColor.Valid {
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
		child.Native().SetStroke(1, v.BorderColor, false)
	}
	if v.UpdateSelectionFunc != nil {
		v.UpdateSelectionFunc(v, cid)
	}
}

func (v *GridListView) setPressed(index int) {
	if !v.MultiSelectable {
		v.pressedIDs[v.IDAtIndexFunc(index)] = true
		return
	}
	if v.appending {
		for id := range v.selectedIDs {
			v.pressedIDs[id] = true
		}
	}
	if v.selectingFromIndex != -1 {
		if !v.appending {
			v.pressedIDs = map[string]bool{}
		}
		min, max := zint.MinMax(index, v.selectingFromIndex)
		for i := min; i <= max; i++ {
			v.pressedIDs[v.IDAtIndexFunc(i)] = true
		}
		if v.appending {
			v.selectingFromIndex = max
		}
	}
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
	if v.CurrentHoverID != "" && v.UpdateCellFunc != nil && v.children[v.CurrentHoverID] != nil {
		v.UpdateCellFunc(v, v.CurrentHoverID)
	}
	v.CurrentHoverID = id
	if v.CurrentHoverID != "" && v.UpdateCellFunc != nil {
		v.UpdateCellFunc(v, v.CurrentHoverID)
	}
	if v.HoverColor.Valid {
		v.updateCellBackgrounds()
	}
	if v.HandleHoverOverFunc != nil {
		v.HandleHoverOverFunc(id)
	}
}

func (v *GridListView) handleUpDownMovedHandler(pos zgeo.Pos, down zbool.BoolInd) bool {
	v.CurrentHoverID = ""
	if !v.Selectable && !v.MultiSelectable {
		return false
	}
	eventHandled := true
	var index int
	id, inside := v.FindCellForPos(pos)
	if id != "" {
		index = v.IndexOfID(id)
		if index == -1 {
			zlog.Info("No index for id:", id)
			return false
		}
	}
	switch down {
	case zbool.True:
		if !inside || id == "" {
			return false
		}
		var interactive bool
		interactive, eventHandled = v.isInsideInteractiveChildCell(id, pos)
		if interactive {
			return false
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
		if !v.appending {
			v.selectingFromIndex = index
		}
		if v.appending && v.selectingFromIndex == -1 && len(v.SelectedIDs()) > 0 {
			v.selectingFromIndex = v.IndexOfID(v.SelectedIDs()[0])
		}
		// zlog.Info("SELINX:", v.selectingFromIndex, v.appending, len(v.SelectedIDs()), index)
		v.setPressed(index)
	case zbool.Unknown:
		if id == "" {
			return false
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
		// v.selectingFromIndex = -1
		v.pressedIDs = map[string]bool{}
	}
	v.updateCellBackgrounds()
	v.ExposeIn(0.0001)
	if v.HandleSelectionChangedFunc != nil {
		v.HandleSelectionChangedFunc()
	}
	return eventHandled
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
			// zlog.Info("box:", cellID, outer, pos)
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
	s := v.MinSize()
	if v.CellCountFunc() == 0 {
		return s
	}
	if v.MakeFullSize {
		return v.CalculatedGridSize(total)
	}
	childSize := v.getAChildSize(total)
	mx := math.Max(1, float64(v.MinColumns))
	w := childSize.W*mx + v.Spacing.W*(mx-1) - v.margin.Size.W
	zfloat.Maximize(&s.W, w)
	return s
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
	s.W += childSize.W*x + v.Spacing.W*(x-1)
	s.H += childSize.H*y + v.Spacing.H*(y-1)
	// zlog.Info(childSize, "CalculatedGridSize:", nx, ny, s, childSize.H*x, v.Spacing.H*(y-1))
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

func (v *GridListView) getAChildSize(total zgeo.Size) zgeo.Size {
	// zlog.Info("getAChild", v.Parent().ObjectName(), v.CreateCellFunc != nil)
	cid := v.IDAtIndexFunc(0)
	child := v.CreateCellFunc(v, cid)
	s := child.CalculatedSize(total)
	if v.CellHeightFunc != nil {
		zfloat.Maximize(&s.H, v.CellHeightFunc(cid))
	}
	return s
}

func (v *GridListView) GetChildren(collapsed bool) (children []zview.View) {
	for _, c := range v.children {
		children = append(children, c)
	}
	return
}

func (v *GridListView) SetRect(rect zgeo.Rect) {
	v.ScrollView.SetRect(rect)
	// zlog.Info("GL: SetRect:", rect, v.Rect(), v.cellsView.Rect())
	at := v.View.(zcontainer.Arranger) // in case we are a stack or something inheriting from GridListView
	at.ArrangeChildren()
}

func (v *GridListView) insertBranchToggle(id string, child zview.View) {
	level, leaf := v.HierarchyLevelFunc(id)
	co, _ := child.(zcontainer.CellsOwner)
	aa, _ := child.(zcontainer.AdvancedAdder)
	if level > 0 {
		w := float64(level-1) * 14
		if v.BranchToggleType != zwidget.BranchToggleNone {
			if !leaf {
				bt := zwidget.BranchToggleViewNew(v.BranchToggleType, id, v.OpenBranches[id])
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

func (v *GridListView) makeOrGetChild(id string) zview.View {
	child := v.children[id]
	if child != nil {
		return child
	}
	child = v.CreateCellFunc(v, id)
	if v.HierarchyLevelFunc != nil {
		v.insertBranchToggle(id, child)
	}
	v.children[id] = child
	v.cellsView.AddChild(child, -1)
	if v.UpdateCellFunc != nil {
		v.UpdateCellFunc(v, id)
	}
	zpresent.CallReady(child, false)
	child.(zview.ExposableType).Expose()
	return child
}

func (v *GridListView) ForEachCell(got func(cellID string, outer, inner zgeo.Rect, x, y int, visible bool) bool) {
	if v.CellCountFunc() == 0 {
		return
	}
	rect := v.LocalRect()
	pos := rect.Pos
	childSize := rect.Size.Plus(v.margin.Size)

	if v.CellHeightFunc == nil {
		childSize = v.getAChildSize(rect.Size)
	} else {
		zlog.Assert(v.MaxColumns <= 1)
	}
	v.columns, v.rows = v.CalculateColumnsAndRows(childSize.W, rect.Size.W)
	var x, y int
	for i := 0; i < v.CellCountFunc(); i++ {
		cellID := v.IDAtIndexFunc(i)
		if v.CellHeightFunc != nil {
			childSize.H = v.CellHeightFunc(cellID)
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
			maxx = -(rect.Max().X - r.Max().X) + 1
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
		}
		maxy -= v.Spacing.H / 2
		marg := zgeo.RectFromXY2(minx, miny, maxx, maxy)
		r.Size.Add(marg.Size.Negative())
		cr := r.Plus(marg)
		visible := (r.Max().Y >= v.YOffset && r.Min().Y <= v.YOffset+v.LocalRect().Size.H)
		r2 := r
		if v.BorderColor.Valid {
			r2.Size.Subtract(zgeo.Size{1, 1})
		}
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
	v.layoutDirty = false
	placed := map[string]bool{}
	v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool {
		if visible {
			// zlog.Info("Arrange:", cid, updateCells)
			child := v.makeOrGetChild(cid)
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
		v.MakeRectVisible(outerRect.Plus(zgeo.RectFromXY2(0, 0, 0, 1)), animate)
	}
}

func (v *GridListView) moveSelection(incX, incY int, mod zkeyboard.Modifier) bool {
	var sx, sy int
	var fid string

	if v.CurrentHoverID != "" {
		hoverID := v.CurrentHoverID
		v.CurrentHoverID = ""
		v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool { // find a selected cell and x,y
			if visible {
				if cid == hoverID {
					v.updateCellBackground(cid, x, y, nil)
					return false
				}
			}
			return true
		})
	}
	append := (mod == zkeyboard.ModifierShift)
	// zlog.Info("moveSelection:", append)
	if append && v.selectingFromIndex != -1 {
		fid = v.IDAtIndexFunc(v.selectingFromIndex)
		if fid != "" {
			sx, sy = v.FindColRowForID(fid)
		}
	}
	if fid == "" {
		v.ForEachCell(func(cid string, outer, inner zgeo.Rect, x, y int, visible bool) bool { // find a selected cell and x,y
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
	}
	if fid != "" {
		if len(v.selectedIDs) == 0 {
			v.selectedIDs[fid] = true
			v.updateCellBackground(fid, sx, sy, nil)
			return true
		} else {
			nx := sx + incX
			ny := sy + incY
			// zlog.Info("moveSel:", v.columns, v.rows, fid, "s:", sx, sy, "n:", nx, ny, ny*v.columns+nx >= len(v.CellIDs))
			if append && nx < 0 && ny > 0 {
				nx = v.columns - 1
				ny--
			} else if append && nx >= v.columns && ny < v.rows-1 {
				nx = 0
				ny++
			} else if nx < 0 || nx >= v.columns || ny < 0 || ny >= v.rows {
				return true
			}
			if ny*v.columns+nx >= v.CellCountFunc() {
				if incY != 1 {
					return true
				}
				nx = 0
			}
			nid := v.FindCellForColRow(nx, ny)
			if append {
				startIndex := v.selectingFromIndex
				if startIndex == -1 {
					startIndex = v.IndexOfID(fid)
				}
				endIndex := v.IndexOfID(nid)
				min, max := zint.MinMax(startIndex, endIndex)
				// zlog.Info(nx, ny, "Append:", v.selectingFromIndex, min, max, fid, nid)
				for i := min; i <= max; i++ {
					v.selectedIDs[v.IDAtIndexFunc(i)] = true
				}
				v.updateCellBackgrounds()
				v.ScrollToCell(nid, false)
				v.selectingFromIndex = endIndex
			} else {
				v.selectedIDs = map[string]bool{}
				v.updateCellBackgrounds()
				// zlog.Info("MoveTo:", nid, len(v.children))
				if nid != "" {
					v.selectedIDs[nid] = true
					v.ScrollToCell(nid, false)
					v.updateCellBackground(nid, nx, ny, nil)
				}
			}
		}
		if v.HandleSelectionChangedFunc != nil {
			v.HandleSelectionChangedFunc()
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
		bt, _ := view.(*zwidget.BranchToggleView)
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

func (v *GridListView) ReadyToShow(beforeWindow bool) {
	// zlog.Info("List ReadyToShow:", v.ObjectName(), v.CreateCellFunc != nil)

	if !beforeWindow && (v.Selectable || v.MultiSelectable) {
		zwindow.GetFromNativeView(&v.NativeView).AddKeypressHandler(v.View, func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
			// zlog.Info("List keypress!", v.ObjectName(), key, mod == zkeyboard.ModifierNone)
			if mod == zkeyboard.ModifierNone || mod == zkeyboard.ModifierShift {
				//					v.doRowPressed(v.highlightedIndex)
				switch key {
				case zkeyboard.KeyUpArrow:
					return v.moveSelection(0, -1, mod)
				case zkeyboard.KeyDownArrow:
					return v.moveSelection(0, 1, mod)
				case zkeyboard.KeyLeftArrow:
					if v.HierarchyLevelFunc != nil && v.MaxColumns == 1 {
						v.toggleBranch(false)
						return true
					}
					return v.moveSelection(-1, 0, mod)
				case zkeyboard.KeyRightArrow:
					if v.HierarchyLevelFunc != nil && v.MaxColumns == 1 {
						v.toggleBranch(true)
						return true
					}
					return v.moveSelection(1, 0, mod)
				case zkeyboard.KeyReturn, zkeyboard.KeyEnter:
					break // do select
				case zkeyboard.KeyEscape:
					v.UnselectAll()
				}
			}
			// zlog.Info("List keypress2", v.HandleKeyFunc != nil)
			if v.HandleKeyFunc != nil {
				return v.HandleKeyFunc(key, mod)
			}
			return false
		})
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
