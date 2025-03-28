package zapp

import (
	"net/url"
	"strings"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdebug"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zerrors"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zstr"
)

var cookies map[string]string

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

func Cookies() map[string]string {
	if cookies == nil {
		params := zdom.DocumentJS.Get("cookie")
		if params.IsUndefined() {
			return map[string]string{}
		}
		cookies = zstr.GetParametersFromURLArgString(params.String())
		zlog.Info("Got Cookies:", cookies)
	}
	return cookies
}

func URLStub() string {
	u := URL()
	u.Path = ""
	u.RawQuery = ""
	return u.String()
}

func guiRestartHandler(err error) {
	if err == nil {
		return
	}
	dict := zdict.Dict{}
	ce := zerrors.MakeContextError(dict, "GUI Restart", err)
	callErr := zrpc.MainClient.Call("AppCalls.SetGUIError", ce, nil)
	zlog.OnError(callErr)
}

// SetUIDefaults sets up an app, uncluding some sensible defaults for rpc communicated with server counterpart
func SetUIDefaults(useRPC bool) (path string, args map[string]string) {
	url := URL()
	args = map[string]string{}
	for k, v := range url.Query() {
		if k == "zdev" && v[0] == "1" {
			zui.DebugOwnerMode = true
		}
		args[k] = v[0]
	}
	url.RawQuery = ""
	DownloadPathPrefix = url.String()
	zwidgets.DocumentationPathPrefix = DownloadPathPrefix + "doc/"
	path, args = MainArgs()
	if zbool.FromString(args["ztest"], false) {
		zdebug.IsInTests = true // for testing gui
	}
	if useRPC {
		surl := url.String()
		zstr.HasSuffix(surl, zrest.AppURLPrefix, &surl)
		zrpc.MainClient = zrpc.NewClient(surl, "")
		zdebug.HandleRestartFunc = guiRestartHandler
	}
	return
}
