//go:build zui

package zcolor

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type ColorView struct {
	zview.NativeView
	valueChangedHandlerFunc func(edited bool)
}

func (v *ColorView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	s = zgeo.SizeD(30, 20)
	return s, s
}

func (v *ColorView) SetValueWithAny(col any) {
	v.SetColor(col.(zgeo.Color))
}

func (v *ColorView) ValueAsAny() any {
	zlog.Info("ColVal GetValue", v.Color())
	return v.Color()
}
