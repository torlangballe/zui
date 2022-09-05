//go:build zui

package zscrollview

import (
	"math"
	"time"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /13/11/15.

type ScrollView struct {
	zcustom.CustomView
	child         zview.View
	YOffset       float64
	ScrollHandler func(pos zgeo.Pos, infiniteDir int)

	lastEdgeScroll time.Time
	ScrolledAt     time.Time
}

func New() *ScrollView {
	v := &ScrollView{}
	v.Init(v, "scrollview")
	return v
}

func (v *ScrollView) AddChild(child zview.View, index int) {
	v.child = child
	v.CustomView.AddChild(child, index)
}

func (v *ScrollView) GetChildren(includeCollapsed bool) []zview.View {
	if v.child != nil {
		return []zview.View{v.child}
	}
	return []zview.View{}
}

func (v *ScrollView) Update() {
	v.SetExposed(false)
	cust, _ := v.child.(*zcustom.CustomView)
	if cust != nil {
		// zlog.Info("SV Update1:", v.ObjectName(), v.Presented, cust.exposed)
		cust.SetExposed(false)
	}
	ct, _ := v.child.(zcontainer.ContainerType)
	var keepOffsetY *float64
	if ct != nil {
		for _, c := range ct.GetChildren(false) {
			if c.Native().Presented {
				diff := c.Rect().Min().Y - v.YOffset
				if diff >= 0 && keepOffsetY == nil {
					y := c.Rect().Min().Y
					keepOffsetY = &y
				}
				// if diff > 0 {
				// 	zlog.Info("SV Update:", i, diff, c.ObjectName())
				// }
			}
		}
	}
	v.ArrangeChildren()
	// if keepOffsetY != nil {
	// 	zlog.Info("SV Update KeepOffset:", *keepOffsetY)
	// 	// v.SetContentOffset(*keepOffsetY, false)
	// }
	v.Expose()
}

func (v *ScrollView) ArrangeChildren() {
	if v.child != nil {
		ls := v.Rect().Size
		ls.H = 20000
		cs := v.child.CalculatedSize(ls)
		cs.W = ls.W
		r := zgeo.Rect{Size: cs}
		v.child.SetRect(r) // this will call arrange children on child if container
		// ct, got := v.child.(ContainerType)
		// if got {
		// 	ct.ArrangeChildren(onlyChild)
		// }
	}
}

func (v *ScrollView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.MinSize()
	if v.child != nil {
		cs := v.child.CalculatedSize(total)
		s.W = cs.W
	}
	// zlog.Info("SV CalculatedSize:", v.ObjectName(), s, v.child != nil)
	return s
}

func (v *ScrollView) SetRect(rect zgeo.Rect) {
	v.CustomView.SetRect(rect)
	if v.child != nil {
		ls := rect.Size
		ls.H = 20000
		cs := v.child.CalculatedSize(ls)
		cs.W = ls.W
		r := zgeo.Rect{Size: cs}
		v.child.SetRect(r)
	}
}

func (v *ScrollView) Expose() {
	v.CustomView.Expose()
	zview.ExposeView(v.child)
}

func (v *ScrollView) ScrollToBottom(animate bool) {
	h := v.child.Rect().Size.H
	h -= v.Rect().Size.H
	h = math.Max(0, h)
	v.SetContentOffset(h, animate)
}

func (v *ScrollView) ScrollToTop(animate bool) {
	v.SetContentOffset(0, animate)
}

func (v *ScrollView) SetScrollHandler(handler func(pos zgeo.Pos, infiniteDir int)) {
	v.ScrollHandler = handler
}

func (v *ScrollView) MakeRectVisible(rect zgeo.Rect, animate bool) {
	var y float64
	h := v.LocalRect().Size.H
	if rect.Min().Y < v.YOffset {
		y = rect.Min().Y
	} else if rect.Max().Y > v.YOffset+h {
		y = rect.Max().Y - h
	} else {
		return
	}
	v.SetContentOffset(y, false)
}

func (v *ScrollView) MakeOffspringVisible(offspring zview.View, animate bool) {
	sp := v.AbsoluteRect().Pos
	or := offspring.Native().AbsoluteRect()
	or.Pos.Subtract(sp)
	or.Pos.Y += v.YOffset
	v.MakeRectVisible(or, animate)
}