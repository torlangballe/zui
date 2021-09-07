// +build zui

package zui

import (
	"math"
	"time"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /13/11/15.

type ScrollView struct {
	CustomView
	child         View
	YOffset       float64
	ScrollHandler func(pos zgeo.Pos, infiniteDir int)

	lastEdgeScroll time.Time
	ScrolledAt     time.Time
}

func ScrollViewNew() *ScrollView {
	v := &ScrollView{}
	v.Init(v, "scrollview")
	return v
}

func (v *ScrollView) AddChild(child View, index int) {
	v.child = child
	v.CustomView.AddChild(child, index)
}

func (v *ScrollView) GetChildren(includeCollapsed bool) []View {
	if v.child != nil {
		return []View{v.child}
	}
	return []View{}
}

func (v *ScrollView) Update() {
	v.exposed = false
	cust, _ := v.child.(*CustomView)
	if cust != nil {
		// zlog.Info("SV Update1:", v.ObjectName(), v.Presented, cust.exposed)
		cust.exposed = false
	}
	ct, _ := v.child.(ContainerType)
	var keepOffsetY *float64
	if ct != nil {
		for _, c := range ct.GetChildren(false) {
			// zlog.Info("SV Update1", c.ObjectName(), ViewGetNative(c).Presented)
			if ViewGetNative(c).Presented {
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
		// zlog.Info("SV Arrange:", r)
		v.child.SetRect(r) // this will call arrange children on child if container
		// ct, got := v.child.(ContainerType)
		// if got {
		// 	ct.ArrangeChildren(onlyChild)
		// }
	}
}

func (v *ScrollView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.minSize
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

func (v *ScrollView) drawIfExposed() {
	// zlog.Info("SV:drawIfExposed")
	if v.child != nil {
		//ViewGetNative(v.child).Presented = false
		PresentViewCallReady(v.child, true)
	}
	v.CustomView.drawIfExposed()
	if v.child != nil {
		et, got := v.child.(ExposableType)
		if got {
			// zlog.Info("SV:drawIfExposed child")
			et.drawIfExposed()
		}
		PresentViewCallReady(v.child, false)
	}
}

func (v *ScrollView) Expose() {
	zlog.Info("SV:Expose")
	v.CustomView.Expose()
	et, _ := v.child.(ExposableType)
	if et != nil {
		// zlog.Info("SV:Expose child!")
		et.Expose()
	}
}

func (v *ScrollView) ScrollToBottom(animate bool) {
	h := v.child.Rect().Size.H
	h -= v.Rect().Size.H
	h = math.Max(0, h)
	// zlog.Info("Scroll2Bottom:", h)
	v.SetContentOffset(h, animate)
}

func (v *ScrollView) ScrollToTop(animate bool) {
	v.SetContentOffset(0, animate)
}

func (v *ScrollView) SetScrollHandler(handler func(pos zgeo.Pos, infiniteDir int)) {
	v.ScrollHandler = handler
}
