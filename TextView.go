package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

//  Originally created by Tor Langballe on /2/11/15.

type TextViewStyle struct {
	KeyboardType  KeyboardType
	AutoCapType   KeyboardAutoCapType
	ReturnKeyType KeyboardReturnKeyType
	IsAutoCorrect bool
}

type TextView struct {
	NativeView
	minWidth      float64
	maxWidth      float64
	maxLines      int
	alignment     zgeo.Alignment
	changed       func(view View)
	pushedBGColor zgeo.Color
	keyPressed    func(view View, key KeyboardKey, mods KeyboardModifier)
	updateTimer   *ztimer.Timer
	//	updated       bool

	Margin zgeo.Size
	// ContinuousUpdateCalls bool
	UpdateSecs float64
}

const TextViewDefaultMargin = 3.0

func TextViewNew(text string, style TextViewStyle) *TextView {
	tv := &TextView{}
	tv.Init(text, style)
	return tv
}

func (v *TextView) CalculatedSize(total zgeo.Size) zgeo.Size {
	s := TextLayoutCalculateSize(v.alignment, v.Font(), "Ajzig", v.maxLines, v.maxWidth, true)
	s.Add(v.Margin.TimesD(2))
	s.MakeInteger()
	return s
}

func (l *TextView) TextAlignment() zgeo.Alignment {
	return l.alignment
}

func (l *TextView) MinWidth() float64 {
	return l.minWidth
}

func (l *TextView) MaxWidth() float64 {
	return l.maxWidth
}

func (l *TextView) MaxLines() int {
	return l.maxLines
}

func (l *TextView) SetMinWidth(min float64) View {
	l.minWidth = min
	return l
}

func (l *TextView) SetMaxWidth(max float64) View {
	l.maxWidth = max
	return l
}

func (l *TextView) SetMaxLines(max int) View {
	l.maxLines = max
	return l
}

func (l *TextView) IsMinimumOneLineHight() bool {
	return true
}
