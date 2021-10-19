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
	path = strings.TrimRight(u.Path, "/")
	for k, v := range u.Query() {
		args[k] = v[0]
	}
	return
}

func AppSetUIDefaults(useTokenAuth bool, rpcPort int) (path string, args map[string]string) {
	url, _ := url.Parse(AppURL())
	host, _ := znet.GetHostAndPort(url)
	args = map[string]string{}
	for k, v := range url.Query() {
		args[k] = v[0]
	}
	DocumentationPathPrefix = "http://" + host + zrest.AppURLPrefix + "doc/"
	zrpc.ToServerClient = zrpc.NewClient(useTokenAuth, 0)
	zrpc.ToServerClient.SetAddressFromHost(url.Scheme, host)
	port, _ := strconv.Atoi(args["zrpcport"])
	if port != 0 {
		rpcPort = port
	}
	if rpcPort != 0 {
		zrpc.ToServerClient.Port = rpcPort
	}
	// fmt.Println("AppSetUIDefaults:", url.Query, args, AppURL(), zrpc.ToServerClient.Port)
	DefaultLocalKeyValueStore = KeyValueStoreNew(true)
	path, args = AppMainArgs()
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
