package zgo

import "github.com/torlangballe/zutil/ztimer"

//  Oriiginally created by Tor Langballe on /2/11/15.

type TextView struct {
	NativeView
	minWidth    float64
	maxWidth    float64
	maxLines    int
	alignment   Alignment
	changed     func(view View)
	updateTimer *ztimer.Timer

	Margin     Size
	UpdateSecs float64
}

const TextViewDefaultMargin = 3.0

func (v *TextView) GetCalculatedSize(total Size) Size {
	var o TextLayoutOwner
	o = v
	s := TextLayoutOwnerCalculateSize(o)
	s.Add(v.Margin.TimesD(2))
	s.MakeInteger()
	return s
}

func (l *TextView) GetTextAlignment() Alignment {
	return l.alignment
}

func (l *TextView) GetMinWidth() float64 {
	return l.minWidth
}

func (l *TextView) GetMaxWidth() float64 {
	return l.maxWidth
}

func (l *TextView) GetMaxLines() int {
	return l.maxLines
}

func (l *TextView) MinWidth(min float64) View {
	l.minWidth = min
	return l
}

func (l *TextView) MaxWidth(max float64) View {
	l.maxWidth = max
	return l
}

func (l *TextView) MaxLines(max int) View {
	l.maxLines = max
	return l
}
