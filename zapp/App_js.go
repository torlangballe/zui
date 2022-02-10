package zapp

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/znet"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc"
)

type nativeApp struct {
}

// URL returns the url that invoked this app
func URL() string {
	return zui.WindowJS.Get("location").Get("href").String()
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

func SetUIDefaults(useTokenAuth bool, rpcPort int) (path string, args map[string]string) {
	url, _ := url.Parse(URL())
	host, _ := znet.GetHostAndPort(url)
	args = map[string]string{}
	for k, v := range url.Query() {
		args[k] = v[0]
	}
	DownloadPathPrefix = "http://" + host + zrest.AppURLPrefix
	zui.DocumentationPathPrefix = DownloadPathPrefix + "doc/"
	zrpc.ToServerClient = zrpc.NewClient(useTokenAuth, 0)
	zrpc.ToServerClient.SetAddressFromHost(url.Scheme, host)
	port, _ := strconv.Atoi(args["zrpcport"])
	if port != 0 {
		rpcPort = port
	}
	if rpcPort != 0 {
		zrpc.ToServerClient.Port = rpcPort
	}
	// fmt.Println("app.SetUIDefaults:", url.Query, args, URL(), zrpc.ToServerClient.Port)
	zui.DefaultLocalKeyValueStore = zui.KeyValueStoreNew(true)
	path, args = MainArgs()
	if zbool.FromString(args["zdebug"], false) {
		zui.DebugMode = true
	}
	if zbool.FromString(args["ztest"], false) {
		zlog.IsInTests = true
	}
	return
}
