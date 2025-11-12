//go:build zui

package ztext

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zgeo"
)

type SearchField struct {
	zcontainer.StackView
	TextView *TextView
}

func SearchFieldNew(style Style, chars int) *SearchField {
	s := &SearchField{}
	s.StackView.Init(s, false, "search-stack")
	// s.SetMargin(zgeo.RectFromXY2(5, 5, -5, -5))
	// s.SetMinSize(zgeo.SizeD(0, 34))
	style.Type = Search
	tv := NewView("", style, chars, 1)
	tv.SetColor(zstyle.DefaultFGColor())
	tv.JSSet("className", "rounded")
	s.TextView = tv
	tv.SetCorner(14)
	tv.SetObjectName("search")
	// tv.SetMargin(zgeo.RectFromXY2(0, 3, -24, -10))
	// tv.SetMargin(zgeo.RectFromXY2(0, 3, -24, -10))
	tv.SetNativePadding(zgeo.RectFromXY2(16, 0, 0, 0))
	tv.JSSet("inputmode", "search")
	tv.SetJSStyle("border", "2px")
	tv.UpdateSecs = 0.2
	iv := zimageview.NewWithCachedPath("images/zcore/magnifier.png", zgeo.SizeD(12, 12))
	iv.MixColorForDarkMode = zgeo.ColorLightGray
	iv.SetAlpha(0.4)
	s.Add(tv, zgeo.CenterLeft|zgeo.VertExpand)
	s.Add(iv, zgeo.CenterLeft, zgeo.SizeD(6, 0)).Free = true
	tv.SetValueHandler("zsearch.showCross", func(edited bool) {
		iv.Show(tv.Text() == "")
		// zlog.Info("Show:", t.Text() == "")
	})
	return s
}

func (v *SearchField) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	s, max = v.StackView.CalculatedSize(total)
	return s, max
}

func (v *SearchField) SetValueHandler(id string, handler func(edited bool)) {
	v.TextView.SetValueHandler(id, handler)
}

func (v *SearchField) Text() string {
	return v.TextView.Text()
}

func (v *SearchField) SetText(str string) {
	v.TextView.SetText(str)
}
