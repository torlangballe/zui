// The server variant of App is an App (program) in it's own right, but also contains functionality to
// serve a wasm app to a browser.
// It is invoked with ServeZUIWasm (below), which uses a FilesRedirector (below) instance to handle serving the wasm, html and assets.

//go:build !js && !catalyst && server

package zapp

import (
	"embed"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmarkdown"
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

//go:embed www
var wwwFS embed.FS

var Calls = new(AppCalls)

func Init() {
	zrpc2.Register(Calls)
	if zfile.NotExist(zrest.StaticFolder) {
		os.Mkdir(zrest.StaticFolder, os.ModeDir|0755)
	}
}

// FilesRedirector's ServeHTTP serves everything in www, handling directories, * wildcards, and auto-translating .md (markdown) files to html
func (r FilesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	const filePathPrefix = zrest.StaticFolder + "/"
	spath := req.URL.Path
	var redirectToDir bool
	if spath == strings.TrimRight(zrest.AppURLPrefix, "/") {
		redirectToDir = true
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
	// if spath != "" {
	// 	if strings.Contains(filepath, "*") {
	// 		files, _ := zfile.GetFilesFromPath(filepath, false)
	// 		if len(files) > 0 {
	// 			filepath = files[0]
	// 		}
	// 	}
	// 	if zfile.IsFolder(filepath) && r.ServeDirectories {
	// 		files, err := zfile.GetFilesFromPath(filepath, true)
	// 		if err != nil {
	// 			zlog.Error(err)
	// 			return
	// 		}
	// 		str := strings.Join(files, "\n")
	// 		zrest.AddCORSHeaders(w, req)
	// 		io.WriteString(w, str)
	// 		return
	// 	}
	// }
	if filepath == zrest.StaticFolder {
		filepath = zrest.StaticFolder + "/index.html"
	}

	if zfile.Exists(filepath) {
		// zlog.Info("FilesServe:", req.URL.Path, filepath, zfile.Exists(filepath))
		http.ServeFile(w, req, filepath)
		return
	}
	if redirectToDir {
		// zlog.Info("Serve embed:", spath)
		localRedirect(w, req, zrest.AppURLPrefix)
		return
	}
	// zlog.Info("FilesRedir2:", req.URL.Path, spath)

	if spath == "" { // hack to replicate how http.ServeFile serves index.html if serving empty folder at root level
		spath = "index.html"
	}
	data, err := wwwFS.ReadFile(zrest.StaticFolder + "/" + spath)
	// zlog.Info("FSREAD:", zrest.StaticFolder+"/"+spath, err, len(data), req.URL.String())
	if err == nil {
		_, err := w.Write(data)
		if err != nil {
			zlog.Error(err, "write to ResponseWriter from embedded")
		}
		return
	}
	// zlog.Info("Serve app:", path, filepath)
}

func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
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
		http.ServeFile(w, r, zrest.StaticFolder+"/favicon.ico")
	})
}

func (c *AppCalls) GetTimeInfo(u zrpc2.Unused, info *LocationTimeInfo) error {
	t := time.Now().Local()
	name, offset := t.Zone()
	info.JSISOTimeString = t.UTC().Format(ztime.JavascriptISO)
	info.ZoneName = name
	info.ZoneOffsetSeconds = offset
	return nil
}

func ManualAsPDF(w http.ResponseWriter, req *http.Request, name string, tableOC bool, parts []string) {
	defer req.Body.Close()
	values := req.URL.Query()
	raw := values.Get("raw")
	md := (raw == "md")
	html := (raw == "html")
	prefix := zrest.StaticFolder + "/doc/"
	fullmd, err := zmarkdown.FlattenMarkdown(prefix, parts, tableOC)
	// zlog.Info("MD:\n", fullmd)
	if err != nil {
		zrest.ReturnAndPrintError(w, req, http.StatusInternalServerError, err, "building pdf", name)
		return
	}
	if md {
		w.Write([]byte(fullmd))
		return
	}
	if html {
		html, err := zmarkdown.ConvertToHTML(fullmd, name, "", GetDocumentationValues())
		if err != nil {
			zrest.ReturnAndPrintError(w, req, http.StatusInternalServerError, err, "converting")
			return
		}
		w.Write([]byte(html))
	}
	spdf, err := zmarkdown.ConvertToPDF(fullmd, "Bridgetech QTT", zrest.StaticFolder+"/doc/", GetDocumentationValues())
	if err != nil {
		zrest.ReturnAndPrintError(w, req, http.StatusInternalServerError, "error converting manual to pdf")
		return
	}
	w.Write([]byte(spdf))
}
