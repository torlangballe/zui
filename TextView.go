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

func TextViewNew(text string, style TextViewStyle, maxLines int) *TextView {
	tv := &TextView{}
	tv.Init(text, style, maxLines)
	return tv
}

func (v *TextView) CalculatedSize(total zgeo.Size) zgeo.Size {
	ti := TextInfoNew()
	ti.Alignment = v.alignment
	ti.Text = v.Text()
	ti.IsMinimumOneLineHight = true
	ti.Font = v.Font()
	ti.MaxLines = v.maxLines
	ti.SetWidthFreeHight(v.maxWidth)
	s := ti.GetBounds()
	s.Add(v.Margin.TimesD(2))
	s.MakeInteger()
	return s
}

func (v *TextView) TextAlignment() zgeo.Alignment {
	return v.alignment
}

func (v *TextView) MinWidth() float64 {
	return v.minWidth
}

func (v *TextView) MaxWidth() float64 {
	return v.maxWidth
}

func (v *TextView) MaxLines() int {
	return v.maxLines
}

func (v *TextView) SetMinWidth(min float64) View {
	v.minWidth = min
	return v
}

func (v *TextView) SetMaxWidth(max float64) View {
	v.maxWidth = max
	return v
}

func (v *TextView) SetMaxLines(max int) View {
	v.maxLines = max
	return v
}

func (v *TextView) IsMinimumOneLineHight() bool {
	return true
}

