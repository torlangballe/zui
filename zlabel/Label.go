//go:build zui

package zlabel

import (
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	zview.NativeView
	zview.LongPresser

	minWidth    float64
	maxWidth    float64
	maxLines    int
	margin      zgeo.Rect
	padding     zgeo.Rect
	alignment   zgeo.Alignment
	text        string // we need to store the text as NativeView's Text() doesn't work right away
	wrap        ztextinfo.WrapType
	pressed     func()
	longPressed func()

	Columns          int
	KeyboardShortcut zkeyboard.KeyMod
}

func New(text string) *Label {
	v := &Label{}
	v.Init(v, text)
	return v
}

func NewLink(name, surl string) *Label {
	v := &Label{}
	v.InitAsLink(v, name, surl)
	return v
}

func (v *Label) GetTextInfo() ztextinfo.Info {
	t := ztextinfo.New()
	t.Alignment = v.alignment
	t.Font = v.Font()
	t.Wrap = v.wrap
	if v.Columns == 0 {
		t.Text = v.Text()
	}
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = v.maxLines
	// t.MinLines = v.maxLines
	return *t
}

func (v *Label) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	var widths []float64
	to := v.View.(ztextinfo.Owner)
	ti := to.GetTextInfo()
	if v.Columns != 0 {
		s = ti.GetColumnsSize(v.Columns)
	} else {
		s, _, widths = ti.GetBounds()
	}
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	if len(widths) == 1 {
		// zlog.Info("LABELINC:", v.Text(), ti.Font.Size/5)
		s.H += ti.Font.Size / 5
	}
	s = s.Ceil()
	s.W += 1
	// zlog.Info("LabelCalcSize:", v.ObjectName(), s, zlog.Full(ti))
	return s, zgeo.SizeD(v.maxWidth, 0) // should we calculate max height?
}

func (v *Label) Text() string {
	return v.text
}

func (v *Label) IsMinimumOneLineHight() bool {
	return v.maxLines > 0
}

func (v *Label) TextAlignment() zgeo.Alignment {
	return v.alignment
}

func (v *Label) MinWidth() float64 {
	return v.minWidth
}

func (v *Label) MaxWidth() float64 {
	return v.maxWidth
}

func (v *Label) MaxLines() int {
	return v.maxLines
}

func (v *Label) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *Label) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func (v *Label) SetWidth(w float64) {
	v.SetMinWidth(w)
	v.SetMaxWidth(w)
}

func (v *Label) PressedHandler() func() {
	return v.pressed
}

func (v *Label) LongPressedHandler() func() {
	return v.longPressed
}

func (v *Label) HandleOutsideShortcut(sc zkeyboard.KeyMod) bool {
	if !v.KeyboardShortcut.IsNull() && sc == v.KeyboardShortcut && v.PressedHandler() != nil {
		v.PressedHandler()()
		return true
	}
	return false
}
