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
	ValueChangedHandlerFunc func()
}

func RegisterWidget() {
	zfields.RegisterWidgeter("zcolor", Widgeter{})
	zlog.Info("zfields.RegisterWidgeter.zcolor")
}

func (v *ColorView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return zgeo.Size{30, 20}
}

// Set up a widgeter interface type for creating colors in zfields:

type Widgeter struct{}

func (a Widgeter) Create(f *zfields.Field) zview.View {
	return New(zgeo.ColorClear)
}

func (a Widgeter) SetValue(view zview.View, val any) {
	view.(*ColorView).SetColor(val.(zgeo.Color))
}

func (a Widgeter) GetValue(view zview.View) any {
	return view.(*ColorView).Color()
}
