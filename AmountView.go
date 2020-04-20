package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type AmountView struct {
	CustomView
	value           float64
	Size            zgeo.Size
	StrokeColor     zgeo.Color
	StrokeWidth     float64
	ColorsFromValue map[float64]zgeo.Color
}

func (v *AmountView) SetValue(value float64) *AmountView {
	v.value = value
	v.Expose()
	return v
}

func (v *AmountView) Value() float64 {
	return v.value
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

func AmountCircleDraw(rect zgeo.Rect, canvas *Canvas, value, strokeWidth float64, color, strokeColor zgeo.Color) {
	path := zgeo.PathNew()
	s := rect.Size.MinusD(strokeWidth).DividedByD(2).TimesD(ScreenMain().SoftScale).MinusD(1)
	w := s.Min()
	path.MoveTo(rect.Center())
	path.ArcDegFromCenter(rect.Center(), zgeo.SizeBoth(w), 0, value*360)
	canvas.SetColor(color, 1)
	canvas.FillPath(path)

	line := zgeo.PathNew()
	line.ArcDegFromCenter(rect.Center(), zgeo.SizeBoth(w), 0, 360)
	canvas.SetColor(strokeColor, 1)
	canvas.StrokePath(line, strokeWidth, zgeo.PathLineRound)
}

func (v *AmountView) drawCircle(rect zgeo.Rect, canvas *Canvas, view View) {
	rect.Size.W = rect.Size.H
	AmountCircleDraw(rect, canvas, v.Value(), v.StrokeWidth, v.getColorForValue(), v.StrokeColor)
}

func AmountViewCircleNew() *AmountView {
	v := &AmountView{}
	v.CustomView.init(v, "amount")
	v.StrokeColor = zgeo.ColorDarkGray
	v.SetColor(zgeo.ColorNewGray(0.7, 1))
	v.CustomView.SetMinSize(zgeo.SizeBoth(24))
	v.StrokeWidth = 2
	v.SetDrawHandler(v.drawCircle)
	v.ColorsFromValue = map[float64]zgeo.Color{}
	return v
}

func (v *AmountView) drawBar(rect zgeo.Rect, canvas *Canvas, view View) {
	val := v.Value()
	if val != -1 {
		rect.Size.W *= val
		if rect.Size.W < 3 {
			rect.Size.W = 3
		}
		path := zgeo.PathNewRect(rect, zgeo.Size{3, 3})
		col := v.getColorForValue()
		canvas.SetColor(col, 1)
		canvas.FillPath(path)
	}
}

func AmountViewBarNew(width float64) *AmountView {
	v := &AmountView{}
	v.CustomView.init(v, "amount")
	v.SetBGColor(zgeo.ColorNewGray(0.9, 1))
	v.SetColor(zgeo.ColorNew(0.4, 0.4, 0.8, 1))
	v.SetCorner(3)
	v.CustomView.SetMinSize(zgeo.Size{width, 10})
	v.SetDrawHandler(v.drawBar)
	v.ColorsFromValue = map[float64]zgeo.Color{}
	return v
}
