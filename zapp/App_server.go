// The server variant of App is an App (program) in it's own right, but also contains functionality to
// serve a wasm app to a browser.
// It is invoked with ServeZUIWasm (below), which uses a FilesRedirector (below) instance to handle serving the wasm, html and assets.

//go:build !js && !catalyst

package zapp

import (
	"io"
	"net/http"
	"strings"

	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmarkdown"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/zstr"
)

// nativeApp is used in App in zapp.go, so must be defined even if empty:
type nativeApp struct {
}

// FilesRedirector is a type that can handle serving files
type FilesRedirector struct {
	Override         func(w http.ResponseWriter, req *http.Request) bool // Override is a method to handle special cases of files, return true if handled
	ServeDirectories bool                                                // if ServeDirectories is true, it serves content list of directory
}

// FilesRedirector's ServeHTTP serves everything in www, handling directories, * wildcards, and auto-translating .md (markdown) files to html
func (r FilesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	const filePathPrefix = "www/"
	// zlog.Info("FilesRedir1:", req.URL.Path)
	if r.Override != nil {
		if r.Override(w, req) {
			return
		}
	}
	path := req.URL.Path
	zstr.HasPrefix(path, zrest.AppURLPrefix, &path)
	filepath := filePathPrefix + path
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
		if strings.HasSuffix(path, ".md") { // strings.HasPrefix(path, "doc/") &&
			html, err := convertMarkdownToHTML(filepath, path)
			if err != nil {
				zrest.ReturnError(w, req, err.Error(), http.StatusInternalServerError)
				return
			}
			zrest.AddCORSHeaders(w, req)
			io.WriteString(w, html)
			return
		}
	}
	zlog.Info("Serve app:", path, filepath)
	http.ServeFile(w, req, filepath)
}

// convertMarkdownToHTML is used by FilesRedirector to convert an .md file to html
func convertMarkdownToHTML(filepath, title string) (string, error) {
	markdown, err := zfile.ReadStringFromFile(filepath)
	if err != nil {
		return "", zlog.Error(err, "read markdown", filepath)
	}
	html, err := zmarkdown.ConvertToHTML(markdown, title)
	if err != nil {
		return "", zlog.Error(err, "convert", filepath)
	}
	return html, nil
}

func ServeZUIWasm(serveDirs bool, override func(w http.ResponseWriter, req *http.Request) bool) {
	f := &FilesRedirector{
		ServeDirectories: serveDirs,
		Override:         override,
	}
	zlog.Info("HandleApp:", zrest.AppURLPrefix)
	http.Handle(zrest.AppURLPrefix, f)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/favicon.ico")
	})
}

// URL returns the url/command that invoked this app
// func URL() string {
// 	return strings.Join(os.Args, " ")
// }
