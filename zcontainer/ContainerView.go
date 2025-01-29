//go:build zui

package zcontainer

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdebug"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztimer"
)

// Original class created by Tor Langballe on 23-sept-2014.
type ContainerView struct {
	zcustom.CustomView
	margin             zgeo.Rect
	singleOrientation  bool
	Cells              []Cell
	InitialFocusedView zview.View
}

type Cell struct {
	zgeo.LayoutCell
	NotInGrid bool
	View      zview.View
	AnyInfo   any
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
	AddAdvanced(view zview.View, align zgeo.Alignment, marg zgeo.Rect, maxSize zgeo.Size, index int, free bool) *Cell
}

type ChildrenOwner interface {
	GetChildren(includeCollapsed bool) []zview.View
}

type Arranger interface {
	ArrangeChildren()
	ArrangeChild(Cell, zgeo.Rect) zgeo.Rect
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

func init() {
	zview.RangeChildrenFunc = func(root zview.View, recursive, includeCollapsed bool, got func(zview.View) bool) {
		// ct, _ := root.(ChildrenOwner)
		// zlog.Info("RangeAllVisibleChildrenFunc:", ct != nil, reflect.TypeOf(root))
		ViewRangeChildren(root, recursive, includeCollapsed, got)
	}
	zview.ChildOfViewFunc = ChildView
	zview.FindLayoutCellForView = func(view zview.View) *zgeo.LayoutCell {
		co, _ := view.Native().Parent().View.(CellsOwner)
		if co != nil {
			for _, c := range *(co.GetCells()) {
				if c.View == view {
					return &c.LayoutCell
				}
			}
		}
		return nil
	}
}

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
		zlog.Error("no parent arranger", view.Native().Hierarchy())
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
		nv = nv.Parent()
		a, _ := nv.View.(Arranger)
		if a != nil {
			return a
		}
	}
}

func (v *ContainerView) CountChildren() int {
	return len(v.Cells)
}

func (v *ContainerView) addCellWithAdder(cell Cell, index int) *Cell {
	a, _ := v.View.(CellAdder)
	return a.AddCell(cell, index)
}

func (v *ContainerView) Add(view zview.View, align zgeo.Alignment, params ...any) (first *Cell) {
	return v.AddBefore(view, nil, align, params...)
}

func (v *ContainerView) AddBefore(view, before zview.View, align zgeo.Alignment, params ...any) (first *Cell) {
	var maxSize zgeo.Size
	var marg zgeo.Rect
	var got bool
	if len(params) > 0 {
		marg, got = params[0].(zgeo.Rect)
		if !got {
			m := params[0].(zgeo.Size)
			marg = zgeo.RectMarginForSizeAndAlign(m, align)
		}
	}
	if len(params) > 1 {
		maxSize = params[1].(zgeo.Size)
	}
	i := -1
	if before != nil {
		_, i = v.FindCellWithView(before)
	}
	return v.AddAdvanced(view, align, marg, maxSize, i, false)
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
	if index < 0 || index >= len(v.Cells) {
		v.Cells = append(v.Cells, cell)
		if cell.View != nil {
			v.AddChild(cell.View, nil)
		}
		return &v.Cells[len(v.Cells)-1]
	}
	before := v.Cells[index].View
	n := append(append([]Cell{}, v.Cells[:index]...), cell) // convoluted way of doing it due to append altering first argument
	v.Cells = append(n, v.Cells[index:]...)
	// for i, c := range v.Cells {
	// 	zlog.Info("AddCell insert:", v.ObjectName(), i, c.View.ObjectName())
	// }
	if cell.View != nil {
		v.AddChild(cell.View, before)
	}
	return &v.Cells[index]
}

// func (v *ContainerView) AddView(view zview.View, align zgeo.Alignment) *Cell {
// 	return v.AddAdvanced(view, align, zgeo.SizeNull, zgeo.SizeNull, -1, false)
// }

