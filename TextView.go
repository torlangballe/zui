// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

//  Originally created by Tor Langballe on /2/11/15.

type TextViewStyle struct {
	IsSearch      bool
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
	changed       func()
	pushedBGColor zgeo.Color
	keyPressed    func(key KeyboardKey, mods KeyboardModifier)
	updateTimer   *ztimer.Timer
	Columns       int
	rows          int
	isSearch      bool
	//	updated       bool

	margin zgeo.Rect
	// ContinuousUpdateCalls bool
	UpdateSecs float64
	editDone   func(canceled bool)
}

var TextViewDefaultMargin = 2.0
var TextViewDefaultColor zgeo.Color       // if undef, don't set, use whatever platform already has
var TextViewDefaultBGColor zgeo.Color     // "
var TextViewDefaultBorderColor zgeo.Color // "

func TextViewNew(text string, style TextViewStyle, cols, rows int) *TextView {
	v := &TextView{}
	zlog.Assert(cols != 0)
	v.Init(v, text, style, rows, cols)
	return v
}

func (v *TextView) SelectAll() {
	v.Select(0, -1)
}

func (v *TextView) IsEditing() bool {
	return v.updateTimer != nil
}

func (v *TextView) CalculatedSize(total zgeo.Size) zgeo.Size {
	ti := TextInfoNew()
	ti.Alignment = v.alignment
	ti.IsMinimumOneLineHight = true
	ti.Font = v.Font()
	ti.MinLines = v.rows
	if v.maxWidth != 0 {
		ti.SetWidthFreeHight(v.maxWidth + v.margin.Size.W*2)
	}

	s := ti.GetColumnsSize(v.Columns)
	// zlog.Info("TextView size:", s, v.margin.Size, v.ObjectName()) //, zlog.GetCallingStackString())
	s.Add(v.margin.Size.Negative())
	if v.isSearch {
		//		s.H += 14
	}
	s = s.Ceil()
	return s
}

func (v *TextView) Margin() zgeo.Rect {
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

func (v *TextView) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *TextView) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func (v *TextView) SetMaxLines(max int) {
	v.rows = max
}

func (v *TextView) IsMinimumOneLineHight() bool {
	return true
}

func (v *TextView) ChangedHandler() func() {
	return v.changed
}
