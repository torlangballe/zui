package zui

import "github.com/torlangballe/zutil/zgeo"

type ProgressCircleView struct {
	CustomView
	FillColor zgeo.Color
	Size      zgeo.Size
}

func ProgressCircleViewNew() *ProgressCircleView {
	v := &ProgressCircleView{}
	v.CustomView.init(v, "progress")
	v.FillColor = zgeo.ColorGray
	v.SetColor(zgeo.ColorDarkGray)
	v.CustomView.SetMinSize(zgeo.Size{24, 24})
	v.DrawHandler(shapeViewDraw)
	return v
}

func (v *ProgressCircleView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	return v.GetMinSize()
}
