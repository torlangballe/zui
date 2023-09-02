// The server variant of App is an App (program) in it's own right, but also contains functionality to
// serve a wasm app to a browser.
// It is invoked with ServeZUIWasm (below), which uses a filesRedirector (below) instance to handle serving the wasm, html and assets.

//go:build !js && !catalyst && server

package zapp

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmarkdown"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
)

// nativeApp is used in App in zapp.go, so must be defined even if empty:
type nativeApp struct {
}

type AppCalls zrpc.CallsBase

// filesRedirector is a type that can handle serving files
type filesRedirector struct {
	Override         func(w http.ResponseWriter, req *http.Request, filepath string) bool // Override is a method to handle special cases of files, return true if handled
	ServeDirectories bool
	Router           *mux.Router // if ServeDirectories is true, it serves content list of directory
}

//go:embed www
var wwwFS embed.FS

var AllWebFS zfile.MultiFS

func Init() {
	var beforeWWW string
	zrpc.Register(AppCalls{})
	stat := zrest.StaticFolderPathFunc("")
	if zfile.NotExist(stat) {
		os.Mkdir(stat, os.ModeDir|0755)
	}
	zstr.HasSuffix(stat, "/www", &beforeWWW) // we remove www because os.DirFS does not include it, but //go:embed www does...
	AllWebFS.Add(os.DirFS(beforeWWW))
	AllWebFS.Add(wwwFS)
}

// filesRedirector's ServeHTTP serves everything in zrest.StaticFolderPathFunc()
func (r filesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	spath := req.URL.Path
	if spath == strings.TrimRight(zrest.AppURLPrefix, "/") {
		localRedirect(w, req, zrest.AppURLPrefix)
		req.Body.Close()
		return
	}
	// zlog.Info("FilesRedir1:", req.URL.Path, spath, strings.Trim(zrest.AppURLPrefix, "/"))
	zstr.HasPrefix(spath, zrest.AppURLPrefix, &spath)
	if r.Override != nil {
		if r.Override(w, req, spath) {
			req.Body.Close()
			return
		}
	}

	// zlog.Info("FilesRedir1:", spath)

	if spath == "" { // hack to replicate how http.ServeFile serves index.html if serving empty folder at root level
		spath = "index.html"
	}
	if spath == "main.wasm.gz" {
		zlog.Info("Serve WASM.gz:", spath)
		// If we are serving the gzip'ed wasm file, set encoding to gzip and type to wasm
		w.Header().Set("Content-Type", "application/wasm")
		w.Header().Set("Content-Encoding", "gzip")
	}
	f, err := AllWebFS.Open("www/" + spath)
	// f, err := AllWebFS.Open("www/" + spath)
	if !zlog.OnError(err, spath) {
		_, err := io.Copy(w, f)
		zlog.OnError(err, spath)
	}
}

// localRedirect redirects empty path to directory (I think)
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

func ServeZUIWasm(router *mux.Router, serveDirs bool, override func(w http.ResponseWriter, req *http.Request, filepath string) bool) {
	f := &filesRedirector{
		ServeDirectories: serveDirs,
		Override:         override,
	}
	zrest.AddSubHandler(router, "", f)

	// zlog.Info("HandleApp:", zrest.AppURLPrefix)
	//	route := router.PathPrefix(zrest.AppURLPrefix)
	//	route.Handler(f)
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, zrest.StaticFolderPathFunc("/favicon.ico"))
	})
}

func (AppCalls) GetTimeInfo(in zrpc.Unused, info *LocationTimeInfo) error {
	t := time.Now().Local()
	name, offset := t.Zone()
	info.JSISOTimeString = t.UTC().Format(ztime.JavascriptISO)
	info.ZoneName = name
	info.ZoneOffsetSeconds = offset
	return nil
}

func ManualAsPDF(w http.ResponseWriter, req *http.Request, name string, tableOC bool, parts []string) {
	values := req.URL.Query()
	raw := values.Get("raw")
	md := (raw == "md")
	html := (raw == "html")
	prefix := zrest.StaticFolderPathFunc("doc/")
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
		html, err := zmarkdown.ConvertToHTML(fullmd, prefix, name, "", GetDocumentationValues())
		if err != nil {
			zrest.ReturnAndPrintError(w, req, http.StatusInternalServerError, err, "converting")
			return
		}
		w.Write([]byte(html))
		return
	}
	spdf, err := zmarkdown.ConvertToPDF(fullmd, prefix, name, prefix, GetDocumentationValues())
	if err != nil {
		zrest.ReturnAndPrintError(w, req, http.StatusInternalServerError, "error converting manual to pdf")
		return
	}
	w.Write([]byte(spdf))
}

func (AppCalls) GetTopImages(args *zrpc.Unused, reply *[]string) error {
	statSlash := zrest.StaticFolderPathFunc("/")
	zfile.Walk(zrest.StaticFolderPathFunc("images/"), "*.png", zfile.WalkOptionsNone, func(fpath string, info os.FileInfo) error {
		zstr.HasPrefix(fpath, statSlash, &fpath)
		*reply = append(*reply, fpath)
		return nil
	})
	for _, f := range AllWebFS {
		fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() || !strings.HasSuffix(path, ".png") {
				return nil
			}
			zstr.HasPrefix(path, statSlash, &path)
			zstr.AddToSet(reply, path)
			return nil
		})
	}
	return nil
}

func handleSetVerbose(w http.ResponseWriter, req *http.Request) {
	on := zrest.GetBoolVal(req.URL.Query(), "on")
	zlog.Info("handleSetVerbose", on, req.Method)
	var set string
	if on {
		zlog.PrintPriority = zlog.VerboseLevel
		set = "verbose"
	} else {
		zlog.PrintPriority = zlog.DebugLevel
		set = "debug"
	}
	fmt.Fprintln(w, "zlog priority set to", set)
}

func SetVerboseLogHandler(router *mux.Router) {
	zrest.AddHandler(router, "zlogverbose", handleSetVerbose).Methods("GET")
}
