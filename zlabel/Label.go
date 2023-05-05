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

	minWidth  float64
	maxWidth  float64
	maxLines  int
	margin    zgeo.Rect
	alignment zgeo.Alignment

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

func (v *Label) GetTextInfo() ztextinfo.Info {
	t := ztextinfo.New()
	t.Alignment = v.alignment
	t.Font = v.Font()
	if v.Columns == 0 {
		t.Text = v.Text()
	}
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = v.maxLines
	return *t
}

func (v *Label) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	to := v.View.(ztextinfo.Owner)
	ti := to.GetTextInfo()
	if v.Columns != 0 {
		s = ti.GetColumnsSize(v.Columns)
	} else {
		s, _, _ = ti.GetBounds()
	}
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s = s.Ceil()
	s.W += 1
	return s
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
