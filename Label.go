// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	NativeView
	LongPresser

	minWidth  float64
	maxWidth  float64
	maxLines  int
	margin    zgeo.Rect
	alignment zgeo.Alignment

	pressed     func()
	longPressed func()

	Columns int
}

func (v *Label) GetTextInfo() TextInfo {
	t := TextInfoNew()
	t.Alignment = v.alignment
	t.Font = v.Font()
	if v.Columns == 0 {
		t.Text = v.Text()
	}
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = v.maxLines
	return *t
}

func (v *Label) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	to := v.View.(TextInfoOwner)
	ti := to.GetTextInfo()
	if v.Columns != 0 {
		s = ti.GetColumnsSize(v.Columns)
	} else {
		s, _, _ = ti.GetBounds()
	}
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	// zlog.Info("L CS:", v.ObjectName(), s, v.minWidth, v.maxWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s = s.Ceil()
	s.W += 1
	return s
}

func (v *Label) IsMinimumOneLineHight() bool {
	return v.maxLines > 0
}

func (v *Label) TextAlignment() zgeo.Alignment {
	return v.alignment
}

func (v *Label) MinWidth() float64 {
	return v.minWidth
}

func (v *Label) MaxWidth() float64 {
	return v.maxWidth
}

func (v *Label) MaxLines() int {
	return v.maxLines
}

func (v *Label) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *Label) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func Labelize(view View, prefix string, minWidth float64, alignment zgeo.Alignment) (label *Label, stack *StackView, viewCell *ContainerViewCell) {
	font := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleBold)
	to, _ := view.(TextInfoOwner)
	if to != nil {
		ti := to.GetTextInfo()
		font = ti.Font
		font.Style = zgeo.FontStyleBold
	}
	title := prefix
	checkBox, isCheck := view.(*CheckBox)
	if checkBox != nil && alignment&zgeo.Right != 0 {
		title = ""
		_, cstack := checkBox.Labelize(prefix)
		view = cstack
		alignment = alignment.FlippedHorizontal()
	}
	label = LabelNew(title)
	label.SetObjectName("$labelize.label " + prefix)
	label.SetTextAlignment(zgeo.Right)
	label.SetFont(font)
	label.SetColor(StyleDefaultFGColor().WithOpacity(0.7))
	stack = StackViewHor("$labelize." + prefix) // give it special name so not easy to mis-search for in recursive search

	stack.AddView(label, zgeo.CenterLeft).MinSize.W = minWidth
	marg := zgeo.Size{}
	if isCheck {
		marg.W = -6 // in html cell has a box around it of 20 pixels
	}
	viewCell = stack.Add(view, alignment, marg)
	return
}

func (v *Label) PressedHandler() func() {
	return v.pressed
}

func (v *Label) LongPressedHandler() func() {
	return v.longPressed
}
