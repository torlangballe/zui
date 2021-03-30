// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type MapView struct {
	NativeView
	baseMapView
	LongPresser
	canvas      *Canvas
	minSize     zgeo.Size
	pressed     func()
	longPressed func()

	draw func(rect zgeo.Rect, canvas *Canvas, view View)
}

func MapViewNew(center zgeo.Pos, zoom int) *MapView {
	v := &MapView{}
	v.Init(v, center, zoom)
	return v
}

func (v *MapView) CalculatedSize(total zgeo.Size) zgeo.Size {
	zlog.Info("mapview cs:", v.minSize)
	return v.minSize
}
