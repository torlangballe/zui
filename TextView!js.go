// +build !js,zui

package zui

import "github.com/torlangballe/zutil/zgeo"

func (tv *TextView) Init(view View, text string, style TextViewStyle, rows, cols int) {
	tv.View = view
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

func (v *TextView) SetChangedHandler(handler func())
func (v *TextView) SetKeyHandler(handler func(key KeyboardKey, mods KeyboardModifier)) {}
func (v *TextView) ScrollToBottom()                                                    {}
func (v *TextView) SetIsStatic(s bool)                                                 {}

func (v *TextView) SetPlaceholder(str string) *TextView {
	return v
}

func (v *TextView) SetMargin(m zgeo.Rect) View {
	return v
}
