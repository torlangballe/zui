package zui

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/znet"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zstr"
)

type nativeApp struct {
}

// AppURL returns the url that invoked this app
func AppURL() string {
	return WindowJS.Get("location").Get("href").String()
}

// MainArgs returns the path of browser and url parameters as args map of max one parameter of each key
func AppMainArgs() (path string, args map[string]string) {
	args = map[string]string{}
	u, err := url.Parse(AppURL())
	zlog.AssertNotError(err)
	path = u.Path
	zstr.HasPrefix(path, zrest.AppURLPrefix+"page/", &path)
	path = strings.TrimRight(path, "/")
	for k, v := range u.Query() {
		args[k] = v[0]
	}
	return
}

func AppSetUIDefaults(useTokenAuth bool) (part string, args map[string]string) {
	url, _ := url.Parse(AppURL())
	host, _ := znet.GetHostAndPort(url)
	args = map[string]string{}
	for k, v := range url.Query() {
		args[k] = v[0]
	}
	// fmt.Println("AppSetUIDefaults:", url.Query, args, AppURL())
	DocumentationPathPrefix = "http://" + host + zrest.AppURLPrefix + "doc/"
	zrpc.ToServerClient = zrpc.NewClient(useTokenAuth, 0)
	zrpc.ToServerClient.SetAddressFromHost(url.Scheme, host)
	port, _ := strconv.Atoi(args["zrpcport"])
	if port != 0 {
		zrpc.ToServerClient.Port = port
	}
	DefaultLocalKeyValueStore = KeyValueStoreNew(true)
	part, args = AppMainArgs()
	if zbool.FromString(args["zdebug"], false) {
		DebugMode = true
	}
	if zbool.FromString(args["ztest"], false) {
		zlog.IsInTests = true
	}
	return
}

func appNew(a *App) {

}
