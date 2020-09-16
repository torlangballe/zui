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
		part := zstr.ExtractStringTilSeparator(&path, "/")
		switch part {
		case "images", "js", "css", "fonts", "templates":
			path = part + "/" + path
		}
	}
	// if path == "main.wasm.gz" {
	// 	zlog.Info("SERVE WASM:", "www/"+path, req.Header)
	// 	w.Header().Set("Content-Encoding", "gzip")
	// 	w.Header().Set("Content-Type", "application/wasm")
	// 	file, err := os.Open("www/" + path)
	// 	if err != nil {
	// 		zlog.Error(err, "open")
	// 	}
	// 	_, err = io.Copy(w, file)
	// 	if err != nil {
	// 		zlog.Error(err, "copy")
	// 	}
	// 	return
	// }
	http.ServeFile(w, req, "www/"+path)
}

func AppServeZUIWasm() {
	http.Handle("/", FilesRedirector{})
}
