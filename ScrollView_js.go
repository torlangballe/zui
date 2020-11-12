package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func (v *ScrollView) Init(view View, name string) {
	v.CustomView.Init(view, name)
	style := v.style()
	style.Set("overflow-x", "hidden")
	//	style.Set("overflow-y", "auto")
	style.Set("overflow-y", "scroll")
	style.Set("overscrollBehavior", "contain")
	v.NativeView.SetScrollHandler(func(pos zgeo.Pos) {
		v.YOffset = pos.Y
		if v.ScrollHandler != nil {
			dir := 0
			if pos.Y < 2 {
				dir = -1
			} else if pos.Y > v.child.Rect().Size.H-v.Rect().Size.H+2 {
				dir = 1
			}
			v.ScrollHandler(pos, dir)
		}
	})
}

func (v *ScrollView) SetContentOffset(y float64, animated bool) {
	if animated {
		zlog.Info("Scroll:", y)
		Animate(v, 2, func(posSecs float64) bool {
			zlog.Info("Animate:", v.YOffset, y, v.YOffset+(y-v.YOffset)*(posSecs/2))
			return true
		})
		return
	}
	v.YOffset = y
	if v.child != nil {
		v.setjs("scrollTop", y)
		// zlog.Info("SetContentOffset:", y)
	}
}
