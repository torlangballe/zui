package zgo

import (
	"fmt"
	"time"
)

// Created by Tor Langballe on /23/9/14.

type ContainerType interface {
	ArrangeChildren(onlyChild *View)
}

type ContainerViewCell struct {
	Alignment Alignment
	Margin    Size
	View      View
	MaxSize   Size
	Collapsed bool
	Free      bool
	//HandleTransition func( size Size,  layout ScreenLayout,  inRect Rect,  alignRect Rect) Rect
}

func CVCell(view View, alignment Alignment) *ContainerViewCell {
	cell := ContainerViewCell{Alignment: alignment, View: view}
	return &cell
}

type ContainerView struct {
	CustomView
	margin            Rect
	singleOrientation bool
	//	view              *CustomView
	cells         []ContainerViewCell
	layoutHandler ViewLayoutProtocol
}

func Container(elements ...interface{}) *ContainerView {
	c := ContainerViewNew(nil)
	c.SetElements(AlignmentNone, elements...)
	return c
}

func calculateAddAlignment(def, a Alignment) Alignment {
	if a&AlignmentVertical != 0 && def&AlignmentVertical != 0 {
		def &= ^AlignmentVertical
	}
	if a&AlignmentHorizontal != 0 && def&AlignmentHorizontal != 0 {
		def &= ^AlignmentHorizontal
	}
	a |= def
	if a&AlignmentVertical == 0 {
		a |= AlignmentTop
	}
	if a&AlignmentHorizontal == 0 {
		a |= AlignmentLeft
	}
	return a
}

func (c *ContainerView) SetElements(defAlignment Alignment, elements ...interface{}) {
	var gotView *View
	var gotAlign Alignment
	var gotMargin Size

	for _, v := range elements {
		if cell, got := v.(ContainerViewCell); got {
			c.AddCell(cell, -1)
			continue
		}
		if view, got := v.(View); got {
			if gotView != nil {
				a := calculateAddAlignment(defAlignment, gotAlign)
				c.Add(*gotView, a, gotMargin, Size{}, -1, false)
				gotView = nil
				gotAlign = AlignmentNone
				gotMargin = Size{}
			}
			gotView = &view
			continue
		}
		if a, got := v.(Alignment); got {
			gotAlign = a
			continue
		}
		if m, got := v.(Size); got {
			gotMargin = m
			continue
		}
	}
	if gotView != nil {
		a := calculateAddAlignment(defAlignment, gotAlign)
		c.Add(*gotView, a, gotMargin, Size{}, -1, false)
	}
}

func ContainerViewNew(view View) *ContainerView {
	c := &ContainerView{}
	if view == nil {
		view = c
	}
	c.CustomView.Init(view)
	return c
}

func (v *ContainerView) LayoutHandler(handler ViewLayoutProtocol) *ContainerView {
	v.layoutHandler = handler
	return v
}

func (v *ContainerView) Margin(margin Rect) *ContainerView {
	v.margin = margin
	return v
}

func (v *ContainerView) MarginS(margin Size) *ContainerView {
	v.margin = RectFromMinMax(margin.Pos(), margin.Pos().Negative())
	return v
}

func (v *ContainerView) SingleOrientation(single bool) *ContainerView {
	v.singleOrientation = single
	return v
}

func (v *ContainerView) AddCell(cell ContainerViewCell, index int) int {
	if index == -1 {
		v.cells = append(v.cells, cell)
		zViewAddView(v, cell.View, -1)
		return len(v.cells) - 1
	} else {
		v.cells = append([]ContainerViewCell{cell}, v.cells...)
		zViewAddView(v, cell.View, index)
		return index
	}
}

func (v *ContainerView) Add(view View, align Alignment, marg Size, maxSize Size, index int, free bool) int {
	collapsed := false
	return v.AddCell(ContainerViewCell{align, marg, view, maxSize, collapsed, free}, index)
}

func (v *ContainerView) Contains(view View) bool {
	for _, c := range v.cells {
		if c.View == view {
			return true
		}
	}
	return false
}

func (v *ContainerView) GetCalculatedSize(total Size) Size {
	return v.GetMinSize()
}

