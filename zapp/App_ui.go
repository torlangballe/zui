//go:build zui && !catalyst

package zapp

import (
	"time"

	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type nativeApp struct {
}

const offsetKey = "zui.ServerTimeOffset"

var (
	ServerTimeDifference time.Duration
	getTimeCount         int
)

func NewCurrentTimeLabel() *zlabel.Label {
	if zlocale.IsDisplayServerTime.Get() && ServerTimezoneName == "" {
		offset, got := zkeyvalue.DefaultStore.GetInt(offsetKey, 0)
		if !got {
			_, offset = time.Now().In(time.Local).Zone()
		}
		ztime.ServerTimezoneOffsetSecs = offset
	}
	label := zlabel.New("")
	label.SetObjectName("time")
	label.SetFont(zgeo.FontDefault(0))
	label.SetColor(zgeo.ColorNewGray(0.3, 1))
	label.SetMinWidth(145)
	label.SetTextAlignment(zgeo.Right)
	label.SetPressedDownHandler("", 0, func() bool {
		toggleTimeZoneMode(label)
		return true
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
	zkeyvalue.DefaultStore.SetInt(info.ZoneOffsetSeconds, offsetKey, true)
	// zlog.Info("fetchTimeInfo:", ServerTimezoneName, ztime.ServerTimezoneOffsetSecs)
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
	if zlocale.IsDisplayServerTime.Get() {
		t = t.Add(ServerTimeDifference)
		loc := time.FixedZone(ServerTimezoneName, ztime.ServerTimezoneOffsetSecs)
		t = t.In(loc)
		// zlog.Info("updateCurrentTime:", t, loc, t.Location(), loc == nil, ServerTimezoneName)
		format += "-07"
		str = "☁️"
	}
	str += t.Format(format)
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
