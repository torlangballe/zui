// The server variant of App is an App (program) in it's own right, but also contains functionality to
// serve a wasm app to a browser.
// It is invoked with ServeZUIWasm (below), which uses a FilesRedirector (below) instance to handle serving the wasm, html and assets.

//go:build !js && !catalyst

package zapp

import (
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zstr"
)

// nativeApp is used in App in zapp.go, so must be defined even if empty:
type nativeApp struct {
}

// FilesRedirector is a type that can handle serving files
type FilesRedirector struct {
	Override         func(w http.ResponseWriter, req *http.Request, filepath string) bool // Override is a method to handle special cases of files, return true if handled
	ServeDirectories bool
	Router           *mux.Router // if ServeDirectories is true, it serves content list of directory
}

// FilesRedirector's ServeHTTP serves everything in www, handling directories, * wildcards, and auto-translating .md (markdown) files to html
func (r FilesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	const filePathPrefix = "www/"
	// zlog.Info("FilesRedir1:", req.URL.Path)
	path := req.URL.Path
	zstr.HasPrefix(path, zrest.AppURLPrefix, &path)
	filepath := filePathPrefix + path
	if r.Override != nil {
		if r.Override(w, req, filepath) {
			return
		}
	}
	// zlog.Info("FilesRedir:", req.URL.Path, filepath, zfile.Exists(filepath))
	if path != "" {
		if strings.Contains(filepath, "*") {
			files, _ := zfile.GetFilesFromPath(filepath, false)
			if len(files) > 0 {
				filepath = files[0]
			}
		}
		if zfile.IsFolder(filepath) && r.ServeDirectories {
			files, err := zfile.GetFilesFromPath(filepath, true)
			if err != nil {
				zlog.Error(err)
				return
			}
			str := strings.Join(files, "\n")
			zrest.AddCORSHeaders(w, req)
			io.WriteString(w, str)
			return
		}
	}
	// zlog.Info("Serve app:", path, filepath)
	http.ServeFile(w, req, filepath)
}

func ServeZUIWasm(router *mux.Router, serveDirs bool, override func(w http.ResponseWriter, req *http.Request, filepath string) bool) {
	f := &FilesRedirector{
		ServeDirectories: serveDirs,
		Override:         override,
	}
	// zlog.Info("HandleApp:", zrest.AppURLPrefix)
	route := router.PathPrefix(zrest.AppURLPrefix)
	route.Handler(f)
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/favicon.ico")
	})
}

// URL returns the url/command that invoked this app
// func URL() string {
// 	return strings.Join(os.Args, " ")
// }
