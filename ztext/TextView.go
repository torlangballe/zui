//go:build zui

package ztext

import (
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

//  Originally created by Tor Langballe on /2/11/15.

type Type string

const (
	Normal Type = ""
	Search Type = "search"
	Date   Type = "date"
)

type Style struct {
	Type                Type
	KeyboardType        zkeyboard.Type
	AutoCapType         zkeyboard.AutoCapType
	ReturnKeyType       zkeyboard.ReturnKeyType
	IsAutoCorrect       bool
	DisableAutoComplete bool
	IsExistingPassword  bool // password type, it's an existing password
}

type TextView struct {
	zview.NativeView
	minWidth      float64
	maxWidth      float64
	alignment     zgeo.Alignment
	changed       zview.ValueHandlers
	pushedBGColor zgeo.Color
	updateTimer   *ztimer.Timer
	Columns       int
	rows          int
	textStyle     Style
	margin        zgeo.Rect
	UpdateSecs    float64
	FilterFunc    func(str string) string // Filter changes text before .SetText() and .Text(). Also called on each key input. return 0 means no rune
	editDone      func(canceled bool)
}

var (
	DefaultMargin      = zgeo.RectFromXY2(4, 3, -3, -2)
	DefaultColor       = zstyle.DefaultFGColor
	DefaultBGColor     = zstyle.GrayF(0.95, 0.3)
	DefaultBorderColor = zstyle.GrayF(0.3, 0.5)
)

func NewView(text string, style Style, cols, rows int) *TextView {
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

func (v *TextView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	ti := ztextinfo.New()
	ti.Alignment = v.alignment
	ti.IsMinimumOneLineHight = true
	ti.Font = v.Font()
	ti.MinLines = v.rows
	if v.maxWidth != 0 {
		ti.SetWidthFreeHight(v.maxWidth + v.margin.Size.W*2)
	}
	s = ti.GetColumnsSize(v.Columns)
	s.Add(v.margin.Size.Negative())
	s = s.Ceil()
	s.W += 6
	// s.H -= 2
	zfloat.Maximize(&s.H, 34)
	if v.maxWidth != 0 {
		max.W = s.W
	}
	return s, max
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
