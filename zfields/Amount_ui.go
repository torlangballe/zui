//go:build zui

package zfields

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidget"
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

func (a AmountBarWidgeter) Create(f *Field) zview.View {
	min := f.MinWidth
	if min == 0 {
		min = 100
	}
	progress := zwidget.AmountViewBarNew(min)
	if f.Styling.FGColor.Valid {
		col := f.Styling.FGColor
		if col.Valid {
			progress.SetColor(col)
		}
	}
	return progress
}

func (a AmountBarWidgeter) SetValue(view zview.View, val any) {
	av := view.(*zwidget.AmountView)
	n, err := zfloat.GetAny(val)
	zlog.Info("AmountSet:", av.Hierarchy(), n, err)
	if !zlog.OnError(err) {
		av.SetValue(n)
	}
}

func (a AmountBarWidgeter) IsStatic() bool {
	return true
}

func (a AmountBarWidgeter) GetValue(view zview.View) any {
	progress := view.(*zwidget.AmountView)
	return progress.Value()
}

type AmountCircleWidgeter struct{} //////////////////////////////////////////////////////////////

func (a AmountCircleWidgeter) Create(f *Field) zview.View {
	if f.Size.IsNull() {
		f.Size = zgeo.SizeBoth(20)
	}
	view := zwidget.AmountViewCircleNew()
	view.SetMinSize(f.Size)
	view.SetColor(zgeo.ColorNew(0, 0.8, 0, 1))
	for i, n := range []float64{0, 70, 90} {
		if len(f.Colors) > 1 {
			if i < len(f.Colors) {
				view.ColorsFromValue[n] = zgeo.ColorFromString(f.Colors[i])
			}
		} else if f.Styling.FGColor.Valid {
			view.ColorsFromValue[0] = f.Styling.FGColor
		}
	}
	return view
}

func (a AmountCircleWidgeter) SetValue(view zview.View, val any) {
	circle := view.(*zwidget.AmountView)
	n, err := zfloat.GetAny(val)
	if !zlog.OnError(err) {
		circle.SetValue(n)
	}
}

func (a AmountCircleWidgeter) IsStatic() bool {
	return true
}

func (a AmountCircleWidgeter) GetValue(view zview.View) any {
	circle := view.(*zwidget.AmountView)
	return circle.Value()
}

type ActivityWidgeter struct{} //////////////////////////////////////////////////////////////

func (a ActivityWidgeter) Create(f *Field) zview.View {
	if f.Size.IsNull() {
		f.Size = zgeo.SizeBoth(20)
	}
	f.SetEdited = false
	av := zwidget.NewActivityView(f.Size)
	av.AlwaysVisible = f.Visible
	return av
}

func (a ActivityWidgeter) SetValue(view zview.View, val any) {
	on := val.(bool)
	activity := view.(*zwidget.ActivityView)
	if on {
		activity.Start()
	} else {
		activity.Stop()
	}
}

func (a ActivityWidgeter) IsStatic() bool {
	return true
}

func (a ActivityWidgeter) GetValue(view zview.View) any {
	activity := view.(*zwidget.ActivityView)
	return activity.IsStopped()
}
