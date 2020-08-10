package zui

import (
	"strings"

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
	alignment     zgeo.Alignment
	changed       func(view View)
	pushedBGColor zgeo.Color
	keyPressed    func(view View, key KeyboardKey, mods KeyboardModifier)
	updateTimer   *ztimer.Timer
	Columns       int
	rows          int
	//	updated       bool

	margin zgeo.Size
	// ContinuousUpdateCalls bool
	UpdateSecs float64
}

const TextViewDefaultMargin = 2.0

func TextViewNew(text string, style TextViewStyle, cols, rows int) *TextView {
	tv := &TextView{}
	tv.Init(text, style, rows, cols)
	return tv
}

func (v *TextView) IsEditing() bool {
	return v.updateTimer != nil
}

func (v *TextView) CalculatedSize(total zgeo.Size) zgeo.Size {
	const letters = "etaoinsrhdlucmfywgpbvkxqjz"
	ti := TextInfoNew()
	ti.Alignment = v.alignment
	ti.IsMinimumOneLineHight = true
	if v.Columns == 0 {
		ti.Text = v.Text()
	} else {
		len := len(letters)
		for i := 0; i < v.Columns; i++ {
			c := string(letters[i%len])
			if i%8 == 4 {
				c = strings.ToUpper(c)
			}
			ti.Text += c
		}
	}
	ti.Font = v.Font()
	ti.MaxLines = v.rows
	if v.maxWidth != 0 {
		ti.SetWidthFreeHight(v.maxWidth - v.margin.W*2)
	}
	s := ti.GetBounds()
	s.Add(v.margin.TimesD(2))
	s.MakeInteger()
	return s
}

func (v *TextView) Margin() zgeo.Size {
	return v.margin
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
	return v.rows
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
	v.rows = max
	return v
}

func (v *TextView) IsMinimumOneLineHight() bool {
	return true
}
