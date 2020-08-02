package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

const jsTextMargin = 3

// https://www.w3schools.com/jsref/dom_obj_textarea.asp

func (v *TextView) Init(text string, style TextViewStyle, maxLines int) {
	v.SetMaxLines(maxLines)
	if maxLines > 1 {
		v.Element = DocumentJS.Call("createElement", "textarea")
	} else {
		v.Element = DocumentJS.Call("createElement", "input")
		v.set("type", "text")
		if style.KeyboardType == KeyboardTypePassword {
			v.set("type", "password")
		} else {
			v.set("type", "text")
		}
	}
	v.Margin = zgeo.SizeBoth(TextViewDefaultMargin)
	v.set("style", "position:absolute")
	v.set("value", text)
	v.View = v
	v.UpdateSecs = 1
	f := FontNice(FontDefaultSize, FontStyleNormal)
	v.SetFont(f)
}

func (v *TextView) SetStatic(s bool) {
	v.set("readOnly", s)
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
	// zlog.Info("TV: Rect:", v.ObjectName(), rect)
	v.NativeView.SetRect(rect)
	return v
}

func (v *TextView) SetText(str string) View {
	if v.Text() != str {
		v.set("value", str)
	}
	return v
}

func (v *TextView) Text() string {
	text := v.get("value").String()
	return text
}

func (v *TextView) BGColor() zgeo.Color {
	str := v.style().Get("backgroundColor").String()
	if str == "" || str == "initial" { // hack since it is white but has no value, so it isn't "initall", which is transparent
		return zgeo.ColorWhite
	}
	c := zgeo.ColorFromString(str)
	return c
}

func (v *TextView) updateDone() {
	if v.updateTimer != nil {
		v.updateTimer.Stop()
	}
	if v.UpdateSecs > 1 {
		v.SetBGColor(v.pushedBGColor)
		v.pushedBGColor = zgeo.Color{}
	}
	v.changed(v)
}

func (v *TextView) startUpdate() {
	if v.UpdateSecs > 1 && !v.pushedBGColor.Valid {
		v.pushedBGColor = v.BGColor()
		v.SetBGColor(zgeo.ColorNew(1, 0.9, 0.9, 1))
	}
	if v.updateTimer != nil {
		v.updateTimer.Stop()
	}
	// zlog.Info("call Text Update", v.UpdateSecs)
	v.updateTimer = ztimer.StartIn(v.UpdateSecs, func() {
		v.updateDone()
	})
	//	v.updated = false
}

func (v *TextView) SetChangedHandler(handler func(view View)) {
	v.changed = handler
	if handler != nil {
		v.set("onkeydown", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
			if v.UpdateSecs != 0 { //  && v.updated
				event := vs[0]
				key := event.Get("which").Int()
				//				zlog.Info("down-key:", key, v.ObjectName())
				if key == KeyboardKeyReturn || key == KeyboardKeyTab {
					//					zlog.Info("push:", v.ContinuousUpdateCalls, v.updated)
					v.updateDone()
				}
			}
			return nil
		}))
		v.set("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
			// v.updated = true
			if v.UpdateSecs == 0 {
				v.changed(v)
			} else {
				v.startUpdate()
			}
			return nil
		}))
	}
}

func (v *TextView) SetKeyHandler(handler func(view View, key KeyboardKey, mods KeyboardModifier)) {
	v.keyPressed = handler
	v.set("onkeyup", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		if handler != nil {
			event := vs[0]
			// key := event.Get("which").Int()
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
			// zlog.Info("KeyUp:", key, mods)
		}
		return nil
	}))
}

func (v *TextView) ScrollToBottom() {
	v.set("scrollTop", v.get("scrollHeight"))
}
