// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /20/10/15.

type StackView struct {
	ContainerView
	spacing  float64
	Vertical bool
}

func StackViewNew(vertical bool, name string) *StackView {
	s := &StackView{}
	s.Init(s, vertical, name)
	return s
}

func (v *StackView) Init(view View, vertical bool, name string) {
	v.ContainerView.Init(view, name)
	v.Vertical = vertical
	v.spacing = 6
}

func StackViewVert(name string) *StackView {
	return StackViewNew(true, name)
}

func StackViewHor(name string) *StackView {
	return StackViewNew(false, name)
}

func (v *StackView) SetSpacing(spacing float64) *StackView {
	v.spacing = spacing
	return v
}

func (v *StackView) Spacing() float64 {
	return v.spacing
}

func (v *StackView) CalculatedSize(total zgeo.Size) zgeo.Size {
	lays := v.getLayoutCells(zgeo.Rect{Size: total})
	s := zgeo.LayoutGetCellsStackedSize(v.Vertical, v.Spacing(), lays)
	s.MaximizeNonZero(v.MinSize())
	s.Subtract(v.Margin().Size)
	// zlog.Info("SV CS:", v.ObjectName(), s)
	return s
}

func (v *StackView) getLayoutCells(rect zgeo.Rect) (lays []zgeo.LayoutCell) {
	// zlog.Info("Layout Stack getCells start", v.ObjectName())
	for _, c := range v.cells {
		l := c.LayoutCell
		l.OriginalSize = c.View.CalculatedSize(rect.Size)
		l.Name = c.View.ObjectName()
		lays = append(lays, l)
	}
	return
}

func (v *StackView) ArrangeChildren(onlyChild *View) {
	// zlog.Info("*********** Stack.ArrangeChildren:", v.ObjectName(), v.Rect(), len(v.cells))
	if v.layoutHandler != nil {
		v.layoutHandler.HandleBeforeLayout()
	}
	zlog.Assert(onlyChild == nil) // going away...
	rm := v.LocalRect().Plus(v.Margin())
	lays := v.getLayoutCells(rm)
	// if v.ObjectName() == "header" {
	// }
	rects := zgeo.LayoutCellsInStack(v.ObjectName(), rm, v.Vertical, v.spacing, lays)
	// for i, r := range rects {
	// 	zlog.Info("R:", i, v.cells[i].View.ObjectName(), r)
	// }
	for i, c := range v.cells {
		r := rects[i]
		// zlog.Info(i, "LAYOUT SETRECT:", r, c.Alignment, c.View.ObjectName())
		if !r.IsNull() {
			c.View.SetRect(r)
		}
	}
	// zlog.Info("*********** Stack.ArrangeChildren Done:", v.ObjectName())
}
