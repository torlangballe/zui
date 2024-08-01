//go:build zui

package zmap

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type MapView struct {
	zview.NativeView
	baseMapView
	zview.LongPresser
	canvas      *zcanvas.Canvas
	minSize     zgeo.Size
	pressed     func()
	longPressed func()

	draw func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View)
}

func MapViewNew(center zgeo.Pos, zoom int) *MapView {
	v := &MapView{}
	v.Init(v, center, zoom)
	return v
}

func (v *MapView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	zlog.Info("mapview cs:", v.minSize)
	return v.minSize, zgeo.Size{}
}
