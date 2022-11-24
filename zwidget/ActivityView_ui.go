//go:build zui

package zwidget

import (
	"time"

	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type ActivityView struct {
	zimageview.ImageView
	rotationSecs  float64
	AlwaysVisible bool
	repeater      *ztimer.Repeater
	start         time.Time
}

func NewActivityView(size zgeo.Size) *ActivityView {
	v := &ActivityView{}
	v.Init(v, nil, "images/activity.png", size)
	v.rotationSecs = 1.5
	v.repeater = ztimer.RepeaterNew()
	v.SetAlpha(0)
	return v
}

func (v *ActivityView) Start() {
	if !v.Presented {
		return
	}
	v.SetAlpha(1)
	v.start = time.Now()
	v.repeater.Set(0.1, false, func() bool {
		t := ztime.Since(v.start)
		deg := 360 * (t / v.rotationSecs)
		v.Rotate(deg)
		return true
	})
}

func (v *ActivityView) Stop() {
	v.repeater.Stop()
	if !v.AlwaysVisible {
		v.SetAlpha(0)
	}
}

func (v *ActivityView) IsStopped() bool {
	return v.repeater.IsStopped()
}
