// The server variant of App is an App (program) in it's own right, but also contains functionality to
// serve a wasm app to a browser.
// It is invoked with ServeZUIWasm (below), which uses a FilesRedirector (below) instance to handle serving the wasm, html and assets.

//go:build !js && !catalyst && server

package zapp

import (
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc2"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
)

// nativeApp is used in App in zapp.go, so must be defined even if empty:
type nativeApp struct {
}

type AppCalls zrpc2.CallsBase

// FilesRedirector is a type that can handle serving files
type FilesRedirector struct {
	Override         func(w http.ResponseWriter, req *http.Request, filepath string) bool // Override is a method to handle special cases of files, return true if handled
	ServeDirectories bool
	Router           *mux.Router // if ServeDirectories is true, it serves content list of directory
}

var Calls = new(AppCalls)

func Init() {
	zrpc2.Register(Calls)
}

// FilesRedirector's ServeHTTP serves everything in www, handling directories, * wildcards, and auto-translating .md (markdown) files to html
func (r FilesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	const filePathPrefix = "www/"
	spath := req.URL.Path
	if spath == strings.TrimRight(zrest.AppURLPrefix, "/") {
		spath += "/"
	}
	// zlog.Info("FilesRedir1:", req.URL.Path, spath, strings.Trim(zrest.AppURLPrefix, "/"))
	zstr.HasPrefix(spath, zrest.AppURLPrefix, &spath)
	filepath := path.Join(filePathPrefix, spath)
	if r.Override != nil {
		if r.Override(w, req, filepath) {
			return
		}
	}
	// zlog.Info("FilesRedir:", req.URL.Path, filepath, zfile.Exists(filepath))
	if spath != "" {
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
	zrest.AddSubHandler(router, "", f)

	// zlog.Info("HandleApp:", zrest.AppURLPrefix)
	//	route := router.PathPrefix(zrest.AppURLPrefix)
	//	route.Handler(f)
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/favicon.ico")
	})
}

func (c *AppCalls) GetTimeInfo(u zrpc2.Unused, info *LocationTimeInfo) error {
	t := time.Now()
	name, offset := t.Zone()
	info.JSISOTimeString = t.Format(ztime.JavascriptISO)
	info.ZoneName = name
	info.ZoneOffsetSeconds = offset
	zlog.Info("GetTimeInfo:", name, offset)
	return nil
}
