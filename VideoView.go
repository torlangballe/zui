//go:build zui

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
	renderCanvas     *Canvas
	maxSize          zgeo.Size
	pressed          func()
	longPressed      func()
	StreamingStarted func()
	Overlay          View

	draw func(rect zgeo.Rect, canvas *Canvas, view View)
}

func VideoViewNew(maxSize zgeo.Size) *VideoView {
	v := &VideoView{}
	v.Init(v, maxSize)
	return v
}

func (v *VideoView) CalculatedSize(total zgeo.Size) zgeo.Size {
	if !v.StreamSize.IsNull() {
		s := v.StreamSize.ShrunkInto(v.maxSize)
		return s
	}
	return v.maxSize
}

func (v *VideoView) SetMaxSize(max zgeo.Size) {
	v.maxSize = max
}

func (v *VideoView) SetRect(rect zgeo.Rect) {
	v.NativeView.SetRect(rect)
	if v.Overlay != nil {
		v.Overlay.SetRect(rect)
	}
}
