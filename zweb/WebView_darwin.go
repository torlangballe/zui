//go:build zui && catalyst

package zweb

// #include <stdlib.h>
// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa
// #cgo LDFLAGS: -framework WebKit
// void *NewWKWebView(int width, int heigt);
// void WebViewSetLogPath(void *w, const char *logPath);
// void WebViewSetURL(void *view, const char *surl);
// void WebViewSetContent(void *view, const char *html);
// void WebViewClearAllCaches();
import "C"

import (
	"unsafe"

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type nativeWebView struct {
	//TODO: general viewptr in NativeView
	webViewPtr unsafe.Pointer
}

type webkitCode struct {
	webKit unsafe.Pointer
	code   int
}

var (
	nativeToWebView = map[unsafe.Pointer]*WebView{}
	errorChan       = make(chan webkitCode)
)

func init() {
	zwindow.GetViewNativePointerFunc = func(v zview.View) unsafe.Pointer {
		wv, got := v.(*WebView)
		zlog.Assert(got)
		return wv.webViewPtr
	}
	go selectLoopForError()
}

func selectLoopForError() {
	for {
		select {
		case wkCode := <-errorChan:
			webView := nativeToWebView[wkCode.webKit]
			zlog.Assert(webView != nil)
			webView.HandleErrorFunc(wkCode.code)
		}
	}
}

//export goErrorkHandler
func goErrorkHandler(wk unsafe.Pointer, code C.int) {
	zlog.Info("goErrorkHandler", wk, code)
	errorChan <- webkitCode{wk, int(code)}
}

func (v *WebView) init(minSize zgeo.Size, isFrame bool) {
	v.webViewPtr = C.NewWKWebView(C.int(minSize.W), C.int(minSize.H))
	nativeToWebView[v.webViewPtr] = v
}

func (v *WebView) SetLogPath(path string) {
	// cpath := C.CString(title)
	// C.WebViewSetLogPath(v.webViewPtr, cpath)
	// C.free(unsafe.Pointer(cpath))
}

func (v *WebView) SetURL(surl string) {
	curl := C.CString(surl)
	C.WebViewSetURL(v.webViewPtr, C.CString(surl))
	C.free(unsafe.Pointer(curl))
}

func WebViewClearAllCaches() {
	C.WebViewClearAllCaches()
}

func (v *WebView) SetHTMLContent(html string) {
	C.WebViewSetContent(v.webViewPtr, C.CString(html))
}
