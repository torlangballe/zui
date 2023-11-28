package zbutton

import (
	"syscall/js"

	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

func New(text string) *Button {
	v := &Button{}
	v.MakeJSElement(v, "button")
	v.SetText(text)
	v.SetObjectName(text)
	v.SetMargin(zgeo.RectFromXY2(12, 4, -12, -4))
	v.SetNativePadding(zgeo.RectFromXY2(-8, -8, 8, 8))
	// style := v.JSStyle()
	// style.Set("textAlign", "left")
	// style.Set("display", "block")
	// style.Set("verticalAlign", "middle")
	// style.Set("whiteSpace", "preWrap")
	f := zgeo.FontNice(zgeo.FontDefaultSize-2, zgeo.FontStyleNormal)
	v.SetFont(f)
	// v.SetKeyDownHandler(func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
	// 	if key.IsReturnish() && mod == zkeyboard.ModifierNone {
	// 		zlog.Info("ButKey:", key, mod)
	// 		if v.PressedHandler != nil {
	// 			zlog.Info("ButKey2:", key, mod)
	// 			v.PressedHandler()
	// 			return true
	// 		}
	// 	}
	// 	return false
	// })
	return v
}

func (v *Button) MakeReturnKeyDefault() {
	v.SetStroke(2, zgeo.ColorNew(0.4, 0.4, 1, 1), true)
	v.SetCorner(6)
	v.SetJSStyle("border", "none")
	// style := v.JSStyle()
	// style.Set("boxSizing", "border-box")

	//	v.margin.Size.H += 4
	ztimer.StartIn(0.01, func() {
		win := zwindow.FromNativeView(&v.NativeView)
		win.AddKeypressHandler(v.View, func(km zkeyboard.KeyMod, down bool) bool {
			// zlog.Info("KeyPress", down, v.RootParent().Hierarchy())
			if !down {
				return false
			}
			top := v.RootParent()
			foc := top.GetFocusedChildView(false)
			if foc != nil {
				tv, _ := foc.(*ztext.TextView)
				if tv != nil {
					if tv.MaxLines() > 1 {
						return false
					}
				} else if foc != v.View {
					return false
				}
			}
			if down && km.Key == zkeyboard.KeyReturn && km.Modifier == zkeyboard.ModifierNone {
				zlog.Info("Default click")
				// v.Element.Call("click")
				v.Press()
				return true
			}
			return false
		})
	})
}

func (v *Button) MakeEscapeCanceler() {
	ztimer.StartIn(0.01, func() {
		win := zwindow.FromNativeView(&v.NativeView)
		win.AddKeypressHandler(v.View, func(km zkeyboard.KeyMod, down bool) bool {
			if down && km.Key == zkeyboard.KeyEscape && km.Modifier == zkeyboard.ModifierNone {
				// v.Element.Call("click")
				v.Press()
				return true
			}
			return false
		})
	})
}

func (v *Button) SetPressedHandler(handler func()) {
	v.pressed = handler
	v.JSSet("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		v.SetStateOnDownPress(e)
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
