//go:build zui
// +build zui

package zui

import (
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

//  Originally created by Tor Langballe on /2/11/15.

type TextViewType string

const (
	TextViewNormal = ""
	TextViewSearch = "search"
	TextViewDate   = "date"
)

type TextViewStyle struct {
	Type          TextViewType
	KeyboardType  zkeyboard.Type
	AutoCapType   zkeyboard.AutoCapType
	ReturnKeyType zkeyboard.ReturnKeyType
	IsAutoCorrect bool
}

type TextView struct {
	NativeView
	minWidth      float64
	maxWidth      float64
	alignment     zgeo.Alignment
	changed       func()
	pushedBGColor zgeo.Color
	keyPressed    func(key zkeyboard.Key, mods zkeyboard.Modifier) bool
	updateTimer   *ztimer.Timer
	Columns       int
	rows          int
	textStyle     TextViewStyle
	//	isSearch      bool
	//	updated       bool

	margin zgeo.Rect
	// ContinuousUpdateCalls bool
	UpdateSecs float64
	editDone   func(canceled bool)
}

var (
	TextViewDefaultMargin      = zgeo.RectFromXY2(4, 3, -3, -2)
	TextViewDefaultColor       = StyleDefaultFGColor
	TextViewDefaultBGColor     = StyleGrayF(0.95, 0.3)
	TextViewDefaultBorderColor = StyleGrayF(0.3, 0.5)
)

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
	ti := ztextinfo.New()
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
