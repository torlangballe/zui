// The server variant of App is an App (program) in itsZ own right, but also contains functionality to
// serve a wasm app to a browser.
// It is invoked with ServeZUIWasm (below), which uses a filesRedirector (below) instance to handle serving the wasm, html and assets.

//go:build !js && !catalyst && server

package zapp

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	ua "github.com/mileusna/useragent"
	"github.com/torlangballe/zutil/zbuild"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zerrors"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zhttp"
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
	Override func(w http.ResponseWriter, req *http.Request, filepath string) bool // Override is a method to handle special cases of files, return true if handled
	Router   *mux.Router
}

//go:embed www
var wwwFS embed.FS

var (
	AllWebFS                    zfile.MultiFS
	RequestRedirector           *filesRedirector
	InlineDocumentationHeaderMD string
	CanBrotlyFunc               func(req *http.Request) bool
	HandleGUIErrorFunc          func(ci *zrpc.ClientInfo, ce zerrors.ContextError, dict zdict.Dict)
)

func Init(executor *zrpc.Executor) {
	if executor != nil {
		executor.Register(AppCalls{})
	}
	AllWebFS.Add(wwwFS, "zapp.www")

	var beforeWWW string
	stat := zrest.StaticFolderPathFunc("")
	zstr.HasSuffix(stat, "/www", &beforeWWW) // we remove www because os.DirFS does not include it in path names, but //go:embed www does...
	if beforeWWW == "" {
		beforeWWW = "."
	}
	AllWebFS.InsertFirst(os.DirFS(beforeWWW), "disk.www") // we insert the disk system first, so we can override embeded
}

func canBrotly(req *http.Request) bool {
	if CanBrotlyFunc != nil && !CanBrotlyFunc(req) {
		return false
	}
	if req.URL.Scheme == "https" || zhttp.HostIsLocal(req.Host) {
		return true
	}
	u := ua.Parse(req.UserAgent())
	return u.Name == ua.Safari
}

// filesRedirector's ServeHTTP serves everything in zrest.StaticFolderPathFunc()
func (r filesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	spath := req.URL.Path
	// zlog.Info("FilesRedir:", spath)
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
	fpath := "www/" + spath
	if spath == "main.wasm" {
		var enc string
		var filesystem string
		es := []string{"gz"}
		if canBrotly(req) {
			es = append([]string{"br"}, es...) // we want if first, to use if possible
		}
	allWebFS:
		for _, f := range AllWebFS {
			for _, ext := range es {
				wpath := fpath + "." + ext
				exists := zfile.CanOpenInFS(f.FS, wpath)
				// zlog.Info("wasm?", wpath, exists, f.FSName)
				if exists {
					filesystem = f.FSName
					fpath = wpath
					enc = ext
					if ext == "gz" {
						enc = "gzip"
					}
					break allWebFS
				}
			}
		}
		if enc == "" {
			zrest.ReturnAndPrintError(w, req, http.StatusNotFound, "No main.wasm for", es, spath, fpath)
			return
		}
		zlog.Info("Serve WASM bin:", fpath, req.RemoteAddr, "filesystem:", filesystem, enc)
		smime = "application/wasm"
		w.Header().Set("Content-Encoding", enc)
		w.Header().Set("Expires", time.Now().Add(time.Hour).Format(time.RFC1123))
		// w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Accept-Ranges", "bytes")
		// w.Header().Set("X-Frame-Options", "sameorigin")
		// w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Age", "586183")
		w.Header().Set("Cache-Control", "public, max-age=604800")
		// w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		// w.Header().Set("Cross-Origin-Embedder-Policy-Report-Only", `credentialless; report-to="geo-earth-eng-team"`)
		// w.Header().Set("Cross-Origin-Opener-Policy-Report-Only", `same-origin; report-to="geo-earth-eng-team"`)
		// w.Header().Set("Cross-Origin-Opener-Policy", `same-origin; report-to="geo-earth-eng-team"`)
		// w.Header().Set("Cache-Control", "public, max-age=604800")
	}
	if !zbuild.Build.At.IsZero() {
		// zlog.Info("LastMod:", spath, zbuild.Build.At.Format(time.RFC1123))
		w.Header().Set("Last-Modified", zbuild.Build.At.Format(time.RFC1123))
		// w.Header().Set("ETag", zstr.HashTo64Hex(zbuild.Build.At.Format(time.RFC1123)))
	}
	f, err := zfile.ReaderFromFileInFS(AllWebFS, fpath)
	if zlog.OnError(err, fpath) {
		return
	}
	info, _, err := AllWebFS.Stat(fpath)
	if err == nil {
		len := info.Size()
		if len != 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(len, 10))
		}
	}
	if smime != "" {
		w.Header().Set("Content-Type", smime)
	}
	urlTick := req.URL.Query().Get("zurltick")
	if spath == "index.html" {
		zlog.Info("TICK:", urlTick)
		data, err := io.ReadAll(f)
		if zlog.OnError(err, spath, fpath) {
			return
		}
		sdata := strings.Replace(string(data), "{{.WasmPostfix}}", "?tick="+urlTick, -1)
		_, err = w.Write([]byte(sdata))
	} else {
		_, err = io.Copy(w, f)
	}
	zlog.OnError(err, spath, fpath)
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
	RequestRedirector = &filesRedirector{
		Override: override,
	}
	zrest.AddSubHandler(router, "", RequestRedirector)

	// zlog.Info("HandleApp:", zrest.AppURLPrefix)
	//	route := router.PathPrefix(zrest.AppURLPrefix)
	//	route.Handler(f)
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, zrest.StaticFolderPathFunc("/favicon.ico"))
	})
}

func handleSetVerbose(w http.ResponseWriter, req *http.Request) {
	on := zrest.GetBoolVal(req.URL.Query(), "on")
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
	m.HeaderMD = InlineDocumentationHeaderMD + "\n"
	return m
}

func appNew(a *App) {
}

func (AppCalls) SetGUIError(ci *zrpc.ClientInfo, ce zerrors.ContextError) error {
	zlog.Info("Got gui error:", zlog.Full(ce), ci.IPAddress)
	if HandleGUIErrorFunc != nil {
		dict := zdict.Dict{
			"Browser IP": ci.IPAddress,
			"UserAgent":  ci.UserAgent,
		}
		if ci.UserID != 0 {
			dict["UserID"] = ci.UserID
		}
		HandleGUIErrorFunc(ci, ce, dict)
	}
	return nil
}

// CheckServeFilesExists sees if paths exists in AllWebFS embeded and dir.
// They should be non-absolute, without www/ prefix.
func (AppCalls) CheckServeFilesExists(paths []string, existPaths *[]string) error {
	var returnErr error
	for _, p := range paths {
		urlPath := "www/" + p
		_, err := AllWebFS.Open(urlPath)
		if err != nil {
			if err != fs.ErrNotExist {
				returnErr = err
			}
			continue
		}
		*existPaths = append(*existPaths, p)
	}
	return returnErr
}

func (AppCalls) GetTimeInfo(a zrpc.Unused, info *TimeInfo) error {
	t := time.Now()
	ServerTimezoneName, info.ZoneOffsetSeconds = t.Zone()
	// zlog.Info("AppCall.GetTimeInfo:", t, ServerTimezoneName, info.ZoneOffsetSeconds)
	info.ZoneName = ServerTimezoneName
	info.JSISOTimeString = t.UTC().Format(ztime.JavascriptISO)
	return nil
}
