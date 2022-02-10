//go:build zui
// +build zui

package zfields

import (
	"github.com/torlangballe/zui"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func init() {
	RegisterWigeter("zamount-bar", AmountBarWidgeter{})
	RegisterWigeter("zamount-circle", AmountCircleWidgeter{})
	RegisterWigeter("zactivity", ActivityWidgeter{})
}

type AmountBarWidgeter struct{} //////////////////////////////////////////////////////////////

func (a AmountBarWidgeter) Create(f *Field) zui.View {
	min := f.MinWidth
	if min == 0 {
		min = 100
	}
	progress := zui.AmountViewBarNew(min)
	if len(f.Colors) != 0 {
		col := zgeo.ColorFromString(f.Colors[0])
		if col.Valid {
			progress.SetColor(col)
		}
	}
	return progress
}

func (a AmountBarWidgeter) SetValue(view zui.View, val interface{}) {
	av := view.(*zui.AmountView)
	n, err := zfloat.GetAny(val)
	zlog.Info("AmountSet:", av.Hierarchy(), n, err)
	if !zlog.OnError(err) {
		av.SetValue(n)
	}
}

func (a AmountBarWidgeter) GetValue(view zui.View) interface{} {
	progress := view.(*zui.AmountView)
	return progress.Value()
}

type AmountCircleWidgeter struct{} //////////////////////////////////////////////////////////////

func (a AmountCircleWidgeter) Create(f *Field) zui.View {
	if f.Size.IsNull() {
		f.Size = zgeo.SizeBoth(20)
	}
	view := zui.AmountViewCircleNew()
	view.SetMinSize(f.Size)
	view.SetColor(zgeo.ColorNew(0, 0.8, 0, 1))
	for i, n := range []float64{0, 70, 90} {
		if i < len(f.Colors) {
			view.ColorsFromValue[n] = zgeo.ColorFromString(f.Colors[i])
		}
	}
	return view
}

func (a AmountCircleWidgeter) SetValue(view zui.View, val interface{}) {
	circle := view.(*zui.AmountView)
	n, err := zfloat.GetAny(val)
	if !zlog.OnError(err) {
		circle.SetValue(n)
	}
}

func (a AmountCircleWidgeter) GetValue(view zui.View) interface{} {
	circle := view.(*zui.AmountView)
	return circle.Value()
}

type ActivityWidgeter struct{} //////////////////////////////////////////////////////////////

func (a ActivityWidgeter) Create(f *Field) zui.View {
	if f.Size.IsNull() {
		f.Size = zgeo.SizeBoth(20)
	}
	f.SetEdited = false
	av := zui.ActivityViewNew(f.Size)
	av.AlwaysVisible = f.Visible
	return av
}

func (a ActivityWidgeter) SetValue(view zui.View, val interface{}) {
	on := val.(bool)
	activity := view.(*zui.ActivityView)
	if on {
		activity.Start()
	} else {
		activity.Stop()
	}
}

func (a ActivityWidgeter) GetValue(view zui.View) interface{} {
	activity := view.(*zui.ActivityView)
	return activity.IsStopped()
}
