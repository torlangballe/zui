package zui

import (
	"syscall/js"

	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

func ButtonNew(text string) *Button {
	v := &Button{}
	v.MakeJSElement(v, "button")
	v.SetText(text)
	v.SetObjectName(text)
	v.SetMargin(zgeo.RectFromXY2(12, 4, -12, -4))
	// style := v.style()
	// style.Set("textAlign", "left")
	// style.Set("display", "block")
	// style.Set("verticalAlign", "middle")
	// style.Set("whiteSpace", "preWrap")
	f := zgeo.FontNice(zgeo.FontDefaultSize-2, zgeo.FontStyleNormal)
	v.SetFont(f)

	return v
}

func (v *Button) MakeEnterDefault() {
	v.SetStroke(1, zgeo.ColorNew(0.3, 0.3, 1, 1))
	v.SetCorner(6)
	//	v.margin.Size.H += 4
	ztimer.StartIn(0.01, func() {
		win := v.GetWindow()
		win.AddKeypressHandler(v.View, func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
			if key == zkeyboard.KeyReturn && mod == zkeyboard.ModifierNone {
				v.Element.Call("click")
				return true
			}
			return false
		})
	})
}

func (v *Button) MakeEscapeCanceler() {
	ztimer.StartIn(0.01, func() {
		win := v.GetWindow()
		win.AddKeypressHandler(v.View, func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
			if key == zkeyboard.KeyEscape && mod == zkeyboard.ModifierNone {
				v.Element.Call("click")
				return true
			}
			return false
		})
	})

}

func (v *Button) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.JSSet("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnClick(v)
		return nil
	}))
	v.JSSet("className", "widget")
}

func (v *Button) SetLongPressedHandler(handler func()) {
	// zlog.Info("Button.SetLongPressedHandler:", v.ObjectName())
	v.longPressed = handler
	v.JSSet("className", "widget")
	v.JSSet("onmousedown", js.FuncOf(func(js.Value, []js.Value) interface{} {
		(&v.LongPresser).HandleOnMouseDown(v)
		return nil
	}))
	v.JSSet("onmouseup", js.FuncOf(func(js.Value, []js.Value) interface{} {
		// fmt.Println("MOUSEUP")
		(&v.LongPresser).HandleOnMouseUp(v)
		return nil
	}))
}

func (v *Button) SetMargin(m zgeo.Rect) {
	v.margin = m
}
