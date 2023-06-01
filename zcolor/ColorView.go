//go:build zui

package zcolor

import (
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type ColorView struct {
	zview.NativeView
	valueChangedHandlerFunc func()
}

func RegisterWidget() {
	zfields.RegisterWidgeter("zcolor", Widgeter{})
}

func (v *ColorView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{30, 20}
}

// Set up a widgeter interface type for creating colors in zfields:
// And get/set methods it uses

type Widgeter struct{}

func (a Widgeter) Create(f *zfields.Field) zview.View {
	return New(zgeo.ColorClear)
}

func (v *ColorView) SetValueWithAny(col any) {
	v.SetColor(col.(zgeo.Color))
}

func (v *ColorView) ValueAsAny() any {
	zlog.Info("ColVal GetValue", v.Color())
	return v.Color()
}
