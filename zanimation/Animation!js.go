//go:build !js && zui

package zanimation

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

func Animate(view zview.View, secs float64, handler func(secsPos float64) bool)        {}
func Transform(nv *zview.NativeView, dir zgeo.Pos, secs float64, removeViewAfter bool) {}
