package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/ztimer"
)

// Original class created by Tor Langballe on 23-sept-2014.

type ContainerViewCell struct {
	Alignment zgeo.Alignment
	Margin    zgeo.Size
	View      View
	MaxSize   zgeo.Size // MaxSize is maximum size of child-view including margin
	MinSize   zgeo.Size // MinSize is minimum size of child-view including margin
	Collapsed bool
	Free      bool // Free Cells are placed using ContainerView method, not "inherited" ArrangeChildren method
	Weight    float64
}

func CVCell(view View, alignment zgeo.Alignment) *ContainerViewCell {
	cell := ContainerViewCell{Alignment: alignment, View: view}
	return &cell
}

type ContainerView struct {
	CustomView
	margin            zgeo.Rect
	singleOrientation bool
	//	view              *CustomView
	cells         []ContainerViewCell
	layoutHandler ViewLayoutProtocol
}

func (v *ContainerView) GetChildren() (children []View) {
	for _, c := range v.cells {
		children = append(children, c.View)
	}
	return
}

func calculateAddAlignment(def, a zgeo.Alignment) zgeo.Alignment {
	if a&zgeo.VertPos != 0 && def&zgeo.VertPos != 0 {
		def &= ^zgeo.Vertical
	}
	if a&zgeo.HorPos != 0 && def&zgeo.HorPos != 0 {
		def &= ^zgeo.HorPos
	}
	a |= def
	if a&zgeo.VertPos == 0 {
		a |= zgeo.Top
	}
	if a&zgeo.HorPos == 0 {
		a |= zgeo.Left
	}
	return a
}

func (c *ContainerView) Add(defAlignment zgeo.Alignment, elements ...interface{}) {
	var gotView *View
	var gotAlign zgeo.Alignment
	var gotMargin zgeo.Size

	for _, v := range elements {
		if cell, got := v.(ContainerViewCell); got {
			c.AddCell(cell, -1)
			continue
		}
		if view, got := v.(View); got {
			if gotView != nil {
				a := calculateAddAlignment(defAlignment, gotAlign)
				c.AddAdvanced(*gotView, a, gotMargin, zgeo.Size{}, -1, false)
				gotView = nil
				gotAlign = zgeo.AlignmentNone
				gotMargin = zgeo.Size{}
			}
			gotView = &view
			continue
		}
		if a, got := v.(zgeo.Alignment); got {
			gotAlign = a
			continue
		}
		if m, got := v.(zgeo.Size); got {
			gotMargin = m
			continue
		}
	}
	if gotView != nil {
		a := calculateAddAlignment(defAlignment, gotAlign)
		c.AddAdvanced(*gotView, a, gotMargin, zgeo.Size{}, -1, false)
	}
}

func ContainerViewNew(view View, name string) *ContainerView {
	v := &ContainerView{}
	if view == nil {
		view = v
	}
	v.init(view, name)
	return v
}

func (v *ContainerView) Init(view View, name string) {
	v.CustomView.init(view, name)
}

func (v *ContainerView) LayoutHandler(handler ViewLayoutProtocol) *ContainerView {
	v.layoutHandler = handler
	return v
}

func (v *ContainerView) SetMargin(margin zgeo.Rect) *ContainerView {
	v.margin = margin
	return v
}

func (v *ContainerView) SetMarginS(margin zgeo.Size) *ContainerView {
	v.margin = zgeo.RectFromMinMax(margin.Pos(), margin.Pos().Negative())
	return v
}

func (v *ContainerView) SingleOrientation(single bool) *ContainerView {
	v.singleOrientation = single
	return v
}

func (v *ContainerView) AddCell(cell ContainerViewCell, index int) *ContainerViewCell {
	if index == -1 {
		v.cells = append(v.cells, cell)
		v.AddChild(cell.View, -1)
		return &v.cells[len(v.cells)-1]
	} else {
		v.cells = append([]ContainerViewCell{cell}, v.cells...)
		v.AddChild(cell.View, index)
		return &v.cells[index]
	}
}

