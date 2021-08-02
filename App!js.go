// +build !js,!catalyst

package zui

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

type nativeApp struct {
}

type FilesRedirector struct {
	Override         func(w http.ResponseWriter, req *http.Request) bool
	ServeDirectories bool
}

const filePathPrefix = "www/"

// AppURLPrefix is the first part of the path to your webapp, everything, including assets etc are within this prefix.

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

func (r FilesRedirector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	zlog.Info("FilesRedir1:", req.URL.Path)
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
	// zlog.Info("Serve app:", path, filepath)
	http.ServeFile(w, req, filepath)
}

func AppServeZUIWasm(serveDirs bool, override func(w http.ResponseWriter, req *http.Request) bool) {
	f := &FilesRedirector{
		ServeDirectories: serveDirs,
		Override:         override,
	}
	http.Handle(zrest.AppURLPrefix, f)
}

func appNew(a *App) {
}

func (a *App) Run() {
}
