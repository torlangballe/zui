//go:build zui
// +build zui

package zui

import (
	"sort"

	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztimer"
)

// Original class created by Tor Langballe on 23-sept-2014.

var GroupingStrokeColor = zgeo.ColorNewGray(0.7, 1)
var GroupingStrokeWidth = 2.0
var GroupingStrokeCorner = 4.0
var GroupingMargin = 10.0
var AlertButtonsOnRight = true

// func CVCell(view View, alignment zgeo.Alignment) *ContainerViewCell {
// 	cell := ContainerViewCell{Alignment: alignment, View: view}
// 	return &cell
// }

type ContainerView struct {
	CustomView
	margin            zgeo.Rect
	singleOrientation bool
	cells             []ContainerViewCell
	layoutHandler     ViewLayoutProtocol
}

type ContainerViewCell struct {
	zgeo.LayoutCell
	View View
}

type ContainerType interface {
	GetChildren(includeCollapsed bool) []View
	ArrangeChildren()
	ReplaceChild(child, with View)
}

type Collapser interface {
	CollapseChild(view View, collapse bool, arrange bool) bool
}

func (v *ContainerView) GetChildren(includeCollapsed bool) (children []View) {
	for _, c := range v.cells {
		if includeCollapsed || !c.Collapsed {
			children = append(children, c.View)
		}
	}
	return
}

func ArrangeParentContainer(view View) {
	parent := ViewGetNative(view).Parent().View.(ContainerType)
	parent.ArrangeChildren()
}

func (v *ContainerView) CountChildren() int {
	return len(v.cells)
}

func (v *ContainerView) Add(elements ...interface{}) (first *ContainerViewCell) {
	var gotView View
	var gotAlign zgeo.Alignment
	var gotMargin zgeo.Size
	var gotIndex = -1

	if len(v.cells) == 1200 {
		zlog.Info("CV ADD1:", v.ObjectName(), zlog.GetCallingStackString())
	}
	for _, e := range elements {
		if cell, got := e.(ContainerViewCell); got {
			cell := v.AddCell(cell, -1)
			if first == nil {
				first = cell
			}
			continue
		}
		if view, got := e.(View); got {
			if gotView != nil {
				// zlog.Info("CV ADD got:", gotView.ObjectName())
				cell := v.AddAdvanced(gotView, gotAlign, gotMargin, zgeo.Size{}, gotIndex, false)
				if first == nil {
					first = cell
				}
				gotAlign = zgeo.AlignmentNone
				gotMargin = zgeo.Size{}
				gotIndex = -1
			}
			gotView = view
			continue
		}
		if n, got := e.(int); got {
			gotIndex = n
		}
		if a, got := e.(zgeo.Alignment); got {
			gotAlign = a
			continue
		}
		if m, got := e.(zgeo.Size); got {
			gotMargin = m
			continue
		}
	}
	if gotView != nil {
		// zlog.Info("CV ADD got end:", gotView.ObjectName())
		cell := v.AddAdvanced(gotView, gotAlign, gotMargin, zgeo.Size{}, gotIndex, false)
		if first == nil {
			first = cell
		}
	}
	return
}

func (v *ContainerView) AddAlertButton(button View) {
	a := zgeo.VertCenter
	if AlertButtonsOnRight {
		a |= zgeo.Right
	} else {
		a |= zgeo.Left
	}
	v.Add(button, a)
}

func ContainerViewNew(name string) *ContainerView {
	v := &ContainerView{}
	v.Init(v, name)
	return v
}

func (v *ContainerView) Init(view View, name string) {
	v.CustomView.Init(view, name)
}

func (v *ContainerView) LayoutHandler(handler ViewLayoutProtocol) {
	v.layoutHandler = handler
}

