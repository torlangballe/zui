package zapp

import (
	"net/url"
	"strings"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zwidget"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc2"
)

type nativeApp struct {
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
func SetUIDefaults(useTokenAuth, useRPC bool) (path string, args map[string]string) {
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
		zrpc2.MainClient = zrpc2.NewClient(url.String(), "")
	}
	// fmt.Println("app.SetUIDefaults:", url.Query, args, URL(), zrpc.ToServerClient.Port)
	zkeyvalue.DefaultStore = zkeyvalue.NewStore(true)
	path, args = MainArgs()
	if zbool.FromString(args["zdebug"], false) {
		zui.DebugMode = true
	}
	if zbool.FromString(args["ztest"], false) {
		zlog.IsInTests = true
	}
	return
}
