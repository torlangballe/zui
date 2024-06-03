//go:build zui

package zscrollview

import (
	"math"
	"time"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /13/11/15.

type ScrollView struct {
	zcustom.CustomView
	YOffset                float64
	ScrollHandler          func(pos zgeo.Pos, infiniteDir int, delta float64)
	ExpandToChildHightUpTo float64
	lastEdgeScroll         time.Time
	ScrolledAt             time.Time
	ShowBar                bool
	child                  zview.View
	overflow               bool
}

var DefaultBarSize = 17.0 //16.0 // 22?

func init() {
	if zdevice.OS() == zdevice.MacOSType {
		DefaultBarSize = 15 // 15
	}
}

func New() *ScrollView {
	v := &ScrollView{}
	v.Init(v, "scrollview")
	return v
}

func (v *ScrollView) BarSize() float64 {
	if v.overflow && v.ShowBar {
		return DefaultBarSize
	}
	return 0
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
		cust.SetExposed(false)
	}
	ct, _ := v.child.(zcontainer.ChildrenOwner)
	var keepOffsetY *float64
	if ct != nil {
		for _, c := range ct.GetChildren(false) {
			if c.Native().IsPresented() {
				diff := c.Rect().Min().Y - v.YOffset
				if diff >= 0 && keepOffsetY == nil {
					y := c.Rect().Min().Y
					keepOffsetY = &y
				}
			}
		}
	}
	v.ArrangeChildren()
	v.Expose()
}

func (v *ScrollView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := v.MinSize()
	if v.child != nil {
		cs := v.child.CalculatedSize(total)
		if v.ExpandToChildHightUpTo != 0 {
			s.H = math.Min(total.H, math.Max(s.H, cs.H))
			zlog.Info("ScrollView.ExpandToChildHight:", v.Hierarchy(), total.H, s.H)
		}
		s.W = cs.W
		s.W += 16
	}
	return s
}

func (v *ScrollView) SetRect(rect zgeo.Rect) {
	var cs zgeo.Size
	if v.child != nil {
		ls := rect.Size
		ls.H = 20000
		cs = v.child.CalculatedSize(ls)
		// zlog.Info("SV SetRect:", v.Hierarchy(), v.child != nil, "content:", cs.H, rect)
		cs.W = ls.W - 16
	}
	v.SetRectWithChildSize(rect, cs)
}

func (v *ScrollView) SetRectWithChildSize(rect zgeo.Rect, cs zgeo.Size) {
	// zlog.Info("SV.SetRectWithChildSize", v.ObjectName(), rect, cs)
	v.CustomView.SetRect(rect)
	if rect.Size.W < 0 {
		zlog.Info("ScrollStRectW:", v.Hierarchy(), rect, cs, zlog.CallingStackString())
	}
	if v.child != nil {
		v.child.SetRect(zgeo.Rect{Size: cs})
	}
	v.overflow = (cs.H > v.LocalRect().Size.H)
	// zlog.Info("ScrollOverflow:", v.Hierarchy(), rect, cs, v.overflow)
}

func (v *ScrollView) ArrangeChildren() {
	v.SetRect(v.Rect())
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

func (v *ScrollView) ScrollPage(up, animate bool) {
	y := v.YOffset
	window := v.Rect().Size.H - 20
	if up {
		y -= window
		y = math.Max(0, y)
		// zlog.Info("Up", animate, v.YOffset, y, window, v.child.Rect().Size.H)
	} else {
		y += window
		end := math.Max(0, v.child.Rect().Size.H-v.Rect().Size.H)
		y = min(y, end)
	}
	v.SetContentOffset(y, animate)
}

func (v *ScrollView) SetScrollHandler(handler func(pos zgeo.Pos, infiniteDir int, delta float64)) {
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
