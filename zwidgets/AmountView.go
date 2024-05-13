//go:build zui

package zwidgets

import (
	"fmt"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdraw"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

var AmountViewCircleDefaultDiameter = 20.0

type AmountView struct {
	zcustom.CustomView
	value           float64
	Size            zgeo.Size
	StrokeColor     zgeo.Color
	StrokeWidth     float64
	ColorsFromValue map[float64]zgeo.Color
}

func (v *AmountView) SetValue(value float64) {
	v.value = value
	v.SetToolTip(fmt.Sprint(value))
	v.Expose()
}

func (v *AmountView) SetValueWithAny(value any) {
	n, err := zfloat.GetAny(value)
	if !zlog.OnError(err) {
		v.SetValue(n)
	}

}

func (v *AmountView) Value() float64 {
	return v.value
}

func (v *AmountView) ValueAsAny() any {
	return v.Value()
}

func (v *AmountView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.MinSize()
}

func (v *AmountView) getColorForValue() zgeo.Color {
	var col = v.Color()
	max := 0.0
	for val, c := range v.ColorsFromValue {
		p := 100 * v.Value()
		if p >= val && max < p {
			col = c
			max = val
		}
	}
	return col
}

func (v *AmountView) drawCircle(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	rect.Size.W = rect.Size.H
	zdraw.DrawAmountPie(rect, canvas, v.Value(), v.StrokeWidth, v.getColorForValue(), v.StrokeColor)
}

func AmountViewCircleNew() *AmountView {
	v := &AmountView{}
	v.CustomView.Init(v, "amount")
	v.StrokeColor = zgeo.ColorDarkGray
	v.SetColor(zgeo.ColorNewGray(0.7, 1))
	v.CustomView.SetMinSize(zgeo.SizeBoth(AmountViewCircleDefaultDiameter))
	v.StrokeWidth = 2
	v.SetDrawHandler(v.drawCircle)
	v.ColorsFromValue = map[float64]zgeo.Color{}
	return v
}

func (v *AmountView) drawBar(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	val := v.Value()
	if val != -1 {
		rect.Size.W *= val
		if rect.Size.W < 3 {
			rect.Size.W = 3
		}
		path := zgeo.PathNewRect(rect, zgeo.SizeD(3, 3))
		col := v.getColorForValue()
		canvas.SetColor(col)
		canvas.FillPath(path)
	}
}

// AmountViewBar ********************************************************

func AmountViewBarNew(width float64) *AmountView {
	v := &AmountView{}
	v.CustomView.Init(v, "amount")
	v.SetBGColor(zgeo.ColorNewGray(0, 0.15))
	v.SetColor(zgeo.ColorNew(0.4, 0.4, 0.8, 1))
	v.SetCorner(3)
	v.CustomView.SetMinSize(zgeo.SizeD(width, 10))
	v.SetDrawHandler(v.drawBar)
	v.ColorsFromValue = map[float64]zgeo.Color{}
	return v
}

// AmountViewCircle *******************************************************************
