package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	NativeView
	minWidth  float64
	maxWidth  float64
	maxLines  int
	margin    zgeo.Rect
	alignment zgeo.Alignment
	pressed   func()
}

func (v *Label) CalculatedSize(total zgeo.Size) zgeo.Size {
	var o TextLayoutOwner
	o = v
	s := TextLayoutOwnerCalculateSize(o)
	s.Add(v.margin.Size.Negative())
	s.MakeInteger()

	// s.W += 10
	s.H += 3
	// fmt.Println("label calcedsize:", v.Text(), s, o.Font())
	return s
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
	o, _ := view.(TextLayoutOwner)
	if o != nil {
		font = o.Font()
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
