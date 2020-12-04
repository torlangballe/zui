package zui

import (
	"fmt"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

const jsTextMargin = 2

// https://www.w3schools.com/jsref/dom_obj_textarea.asp

func (v *TextView) Init(text string, style TextViewStyle, rows, cols int) {
	v.SetMaxLines(rows)
	if rows > 1 {
		v.Element = DocumentJS.Call("createElement", "textarea")
	} else {
		v.Element = DocumentJS.Call("createElement", "input")

		if style.KeyboardType == KeyboardTypePassword {
			v.setjs("type", "password")
		} else if style.IsSearch {
			v.setjs("type", "search")
		} else {
			v.setjs("type", "text")
		}
	}

	v.Columns = cols
	v.setjs("style", "position:absolute")
	v.style().Set("boxSizing", "border-box")  // this is incredibly important; Otherwise a box outside actual rect is added. But NOT in programatically made windows!!
	v.style().Set("-webkitBoxShadow", "none") // doesn't work
	if rows <= 1 {
		v.SetMargin(zgeo.RectFromXY2(TextViewDefaultMargin, TextViewDefaultMargin, -TextViewDefaultMargin, -TextViewDefaultMargin))
	}
	v.setjs("value", text)
	v.View = v
	v.UpdateSecs = 1
	f := FontNice(FontDefaultSize, FontStyleNormal)
	v.SetFont(f)

	v.setjs("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("stopPropagation")
		//		zlog.Info("onclick textview:", args)
		// event.stopPropagation();
		return nil
	}))
}

func (v *TextView) SetStatic(s bool) {
	v.setjs("readOnly", s)
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
	v.setjs("readonly", is)
	return v
}

func (v *TextView) SetMin(min float64) *TextView {
	v.setjs("min", min)
	return v
}

func (v *TextView) SetMax(max float64) *TextView {
	v.setjs("max", max)
	return v
}

func (v *TextView) SetStep(step float64) *TextView {
	v.setjs("step", step)
	return v
}

func (v *TextView) SetPlaceholder(str string) *TextView {
	v.setjs("placeholder", str)
	return v
}

func (v *TextView) SetMargin(m zgeo.Rect) View {
	v.margin = m
	// zlog.Info("TextView. SetMargin:", v.ObjectName(), m)
	style := v.style()
	style.Set("paddingLeft", fmt.Sprintf("%dpx", int(m.Min().X)))
	style.Set("paddingRight", fmt.Sprintf("%dpx", int(m.Max().X)))
	style.Set("paddingTop", fmt.Sprintf("%dpx", -int(m.Min().Y)))
	style.Set("paddingBottom", fmt.Sprintf("%dpx", -int(m.Max().Y)))
	return v
}

// func (v *TextView) SetRect(rect zgeo.Rect) View {
// 	// m := v.margin.Maxed(zgeo.SizeBoth(jsTextMargin))
// 	// rect = rect.Expanded(m.Negative())
// 	if v.MaxLines() <= 1 {
// 		rect = rect.Expanded(zgeo.Size{-6, -4})
// 		rect.Pos.Y -= 3
// 	}
// 	zlog.Info("TextView. SetRect:", v.ObjectName(), rect.Size)
// 	v.NativeView.SetRect(rect)
// 	return v
// }

func (v *TextView) SetText(str string) View {
	if v.Text() != str {
		v.setjs("value", str)
	}
	return v
}

func (v *TextView) Text() string {
	text := v.getjs("value").String()
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
		v.updateTimer = nil
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
	// zlog.Info("call Text Update", v.UpdateSecs)
	if v.updateTimer != nil {
		v.updateTimer.Stop()
	}
	v.updateTimer = ztimer.StartIn(v.UpdateSecs, func() {
		v.updateDone()
	})
	//	v.updated = false
}

func (v *TextView) SetChangedHandler(handler func(view View)) {
	v.changed = handler
	if handler != nil {
		v.setjs("onkeydown", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
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
		v.setjs("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
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
	v.setjs("onkeyup", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
		// zlog.Info("KeyUp")
		if handler != nil {
			event := vs[0]
			key := KeyboardKey(event.Get("which").Int())
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
			handler(v, key, mods)
			// zlog.Info("KeyUp:", key, mods)
		}
		return nil
	}))
}

func (v *TextView) ScrollToBottom() {
	v.setjs("scrollTop", v.getjs("scrollHeight"))
}
