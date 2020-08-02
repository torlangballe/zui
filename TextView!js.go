// +build !js

package zui

import "github.com/torlangballe/zutil/zgeo"

func (tv *TextView) Init(text string, style TextViewStyle, maxLines int) {
	tv.View = tv
	f := FontNice(FontDefaultSize, FontStyleNormal)
	tv.SetFont(f)
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

func (v *TextView) SetChangedHandler(handler func(view View))                                     {}
func (v *TextView) SetKeyHandler(handler func(view View, key KeyboardKey, mods KeyboardModifier)) {}

func (v *TextView) SetPlaceholder(str string) *TextView {
	return v
}

func (v *TextView) ScrollToBottom() {}
