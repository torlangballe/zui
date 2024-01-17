//go:build zui && !catalyst

package zapp

import (
	"time"

	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type nativeApp struct {
}

var ServerTimeDifference time.Duration

func init() {
	ServerTimeDifferenceSeconds.AddChangedHandler(func() {
		handleTimeInfoChanged()
	})
	ServerTimezoneName.AddChangedHandler(func() {
		handleTimeInfoChanged()
	})
	ServerTimeJSISO.AddChangedHandler(func() {
		handleTimeInfoChanged()
	})
	zlocale.IsDisplayServerTime.AddChangedHandler(func() {
		handleTimeInfoChanged()
	})
}

func handleTimeInfoChanged() error {
	// zlog.Info("handleTimeInfoChanged:", ServerTimeDifferenceSeconds.Get(), ServerTimezoneName.Get(), ServerTimeJSISO.Get(), zlocale.IsDisplayServerTime.Get())
	stime := ServerTimeJSISO.Get()
	if stime != "" {
		t, err := time.Parse(ztime.JavascriptISO, stime)
		if err != nil {
			return zlog.Error(err, "parse")
		}
		ServerTimeDifference = t.Sub(time.Now()) / 2
	}
	return nil
}

func NewCurrentTimeLabel() *zlabel.Label {
	label := zlabel.New("")
	label.SetObjectName("time")
	label.SetFont(zgeo.FontDefault(-2))
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
	d := !zlocale.IsDisplayServerTime.Get()
	zlocale.IsDisplayServerTime.Set(d, true)
	updateCurrentTime(label)
	zwindow.GetMain().Reload()
	ztimer.StartIn(2, func() {
		zlog.Info("toggleTimeZoneMode", d)
	})
}

func updateCurrentTime(label *zlabel.Label) {
	t := time.Now()
	t = t.Add(ServerTimeDifference)
	// zlog.Info("updateCurrentTime:", ServerTimezoneName.Get())
	if ServerTimezoneName.Get() != "" {
		loc, _ := time.LoadLocation(ServerTimezoneName.Get())
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
