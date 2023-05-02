package ztext

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

// https://www.w3schools.com/jsref/dom_obj_textarea.asp

func (v *TextView) Init(view zview.View, text string, textStyle Style, rows, cols int) {
	v.textStyle = textStyle
	v.SetMaxLines(rows)
	if rows > 1 {
		v.Element = zdom.DocumentJS.Call("createElement", "textarea")
	} else {
		v.Element = zdom.DocumentJS.Call("createElement", "input")
		stype := "text"
		switch textStyle.KeyboardType {
		case zkeyboard.TypePassword:
			stype = "password"
		case zkeyboard.TypeEmailAddress:
			stype = "email"
		case zkeyboard.TypeFloat:
			v.JSSet("pattern", `[0-9]*`)
			stype = "number"
			v.JSSet("inputmode", "numeric")
			v.JSSet("digitOnly", "true")
		case zkeyboard.TypeInteger:
			stype = "number"
			v.JSSet("pattern", `[0-9]*`)
		}
		if textStyle.Type == Search {
			stype = "search"
		}
		if textStyle.Type == Date {
			stype = "date"
		}
		// zlog.Info("TextView:", v.Hierarchy(), stype)
		v.JSSet("type", stype)
	}
	v.SetMargin(zgeo.RectFromXY2(0, 3, -10, -12))
	v.SetObjectName("textview")
	v.SetAboveParent(true)
	v.Columns = cols
	css := v.JSStyle()
	css.Set("position", "absolute")
	css.Set("resize", "none")
	// css.Set("boxSizing", "border-box")  // this is incredibly important; Otherwise a box outside actual rect is added. But NOT in programatically made windows!!
	css.Set("-webkitBoxShadow", "none") // doesn't work
	// css.Set("outlineOffset", "-2px")
	css.Set("overflow", "visible")
	// css.Set("margin-top", "1px")

	if textStyle.DisableAutoComplete {
		v.JSSet("autocomplete", "off")
	}
	// if rows <= 1 {
	// style := v.JSStyle()
	// style.Set("margin", "4px")
	// v.SetMargin(DefaultMargin)
	// }
	v.JSSet("value", text)
	// v.JSSet("className", "texter")
	v.View = view
	v.UpdateSecs = 1
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	v.SetFont(f)
	v.JSCall("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		args[0].Call("stopPropagation")
		return nil
	}))
	// v.JSSet("onkeyup", js.FuncOf(func(val js.Value, args []js.Value) interface{} {
	// 	args[0].Call("stopPropagation")
	// 	zlog.Info("key in textview:", args)
	// 	return nil
	// }))
	if DefaultBGColor().Valid {
		v.SetBGColor(DefaultBGColor())
	}
	if DefaultBorderColor().Valid {
		v.SetStroke(1, DefaultBorderColor(), true)
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
	v.JSSet("readOnly", s)
}

func (v *TextView) SetTextAlignment(a zgeo.Alignment) {
	v.alignment = a
	str := "left"
	if a&zgeo.Right != 0 {
		str = "right"
	} else if a&zgeo.HorCenter != 0 {
		str = "center"
	}
	v.JSStyle().Set("textAlign", str)
}

func (v *TextView) SetReadOnly(is bool) {
	v.JSSet("readonly", is)
}

func (v *TextView) SetMin(min float64) {
	v.JSSet("min", min)
}

func (v *TextView) SetMax(max float64) {
	v.JSSet("max", max)
}

func (v *TextView) SetStep(step float64) {
	v.JSSet("step", step)
}

func (v *TextView) SetPlaceholder(str string) {
	v.JSSet("placeholder", str)
}

func (v *TextView) SetRect(rect zgeo.Rect) {
	rect.Add(v.margin)
	// rect.Pos.Y -= 1
	// zlog.Info("TV SetRect", v.ObjectName(), rect, v.Text())
	// rect = rect.Plus(zgeo.RectFromXY2(2, 4, -10, -10))
	v.NativeView.SetRect(rect)
}

func (v *TextView) SetMargin(m zgeo.Rect) {
	v.margin = m
	v.NativeView.SetNativeMargin(m)
}

func (v *TextView) SetText(text string) {
	if v.Text() != text {
		v.JSSet("value", text)
	}
}

func (v *TextView) Text() string {
	text := v.JSGet("value").String()
	return text
}

func (v *TextView) BGColor() zgeo.Color {
	str := v.JSStyle().Get("backgroundColor").String()
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
			col = DefaultBGColor()
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
		col := DefaultBGColor().Mixed(zgeo.ColorNew(1, 0.3, 0.3, 1), 0.3)
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
		v.JSSet("oninput", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
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
			args[0].Call("stopPropagation")
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
		v.JSSet("onkeydown", js.FuncOf(func(val js.Value, vs []js.Value) interface{} {
			event := vs[0]
			w := event.Get("which")
			if w.IsUndefined() {
				return nil
			}
			key := w.Int()
			if key == zkeyboard.KeyReturn || key == zkeyboard.KeyTab {
				if v.editDone != nil {
					v.editDone(false)
				}
				if v.UpdateSecs != 0 { //  && v.updated
					//				zlog.Info("down-key:", key, v.ObjectName())
					//					zlog.Info("push:", v.ContinuousUpdateCalls, v.updated)
					v.updateDone()
				}
			}
			if key == zkeyboard.KeyEscape {
				if v.editDone != nil {
					v.editDone(true)
				}
			}
			return nil
		}))
	}
}

func (v *TextView) SetKeyHandler(handler func(km zkeyboard.KeyMod, down bool) bool) {
	// v.keyPressed = handler
	v.NativeView.SetKeyHandler(handler)
}

func (v *TextView) ScrollToBottom() {
	v.JSSet("scrollTop", v.JSGet("scrollHeight"))
}

func (v *TextView) ConsumesKey(sc zkeyboard.KeyMod) bool {
	if sc.Modifier != zkeyboard.ModifierNone {
		return false
	}
	switch sc.Key {
	case zkeyboard.KeyEscape:
		return false
	case zkeyboard.KeyReturn, zkeyboard.KeyEnter, zkeyboard.KeyUpArrow, zkeyboard.KeyDownArrow:
		return v.Element.Get("type").String() == "textarea"
	}
	return true
}
