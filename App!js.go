// +build !js

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

type FilesRedirector struct {
	Override func(w http.ResponseWriter, req *http.Request) bool
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
	if r.Override != nil {
		if r.Override(w, req) {
			return
		}
	}
	path := req.URL.Path
	zstr.HasPrefix(path, zrest.AppURLPrefix, &path)
	if zstr.HasPrefix(path, "page/", &path) {
		part := zstr.ExtractStringTilSeparator(&path, "/")
		switch part {
		case "images", "js", "css", "fonts", "templates":
			path = part + "/" + path
		}
	}
	filepath := filePathPrefix + path
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
	// zlog.Info("Serve app:", path, filepath)
	http.ServeFile(w, req, filepath)
}

func AppServeZUIWasm(override func(w http.ResponseWriter, req *http.Request) bool) {
	f := &FilesRedirector{
		Override: override,
	}
	http.Handle(zrest.AppURLPrefix, f)
}
