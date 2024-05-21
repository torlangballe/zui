//go:build zui

package ztext

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zutil/zgeo"
)

type SearchField struct {
	TextView *TextView
	zcontainer.StackView
}

func SearchFieldNew(style Style, chars int) *SearchField {
	s := &SearchField{}
	s.StackView.Init(s, false, "search-stack")
	// s.SetMargin(zgeo.RectFromXY2(5, 5, -5, -5))
	s.SetMinSize(zgeo.SizeD(0, 34))
	style.Type = Search
	t := NewView("", style, chars, 1)
	t.JSSet("className", "rounded")
	s.TextView = t
	t.SetCorner(14)
	t.SetObjectName("search")
	t.SetMargin(zgeo.RectFromXY2(0, 3, -24, -10))
	t.SetNativePadding(zgeo.RectFromXY2(16, 0, -0, -0))
	t.JSSet("inputmode", "search")
	t.UpdateSecs = 0.2
	iv := zimageview.NewWithCachedPath("images/zcore/magnifier.png", zgeo.SizeD(12, 12))
	iv.SetAlpha(0.4)
	s.Add(t, zgeo.CenterLeft|zgeo.VertExpand)
	s.Add(iv, zgeo.CenterLeft, zgeo.SizeD(5, 0)).Free = true
	old := t.ChangedHandler()
	t.SetValueHandler(func(edited bool) {
		iv.Show(t.Text() == "")
		// zlog.Info("Show:", t.Text() == "")
		if old != nil {
			old(edited)
		}
	})
	return s
}

func (v *SearchField) Text() string {
	return v.TextView.Text()
}

func (v *SearchField) SetText(str string) {
	v.TextView.SetText(str)
}
