//go:build zui

package zwidgets

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
	v.Init(v, true, nil, "images/zcore/activity.png", size)
	v.rotationSecs = 1.5
	v.repeater = ztimer.RepeaterNew()
	v.SetAlpha(0)
	return v
}

func (v *ActivityView) Start() {
	v.SetAlpha(1)
	if !v.start.IsZero() {
		return
	}
	v.start = time.Now()
	v.repeater.Set(0.1, false, func() bool {
		t := ztime.Since(v.start)
		deg := 360 * (t / v.rotationSecs)
		v.Rotate(deg)
		return true
	})
}

func (v *ActivityView) Stop() {
	v.start = time.Time{}
	v.repeater.Stop()
	if !v.AlwaysVisible {
		v.SetAlpha(0)
	}
}

func (v *ActivityView) IsStopped() bool {
	return v.repeater.IsStopped()
}

func (v *ActivityView) SetValueWithAny(val any) {
	on := val.(bool)
	// zlog.Info("ActivityView.SetValueWithAny", on)
	if on {
		v.Start()
	} else {
		v.Stop()
	}
}
