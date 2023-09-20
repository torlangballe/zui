// The server variant of App is an App (program) in it's own right, but also contains functionality to
// serve a wasm app to a browser.
// It is invoked with ServeZUIWasm (below), which uses a filesRedirector (below) instance to handle serving the wasm, html and assets.

//go:build !js && !catalyst && server

package zapp

import (
	"embed"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/torlangballe/zutil/zbuild"
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
	zrpc.Register(AppCalls{})
	zlog.Info("add zapp fs")
	AllWebFS.Add(wwwFS)

	var beforeWWW string
	stat := zrest.StaticFolderPathFunc("")
	zstr.HasSuffix(stat, "/www", &beforeWWW)                           // we remove www because os.DirFS does not include it in path names, but //go:embed www does...
	AllWebFS = append(zfile.MultiFS{os.DirFS(beforeWWW)}, AllWebFS...) // we insert the disk system first, so we can override embeded
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
	if strings.HasSuffix(spath, ".md") {
		m := MakeMarkdownConverter()
		m.ServeAsHTML(w, req, "www/"+spath)
		return
	}

	if spath == "" { // hack to replicate how http.ServeFile serves index.html if serving empty folder at root level
		spath = "index.html"
	}
	smime := mime.TypeByExtension(path.Ext(spath))
	if spath == "main.wasm.gz" {
		zlog.Info("Serve WASM.gz:", spath, req.Method)
		// If we are serving the gzip'ed wasm file, set encoding to gzip and type to wasm
		smime = "application/wasm"
		w.Header().Set("Content-Encoding", "gzip")
		// w.Header().Set("Expires", time.Now().Add(time.Hour).Format(time.RFC1123))
		// w.Header().Set("Vary", "Accept-Encoding")
		// w.Header().Set("Accept-Ranges", "bytes")
		// w.Header().Set("X-Frame-Options", "sameorigin")
		// w.Header().Set("X-Content-Type-Options", "nosniff")
		// w.Header().Set("Age", "586183")
		// w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		// w.Header().Set("Cross-Origin-Embedder-Policy-Report-Only", `credentialless; report-to="geo-earth-eng-team"`)
		// w.Header().Set("Cross-Origin-Opener-Policy-Report-Only", `same-origin; report-to="geo-earth-eng-team"`)
		// w.Header().Set("Cross-Origin-Opener-Policy", `same-origin; report-to="geo-earth-eng-team"`)
	}
	if smime != "" {
		w.Header().Set("Content-Type", smime)
	}
	w.Header().Set("Cache-Control", "public, max-age=604800")
	if !zbuild.Build.At.IsZero() {
		w.Header().Set("Last-Modified", zbuild.Build.At.Format(time.RFC1123))
		// w.Header().Set("ETag", zstr.HashTo64Hex(zbuild.Build.At.Format(time.RFC1123)))
	}
	//f, err := AllWebFS.Open("www/" + spath)
	f, len, err := zfile.ReaderFromFileInFS(AllWebFS, "www/"+spath)
	if len != 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(len, 10))
	}
	// zlog.Info("FilesRedir2:", spath, err)
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

func ManualFlattened(m *zmarkdown.MarkdownConverter, w io.Writer, name string, output zmarkdown.OutputType) error {
	return m.Convert(w, name, output)
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

func MakeMarkdownConverter() zmarkdown.MarkdownConverter {
	var m zmarkdown.MarkdownConverter
	m.Variables = GetDocumentationValues()
	m.Dir = "www/doc/"
	m.FileSystem = AllWebFS
	// zlog.Info("makeMarkdownConverter fs:", m.FileSystem)
	m.PartNames = []string{
		"setup.shared.md",
	}
	return m
}
