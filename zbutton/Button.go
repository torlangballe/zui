//go:build zui

package zbutton

import (
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /19/apr/21.

type Button struct {
	zview.NativeView
	minWidth float64
	maxWidth float64
	margin   zgeo.Rect
}

func (v *Button) GetTextInfo() ztextinfo.Info {
	t := ztextinfo.New()
	t.Font = v.Font()
	t.Text = v.Text()
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = 1
	t.IsMinimumOneLineHight = true
	return *t
}

func (v *Button) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	to := v.View.(ztextinfo.Owner)
	ti := to.GetTextInfo()
	s, _, _ = ti.GetBounds()
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s = s.Ceil()
	// zlog.Info("Button CS:", v.ObjectName(), s)
	return s, zgeo.SizeD(0, s.H)
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
