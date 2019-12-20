package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
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
	tv.SetFont(f)
	return tv
}

func (v *TextView) TextAlignment(a zgeo.Alignment) View {
	v.alignment = a
	str := "left"
	if a&zgeo.Right != 0 {
		str = "right"
	} else if a&zgeo.HorCenter != 0 {
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

func (v *TextView) callUpdate() {
	zlog.Info("callUpdate")
	if v.changed != nil {
		if v.UpdateSecs == 0 {
			v.changed(v)
		}
		if v.updateTimer != nil {
			v.updateTimer.Stop()
		}
		v.updateTimer = ztimer.StartIn(v.UpdateSecs, true, func() {
			v.changed(v)
		})
	}
	v.updated = false
}

func (v *TextView) ChangedHandler(handler func(view View)) {
	v.changed = handler
	if handler != nil {
		v.set("onkeydown", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
			if !v.ContinuousUpdateCalls && v.updated {
				event := vs[0]
				key := event.Get("which").Int()
				//				zlog.Info("down-key:", key, v.GetObjectName())
				if key == KeyboardKeyReturn || key == KeyboardKeyTab {
					//					zlog.Info("push:", v.ContinuousUpdateCalls, v.updated)
					v.callUpdate()
				}
			}
			return nil
		}))
		v.set("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
			v.updated = true
			//			zlog.Info("UPDATED", v.updated)
			if v.ContinuousUpdateCalls {
				v.callUpdate()
			}
			return nil
		}))
	}
}

func (v *TextView) KeyHandler(handler func(view View, key KeyboardKey, mods KeyboardModifier)) {
	v.keyPressed = handler
	v.set("onkeyup", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		if handler != nil {
			event := vs[0]
			key := event.Get("which").Int()
			var mods KeyboardModifier
			if event.Get("altKey").Bool() {
				mods |= KeyboardModifierAlt
			}
			if event.Get("ctrlKey").Bool() {
				mods |= KeyboardModifierControl
			}
			if event.Get("metaKey").Bool() {
				mods |= KeyboardModifierMeta
			}
			if event.Get("shiftKey").Bool() {
				mods |= KeyboardModifierShift
			}
			zlog.Dummy("KeyUp:", key, mods)
		}
		return nil
	}))
}
