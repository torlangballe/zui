//go:build zui && !catalyst

package zapp

import (
	"time"

	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type nativeApp struct {
}

var (
	ServerTimeDifference time.Duration
	getTimeCount         int
)

func NewCurrentTimeLabel() *zlabel.Label {
	label := zlabel.New("")
	label.SetObjectName("time")
	label.SetFont(zgeo.FontDefault(0))
	label.SetColor(zgeo.ColorNewGray(0.3, 1))
	label.SetMinWidth(145)
	label.SetTextAlignment(zgeo.Right)
	label.SetPressedDownHandler("", func() {
		toggleTimeZoneMode(label)
	})
	ztimer.RepeatForeverNow(1, func() {
		if getTimeCount%10 == 0 {
			if !fetchTimeInfo() {
				getTimeCount = 0
				return
			}
		}
		getTimeCount++
		updateCurrentTime(label)
	})
	return label
}

func fetchTimeInfo() bool {
	var info TimeInfo
	start := time.Now()
	err := zrpc.MainClient.Call("AppCalls.GetTimeInfo", nil, &info)
	if zlog.OnError(err) {
		return false
	}
	since := time.Since(start)
	ServerTimezoneName = info.ZoneName
	ztime.ServerTimezoneOffsetSecs = info.ZoneOffsetSeconds
	if since > time.Second {
		return false
	}
	t, err := time.Parse(ztime.JavascriptISO, info.JSISOTimeString)
	if err != nil {
		zlog.Error("parse", err)
		return false
	}
	mid := start.Add(since / 2)
	ServerTimeDifference = mid.Sub(t)
	return true
}

func updateCurrentTime(label *zlabel.Label) {
	if ServerTimezoneName == "" {
		label.SetText("")
		return
	}
	format := "15:04:05"
	t := time.Now()
	str := ""
	// zlog.Info("updateCurrentTime:", zlocale.IsDisplayServerTime.Get())
	if zlocale.IsDisplayServerTime.Get() {
		t = t.Add(ServerTimeDifference)
		loc, _ := time.LoadLocation(ServerTimezoneName)
		if loc != nil {
			t = t.In(loc)
		}
		format += "-07"
		str = "☁️"
	}
	str += time.Now().Format(format)
	col := zgeo.ColorBlack
	if ServerTimeDifference > time.Second*2 {
		col = zgeo.ColorRed
	}
	label.SetColor(col)
	label.SetText(str)
}

func appNew(a *App) {
}

func toggleTimeZoneMode(label *zlabel.Label) {
	zlog.Info("toggleTimeZoneMode:", zlocale.IsDisplayServerTime != nil)
	zlocale.IsDisplayServerTime.Set(!zlocale.IsDisplayServerTime.Get(), false)
	zwindow.GetMain().Reload()
}
