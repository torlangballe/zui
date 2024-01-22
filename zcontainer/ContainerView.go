//go:build zui

package zcontainer

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztimer"
)

// Original class created by Tor Langballe on 23-sept-2014.
type ContainerView struct {
	zcustom.CustomView
	margin            zgeo.Rect
	singleOrientation bool
	Cells             []Cell
}

type Cell struct {
	zgeo.LayoutCell
	View    zview.View
	AnyInfo any
}

type CellsCounter interface {
	CountChildren() int
}

type CellsOwner interface {
	GetCells() *[]Cell
}

type CellAdder interface {
	AddCell(cell Cell, index int) *Cell
}

type AdvancedAdder interface {
	AddAdvanced(view zview.View, align zgeo.Alignment, marg zgeo.Size, maxSize zgeo.Size, index int, free bool) *Cell
}

type ChildrenOwner interface {
	GetChildren(includeCollapsed bool) []zview.View
}

type Arranger interface {
	ArrangeChildren()
}

type Collapser interface {
	CollapseChild(view zview.View, collapse bool, arrange bool) bool
}

var (
	GroupingStrokeColor  = zgeo.ColorNewGray(0.7, 1)
	GroupingStrokeWidth  = 2.0
	GroupingStrokeCorner = 4.0
	GroupingMargin       = 10.0
	AlertButtonsOnRight  = true
)

func New(name string) *ContainerView {
	v := &ContainerView{}
	v.Init(v, name)
	return v
}

func (v *ContainerView) Init(view zview.View, name string) {
	v.CustomView.Init(view, name)
}

func CountChildren(v zview.View) int {
	ct, _ := v.(CellsCounter)
	if ct == nil {
		return 0
	}
	return ct.CountChildren()
}

func (v *ContainerView) GetChildren(includeCollapsed bool) (children []zview.View) {
	for _, c := range v.Cells {
		if (includeCollapsed || !c.Collapsed) && c.View != nil {
			children = append(children, c.View)
		}
	}
	return
}

func (v *ContainerView) GetCells() *[]Cell {
	return &v.Cells
}

func ArrangeAncestorContainer(view zview.View) {
	a := FindAncestorArranger(view)
	if a == nil {
		zlog.Error(nil, "no parent arranger", view.Native().Hierarchy())
		return
	}
	a.ArrangeChildren()
}

func FindAncestorArranger(view zview.View) Arranger {
	nv := view.Native()
	for {
		if nv.Parent() == nil {
			return nil
		}
		a, _ := nv.View.(Arranger)
		if a != nil {
			return a
		}
		nv = nv.Parent()
	}
}

func (v *ContainerView) CountChildren() int {
	return len(v.Cells)
}

func (v *ContainerView) addCellWithAdder(cell Cell, index int) *Cell {
	a, _ := v.View.(CellAdder)
	return a.AddCell(cell, index)
}

func (v *ContainerView) Add(view zview.View, align zgeo.Alignment, sizes ...zgeo.Size) (first *Cell) {
	var marg, maxSize zgeo.Size
	if len(sizes) > 0 {
		marg = sizes[0]
	}
	if len(sizes) > 1 {
		maxSize = sizes[1]
	}
	return v.AddAdvanced(view, align, marg, maxSize, -1, false)
}

func (v *ContainerView) AddAlertButton(button zview.View) {
	a := zgeo.VertCenter
	if AlertButtonsOnRight {
		a |= zgeo.Right
	} else {
		a |= zgeo.Left
	}
	v.Add(button, a)
}

func (v *ContainerView) SetMargin(margin zgeo.Rect) {
	// zlog.Info("CV SetMargin:", v.ObjectName(), margin)
	v.margin = margin
}

func (v *ContainerView) Margin() zgeo.Rect {
	return v.margin
}

func (v *ContainerView) SetMarginS(margin zgeo.Size) {
	v.margin = zgeo.RectFromMinMax(margin.Pos(), margin.Pos().Negative())
}

