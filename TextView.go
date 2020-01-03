package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

//  Oriiginally created by Tor Langballe on /2/11/15.

type TextViewStyle struct {
	KeyboardType  KeyboardType
	AutoCapType   KeyboardAutoCapType
	ReturnKeyType KeyboardReturnKeyType
	IsAutoCorrect bool
}

type TextView struct {
	NativeView
	minWidth    float64
	maxWidth    float64
	maxLines    int
	alignment   zgeo.Alignment
	changed     func(view View)
	keyPressed  func(view View, key KeyboardKey, mods KeyboardModifier)
	updateTimer *ztimer.Timer
	updated     bool

	Margin                zgeo.Size
	ContinuousUpdateCalls bool
	UpdateSecs            float64
}

const TextViewDefaultMargin = 3.0

func (v *TextView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	var o TextLayoutOwner
	o = v
	s := TextLayoutOwnerCalculateSize(o)
	s.Add(v.Margin.TimesD(2))
	s.MakeInteger()
	return s
}

func (l *TextView) GetTextAlignment() zgeo.Alignment {
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
