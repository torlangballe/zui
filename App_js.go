package zui

import (
	"net/url"
	"strings"

	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
)

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
	zstr.HasPrefix(path, "/page/", &path)
	path = strings.TrimRight(path, "/")
	for k, v := range u.Query() {
		args[k] = v[0]
	}
	return
}