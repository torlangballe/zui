package zui

import (
	"fmt"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

const jsTextMargin = 2

// https://www.w3schools.com/jsref/dom_obj_textarea.asp

func (v *TextView) Init(view View, text string, textStyle TextViewStyle, rows, cols int) {
	v.textStyle = textStyle
	v.SetMaxLines(rows)
	if rows > 1 {
		v.Element = DocumentJS.Call("createElement", "textarea")
	} else {
		v.Element = DocumentJS.Call("createElement", "input")
		stype := "text"
		switch textStyle.KeyboardType {
		case KeyboardTypePassword:
			stype = "password"
		case KeyboardTypeEmailAddress:
			stype = "email"
		}
		if textStyle.Type == TextViewSearch {
			stype = "search"
		}
		if textStyle.Type == TextViewDate {
			stype = "date"
		}
		v.setjs("type", stype)
	}

	v.SetObjectName("textview")
	v.Columns = cols
	v.setjs("style", "position:absolute")
	css := v.style()
	css.Set("boxSizing", "border-box")  // this is incredibly important; Otherwise a box outside actual rect is added. But NOT in programatically made windows!!
	css.Set("-webkitBoxShadow", "none") // doesn't work

	// if rows <= 1 {
	v.SetMargin(TextViewDefaultMargin)
	// }
	v.setjs("value", text)
	v.setjs("className", "texter")
	v.View = view
	v.UpdateSecs = 1
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	v.SetFont(f)

	v.setjs("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("stopPropagation")
		//		zlog.Info("onclick textview:", args)
		// event.stopPropagation();
		return nil
	}))
	if TextViewDefaultColor().Valid {
		v.SetColor(TextViewDefaultColor())
	}
	if TextViewDefaultBGColor().Valid {
		v.SetBGColor(TextViewDefaultBGColor())
	}
	if TextViewDefaultBorderColor().Valid {
		v.SetStroke(1, TextViewDefaultBorderColor())
	}
}

func (v *TextView) Select(from, to int) {
	if to == -1 {
		to = len(v.Text())
	}
	v.Element.Call("setSelectionRange", from, to)
}

func (v *TextView) SetBGColor(c zgeo.Color) {
	v.NativeView.SetBGColor(c)
}

func (v *TextView) SetIsStatic(s bool) {
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
	style := v.style()
	str := fmt.Sprintf("%dpx", int(m.Min().X))
	style.Set("paddingLeft", str)
	style.Set("-webkit-padding-start", str)
	style.Set("paddingRight", fmt.Sprintf("%dpx", int(m.Max().X)))
	style.Set("paddingTop", fmt.Sprintf("%dpx", -int(m.Min().Y)))
	style.Set("paddingBottom", fmt.Sprintf("%dpx", -int(m.Max().Y)))
	return v
}

func (v *TextView) SetText(text string) {
	if v.Text() != text {
		v.setjs("value", text)
	}
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
		col := v.pushedBGColor
		if !col.Valid {
			col = TextViewDefaultBGColor()
		}
		v.SetBGColor(col)
		v.pushedBGColor = zgeo.Color{}
	}
	if v.changed != nil {
		v.changed()
	}
}

func (v *TextView) startUpdate() {
	if v.UpdateSecs > 1 && !v.pushedBGColor.Valid {
		v.pushedBGColor = v.BGColor()
		col := TextViewDefaultBGColor().Mixed(zgeo.ColorNew(1, 0.3, 0.3, 1), 0.3)
		v.SetBGColor(col)
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

func (v *TextView) SetChangedHandler(handler func()) {
	v.changed = handler
	if handler != nil {
		v.updateEnterHandlers()
		v.setjs("oninput", js.FuncOf(func(js.Value, []js.Value) interface{} {
			// v.updated = true
			if v.UpdateSecs < 0 {
				return nil
			}
			if v.UpdateSecs == 0 {
				if v.changed != nil {
					v.changed()
				}
			} else {
				v.startUpdate()
			}
			return nil
		}))
	}
}

func (v *TextView) SetEditDoneHandler(handler func(canceled bool)) {
	v.editDone = handler
	v.updateEnterHandlers()
}

func (v *TextView) updateEnterHandlers() {
	if v.changed != nil || v.editDone != nil {
		v.setjs("onkeydown", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
			event := vs[0]
			key := event.Get("which").Int()
			if key == KeyboardKeyReturn || key == KeyboardKeyTab {
				if v.editDone != nil {
					v.editDone(false)
				}
				if v.UpdateSecs != 0 { //  && v.updated
					//				zlog.Info("down-key:", key, v.ObjectName())
					//					zlog.Info("push:", v.ContinuousUpdateCalls, v.updated)
					v.updateDone()
				}
			}
			if key == KeyboardKeyEscape {
				if v.editDone != nil {
					v.editDone(true)
				}
			}
			return nil
		}))
	}
}

// TODO: Replace with NativeView version
func (v *TextView) SetKeyHandler(handler func(key KeyboardKey, mods KeyboardModifier) bool) {
	v.keyPressed = handler
	v.NativeView.SetKeyHandler(handler)
	/*
		v.setjs("onkeyup", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
			// zlog.Info("KeyUp")
			if handler != nil {
				event := vs[0]
				key, mods := getKeyAndModsFromEvent(event) // replaces below code?
				handler(key, mods)
				// zlog.Info("KeyUp:", key, mods)
			}
			return nil
		}))
	*/
}

func (v *TextView) ScrollToBottom() {
	v.setjs("scrollTop", v.getjs("scrollHeight"))
}
