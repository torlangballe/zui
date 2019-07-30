package zgo

import (
	"fmt"
)

// Created by Tor Langballe on /23/9/14.

type ContainerType interface {
	ArrangeChildren(onlyChild *View)
}

type ContainerViewCell struct {
	Alignment Alignment
	Margin    Size
	View      ViewSimple
	MaxSize   Size
	Collapsed bool
	Free      bool
	//HandleTransition func( size Size,  layout ScreenLayout,  inRect Rect,  alignRect Rect) Rect
}

func CVCell(view ViewSimple, alignment Alignment) *ContainerViewCell {
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

func ContainerViewNew() *ContainerView {
	c := &ContainerView{}
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

func (v *ContainerView) Add(view ViewSimple, align Alignment, marg Size, maxSize Size, index int, free bool) int {
	collapsed := false
	return v.AddCell(ContainerViewCell{align, marg, view, maxSize, collapsed, free}, index)
}

func (v *ContainerView) Contains(view ViewSimple) bool {
	for _, c := range v.cells {
		if c.View == view {
			return true
		}
	}
	return false
}

func (v *ContainerView) CalculateSize(total Size) Size {
	return v.MinSize
}

func (v *ContainerView) SetAsFullView(useableArea bool) {
	layout := true
	zViewSetRect(v.GetView(), ScreenMainRect(), layout)
	v.MinSize = ScreenMainRect().Size
	if !DefinesIsTVBox() {
		h := ScreenStatusBarHeight()
		r := v.GetView().GetRect()
		if h > 20 && !ScreenHasNotch() {
			r.Size.H -= h
			zViewSetRect(v.GetView(), r, layout)
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
	s := zConvertViewSizeThatFitstToSize(c.View.GetView(), ir.Size)
	var rv = r.Align(s, c.Alignment, c.Margin, c.MaxSize)
	// if c.handleTransition != nil {
	//     if let r = c.handleTransition(s, Screen.Orientation(), r, rv) {
	//         rv = r
	//     }
	// }
	zViewSetRect(c.View.GetView(), rv, false)
}

func (v *ContainerView) ArrangeChildren(onlyChild *View) {
	fmt.Println("ContainerView ArrangeChildren")
	v.layoutHandler.HandleBeforeLayout()
	r := RectFromSize(v.GetView().GetRect().Size).Plus(v.margin)
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

func (v *ContainerView) CollapseChild(view ViewSimple, collapse bool, arrange bool) bool {
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

func (v *ContainerView) RangeChildren(subViews bool, foreach func(v ViewSimple) bool) {
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

func (v *ContainerView) FindViewWithName(name string, recursive bool) *ViewSimple {
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

func (v *ContainerView) FindCellWithView(view ViewSimple) int {
	for i, c := range v.cells {
		if c.View == view {
			return i
		}
	}
	return -1
}

func (v *ContainerView) RemoveChild(subView ViewSimple) {
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

func (v *ContainerView) DetachChild(subView ViewSimple) {
	for i, c := range v.cells {
		if c.View == subView {
			UtilRemoveAt(v.cells, i)
			break
		}
	}
}