func (v *ContainerView) AddAdvanced(view zview.View, align zgeo.Alignment, marg zgeo.Rect, maxSize zgeo.Size, index int, free bool) *Cell {
	collapsed := false
	// zlog.Info("CV AddAdvancedView:", view != nil, view.Native() != nil)
	name := "nil"
	if view != nil {
		name = view.ObjectName()
	}
	lc := zgeo.LayoutCell{
		Alignment: align,
		Margin:    marg,
		MaxSize:   maxSize,
		Collapsed: collapsed,
		Free:      free,
		Name:      name,
	}
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

func (v *ContainerView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	s = v.MinSize()
	return s, zgeo.Size{}
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

func (v *ContainerView) ReadyToShow(beforeWindow bool) {
	v.CustomView.ReadyToShow(beforeWindow)
	// zlog.Info("Container Ready0", v.Hierarchy(), beforeWindow, v.InitialFocusedView != nil)
	if beforeWindow {
		return
	}
	if v.InitialFocusedView == nil {
		return
	}
	foc := v.Native().GetFocusedChildView(true)
	// zlog.Info("Container Ready1", v.Hierarchy(), foc != nil)
	if foc != nil {
		return
	}
	has := (v.InitialFocusedView == v.View)
	if !has {
		has = v.HasViewRecursive(v.InitialFocusedView)
	}
	// zlog.Info("Container Ready", v.Hierarchy(), has, zlog.Pointer(v.InitialFocusedView), zlog.Pointer(v.View))
	if has {
		v.InitialFocusedView.Native().Focus(true)
	}
}

func (v *ContainerView) ArrangeChildrenAnimated() {
	//        ZAnimation.Do(duration 0.6, animations  { [weak self] () in
	v.ArrangeChildren()
	//        })
}

func (v *ContainerView) ArrangeChild(c Cell, r zgeo.Rect) zgeo.Rect {
	if c.Alignment == zgeo.AlignmentNone {
		return zgeo.RectNull
	}
	ir := r.ExpandedD(-1) // -2?
	s, _ := c.View.CalculatedSize(ir.Size)
	if c.RelativeToName != "" {
		rv, _ := v.FindCellWithName(c.RelativeToName)
		// zlog.Info("CV Arrange relname:", c.View.Native().Hierarchy(), c.RelativeToName, rv != nil)
		if rv != nil && rv.View != nil {
			r = rv.View.Rect()
		}
	}
	var rv = r.AlignPro(s, c.Alignment, c.Margin, c.MaxSize, zgeo.SizeNull)
	// if c.View != nil && c.View.ObjectName() == "left-pole" {
	// 	zlog.Info("ALIGN:", r, s, c.Alignment, c.Margin, "->", rv)
	// }
	c.View.SetRect(rv)
	return rv
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
		if c.View == nil {
			continue
		}
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
	// zlog.Info("COLLAPSE:", collapse, "changed:", changed, "arrange:", arrange, v.Hierarchy(), view.ObjectName())
	// }
	if collapse {
		cell.View.Show(false)
	}
	nc, _ := view.(*zcustom.CustomView)

	if changed {
		cell.Collapsed = collapse
		if collapse {
			//detachFromContainer := false
			v.CustomView.RemoveChild(view, false)
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
			v.AddChild(cell.View, nil)
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
			zlog.Fatal("nil child in range", view.Native().Hierarchy())
		}
		cont := foreach(c)
		// zlog.Info("ContainerViewRangeChildren2:", c.ObjectName(), subViews, cont)
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

func (v *ContainerView) RemoveNamedChild(name string, all, callRemoveFuncs bool) bool {
	for {
		removed := false
		for _, c := range v.Cells {
			if c.View == nil {
				continue
			}
			if c.View.ObjectName() == name {
				v.RemoveChild(c.View, callRemoveFuncs)
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
}

func (v *ContainerView) HasViewRecursive(view zview.View) bool {
	var has bool
	// zlog.Info("HasViewRecursive0:", reflect.TypeOf(v), reflect.TypeOf(v.View), ":", reflect.TypeOf(view))
	if v.View == view {
		return true
	}
	includeCollapsed := true
	recursive := true
	// zlog.Info("HasViewRecursive1:", zlog.Pointer(view), view.ObjectName())
	ViewRangeChildren(v.View, recursive, includeCollapsed, func(child zview.View) bool {
		// zlog.Info("HasViewRecursive:", view.ObjectName(), zlog.Pointer(view), "==", child.ObjectName(), zlog.Pointer(child))
		if child == view {
			has = true
			return false
		}
		return true
	})
	return has
}

func (v *ContainerView) FindViewWithName(name string, recursive bool) (zview.View, int) {
	return ContainerOwnerFindViewWithName(v, name, recursive)
}

func ContainerOwnerFindViewWithName(view zview.View, name string, recursive bool) (zview.View, int) {
	var found zview.View
	ct, _ := view.(ChildrenOwner)
	if ct == nil {
		zlog.Fatal("view is not container")
		return nil, -1
	}
	fi := -1
	i := 0
	includeCollapsed := true
	ViewRangeChildren(view, recursive, includeCollapsed, func(view zview.View) bool {
		// zlog.Info("FindViewWithName:", name, "==", view.ObjectName())
		if view.ObjectName() == name {
			found = view
			fi = i
			return false
		}
		i++
		return true
	})
	return found, fi
}

func (v *ContainerView) FindCellWithName(name string) (*Cell, int) {
	for i, c := range v.Cells {
		if c.View == nil {
			continue
		}
		if c.View.ObjectName() == name {
			return &v.Cells[i], i
		}
	}
	return nil, -1
}

func (v *ContainerView) FindCellWithView(view zview.View) (*Cell, int) {
	for i, c := range v.Cells {
		if c.View == nil {
			continue
		}
		// fmt.Printf("FindCellWithView: %d) %s %p == %p %v\n", i, c.View.ObjectName(), c.View.Native(), view.Native(), c.View.Native() == view.Native())
		// zlog.Info("FindCellWithView:", i, c.View.ObjectName(), reflect.TypeOf(c.View), reflect.TypeOf(view))
		if c.View.Native() == view.Native() {
			return &v.Cells[i], i
		}
	}
	return nil, -1
}

func (v *ContainerView) RemoveChild(subView zview.View, callRemoveFuncs bool) {
	v.DetachChild(subView)
	v.CustomView.RemoveChild(subView, callRemoveFuncs)
}

func (v *ContainerView) RemoveChildrenFunc(remove func(cell Cell) bool) {
	for i := 0; i < len(v.Cells); i++ {
		c := v.Cells[i]
		if remove(c) {
			v.CustomView.RemoveChild(c.View, true)
			zslice.RemoveAt(&v.Cells, i)
			i--
		}
	}
}

func (v *ContainerView) RemoveAllChildren() {
	for _, c := range v.Cells {
		if c.View == nil {
			continue
		}
		v.CustomView.RemoveChild(c.View, true)
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
	// zlog.Info("CV ReplaceChild:", v.Hierarchy())
	c, _ := v.FindCellWithView(child)
	if c == nil {
		zlog.Error("CV ReplaceChild: old not found:", child.Native().Hierarchy(), "in:", v.Hierarchy(), zdebug.CallingStackString())
		for _, c := range v.GetChildren(true) {
			fmt.Printf("Children: %s %p != %p %v\n", c.ObjectName(), c, child, c == child)
		}
		return
	}
	v.NativeView.ReplaceChild(child, with)
	c.View = with
}

// CollapseView collapses/uncollapses a view in its parent which is Collapsable type. (ContainerView)
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
		zlog.Error("ChildView from non-container", v.Native().Hierarchy(), reflect.TypeOf(v))
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

func ArrangeChildrenAtRootContainer(view zview.View, toModalWindowRootOnly bool) {
	root := view.Native().RootParent(toModalWindowRootOnly)
	a, _ := root.View.(Arranger)
	if a != nil {
		a.ArrangeChildren()
		return
	}
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

func DumpHierarchy(view zview.View, add string) {
	s, _ := view.CalculatedSize(zgeo.SizeD(4000, 4000))
	fmt.Println(add+view.ObjectName(), reflect.TypeOf(view), view.Rect().Size, s)
	co, _ := view.(CellsOwner)
	if co == nil {
		return
	}
	add += "  "
	for _, cell := range *co.GetCells() {
		if cell.View == nil {
			fmt.Println(add + "nil")
			continue
		}
		if cell.Collapsed {
			fmt.Print("collapsed: ")
		}
		DumpHierarchy(cell.View, add)
	}
}
