// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /19/apr/21.

type Button struct {
	NativeView
	LongPresser

	minWidth float64
	maxWidth float64
	margin   zgeo.Rect

	pressed     func()
	longPressed func()
}

func (v *Button) GetTextInfo() TextInfo {
	t := TextInfoNew()
	t.Font = v.Font()
	t.Text = v.Text()
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = 1
	t.IsMinimumOneLineHight = true
	return *t
}

func (v *Button) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	to := v.View.(TextInfoOwner)
	ti := to.GetTextInfo()
	s, _, _ = ti.GetBounds()
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s = s.Ceil()
	// zlog.Info("Button CS:", s)
	return s
}

func (v *Button) MinWidth() float64 {
	return v.minWidth
}

func (v *Button) MaxWidth() float64 {
	return v.maxWidth
}

func (v *Button) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *Button) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func (v *Button) PressedHandler() func() {
	return v.pressed
}

func (v *Button) LongPressedHandler() func() {
	return v.longPressed
}