func (v *ContainerView) AddView(view View, align zgeo.Alignment) *ContainerViewCell {
	return v.AddAdvanced(view, align, zgeo.Size{}, zgeo.Size{}, -1, false)
}

func (v *ContainerView) AddAdvanced(view View, align zgeo.Alignment, marg zgeo.Size, maxSize zgeo.Size, index int, free bool) *ContainerViewCell {
	collapsed := false
	return v.AddCell(ContainerViewCell{align, marg, view, maxSize, zgeo.Size{}, collapsed, free, 0.0}, index)
}

func (v *ContainerView) Contains(view View) bool {
	for _, c := range v.cells {
		if c.View == view {
			return true
		}
	}
	return false
}

func (v *ContainerView) SetRect(rect zgeo.Rect) View {
	v.CustomView.SetRect(rect)
	ct, got := v.View.(ContainerType)
	// fmt.Println("CV: Rect", got)
	if got {
		ct.ArrangeChildren(nil)
	}
	return v
}

func (v *ContainerView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.MinSize()
}

func (v *ContainerView) SetAsFullView(useableArea bool) {
	v.SetRect(ScreenMain().Rect)
	v.SetMinSize(ScreenMain().Rect.Size)
	if !DefinesIsTVBox() {
		h := ScreenStatusBarHeight()
		r := v.Rect()
		if h > 20 && !ScreenHasNotch() {
			r.Size.H -= h
			v.SetRect(r)
		} else if useableArea {
			v.margin.SetMinY(float64(h))
		}
	}
}

func (v *ContainerView) ArrangeChildrenAnimated(onlyChild *View) {
	//        ZAnimation.Do(duration 0.6, animations  { [weak self] () in
	v.ArrangeChildren(onlyChild)
	//        })
}

func (v *ContainerView) arrangeChild(c ContainerViewCell, r zgeo.Rect) {
	ir := r.Expanded(c.Margin.MinusD(2.0))
	s := c.View.CalculatedSize(ir.Size)
	var rv = r.Align(s, c.Alignment, c.Margin, c.MaxSize)
	c.View.SetRect(rv)
}

func ContainerIsLoading(ct ContainerType) bool {
	// fmt.Println("ContainerIsLoading1", len(ct.GetChildren()))
	for _, v := range ct.GetChildren() {
		iowner, got := v.(ImageOwner)
		if got {
			image := iowner.GetImage()
			if image != nil && image.path == "images/buttons/gray@2x.png" {
				// fmt.Println("CV IsLoading:", image.path, image.loading)
			}
			if image != nil && image.loading {
				// fmt.Println("ContainerIsLoading image loading", len(ct.GetChildren()))
				return true
			}
		} else {
			ct, _ := v.(ContainerType)
			// fmt.Println("CV Sub IsLoading:", v.ObjectName(), v.ObjectName(), ct != nil)
			if ct != nil {
				if ContainerIsLoading(ct) {
					// fmt.Println("ContainerIsLoading sub loading", len(ct.GetChildren()))
					return true
				}
			}
		}
	}
	// fmt.Println("ContainerIsLoading Done", len(ct.GetChildren()))
	return false
}

// WhenContainerLoaded waits for all sub-parts images etc to be loaded before calling done.
// done received waited=true if it had to wait
func WhenContainerLoaded(ct ContainerType, done func(waited bool)) {
	if !ContainerIsLoading(ct) {
		if done != nil {
			done(false)
		}
		return
	}
	ztimer.Repeat(0.1, true, true, func() bool {
		if ContainerIsLoading(ct) {
			return true
		}
		if done != nil {
			done(true)
		}
		return false
	})
}

