// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	// NativeView
	// LongPresser
	CustomView

	minWidth  float64
	maxWidth  float64
	maxLines  int
	margin    zgeo.Rect
	alignment zgeo.Alignment
	text      string

	// pressed     func()
	// longPressed func()

	Columns int
}

func LabelNew(text string) *Label {
	v := &Label{}
	v.Init(v, zstr.Head(text, 100))
	v.text = text
	v.SetObjectName(text)
	v.SetMaxLines(1)
	v.alignment = zgeo.CenterLeft
	v.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	v.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
		ti := v.GetTextInfo()
		ti.Rect = v.LocalRect()
		ti.Draw(canvas)
	})
	return v
}

func (v *Label) SetText(text string) {
	v.text = text
	v.Expose()
}

func (v *Label) Text() string {
	return v.text
}

func (v *Label) GetTextInfo() TextInfo {
	t := TextInfoNew()
	t.Alignment = v.alignment
	t.Font = v.Font()
	t.Color = v.Color()
	t.Text = v.Text()
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
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s = s.Ceil()
	// if v.ObjectName() == "thumbSetTime" {
	// 	zlog.Info("CS:", s, v.Columns, v.maxWidth, v.minWidth, ti.Text)
	// }
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
	font := FontNice(FontDefaultSize, FontStyleBold)
	to, _ := view.(TextInfoOwner)
	if to != nil {
		ti := to.GetTextInfo()
		font = ti.Font
		font.Style = FontStyleBold
	}
	title := prefix
	checkBox, _ := view.(*CheckBox)
	if checkBox != nil {
		title = ""
		clabel, cstack := checkBox.Labelize(prefix)
		clabel.SetFont(font)
		clabel.SetColor(zgeo.ColorDefaultForeground.WithOpacity(0.7))
		view = cstack
	}
	label = LabelNew(title)
	label.SetObjectName("$labelize.label " + prefix)
	label.SetTextAlignment(zgeo.Right)
	label.SetFont(font)
	label.SetColor(zgeo.ColorDefaultForeground.WithOpacity(0.7))
	stack = StackViewHor("$labelize." + prefix) // give it special name so not easy to mis-search for in recursive search

	stack.AddView(label, zgeo.CenterLeft).MinSize.W = minWidth
	viewCell = stack.AddView(view, alignment)
	return
}

func (v *Label) SetMaxLines(max int) {
	// zlog.Info("Label.SetMaxLines:", max, v.ObjectName())
	v.maxLines = max
	v.Expose()
}

func (v *Label) SetTextAlignment(a zgeo.Alignment) {
	v.alignment = a
	v.Expose()
}

func (v *Label) SetMargin(m zgeo.Rect) {
	v.margin = m
	v.Expose()
}
