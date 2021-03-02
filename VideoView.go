// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type VideoView struct {
	NativeView
	baseVideoView
	Out *ImageView
	LongPresser
	streaming        bool
	StreamSize       zgeo.Size
	canvas           *Canvas
	maxSize          zgeo.Size
	pressed          func()
	longPressed      func()
	StreamingStarted func()

	draw func(rect zgeo.Rect, canvas *Canvas, view View)
}

func VideoViewNew(maxSize zgeo.Size) *VideoView {
	v := &VideoView{}
	v.Init(v, maxSize)
	return v
}

func (v *VideoView) CalculatedSize(total zgeo.Size) zgeo.Size {
	if !v.StreamSize.IsNull() {
		return v.StreamSize.ScaledInto(v.maxSize)
	}
	return v.maxSize
}
