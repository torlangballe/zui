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
	ServerTimezoneName   string
	ServerTimeDifference time.Duration
)

func GetTimeInfoFromServer() error {
	var info LocationTimeInfo
	err := zrpc.MainClient.Call("AppCalls.GetTimeInfo", nil, &info)
	if err != nil {
		return zlog.Error(err, "call")
	}
	ServerTimezoneName = info.ZoneName
	ztime.ServerTimezoneOffsetSecs = info.ZoneOffsetSeconds
	t, err := time.Parse(ztime.JavascriptISO, info.JSISOTimeString)
	if err != nil {
		return zlog.Error(err, "parse")
	}
	ServerTimeDifference = t.Sub(time.Now()) / 2
	// zlog.Info("Got Time:", info.JSISOTimeString, time.Now(), "shift:", ServerTimeDifference, "offset:", ztime.ServerTimezoneOffsetSecs)
	return nil
}

func NewCurrentTimeLabel() *zlabel.Label {
	label := zlabel.New("")
	label.SetObjectName("time")
	label.SetFont(zgeo.FontDefault().NewWithSize(zgeo.FontDefaultSize - 2))
	label.SetColor(zgeo.ColorNewGray(0.5, 1))
	label.SetMinWidth(145)
	label.SetTextAlignment(zgeo.Right)
	label.SetPressedDownHandler(func() {
		toggleTimeZoneMode(label)
	})
	updateCurrentTime(label)
	ztimer.RepeatForever(1, func() {
		updateCurrentTime(label)
	})
	return label
}

func toggleTimeZoneMode(label *zlabel.Label) {
	d := !zlocale.DisplayServerTime.Get()
	zlog.Info("toggleTimeZoneMode", d)
	zlocale.DisplayServerTime.Set(d)
	updateCurrentTime(label)
	zwindow.GetMain().Reload()
}

func updateCurrentTime(label *zlabel.Label) {
	t := time.Now()
	t = t.Add(ServerTimeDifference)
	if ServerTimezoneName != "" {
		loc, _ := time.LoadLocation(ServerTimezoneName)
		if loc != nil {
			t = t.In(loc)
		}
	}
	str := ztime.GetNice(time.Now(), true)
	label.SetText(str)
	col := zgeo.ColorBlack
	if ServerTimeDifference > time.Second*4 {
		col = zgeo.ColorRed
	}
	label.SetColor(col.WithOpacity(0.7))
}
