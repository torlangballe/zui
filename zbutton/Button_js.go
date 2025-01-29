package zbutton

import (
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
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

	// style := v.JSStyle()
	// style.Set("boxSizing", "border-box")

	//	v.margin.Size.H += 4
	ztimer.StartIn(0.01, func() {
		win := zwindow.FromNativeView(&v.NativeView)
		win.AddKeyPressHandler(v.View, zkeyboard.KeyMod{Key: zkeyboard.KeyReturn}, true, func() {
			// zlog.Info("KeyPress", down, v.RootParent().Hierarchy())
			toModalWindowOnly := true
			top := v.RootParent(toModalWindowOnly)
			foc := top.GetFocusedChildView(false)
			if foc != nil {
				tv, _ := foc.(*ztext.TextView)
				if tv != nil {
					if tv.MaxLines() > 1 {
						return
					}
				} else if foc != v.View {
					return
				}
			}
			v.Click()
		})
	})
}

func (v *Button) MakeEscapeCanceler() {
	ztimer.StartIn(0.01, func() {
		win := zwindow.FromNativeView(&v.NativeView)
		win.AddKeyPressHandler(v.View, zkeyboard.KeyMod{Key: zkeyboard.KeyEscape}, true, func() {
			v.Click()
		})
	})
}

func (v *Button) SetMargin(m zgeo.Rect) {
	v.margin = m
}
