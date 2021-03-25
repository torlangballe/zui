// +build zui

package zui

import (
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
	// t.SetMargin(zgeo.RectFromXY2(24, 0, 0, 0))
	t.setjs("inputmode", "search")
	iv := ImageViewNew(nil, "images/magnifier.png", zgeo.Size{10, 10})
	iv.SetAlpha(0.4)
	s.Add(t, zgeo.TopLeft)
	s.Add(iv, zgeo.CenterLeft, zgeo.Size{3, 0}).Free = true
	t.SetOnInputHandler(func() {
		iv.Show(t.Text() == "")
	})
	// Call("addEventListener", "input", js.FuncOf(func(js.Value, []js.Value) interface{} {
	// 	iv.Show(t.Text() == "")
	// 	return nil
	// }))
	return s
}

func (v *SearchField) Text() string {
	return v.TextView.Text()
}

func (v *SearchField) SetText(str string) {
	v.TextView.SetText(str)
}