func (v *ContainerView) ArrangeChildren(onlyChild *View) {
	// fmt.Println("CV ArrangeChildren", v.ObjectName())
	if v.layoutHandler != nil {
		v.layoutHandler.HandleBeforeLayout()
	}
	r := zgeo.Rect{Size: v.Rect().Size}.Plus(v.margin)
	for _, c := range v.cells {
		cv, got := c.View.(*ContainerView)
		if got && v.layoutHandler != nil {
			cv.layoutHandler.HandleBeforeLayout()
		}
		if c.Alignment != zgeo.AlignmentNone {
			if onlyChild == nil || c.View == *onlyChild {
				v.arrangeChild(c, r)
			}
			ccv, cgot := c.View.(*ContainerView)
			if cgot {
				ccv.ArrangeChildren(onlyChild)
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

func (v *ContainerView) CollapseChild(view View, collapse bool, arrange bool) bool {
	i := v.FindCellWithView(view)

	changed := (v.cells[i].Collapsed != collapse)
	if changed {
		if collapse {
			//detachFromContainer := false
			v.RemoveChild(v.cells[i].View) //, detachFromContainer)
		} else {
			v.AddChild(v.cells[i].View, -1)
		}
	}
	v.cells[i].Collapsed = collapse
	if arrange {
		v.ArrangeChildren(nil)
	}
	return changed
}

func (v *ContainerView) CollapseChildWithName(name string, collapse bool, arrange bool) bool {
	view := v.FindViewWithName(name, false)
	if view != nil {
		return v.CollapseChild(*view, collapse, arrange)
	}
	return false
}

func ContainerTypeRangeChildren(ct ContainerType, subViews bool, foreach func(view View) bool) {
	for _, c := range ct.GetChildren() {
		// fmt.Println("ContainerViewRangeChildren1:", c.ObjectName(), subViews)
		if !foreach(c) {
			return
		}
	}
	if !subViews {
		return
	}
	for _, c := range ct.GetChildren() {
		sub, got := c.(ContainerType)
		if got {
			ContainerTypeRangeChildren(sub, subViews, foreach)
		}
	}
}

func (v *ContainerView) RemoveNamedChild(name string, all bool) bool {
	for _, c := range v.cells {
		if c.View.ObjectName() == name {
			v.RemoveChild(c.View)
			if !all {
				return true
			}
		}
	}
	return false
}

func (v *ContainerView) FindViewWithName(name string, recursive bool) *View {
	var found *View
	ContainerTypeRangeChildren(v, recursive, func(view View) bool {
		if view.ObjectName() == name {
			found = &view
			return false
		}
		return true
	})
	return found
}

func (v *ContainerView) FindCellWithName(name string) int {
	for i, c := range v.cells {
		if c.View.ObjectName() == name {
			return i
		}
	}
	return -1
}

func (v *ContainerView) FindCellWithView(view View) int {
	for i, c := range v.cells {
		if c.View == view {
			return i
		}
	}
	return -1
}

func (v *ContainerView) RemoveChild(subView View) {
	v.CustomView.RemoveChild(subView)
	v.DetachChild(subView)
}

func (v *ContainerView) RemoveAllChildren() {
	for _, c := range v.cells {
		v.DetachChild(c.View)
		v.CustomView.RemoveChild(c.View)
	}
}

func (v *ContainerView) DetachChild(subView View) {
	for i, c := range v.cells {
		// fmt.Println("detach?:", c.View.ObjectName(), c.View == subView, len(v.cells))
		if c.View == subView {
			zslice.RemoveAt(&v.cells, i)
			// fmt.Println("detach2:", c.View.ObjectName(), len(v.cells))
			break
		}
	}
}

func (v *ContainerView) drawIfExposed() {
	// fmt.Println("CoV drawIf:", v.ObjectName())
	v.CustomView.drawIfExposed()
	for _, c := range v.cells {
		et, got := c.View.(ExposableType)
		// fmt.Println("CoV drawIf:", c.View.ObjectName(), got)
		if got {
			et.drawIfExposed()
		}
	}
}

func (v *ContainerView) ReplaceView(oldView, newView View) {
	i := v.FindCellWithView(oldView)
	if i != -1 {
		c := v.cells[i]
		r := v.Rect()
		newView.SetRect(r)
		v.AddChild(newView, -1)
		v.RemoveChild(c.View)
		c.View = newView
	}
}