func (v *ContainerView) SetAsFullView(useableArea bool) {
	layout := true
	zViewSetRect(v, ScreenMain().Rect, layout)
	v.MinSize(ScreenMain().Rect.Size)
	if !DefinesIsTVBox() {
		h := ScreenStatusBarHeight()
		r := v.GetView().GetRect()
		if h > 20 && !ScreenHasNotch() {
			r.Size.H -= h
			zViewSetRect(v, r, layout)
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

func (v *ContainerView) arrangeChild(c ContainerViewCell, r Rect) {
	ir := r.Expanded(c.Margin.MinusD(2.0))
	// s := zConvertViewSizeThatFitstToSize(c.View.GetView(), ir.Size)
	s := c.View.GetCalculatedSize(ir.Size)
	var rv = r.Align(s, c.Alignment, c.Margin, c.MaxSize)
	// if c.handleTransition != nil {
	//     if let r = c.handleTransition(s, Screen.Orientation(), r, rv) {
	//         rv = r
	//     }
	// }
	zViewSetRect(c.View, rv, false)
}

func (v *ContainerView) isLoading() bool {
	for _, c := range v.cells {
		io, got := c.View.(ImageOwner)
		if got {
			image := io.GetImage()
			fmt.Println("IO:", c.View.GetObjectName(), io.GetImage(), c.View.GetCalculatedSize(v.GetLocalRect().Size))
			if image != nil && image.loading {
				return true
			}
		}
	}
	return false
}

func (v *ContainerView) ArrangeChildren(onlyChild *View) {
	fmt.Println("ContainerView ArrangeChildren")
	for v.isLoading() {
		time.Sleep(time.Millisecond * 10)
	}
	v.layoutHandler.HandleBeforeLayout()
	r := Rect{Size: v.GetView().GetRect().Size}.Plus(v.margin)
	for _, c := range v.cells {
		cv, got := c.View.(*ContainerView)
		if got {
			cv.layoutHandler.HandleBeforeLayout()
		}
		if c.Alignment != AlignmentNone {
			if onlyChild == nil || c.View == *onlyChild {
				v.arrangeChild(c, r)
			}
			ccv, cgot := c.View.(*ContainerView)
			if cgot {
				ccv.ArrangeChildren(onlyChild)
			}
		}
	}
	v.layoutHandler.HandleAfterLayout()
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
			detachFromContainer := false
			zRemoveViewFromSuper(v.cells[i].View.GetView(), detachFromContainer)
		} else {
			zViewAddView(v.cells[i].View, v, -1)
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

func (v *ContainerView) RangeChildren(subViews bool, foreach func(v View) bool) {
	for _, c := range v.cells {
		if !foreach(c.View) {
			return
		}
		if subViews {
			cv, got := c.View.(*ContainerView)
			if got {
				cv.RangeChildren(subViews, foreach)
			}
		}
	}
}

func (v *ContainerView) RemoveNamedChild(name string, all bool) bool {
	for _, c := range v.cells {
		if c.View.GetObjectName() == name {
			v.RemoveChild(c.View)
			if !all {
				return true
			}
		}
	}
	return false
}

func (v *ContainerView) FindViewWithName(name string, recursive bool) *View {
	for _, c := range v.cells {
		if c.View.GetObjectName() == name {
			return &c.View
		}
		if recursive {
			cv, got := c.View.(*ContainerView)
			if got {
				vn := cv.FindViewWithName(name, true)
				if vn != nil {
					return vn
				}
			}
		}
	}

	return nil
}

func (v *ContainerView) FindCellWithName(name string) int {
	for i, c := range v.cells {
		if c.View.GetObjectName() == name {
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
	detachFromContainer := false
	zRemoveViewFromSuper(subView.GetView(), detachFromContainer)
	v.DetachChild(subView)
}

func (v *ContainerView) RemoveAllChildren() {
	for _, c := range v.cells {
		v.DetachChild(c.View)
		detachFromContainer := false
		zRemoveViewFromSuper(c.View.GetView(), detachFromContainer)
	}
}

func (v *ContainerView) DetachChild(subView View) {
	for i, c := range v.cells {
		if c.View == subView {
			UtilRemoveAt(v.cells, i)
			break
		}
	}
}

func (v *ContainerView) drawAllIfExposed() {
	v.drawIfExposed()
	for _, c := range v.cells {
		cov, _ := c.View.(*ContainerView)
		if cov != nil {
			cov.drawAllIfExposed()
		} else {
			cvp, _ := c.View.(CustomViewProtocol)
			if cvp != nil {
				fmt.Println("drawAllIfExposed c:", c.View.GetObjectName())
				cvp.drawIfExposed()
			}
		}
	}
}