func (v *ContainerView) SetSingleOrientation(single bool) {
	v.singleOrientation = single
}

func (v *ContainerView) AddCell(cell Cell, index int) (cvs *Cell) {
	// zlog.Info("AddCell:", v.ObjectName(), index)
	if index < 0 || index >= len(v.Cells) {
		v.Cells = append(v.Cells, cell)
		if cell.View != nil {
			v.AddChild(cell.View, -1)
		}
		return &v.Cells[len(v.Cells)-1]
	}
	n := append(append([]Cell{}, v.Cells[:index]...), cell) // convoluted way of doing it due to append altering first argument
	v.Cells = append(n, v.Cells[index:]...)
	// for i, c := range v.Cells {
	// 	zlog.Info("AddCell insert:", v.ObjectName(), i, c.View.ObjectName())
	// }
	if cell.View != nil {
		v.AddChild(cell.View, -1)
	}
	return &v.Cells[index]
}

// func (v *ContainerView) AddView(view zview.View, align zgeo.Alignment) *Cell {
// 	return v.AddAdvanced(view, align, zgeo.Size{}, zgeo.Size{}, -1, false)
// }

func (v *ContainerView) AddAdvanced(view zview.View, align zgeo.Alignment, marg zgeo.Size, maxSize zgeo.Size, index int, free bool) *Cell {
	collapsed := false
	// zlog.Info("CV AddAdvancedView:", view != nil, view.Native() != nil)
	name := "nil"
	if view != nil {
		name = view.ObjectName()
	}
	lc := zgeo.LayoutCell{align, marg, maxSize, zgeo.Size{}, collapsed, free, zgeo.Size{}, 0.0, name}
	return v.addCellWithAdder(Cell{LayoutCell: lc, View: view}, index)
}

func (v *ContainerView) Contains(view zview.View) bool {
	for _, c := range v.Cells {
		if c.View == view {
			return true
		}
	}
	return false
}

func (v *ContainerView) SetRect(rect zgeo.Rect) {
	// zlog.Info("CV SetRect2", v.ObjectName(), rect)
	v.CustomView.SetRect(rect)
	// zlog.Info("CV SetRect2", v.ObjectName(), v.CustomView.LocalRect())
	at := v.View.(Arranger) // in case we are a stack or something inheriting from ContainerView
	//	start := time.Now()
	at.ArrangeChildren()
	// d := time.Since(start)
	// if d > time.Millisecond*50 {
	// 	zlog.Info("CV SetRect2", v.ObjectName(), d)
	// }
}

func (v *ContainerView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.MinSize()
}

// func (v *ContainerView) SetAsFullView(useableArea bool) {
// 	sm := zscreen.GetMain()
// 	r := sm.Rect
// 	if useableArea {
// 		r = sm.UsableRect
// 	}
// 	v.SetRect(r)
// 	v.SetMinSize(r.Size)
// }

func (v *ContainerView) ArrangeChildrenAnimated() {
	//        ZAnimation.Do(duration 0.6, animations  { [weak self] () in
	v.ArrangeChildren()
	//        })
}

func (v *ContainerView) ArrangeChild(c Cell, r zgeo.Rect) {
	if c.Alignment != zgeo.AlignmentNone {
		ir := r.Expanded(c.Margin.MinusD(2.0))
		s := c.View.CalculatedSize(ir.Size)
		var rv = r.AlignPro(s, c.Alignment, c.Margin, c.MaxSize, zgeo.Size{})
		c.View.SetRect(rv)
	}
}

