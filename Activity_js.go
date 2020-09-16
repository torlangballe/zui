package zui

import (
	"fmt"
	"math"

	"github.com/torlangballe/zutil/zgeo"
)

type ActivityView struct {
	CustomView
	stop         bool
	rotationSecs float64
}

func ActivityNew(colCircle, colPart zgeo.Color, size float64) *ActivityView {
	v := &ActivityView{}
	v.CustomView.init(v, "activity")
	v.SetMinSize(zgeo.SizeBoth(size))
	v.SetCorner(size)
	v.stop = true
	w := math.Floor(size / 5)
	v.SetStroke(w, colCircle)
	style := v.style()
	style.Set("borderTop", fmt.Sprintf("%dpx solid %s", int(w), makeRGBAString(colPart)))
	style.Set("boxSizing", "border-box")
	v.rotationSecs = 1.5

	return v
}

func (v *ActivityView) ReadyToShow(beforeWindow bool) {
	// zlog.Info("AV ReadyToShow!:", v.stop, beforeWindow)
	if !v.stop && !beforeWindow {
		v.Start()
	}
}

func (v *ActivityView) Start() {
	v.stop = false
	if !v.Presented {
		return
	}
	v.Show(true)
	Animate(v, 99999, func(secPos float64) bool {
		if v.stop {
			return false
		}
		v.RotateDeg(secPos * 360 / v.rotationSecs)
		return true
	})
}

func (v *ActivityView) Stop() {
	v.stop = true
	v.Show(false)
}
