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
	style := v.JSStyle()
	style.Set("overflow-x", "hidden")
	style.Set("overflow-y", "auto")
	v.SetCanTabFocus(false)
	style.Set("overscrollBehavior", "contain")
	v.SetMinSize(zgeo.SizeBoth(10))
	v.ShowBar = true
	v.NativeView.SetScrollHandler(func(pos zgeo.Pos) {
		v.YOffset = pos.Y
		if v.ScrollHandler != nil {
			now := time.Now()
			dir := 0
			if time.Since(v.lastEdgeScroll) >= time.Second {
				// zlog.Info("ScrollY:", pos.Y, scrollOutsideDelta, pos.Y < scrollOutsideDelta)
				if pos.Y < scrollOutsideDelta {
					dir = -1
					// zlog.Info("Infin-scroll up:", pos.Y)
					v.lastEdgeScroll = now
				} else if v.child != nil && pos.Y > v.child.Rect().Size.H-v.Rect().Size.H-scrollOutsideDelta {
					dir = 1
					// zlog.Info("Infin-scroll down:", pos.Y, v.child.Rect().Size.H, v.Rect().Size.H, scrollOutsideDelta)
					v.lastEdgeScroll = now
				}
			}
			v.ScrollHandler(pos, dir)
		}
	})
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
