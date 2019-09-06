package zgo

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	NativeView
	minWidth  float64
	maxWidth  float64
	maxHeight float64
	maxLines  int
	alignment Alignment
	Margin    Rect
	pressed   func(pos Pos)
}

func (v *Label) GetCalculatedSize(total Size) Size {
	var o TextLayoutOwner
	o = v
	return CalculateSize(o, total)
}

func (l *Label) TextAlignment(a Alignment) View {
	l.alignment = a
	return l
}

func (l *Label) GetTextAlignment() Alignment {
	return l.alignment
}

func (l *Label) GetMinWidth() float64 {
	return l.minWidth
}

func (l *Label) GetMaxWidth() float64 {
	return l.maxWidth
}

func (l *Label) GetMaxLines() int {
	return l.maxLines
}

func (l *Label) MinWidth(min float64) View {
	l.minWidth = min
	return l
}

func (l *Label) MaxWidth(max float64) View {
	l.maxWidth = max
	return l
}

func (l *Label) MaxLines(max int) View {
	l.maxLines = max
	return l
}