func (v *ContainerView) SetMargin(margin zgeo.Rect) {
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

func (v *ContainerView) AddCell(cell ContainerViewCell, index int) (cvs *ContainerViewCell) {
	if index == -1 {
		v.cells = append(v.cells, cell)
		v.AddChild(cell.View, -1)
		cvs = &v.cells[len(v.cells)-1]
	} else {
		zlog.Assert(index == 0)
		v.cells = append([]ContainerViewCell{cell}, v.cells...)
		v.AddChild(cell.View, index)
		cvs = &v.cells[index]
	}
	// zlog.Info("CV ADD CELL:", cell.View.ObjectName(), v.Presented)
	return cvs
}

func (v *ContainerView) AddView(view View, align zgeo.Alignment) *ContainerViewCell {
	return v.AddAdvanced(view, align, zgeo.Size{}, zgeo.Size{}, -1, false)
}

func (v *ContainerView) AddAdvanced(view View, align zgeo.Alignment, marg zgeo.Size, maxSize zgeo.Size, index int, free bool) *ContainerViewCell {
	collapsed := false
	// zlog.Info("CV AddAdvancedView:", v.ObjectName(), view.ObjectName())
	lc := zgeo.LayoutCell{align, marg, maxSize, zgeo.Size{}, collapsed, free, false, zgeo.Size{}, 0.0, view.ObjectName()}
	return v.AddCell(ContainerViewCell{LayoutCell: lc, View: view}, index)
}

func (v *ContainerView) Contains(view View) bool {
	for _, c := range v.cells {
		if c.View == view {
			return true
		}
	}
	return false
}

func (v *ContainerView) SetRect(rect zgeo.Rect) {
	// zlog.Info("CV SetRect2", v.ObjectName(), rect)
	v.CustomView.SetRect(rect)
	ct := v.View.(ContainerType) // in case we are a stack or something inheriting from ContainerView
	//	start := time.Now()
	ct.ArrangeChildren()
	// d := time.Since(start)
	// if d > time.Millisecond*50 {
	// 	zlog.Info("CV SetRect2", v.ObjectName(), d)
	// }
}

func (v *ContainerView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.MinSize()
}

func (v *ContainerView) SetAsFullView(useableArea bool) {
	sm := zscreen.GetMain()
	r := sm.Rect
	if useableArea {
		r = sm.UsableRect
	}
	v.SetRect(r)
	v.SetMinSize(r.Size)
}

func (v *ContainerView) ArrangeChildrenAnimated() {
	//        ZAnimation.Do(duration 0.6, animations  { [weak self] () in
	v.ArrangeChildren()
	//        })
}

func (v *ContainerView) ArrangeChild(c ContainerViewCell, r zgeo.Rect) {
	if c.Alignment != zgeo.AlignmentNone {
		ir := r.Expanded(c.Margin.MinusD(2.0))
		s := c.View.CalculatedSize(ir.Size)
		var rv = r.AlignPro(s, c.Alignment, c.Margin, c.MaxSize, zgeo.Size{})
		c.View.SetRect(rv)
		// ViewGetNative(c.View).Presented = true
	}
}

func ContainerIsLoading(ct ContainerType) bool {
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
			ct, _ := v.(ContainerType)
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
func WhenContainerLoaded(ct ContainerType, done func(waited bool)) {
	// start := time.Now()
	ztimer.RepeatNow(0.1, func() bool {
		// zlog.Info("WhenContainerLoaded", ct.(View).ObjectName())
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
	if v.layoutHandler != nil {
		v.layoutHandler.HandleBeforeLayout()
	}
	r := zgeo.Rect{Size: v.Rect().Size}.Plus(v.margin)
	for _, c := range v.cells {
		cv, got := c.View.(*ContainerView)
		if got && v.layoutHandler != nil {
			cv.layoutHandler.HandleBeforeLayout()
		}
		// zlog.Info("ArrangeAdvanced:", v.ObjectName(), c.View.ObjectName(), c.Free, freeOnly)
		if c.Alignment != zgeo.AlignmentNone && (!freeOnly || c.Free) {
			v.ArrangeChild(c, r)
			// zlog.Info("ArrangeAdvanced inside:", v.ObjectName(), c.View.ObjectName(), c.Free, freeOnly, c.View.Rect())
			ct, _ := c.View.(ContainerType) // we might be "inherited" by StackView or something
			if ct != nil {
				ct.ArrangeChildren()
			}
		}
	}
	if v.layoutHandler != nil {
		v.layoutHandler.HandleAfterLayout()
	}
	for _, c := range v.cells {
		cv, got := c.View.(*ContainerView)
		if got {
			cv.layoutHandler.HandleAfterLayout()
		}
	}
}

func (v *ContainerView) CollapseChild(view View, collapse bool, arrange bool) (changed bool) {
	cell, _ := v.FindCellWithView(view)
	if cell == nil {
		return false
	}
	changed = (cell.Collapsed != collapse)
	// if cell.View.ObjectName() == "xxx" {
	// 	zlog.Info("COLLAPSE:", collapse, changed, view.ObjectName(), cell.View.ObjectName())
	// }
	if collapse {
		cell.View.Show(false)
	}
	nc, _ := view.(*CustomView)

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
				nc.visible = true
			}
			v.AddChild(cell.View, -1)
		}
	}
	if arrange && v.Presented {
		ct := v.View.(ContainerType) // we might be "inherited" by StackView or something
		ct.ArrangeChildren()
	}
	if !collapse {
		cell.View.Show(true)
		// cv, _ := cell.View.(*ShapeView)
		// if cell.View.ObjectName() == "xxx" {
		// 	zlog.Info("Uncollapse Disco", cv != nil)
		// }
		// if cv != nil && cv.Presented {
		// 	cv.visible = true
		// }
		ExposeView(cell.View)
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

func ViewRangeChildren(view View, subViews, includeCollapsed bool, foreach func(view View) bool) {
	ct, _ := view.(ContainerType)
	if ct != nil {
		ContainerTypeRangeChildren(ct, subViews, includeCollapsed, foreach)
	}
}

func ContainerTypeRangeChildren(ct ContainerType, subViews, includeCollapsed bool, foreach func(view View) bool) {
	children := ct.GetChildren(includeCollapsed)
	for _, c := range children {
		// zlog.Info("ContainerViewRangeChildren1:", c.ObjectName(), subViews)
		if !foreach(c) {
			return
		}
	}
	if !subViews {
		return
	}
	for _, c := range children {
		sub, got := c.(ContainerType)
		if got {
			ContainerTypeRangeChildren(sub, subViews, includeCollapsed, foreach)
		}
	}
}

func (v *ContainerView) RemoveNamedChild(name string, all bool) bool {
	for {
		removed := false
		for _, c := range v.cells {
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

func (v *ContainerView) FindViewWithName(name string, recursive bool) (View, int) {
	return ContainerTypeFindViewWithName(v, name, recursive)
}

func ContainerTypeFindViewWithName(ct ContainerType, name string, recursive bool) (View, int) {
	var found View

	i := 0
	includeCollapsed := true
	ContainerTypeRangeChildren(ct, recursive, includeCollapsed, func(view View) bool {
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

func (v *ContainerView) FindCellWithName(name string) (*ContainerViewCell, int) {
	for i, c := range v.cells {
		if c.View.ObjectName() == name {
			return &v.cells[i], i
		}
	}
	return nil, -1
}

func (v *ContainerView) FindCellWithView(view View) (*ContainerViewCell, int) {
	for i, c := range v.cells {
		if c.View == view {
			return &v.cells[i], i
		}
	}
	return nil, -1
}

func (v *ContainerView) RemoveChild(subView View) {
	v.DetachChild(subView)
	v.CustomView.RemoveChild(subView)
}

func (v *ContainerView) RemoveAllChildren() {
	for _, c := range v.cells {
		v.CustomView.RemoveChild(c.View)
	}
	v.cells = v.cells[:0]
}

func (v *ContainerView) DetachChild(subView View) {
	for i, c := range v.cells {
		//zlog.Info("detach?:", ViewGetNative(c.View).Element, c.View == subView, len(v.cells))
		if c.View == subView {
			zslice.RemoveAt(&v.cells, i)
			// zlog.Info("detach2:", c.View.ObjectName(), len(v.cells))
			break
		}
	}
}

/*
func (v *ContainerView) drawIfExposed() {
	v.CustomView.drawIfExposed()
	// zlog.Info("CoV drawIfExp:", v.Hierarchy())
	for _, c := range v.cells {
		if !c.Collapsed {
			et, got := c.View.(ExposableType)
			if got {
				et.drawIfExposed()
			}
		}
	}
	v.exposed = false
}
*/

func (v *ContainerView) ReplaceChild(child, with View) {
	c, _ := v.FindCellWithView(child)
	if c != nil {
		v.ReplaceChild(child, with)
		c.View = with
	}
}

// CollapseView collapses/uncollapses a view in it's parent which is Collapsable type. (ContainerView)
func CollapseView(v View, collapse, arrange bool) bool {
	p := ViewGetNative(v).Parent()
	c := p.View.(Collapser) // crash if parent isn't ContainerView of some sort
	return c.CollapseChild(v, collapse, arrange)
}

// SortChildren re-arranges order of cells with views in them.
// It does not curr
func (v *ContainerView) SortChildren(less func(a, b View) bool) {
	sort.Slice(v.cells, func(i, j int) bool {
		return less(v.cells[i].View, v.cells[j].View)
	})
}
