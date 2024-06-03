//go:build zui

package zscrollview

import (
	"time"

	"github.com/torlangballe/zui/zanimation"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

var scrollOutsideDelta = 2.0

func (v *ScrollView) Init(view zview.View, name string) {
	v.CustomView.Init(view, name)
	v.ShowScrollBars(false, true)
	v.SetCanTabFocus(false)
	v.JSStyle().Set("overscrollBehavior", "contain")
	v.SetMinSize(zgeo.SizeBoth(10))
	v.ShowBar = true
	v.NativeView.SetScrollHandler(func(pos zgeo.Pos) {
		delta := pos.Y - v.YOffset
		v.YOffset = pos.Y
		if v.ScrollHandler != nil {
			now := time.Now()
			dir := 0
			if time.Since(v.lastEdgeScroll) >= time.Second {
				if pos.Y < -scrollOutsideDelta {
					dir = -1
					v.lastEdgeScroll = now
				} else if v.child != nil && pos.Y > v.OffsetAtBottom()+scrollOutsideDelta {
					dir = 1
					v.lastEdgeScroll = now
				}
			}
			v.ScrollHandler(pos, dir, delta)
		}
	})
}

func (v *ScrollView) OffsetAtBottom() float64 {
	return v.child.Rect().Size.H - v.Rect().Size.H
}

func (v *ScrollView) SetContentOffset(y float64, animated bool) {
	v.ScrolledAt = time.Now()
	if animated {
		zanimation.Animate(v, 0.5, func(t float64) bool {
			ay := v.YOffset + (y-v.YOffset)*t
			// zlog.Info("Animate:", t, v.YOffset, y, ay)
			v.SetContentOffset(ay, false)
			return true
		})
		return
	}
	v.YOffset = y
	if v.child != nil {
		v.JSSet("scrollTop", y)
	}
}
