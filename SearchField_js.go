package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

type SearchField struct {
	TextView *TextView
	StackView
}

func SearchFieldNew(style TextViewStyle, chars int) *SearchField {
	s := &SearchField{}
	s.StackView.Init(s, false, "search-stack")
	style.IsSearch = true
	t := TextViewNew("", style, chars, 1)
	s.TextView = t
	// size := t.CalculatedSize(zgeo.Size{300, 50})
	// t.SetCorner(size.H / 2)
	t.SetObjectName("search")
	t.SetToolTip("Type search string, or hh:mm dd-mm")
	// t.SetMargin(zgeo.RectFromXY2(24, 0, 0, 0))
	t.setjs("inputmode", "search")
	iv := ImageViewNew(nil, "images/magnifier.png", zgeo.Size{12, 12})
	iv.SetAlpha(0.4)
	s.Add(zgeo.TopLeft, t)
	s.Add(zgeo.CenterLeft, iv, zgeo.Size{4, 0}).Free = true
	t.Element.Call("addEventListener", "input", js.FuncOf(func(js.Value, []js.Value) interface{} {
		iv.Show(t.Text() == "")
		return nil
	}))

	return s
}

func (v *SearchField) Text() string {
	return v.TextView.Text()
}

func (v *SearchField) SetText(str string) {
	v.TextView.SetText(str)
}
