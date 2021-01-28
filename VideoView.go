package zui

import "github.com/torlangballe/zutil/zgeo"

type VideoView struct {
	NativeView
	baseVideoView
	LongPresser
	canvas      *Canvas
	minSize     zgeo.Size
	pressed     func()
	longPressed func()

	draw func(rect zgeo.Rect, canvas *Canvas, view View)
}

func VideoViewNew(minSize zgeo.Size) *VideoView {
	v := &VideoView{}
	v.Init(v, minSize)
	return v
}

func (v *VideoView) CalculatedSize(total zgeo.Size) zgeo.Size {
	return v.minSize
}
