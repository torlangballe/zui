//go:build zui

package zwidgets

import (
	"time"

	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zstyle"
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

func NewActivityView(size zgeo.Size, col zgeo.Color) *ActivityView {
	var post string
	if !col.Valid {
		col = zgeo.ColorBlack
		if zstyle.Dark {
			col = zgeo.ColorWhite
		}
	}
	if col != zgeo.ColorBlack {
		post = "-white"
	}
	path := "images/zcore/activity" + post + ".png"
	v := &ActivityView{}
	if col != zgeo.ColorBlack && col != zgeo.ColorWhite {
		v.TintColor = col
	}
	v.Init(v, true, nil, path, size)
	v.rotationSecs = 1.5
	v.repeater = ztimer.RepeaterNew()
	v.SetInteractive(false)
	v.SetAlpha(0)
	return v
}

func (v *ActivityView) StopOrStart(start bool) {
	if start {
		v.Start()
	} else {
		v.Stop()
	}
}

func (v *ActivityView) Start() {
	if !v.start.IsZero() {
		return
	}
	v.start = time.Now()
	v.repeater.Set(0.1, false, func() bool {
		v.SetAlpha(1)
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
