package ztext

import (
	"strconv"
	"syscall/js"
	"unicode/utf8"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

// https://www.w3schools.com/jsref/dom_obj_textarea.asp

func (v *TextView) Init(view zview.View, text string, textStyle Style, cols, rows int) {
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
			str := "new-password"
			if textStyle.IsExistingPassword {
				str = "current-password"
			}
			v.JSSet("autocomplete", str)
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
	v.minValue = zfloat.Undefined
	v.maxValue = zfloat.Undefined
	//!! v.SetMargin(zgeo.RectFromXY2(0, 2, -10, -5))
	v.SetObjectName("textview")
	// zlog.Info("TV.Init", v.Hierarchy(), cols, rows)

	v.Columns = cols
	css := v.JSStyle()
	css.Set("position", "absolute")
	css.Set("resize", "none")
	// if zwindow.FromNativeView(&v.NativeView) == zwindow.GetMain() {
	css.Set("boxSizing", "border-box") // this is incredibly important; Otherwise a box outside actual rect is added. But NOT in programatically made windows!!
	// }
	css.Set("-webkitBoxShadow", "none") // doesn't work
	// css.Set("outlineOffset", "-2px")
	// css.Set("overflow", "visible")
	// css.Set("margin-top", "1px")

	if textStyle.DisableAutoComplete {
		v.JSSet("autocomplete", "off")
	}
	// if rows <= 1 {
	// style := v.JSStyle()
	// style.Set("margin", "4px")
	// v.SetMargin(DefaultMargin)
	// }
	if text != "" {
		v.JSSet("value", text)
	}
	// v.JSSet("className", "texter")
	v.View = view
	v.UpdateSecs = 1
	f := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal)
	v.SetFont(f)
	v.JSCall("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) any {
		args[0].Call("stopPropagation")
		return nil
	}))
	if DefaultBGColor().Valid {
		v.SetBGColor(DefaultBGColor())
	}
	v.setHandlers()
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

func (v *TextView) rawSetText(text string) bool {
	if v.Text() != text {
		v.JSSet("value", text)
		return true
	}
	return false
}

func (v *TextView) SetText(text string) {
	if v.FilterFunc != nil {
		text = v.FilterFunc(text)
	}
	if v.rawSetText(text) {
		v.changed.CallAll(false)
	}
	if zkeyvalue.DefaultStore != nil && v.storeKey != "" {
		zkeyvalue.DefaultStore.SetString(v.Text(), v.storeKey, true)
	}

}

func (v *TextView) Text() string {
	text := v.JSGet("value").String()
	// if v.FilterFunc != nil {
	// 	text = v.FilterFunc(text)
	// }
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
	v.changed.CallAll(true)
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

func (v *TextView) SetValueHandler(id string, handler func(edited bool)) {
	v.changed.Add(id, handler)
}

func (v *TextView) setHandlers() {
	v.JSSet("oninput", js.FuncOf(func(this js.Value, args []js.Value) any {
		if v.FilterFunc != nil {
			text := v.Text()
			ftext := v.FilterFunc(text)
			if text != ftext {
				v.rawSetText(ftext)
				v.changed.CallAll(true)
				i := utf8.RuneCountInString(ftext)
				v.Select(i, i)
			}
		}
		if (v.textStyle.KeyboardType == zkeyboard.TypeInteger || v.textStyle.KeyboardType == zkeyboard.TypeFloat) && v.minValue != zfloat.Undefined || v.maxValue != zfloat.Undefined {
			f, err := strconv.ParseFloat(v.Text(), 64)
			zlog.Info("Max?", f, err, v.minValue, v.maxValue)
			n := f
			if err == nil {
				if v.minValue != zfloat.Undefined && n < v.minValue {
					n = v.minValue
					zlog.Info("Min:", f, n)
				} else if v.maxValue != zfloat.Undefined && n > v.maxValue {
					zlog.Info("Max:", f, n)
					n = v.maxValue
				}
			}
			if n != f {
				zlog.Info("n!=f:", f, n)
				if v.textStyle.KeyboardType == zkeyboard.TypeInteger {
					v.SetInt64(int64(n))
				} else {
					v.SetDouble(n)
				}
			}
		}
		if v.UpdateSecs < 0 {
			return nil
		}
		if v.UpdateSecs == 0 {
			// zlog.Info("Edited:", v.ObjectName(), v.Text(), v.changed.Count())
			v.changed.CallAll(true)
		} else {
			v.startUpdate()
		}
		args[0].Call("stopPropagation")
		return nil
	}))
	v.JSSet("onkeydown", js.FuncOf(func(val js.Value, vs []js.Value) any {
		event := vs[0]
		if v.changed.Count() == 0 && v.editDone == nil && v.FilterFunc == nil {
			return nil
		}
		km := zkeyboard.GetKeyModFromEvent(event)
		if km.Key == 0 {
			return nil
		}
		propagate := false
		if km.Key.IsReturnish() || km.Key == zkeyboard.KeyTab {
			propagate = true
			if v.editDone != nil {
				v.editDone(false)
			}
			if v.UpdateSecs != 0 {
				v.updateDone()
			}
		}
		if km.Key == zkeyboard.KeyEscape {
			propagate = true
			if v.editDone != nil {
				v.editDone(true)
			}
		}
		if !propagate {
			event.Call("stopPropagation") // We'll be using the key, so don't allow other listeners to also react to it
		}
		return nil
	}))
	v.JSSet("onkeypress", js.FuncOf(func(val js.Value, vs []js.Value) any {
		if v.FilterFunc == nil {
			return nil
		}
		event := vs[0]
		km := zkeyboard.GetKeyModFromEvent(event)
		if km.Key == 0 {
			return nil
		}
		if km.Modifier&zkeyboard.ModifierMenu != 0 {
			return nil
		}
		r := rune(event.Get("charCode").Int())
		if r < ' ' && r != '\t' || r == 127 {
			return nil
		}
		if v.FilterFunc != nil {
			in := string(r)
			out := v.FilterFunc(in)
			if out != in {
				event.Call("preventDefault")
				if out != "" {
					v.JSSet("value", v.JSGet("value").String()+out)
				}
			}
			return nil
		}
		return nil
	}))
}

func (v *TextView) SetEditDoneHandler(handler func(canceled bool)) {
	v.editDone = handler
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

// func (v *TextView) InsertionPos() int {
// 	old := v.JSGet("type")
// 	v.JSSet("type", "text")
// 	s := v.Element.Get("selectionStart").Int()
// 	v.JSSet("type", old)
// 	return s
// 	// oSel := zdom.DocumentJS.Get("selection").Call("createRange");

// 	// // Move selection start to 0 position
// 	// oSel.Call("moveStart", "character", -oField.value.length);

// 	// // The caret position is selection length
// 	// iCaretPos = oSel.text.length;
// }
