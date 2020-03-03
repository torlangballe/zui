// +build !js

package zui

import "github.com/torlangballe/zutil/zgeo"

func TextViewNew(text string, style TextViewStyle) *TextView {
	tv := &TextView{}
	// tv.Element = DocumentJS.Call("createElement", "INPUT")
	// tv.set("style", "position:absolute")
	// tv.set("type", "text")
	// tv.set("value", text)
	tv.View = tv
	f := FontNice(FontDefaultSize, FontStyleNormal)
	tv.SetFont(f)
	return tv
}

func (v *TextView) SetTextAlignment(a zgeo.Alignment) View {
	v.alignment = a
	return v
}

func (v *TextView) IsReadOnly(is bool) *TextView {
	return v
}

func (v *TextView) Placeholder(str string) *TextView {
	return v
}

func (v *TextView) IsPassword(is bool) *TextView {
	return v
}

func (v *TextView) ChangedHandler(handler func(view View))                                     {}
func (v *TextView) KeyHandler(handler func(view View, key KeyboardKey, mods KeyboardModifier)) {}