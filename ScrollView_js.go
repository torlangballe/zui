package zui

import "github.com/torlangballe/zutil/zlog"

func (v *ScrollView) init(view View, name string) {
	v.CustomView.init(view, name)
	style := v.style()
	style.Set("overflow-x", "hidden")
	//	style.Set("overflow-y", "auto")
	style.Set("overflow-y", "scroll")
	style.Set("overscrollBehavior", "contain")
}

func (v *ScrollView) SetContentOffset(y float64, animated bool) {
	if animated {
		zlog.Info("Scroll:", y)
		Animate(2, func(posSecs float64) bool {
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
