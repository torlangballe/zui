//go:build !js && zui

package zmap

import (
	"github.com/torlangballe/zutil/zgeo"
)

type baseMapView struct {
}

func (v *MapView) Init(view zview.View, center zgeo.Pos, zoom int) {
}
