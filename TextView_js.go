package zgo

import (
	"fmt"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

const jsTextMargin = 3

func TextViewNew(text string) *TextView {
	tv := &TextView{}
	tv.Element = DocumentJS.Call("createElement", "INPUT")
	tv.Margin = zgeo.SizeBoth(TextViewDefaultMargin)
	tv.set("style", "position:absolute")
	tv.set("type", "text")
	tv.set("value", text)
	tv.View = tv
	tv.UpdateSecs = 1
	f := FontNice(FontDefaultSize, FontStyleNormal)
	tv.Font(f)
	return tv
}

func (v *TextView) TextAlignment(a zgeo.Alignment) View {
	v.alignment = a
	str := "left"
	if a&zgeo.AlignmentRight != 0 {
		str = "right"
	} else if a&zgeo.AlignmentHorCenter != 0 {
		str = "center"
	}
	v.style().Set("textAlign", str)
	return v
}

func (v *TextView) IsReadOnly(is bool) *TextView {
	v.set("readOnly", is)
	return v
}

func (v *TextView) Placeholder(str string) *TextView {
	v.set("placeholder", str)
	return v
}

func (v *TextView) IsPassword(is bool) *TextView {
	v.set("type", "password")
	return v
}

func (v *TextView) Rect(rect zgeo.Rect) View {
	m := v.Margin.Maxed(zgeo.SizeBoth(jsTextMargin))
	rect = rect.Expanded(m.Negative())
	rect.Pos.Y -= 3
	// fmt.Println("TV: Rect:", v.GetObjectName(), rect)
	v.NativeView.Rect(rect)
	return v
}

func (v *TextView) GetText() string {
	text := v.get("value").String()
	return text
}

func (v *TextView) ChangedHandler(handler func(view View)) {
	v.changed = handler
	if handler != nil {
		v.set("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
			if v.changed != nil {
				if v.UpdateSecs == 0 {
					v.changed(v)
					return nil
				}
				if v.updateTimer != nil {
					v.updateTimer.Stop()
				}
				v.updateTimer = ztimer.StartIn(v.UpdateSecs, true, func() {
					v.changed(v)
				})
			}
			return nil
		}))
	}
}

func (v *TextView) KeyHandler(handler func(view View, key int)) {
	v.keyPressed = handler
	if handler != nil {
		v.set("onkeyup", js.FuncOf(func(js.Value, []js.Value) interface{} {
			fmt.Println("KeyUp!")
			return nil
		}))
	}
}
