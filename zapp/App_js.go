package zapp

import (
	"net/url"
	"strings"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdebug"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc"
)

// URL returns the url that invoked this app
func URL() *url.URL {
	u, err := url.Parse(URLString())
	zlog.AssertNotError(err)
	return u
}

func URLString() string {
	return zdom.WindowJS.Get("location").Get("href").String()
}

// MainArgs returns the path of browser and url parameters as args map of max one parameter of each key
func MainArgs() (path string, args map[string]string) {
	args = map[string]string{}
	u := URL()
	path = strings.TrimRight(u.Path, "/")
	for k, v := range u.Query() {
		args[k] = v[0]
	}
	return
}

func URLStub() string {
	u := URL()
	u.Path = ""
	u.RawQuery = ""
	return u.String()
}

// SetUIDefaults sets up an app, uncluding some sensible defaults for rpc communicated with server counterpart
func SetUIDefaults(useRPC bool) (path string, args map[string]string) {
	url := URL()
	url.Path = ""
	host := url.Host
	args = map[string]string{}
	for k, v := range url.Query() {
		if k == "zdev" && v[0] == "1" {
			zui.DebugOwnerMode = true
		}
		args[k] = v[0]
	}
	DownloadPathPrefix = url.Scheme + "://" + host + zrest.AppURLPrefix
	zwidgets.DocumentationPathPrefix = DownloadPathPrefix + "doc/"
	zlog.Info("zapp: DocumentationPathPrefix", zwidgets.DocumentationPathPrefix)
	if useRPC {
		url.RawQuery = ""
		url.Path = ""
		zrpc.MainClient = zrpc.NewClient(url.String(), "")
	}
	path, args = MainArgs()
	if zbool.FromString(args["zdebug"], false) {
		zui.DebugMode = true
	}
	if zbool.FromString(args["ztest"], false) {
		zdebug.IsInTests = true // for testing gui
	}
	return
}
