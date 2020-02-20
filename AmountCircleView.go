package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type AmountCircleView struct {
	CustomView
	value       float64
	Size        zgeo.Size
	StrokeColor zgeo.Color
	StrokeWidth float64
	ColorsFromValue map[float64]zgeo.Color
}

func AmountCircleViewNew() *AmountCircleView {
	v := &AmountCircleView{}
	v.CustomView.init(v, "amount")
	v.StrokeColor = zgeo.ColorDarkGray
	v.SetColor(zgeo.ColorNewGray(0.7, 1))
	v.CustomView.SetMinSize(zgeo.SizeBoth(24))
	v.StrokeWidth = 2
	v.SetDrawHandler(v.amountViewCircleDraw)
	v.ColorsFromValue = map[float64]zgeo.Color{}
	return v
}

func (v *AmountCircleView) SetValue(value float64) *AmountCircleView {
	v.value = value
	v.Expose()	
	return v
}

func (v *AmountCircleView) Value() float64 {
	return v.value
}

func (v *AmountCircleView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.MinSize()
}

func (v *AmountCircleView) amountViewCircleDraw(rect zgeo.Rect, canvas *Canvas, view View) {
	rect.Size.W = rect.Size.H
	path := zgeo.PathNew()

	var col = v.Color()
	max := 0.0
	for val, c := range v.ColorsFromValue {
		p := 100 * v.Value()
		if  p >= val && max < p {
			col = c
			max = val
		}
	}
	s := rect.Size.MinusD(v.StrokeWidth).DividedByD(2).TimesD(ScreenMain().SoftScale).MinusD(1)
	w := s.Min()

	deg := v.Value() * 360
	path.MoveTo(rect.Center())
	path.ArcDegFromCenter(rect.Center(), zgeo.SizeBoth(w), 0, deg)
	canvas.SetColor(col, 1)
	canvas.FillPath(path)

	line := zgeo.PathNew()
	line.ArcDegFromCenter(rect.Center(), zgeo.SizeBoth(w), 0, 360)
	canvas.SetColor(v.StrokeColor, 1)
	canvas.StrokePath(line, v.StrokeWidth, zgeo.PathLineRound)
}
