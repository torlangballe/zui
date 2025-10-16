//go:build zui

package ztext

import (
	"strconv"

	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
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
	storeKey      string
	UpdateSecs    float64
	minValue      float64
	maxValue      float64
	FilterFunc    func(str string) string // Filter changes text before .SetText() and .Text(). Also called on each key input. return 0 means no rune
	editDone      func(canceled bool)
}

var (
	DefaultMargin      = zgeo.RectFromXY2(4, 3, -3, -2)
	DefaultColor       = zstyle.DefaultFGColor
	DefaultBGColor     = zstyle.GrayCur(0.95, 0.3)
	DefaultBorderColor = zstyle.GrayCur(0.3, 0.5)
)

func NewView(text string, style Style, cols, rows int) *TextView {
	v := &TextView{}
	zlog.Assert(cols != 0)
	v.Init(v, text, style, cols, rows)
	return v
}

func NewInteger(style Style, cols int) *TextView {
	style.KeyboardType = zkeyboard.TypeInteger
	v := NewView("", style, cols, 1)
	return v
}

func (v *TextView) SetMinValue(m float64) {
	v.minValue = m
}

func (v *TextView) SetMaxValue(m float64) {
	v.maxValue = m
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
	// zlog.Info("TV.CalculatedSize", v.Hierarchy(), v.Columns, s)
	s.Add(v.margin.Size.Negative())
	s = s.Ceil()
	s.W += 6
	// s.H -= 2
	zfloat.Maximize(&s.H, 22)
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

func (v *TextView) SetStoreKey(key, defaultValue string) {
	v.storeKey = key
	if zkeyvalue.DefaultStore != nil {
		str, got := zkeyvalue.DefaultStore.GetString(v.storeKey)
		if !got {
			str = defaultValue
		}
		v.SetText(str)
	}
}

func (v *TextView) Int64() (int64, error) {
	str := v.Text()
	return strconv.ParseInt(str, 10, 64)
}

func (v *TextView) SetInt64(n int64) {
	str := strconv.FormatInt(n, 10)
	v.SetText(str)
}

func (v *TextView) Int() (int, error) {
	n, err := v.Int64()
	return int(n), err
}

func (v *TextView) SetInt(n int) {
	v.SetInt64(int64(n))
}

func (v *TextView) Double() (float64, error) {
	str := v.Text()
	return strconv.ParseFloat(str, 64)
}

func (v *TextView) SetDouble(n float64) {
	str := strconv.FormatFloat(n, 'f', -1, 64)
	v.SetText(str)
}
