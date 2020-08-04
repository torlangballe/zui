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
}

func (v *Label) GetTextInfo() TextInfo {
	t := TextInfoNew()
	t.Alignment = v.alignment
	t.Font = v.Font()
	t.Text = v.Text()
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = v.maxLines
	return *t
}

func (v *Label) CalculatedSize(total zgeo.Size) zgeo.Size {
	to := v.View.(TextInfoOwner)
	ti := to.GetTextInfo()
	s := ti.GetBounds()
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	s.MakeInteger()

	// s.W += 10
	//s.H += 3
	// zlog.Info("label calcedsize:", v.Text(), s, o.Font())
	return s
}

func (l *Label) IsMinimumOneLineHight() bool {
	return true
}

func (l *Label) TextAlignment() zgeo.Alignment {
	return l.alignment
}

func (l *Label) MinWidth() float64 {
	return l.minWidth
}

func (l *Label) MaxWidth() float64 {
	return l.maxWidth
}

func (l *Label) MaxLines() int {
	return l.maxLines
}

func (l *Label) SetMinWidth(min float64) View {
	l.minWidth = min
	return l
}

func (l *Label) SetMaxWidth(max float64) View {
	l.maxWidth = max
	return l
}

func (l *Label) SetMaxLines(max int) View {
	l.maxLines = max
	return l
}

func Labelize(view View, prefix string, minWidth float64) (label *Label, stack *StackView, viewCell *ContainerViewCell) {
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
		clabel.SetFont(font).SetColor(zgeo.ColorDefaultForeground.OpacityChanged(0.7))
		view = cstack
	}
	label = LabelNew(title)
	label.SetObjectName("$labelize.label " + prefix)
	label.SetTextAlignment(zgeo.Right)
	label.SetFont(font).SetColor(zgeo.ColorDefaultForeground.OpacityChanged(0.7))
	stack = StackViewHor("$labelize." + prefix) // give it special name so not easy to mis-search for in recursive search

	stack.AddView(label, zgeo.Left|zgeo.VertCenter).MinSize.W = minWidth
	viewCell = stack.AddView(view, zgeo.Left|zgeo.VertCenter)
	return
}

// SetLongPressedHandler sets the handdler function for when label is long-pressed.
// Currently not implemented (see customview for implementation), and doesn't print as zfielsd sets this for Pressable type

func (v *Label) PressedHandler() func() {
	return v.pressed
}

func (v *Label) LongPressedHandler() func() {
	return v.longPressed
}
