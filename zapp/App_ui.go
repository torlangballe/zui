//go:build zui

package zapp

import (
	"net/url"
	"strings"
	"time"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zcolor"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zwidget"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/ztimer"

	"github.com/torlangballe/zutil/ztime"
)

type nativeApp struct {
}

var (
	ServerTimezoneName   string
	ServerTimeDifference time.Duration
)

func init() {
	zcolor.RegisterWidget()
}

// URL returns the url that invoked this app
func URL() string {
	return zdom.WindowJS.Get("location").Get("href").String()
}

// MainArgs returns the path of browser and url parameters as args map of max one parameter of each key
func MainArgs() (path string, args map[string]string) {
	args = map[string]string{}
	u, err := url.Parse(URL())
	zlog.AssertNotError(err)
	path = strings.TrimRight(u.Path, "/")
	for k, v := range u.Query() {
		args[k] = v[0]
	}
	return
}

// SetUIDefaults sets up an app, uncluding some sensible defaults for rpc communicated with server counterpart
func SetUIDefaults(useRPC bool) (path string, args map[string]string) {
	url, _ := url.Parse(URL())
	// host, _ := znet.GetHostAndPort(url)
	url.Path = ""
	host := url.Host
	args = map[string]string{}
	for k, v := range url.Query() {
		args[k] = v[0]
	}
	DownloadPathPrefix = "http://" + host + zrest.AppURLPrefix
	zwidget.DocumentationPathPrefix = DownloadPathPrefix + "doc/"
	if useRPC {
		url.RawQuery = ""
		url.Path = ""
		zrpc.MainClient = zrpc.NewClient(url.String(), "")
	}
	zkeyvalue.DefaultStore = zkeyvalue.NewStore(false)
	zkeyvalue.DefaultSessionStore = zkeyvalue.NewStore(true)
	path, args = MainArgs()
	if zbool.FromString(args["zdebug"], false) {
		zui.DebugMode = true
	}
	if zbool.FromString(args["ztest"], false) {
		zlog.IsInTests = true
	}
	return
}

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
