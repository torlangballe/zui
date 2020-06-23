// +build !js

package zui

import (
	"net/http"

	"github.com/torlangballe/zutil/zstr"
)

type FilesRedirector struct {
}

func (r FilesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[1:]
	if zstr.HasPrefix(path, "page/", &path) {
		zstr.ExtractStringTilSeparator(&path, "/")
	}
	http.ServeFile(w, req, "www/"+path)
}

func AppServeZUIWasm() {
	http.Handle("/", FilesRedirector{})
}
