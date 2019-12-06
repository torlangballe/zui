// +build !js

package zgo

import "github.com/torlangballe/zutil/zgeo"

func TextViewNew(text string) *TextView {
	tv := &TextView{}
	// tv.Element = DocumentJS.Call("createElement", "INPUT")
	// tv.set("style", "position:absolute")
	// tv.set("type", "text")
	// tv.set("value", text)
	tv.View = tv
	f := FontNice(FontDefaultSize, FontStyleNormal)
	tv.Font(f)
	return tv
}

func (v *TextView) TextAlignment(a zgeo.Alignment) View {
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

func (v *TextView) ChangedHandler(handler func(view View)) {}

func (v *TextView) KeyHandler(handler func(view View, key int)) {}
