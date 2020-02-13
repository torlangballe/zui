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
	// fmt.Println("label calcedsize:", v.GetText(), s, o.Font())
	return s
}

func (l *Label) GetTextAlignment() zgeo.Alignment {
	return l.alignment
}

func (l *Label) GetMinWidth() float64 {
	return l.minWidth
}

func (l *Label) MaxWidth() float64 {
	return l.maxWidth
}

func (l *Label) GetMaxLines() int {
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

func Labelize(grid *StackView, view View, prefix string, minWidth float64) *Label {
	font := FontNice(FontDefaultSize, FontStyleBold)
	o := view.(TextLayoutOwner)
	if o != nil {
		font = o.Font()
		font.Style = FontStyleBold
	}
	label := LabelNew(prefix)
	label.SetTextAlignment(zgeo.Right)
	label.SetFont(font).SetColor(zgeo.ColorDefaultForeground.OpacityChanged(0.7))
	stack := StackViewHor("labelize: " + prefix)
	stack.AddView(label, zgeo.Left|zgeo.VertCenter).MinSize.W = minWidth
	stack.AddView(view, zgeo.Left|zgeo.VertCenter).MinSize.W = minWidth
	grid.AddView(stack, zgeo.Left|zgeo.VertCenter)
	return label
}
