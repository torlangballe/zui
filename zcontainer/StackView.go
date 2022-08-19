//go:build zui

package zcontainer

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
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

func (v *StackView) Init(view zview.View, vertical bool, name string) {
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
	s := zgeo.LayoutGetCellsStackedSize(v.ObjectName(), v.Vertical, v.Spacing(), lays)
	s.MaximizeNonZero(v.MinSize())
	s.Subtract(v.Margin().Size)
	// if v.ObjectName() == "nameValues-stack" {
	// 	zlog.Info("SV CS namevals stack:", v.Hierarchy(), s, len(v.Cells), v.Spacing())
	// }
	return s
}

func (v *StackView) getLayoutCells(rect zgeo.Rect) (lays []zgeo.LayoutCell) {
	// zlog.Info("Layout Stack getCells start", v.ObjectName())
	for _, c := range v.Cells {
		l := c.LayoutCell
		l.OriginalSize = c.View.CalculatedSize(rect.Size)
		// zlog.Info("StackView getLayoutCells:", v.ObjectName(), c.View.ObjectName(), l.OriginalSize)
		l.Name = c.View.ObjectName()
		div, _ := c.View.(*DividerView)
		if div != nil {
			l.Divider = div.Delta
		}
		lays = append(lays, l)
	}
	return
}

func (v *StackView) ArrangeChildren() {
	// zlog.Info("*********** Stack.ArrangeChildren:", v.Hierarchy(), v.Rect(), len(v.Cells))
	// zlog.PushProfile(v.ObjectName())
	layouter, _ := v.View.(zview.Layouter)
	if layouter != nil {
		layouter.HandleBeforeLayout()
	}
	rm := v.LocalRect().Plus(v.Margin())
	lays := v.getLayoutCells(rm)
	rects := zgeo.LayoutCellsInStack(v.ObjectName(), rm, v.Vertical, v.spacing, lays)
	// zlog.ProfileLog("did layout")
	// for i, r := range rects {
	// 	zlog.Info("R:", i, v.cells[i].View.ObjectName(), r)
	// }
	for i, c := range v.Cells {
		r := rects[i]
		// zlog.Info("Stack.ArrangeChild:", c.View.ObjectName(), r.IsNull())
		if !r.IsNull() {
			c.View.SetRect(r)
		}
	}
}

func MakeLinkedStack(surl, name string, add zview.View) *StackView {
	v := StackViewHor("#type:a") // #type is hack that sets html-type of stack element
	v.MakeLink(surl, name)
	v.Add(add, zgeo.TopLeft, zgeo.Size{})
	return v
}