func ContainerIsLoading(ct ChildrenOwner) bool {
	// zlog.Info("ContainerIsLoading1", ct.(View).ObjectName(), len(ct.GetChildren(false)))
	for _, v := range ct.GetChildren(false) {
		iloader, got := v.(zimage.Loader)
		if got {
			loading := iloader.IsLoading()
			// zlog.Info("ContainerIsLoading image loading", v.ObjectName(), loading)
			if loading {
				return true
			}
		} else {
			ct, _ := v.(ChildrenOwner)
			// zlog.Info("CV Sub IsLoading:", v.ObjectName(), v.ObjectName(), ct != nil)
			if ct != nil {
				if ContainerIsLoading(ct) {
					// zlog.Info("ContainerIsLoading sub loading", len(ct.GetChildren()))
					return true
				}
			}
		}
	}
	// zlog.Info("ContainerIsLoading Done", ct.(View).ObjectName())
	return false
}

// WhenContainerLoaded waits for all sub-parts images etc to be loaded before calling done.
// done received waited=true if it had to wait
func WhenContainerLoaded(ct ChildrenOwner, done func(waited bool)) {
	// start := time.Now()
	ztimer.RepeatAtMostEvery(0.1, func() bool {
		if ContainerIsLoading(ct) {
			// zlog.Info("Wait:", ct.(View).ObjectName())
			return true
		}
		if done != nil {
			// zlog.Info("Waited:", time.Since(start), ct.(View).ObjectName())
			done(true)
		}
		return false
	})
}

func (v *ContainerView) ArrangeChildren() {
	v.ArrangeAdvanced(false)
}

func (v *ContainerView) ArrangeAdvanced(freeOnly bool) {
	// zlog.Info("*********** ContainerView.ArrangeChildren:", v.Hierarchy(), v.Rect(), len(v.Cells))
	layouter, _ := v.View.(zview.Layouter)
	if layouter != nil {
		layouter.HandleBeforeLayout()
	}
	r := zgeo.Rect{Size: v.Rect().Size}.Plus(v.margin)
	for _, c := range v.Cells {
		clayouter, _ := c.View.(zview.Layouter)
		if clayouter != nil {
			clayouter.HandleBeforeLayout()
		}
		if c.Alignment != zgeo.AlignmentNone && (!freeOnly || c.Free) {
			// zlog.Info("ArrangeAdvanced:", v.ObjectName(), v.Rect(), c.View.ObjectName(), c.View.Rect().Size, c.Alignment)
			v.ArrangeChild(c, r)
			at, _ := c.View.(Arranger) // we might be "inherited" by StackView or something
			if at != nil {
				at.ArrangeChildren()
			}
		}
	}
	if layouter != nil {
		layouter.HandleAfterLayout()
	}
	for _, c := range v.Cells {
		clayouter, _ := c.View.(zview.Layouter)
		if clayouter != nil {
			clayouter.HandleAfterLayout()
		}
	}
}

func (v *ContainerView) CollapseChild(view zview.View, collapse bool, arrange bool) (changed bool) {
	cell, _ := v.FindCellWithView(view)
	if cell == nil {
		return false
	}
	changed = (cell.Collapsed != collapse)
	// if cell.View.ObjectName() == "xxx" {
	// zlog.Info("COLLAPSE:", cell.Collapsed, collapse, changed, v.Hierarchy(), view.ObjectName(), cell.View.ObjectName())
	// }
	if collapse {
		cell.View.Show(false)
	}
	nc, _ := view.(*zcustom.CustomView)

	if changed {
		cell.Collapsed = collapse
		if collapse {
			//detachFromContainer := false
			v.CustomView.RemoveChild(view)
			// if view.ObjectName() == "SL01@dash" {
			// 	zlog.Info("Collapse:", view.ObjectName())
			// }
			// v.RemoveChild(view)
			cell = nil // force this to avoid use from here on
			// zlog.Info("COLLAPSED:", view.ObjectName())
		} else {
			if nc != nil {
				nc.SetVisible(true)
			}
			v.AddChild(cell.View, -1)
		}
	}
	if arrange && v.IsPresented() {
		at := v.View.(Arranger) // we might be "inherited" by StackView or something
		at.ArrangeChildren()
	}
	if !collapse {
		cell.View.Show(true)
		zview.ExposeView(cell.View)
	}
	return changed
}

