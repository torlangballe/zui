package zui

import (
	"math"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /13/11/15.

type ScrollView struct {
	CustomView
	child   View
	YOffset float64
}

func ScrollViewNew() *ScrollView {
	v := &ScrollView{}
	v.init(v, "scrollview")
	return v
}

func (v *ScrollView) AddChild(child View, index int) {
	v.child = child
	v.CustomView.AddChild(child, index)
}

func (v *ScrollView) GetChildren() []View {
	if v.child != nil {
		return []View{v.child}
	}
	return []View{}
}

func (v *ScrollView) Update() {
	v.exposed = false
	cust, _ := v.child.(*CustomView)
	if cust != nil {
		zlog.Info("SV Update1:", v.ObjectName(), v.Presented, cust.exposed)
		cust.exposed = false
	}
	ct, _ := v.child.(ContainerType)
	if ct != nil {
		for i, c := range ct.GetChildren() {
			// zlog.Info("SV Update1", c.ObjectName(), ViewGetNative(c).Presented)
			if ViewGetNative(c).Presented {
				zlog.Info("SV Update:", i, c.Rect().Min().Y, c.ObjectName())
			}
		}
	}
	v.ArrangeChildren(nil)
	v.Expose()
}

func (v *ScrollView) ArrangeChildren(onlyChild *View) {
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
	return s
}

func (v *ScrollView) SetRect(rect zgeo.Rect) View {
	v.CustomView.SetRect(rect)
	if v.child != nil {
		ls := rect.Size
		ls.H = 20000
		cs := v.child.CalculatedSize(ls)
		cs.W = ls.W
		r := zgeo.Rect{Size: cs}
		v.child.SetRect(r)
	}
	return v
}

func (v *ScrollView) drawIfExposed() {
	zlog.Info("SV:drawIfExposed")
	if v.child != nil {
		ViewGetNative(v.child).Presented = false
		presentViewCallReady(v.child)
	}
	v.CustomView.drawIfExposed()
	if v.child != nil {
		et, got := v.child.(ExposableType)
		if got {
			// zlog.Info("SV:drawIfExposed child")
			et.drawIfExposed()
		}
	}
}

func (v *ScrollView) Expose() {
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
	v.SetContentOffset(h, animate)
}

func (v *ScrollView) ScrollToTop(animate bool) {
	v.SetContentOffset(0, animate)
}

func (v *ScrollView) SetScrollHandler(handler func(pos zgeo.Pos, infiniteDir int)) {
	v.NativeView.SetScrollHandler(func(pos zgeo.Pos) {
		if handler != nil {
			dir := 0
			if pos.Y < 8 {
				dir = -1
			} else if pos.Y > v.child.Rect().Size.H-v.Rect().Size.H+8 {
				dir = 1
			}
			handler(pos, dir)
		}
	})
}
