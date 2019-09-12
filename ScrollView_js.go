package zgo

import (
	"syscall/js"
)

func (v *ScrollView) init(view View, name string) {
	v.CustomView.init(view, name)
	style := v.style()
	style.Set("overflow-x", "hidden")
	style.Set("overflow-y", "auto")
	v.set("onscroll", js.FuncOf(func(js.Value, []js.Value) interface{} {
		y := v.get("scrollTop").Float()
		if v.HandleScroll != nil {
			v.HandleScroll(Pos{0, y})
		}
		return nil
	}))
}