func (v *ContainerView) CollapseChildWithName(name string, collapse bool, arrange bool) bool {
	view, _ := v.FindViewWithName(name, false)
	// zlog.Info("Collapse:", name, collapse, view != nil)
	if view != nil {
		return v.CollapseChild(view, collapse, arrange)
	}
	return false
}

// ViewRangeChildren loops through *view's* children (if it's a container) calling foreach for each one.
// If foreach returns false it stops, and returns false itself
func ViewRangeChildren(view zview.View, subViews, includeCollapsed bool, foreach func(view zview.View) bool) bool {
	ct, _ := view.(ChildrenOwner)
	// zlog.Info("ContainerViewRangeChildren1:", view.Native().Hierarchy(), ct != nil)
	if ct == nil {
		return true // we return true here, as it wasn't a container
	}
	children := ct.GetChildren(includeCollapsed)
	for _, c := range children {
		if c == nil {
			zlog.Fatal(nil, "nil child in range", view.Native().Hierarchy())
		}
		cont := foreach(c)
		// zlog.Info("ContainerViewRangeChildren1:", c.ObjectName(), subViews, cont)
		if !cont {
			return false
		}
	}
	if !subViews {
		return true
	}
	for _, c := range children {
		if !ViewRangeChildren(c, subViews, includeCollapsed, foreach) {
			return false
		}
	}
	return true
}

func (v *ContainerView) RemoveNamedChild(name string, all bool) bool {
	for {
		removed := false
		for _, c := range v.Cells {
			if c.View.ObjectName() == name {
				v.RemoveChild(c.View)
				removed = true
				if !all {
					return true
				}
			}
		}
		if !removed {
			return false
		}
	}
	return true
}

func (v *ContainerView) FindViewWithName(name string, recursive bool) (zview.View, int) {
	return ContainerOwnerFindViewWithName(v, name, recursive)
}

func ContainerOwnerFindViewWithName(view zview.View, name string, recursive bool) (zview.View, int) {
	var found zview.View

	ct, _ := view.(ChildrenOwner)
	if ct == nil {
		zlog.Fatal(nil, "view is not container")
		return nil, -1
	}
	i := 0
	includeCollapsed := true
	ViewRangeChildren(view, recursive, includeCollapsed, func(view zview.View) bool {
		// zlog.Info("FindViewWithName:", name, "==", view.ObjectName())
		if view.ObjectName() == name {
			found = view
			return false
		}
		i++
		return true
	})
	return found, i
}

func (v *ContainerView) FindCellWithName(name string) (*Cell, int) {
	for i, c := range v.Cells {
		if c.View.ObjectName() == name {
			return &v.Cells[i], i
		}
	}
	return nil, -1
}

func (v *ContainerView) FindCellWithView(view zview.View) (*Cell, int) {
	for i, c := range v.Cells {
		// fmt.Printf("FindCellWithView: %s %p != %p %v\n", c.View.ObjectName(), c.View.Native(), view.Native(), c.View == view)
		// zlog.Info("FindCellWithView:", c.View.ObjectName(), reflect.TypeOf(c.View), reflect.TypeOf(view))
		if c.View == view {
			return &v.Cells[i], i
		}
	}
	return nil, -1
}

func (v *ContainerView) RemoveChild(subView zview.View) {
	v.DetachChild(subView)
	v.CustomView.RemoveChild(subView)
}

func (v *ContainerView) RemoveAllChildren() {
	for _, c := range v.Cells {
		v.CustomView.RemoveChild(c.View)
	}
	v.Cells = v.Cells[:0]
}

func (v *ContainerView) DetachChild(subView zview.View) {
	for i, c := range v.Cells {
		if c.View == subView {
			zslice.RemoveAt(&v.Cells, i)
			break
		}
	}
}

