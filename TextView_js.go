package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

const jsTextMargin = 3

func TextViewNew(text string, style TextViewStyle) *TextView {
	tv := &TextView{}
	stype := "text"
	tv.Element = DocumentJS.Call("createElement", "INPUT")
	tv.Margin = zgeo.SizeBoth(TextViewDefaultMargin)
	tv.set("style", "position:absolute")
	if style.KeyboardType == KeyboardTypePassword {
		stype = "password"
	}
	tv.set("type", stype)
	tv.set("value", text)
	tv.View = tv
	tv.UpdateSecs = 1
	f := FontNice(FontDefaultSize, FontStyleNormal)
	tv.SetFont(f)
	return tv
}

func (v *TextView) SetTextAlignment(a zgeo.Alignment) View {
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

func (v *TextView) SetReadOnly(is bool) *TextView {
	v.set("readonly", is)
	return v
}

func (v *TextView) SetMin(min float64) *TextView {
	v.set("min", min)
	return v
}

func (v *TextView) SetMax(max float64) *TextView {
	v.set("max", max)
	return v
}

func (v *TextView) SetStep(step float64) *TextView {
	v.set("step", step)
	return v
}

func (v *TextView) SetPlaceholder(str string) *TextView {
	v.set("placeholder", str)
	return v
}

func (v *TextView) SetRect(rect zgeo.Rect) View {
	m := v.Margin.Maxed(zgeo.SizeBoth(jsTextMargin))
	rect = rect.Expanded(m.Negative())
	rect.Pos.Y -= 3
	// fmt.Println("TV: Rect:", v.ObjectName(), rect)
	v.NativeView.SetRect(rect)
	return v
}

func (v *TextView) SetText(str string) View {
	v.set("value", str)
	return v
}

func (v *TextView) GetText() string {
	text := v.get("value").String()
	return text
}

func (v *TextView) callUpdate() {
	// zlog.Info("call Text Update")
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
				//				zlog.Info("down-key:", key, v.ObjectName())
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
