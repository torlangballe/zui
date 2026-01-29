//go:build zui

package zconsole

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zutil/zgeo"
)

type ConsoleView struct {
	zcontainer.StackView
	TextView *ztext.TextView
	MaxBytes int
}

func NewView(text string, cols, rows int) *ConsoleView {
	v := &ConsoleView{}
	v.StackView.Init(v, true, "console-view")
	v.SetSpacing(0)

	v.TextView = ztext.NewView(text, ztext.Style{}, cols, rows)
	// v.TextView.SetMargin(zgeo.RectFromXY2(20, 20, -20, -20))
	font := zgeo.FontNew("Menlo", 13, zgeo.FontStyleNormal)
	v.TextView.SetFont(font)
	v.TextView.SetBGColor(zgeo.ColorNewGray(0.15, 1))
	v.TextView.SetColor(zgeo.ColorNew(0.6, 8, 0.6, 1))
	v.TextView.SetStroke(0, zgeo.ColorClear, false)

	// v.SetUsable(false)
	v.MaxBytes = 10000
	v.TextView.SetIsStatic(true)

	v.Add(v.TextView, zgeo.TopCenter|zgeo.Expand, zgeo.SizeNull) // -1, -1
	return v
}

func (v *ConsoleView) AddText(str string) {
	old := v.TextView.Text()
	text := old + str
	length := len(text)
	if v.MaxBytes != 0 && length > v.MaxBytes {
		text = text[length-v.MaxBytes:]
	}
	v.SetText(text)
}

func (v *ConsoleView) SetText(str string) {
	v.TextView.SetText(str)
	v.TextView.ScrollToBottom()
}

func (v *ConsoleView) SetValueWithAny(value any) {
	v.SetText(value.(string))
}