func (v *ContainerView) ReplaceChild(child, with zview.View) {
	c, _ := v.FindCellWithView(child)
	if c == nil {
		zlog.Error(nil, "CV ReplaceChild: old not found:", child.Native().Hierarchy(), "in:", v.Hierarchy())
		for _, c := range v.GetChildren(true) {
			fmt.Printf("Children: %s %p != %p\n", c.ObjectName(), c, child)
		}
		return
	}
	v.NativeView.ReplaceChild(child, with)
	c.View = with
}

// CollapseView collapses/uncollapses a view in it's parent which is Collapsable type. (ContainerView)
func CollapseView(v zview.View, collapse, arrange bool) bool {
	p := v.Native().Parent()
	c := p.View.(Collapser) // crash if parent isn't ContainerView of some sort
	return c.CollapseChild(v, collapse, arrange)
}

// SortChildren re-arranges order of Cells with views in them.
// It does not curr
func (v *ContainerView) SortChildren(less func(a, b zview.View) bool) {
	sort.Slice(v.Cells, func(i, j int) bool {
		return less(v.Cells[i].View, v.Cells[j].View)
	})
}

func ChildView(v zview.View, path string) zview.View {
	if path == "" {
		return v
	}
	parts := strings.Split(path, "/")
	// zlog.Info("ViewChild:", v.ObjectName(), path)
	name := parts[0]
	if name == ".." {
		path = strings.Join(parts[1:], "/")
		parent := v.Native().Parent()
		if parent != nil {
			return ChildView(parent, path)
		}
		return nil
	}
	ct, _ := v.(ChildrenOwner)
	if ct == nil {
		zlog.Error(nil, "ChildView from non-container", v.Native().Hierarchy(), reflect.TypeOf(v))
		return nil
	}
	for _, ch := range ct.GetChildren(true) {
		// zlog.Info("Childs:", name, "'"+ch.ObjectName()+"'")
		if name == "*" || ch.ObjectName() == name {
			if len(parts) == 1 {
				return ch
			}
			path = strings.Join(parts[1:], "/")
			gotView := ChildView(ch, path)
			if gotView != nil || name != "*" {
				return gotView
			}
		}
	}
	return nil
}

func ArrangeChildrenAtRootContainer(view zview.View) {
	for _, p := range view.Native().AllParents() {
		a, _ := p.View.(Arranger)
		if a != nil {
			a.ArrangeChildren()
			return
		}
	}
}

func HandleOutsideShortcutRecursively(view zview.View, sc zkeyboard.KeyMod) bool {
	var handled bool
	sh, _ := view.(zkeyboard.ShortcutHandler)
	if sh != nil && sh.HandleOutsideShortcut(sc) {
		return true
	}
	ViewRangeChildren(view, true, false, func(v zview.View) bool {
		sh, _ := v.(zkeyboard.ShortcutHandler)
		if sh != nil {
			if sh.HandleOutsideShortcut(sc) {
				handled = true
				return false
			}
		}
		return true
	})
	return handled
}

func FocusNext(view zview.View, recursive, loop bool) {
	foc := view.Native().GetFocusedChildView(false)
	var first *zview.NativeView
	ViewRangeChildren(view, recursive, false, func(view zview.View) bool {
		if !view.Native().CanTabFocus() || !view.Native().IsShown() {
			return true
		}
		first = view.Native()
		if foc != nil && view == foc {
			foc = nil
			return true
		}
		if foc == nil {
			view.Native().Focus(true)
			first = nil
			return false
		}
		return true
	})
	if loop && first != nil {
		first.Focus(true)
	}
}

func init() {
	zview.RangeAllChildrenFunc = func(root zview.View, visible bool, got func(zview.View) bool) {
		// ct, _ := root.(ContainerOwner)
		// zlog.Info("RangeAllVisibleChildrenFunc:", ct != nil, reflect.TypeOf(root))
		recursive := true
		includeCollapsed := false
		ViewRangeChildren(root, recursive, includeCollapsed, got)
	}
	zview.ChildOfViewFunc = ChildView
}
