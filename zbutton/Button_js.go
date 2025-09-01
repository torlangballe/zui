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
	// v.SetMargin(zgeo.RectFromXY2(3, 3, -3, -3))
	v.SetMargin(zgeo.RectFromXY2(12, 4, -12, -8))
	// v.SetNativePadding(zgeo.RectFromXY2(-8, -8, 8, 8))
	// style := v.JSStyle()
	// style.Set("textAlign", "left")
	// style.Set("display", "block")
	// style.Set("verticalAlign", "middle")
	// style.Set("whiteSpace", "preWrap")
	f := zgeo.FontNice(zgeo.FontDefaultSize-2, zgeo.FontStyleNormal)
	v.SetFont(f)
	v.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if down && km.Key.IsReturnish() && km.Modifier == zkeyboard.ModifierNone {
			v.ClickAll()
		}
		return false
	})
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
		f := func() bool {
			toModalWindowOnly := true
			top := v.RootParent(toModalWindowOnly)
			foc := top.GetFocusedChildView(false)
			if foc != nil {
				tv, _ := foc.(*ztext.TextView)
				if tv != nil {
					if tv.MaxLines() > 1 {
						return false
					}
				} else if foc != v.View {
					// return false
				}
			}
			v.ClickAll()
			return true
		}
		win.AddKeyPressHandler(v.View, zkeyboard.KeyMod{Key: zkeyboard.KeyReturn}, true, f)
		win.AddKeyPressHandler(v.View, zkeyboard.KeyMod{Key: zkeyboard.KeyEnter}, true, f)
	})
}

func (v *Button) MakeEscapeCanceler() {
	ztimer.StartIn(0.01, func() {
		win := zwindow.FromNativeView(&v.NativeView)
		win.AddKeyPressHandler(v.View, zkeyboard.KeyMod{Key: zkeyboard.KeyEscape}, true, func() bool {
			v.ClickAll()
			return true
		})
	})
}

func (v *Button) SetMargin(m zgeo.Rect) {
	v.margin = m
}

func (v *Button) SetRect(r zgeo.Rect) {
	rm := r.Plus(zgeo.RectFromMarginSize(zgeo.SizeBoth(3)))
	v.NativeView.SetRect(rm)
}

func (v *Button) Text() string {
	return v.NativeView.InnerText()
}

func (v *Button) SetText(str string) {
	v.NativeView.SetInnerText(str)
}
