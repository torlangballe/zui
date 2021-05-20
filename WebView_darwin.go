// +build zui,catalyst

package zui

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa
// #cgo LDFLAGS: -framework WebKit
// void *NewWKWebView(int width, int heigt);
// void WebViewSetLogPath(void *w, const char *logPath);
// void WebViewSetURL(void *view, const char *surl);
// void WebViewClearAllCaches();
import "C"

import (
	"unsafe"

	"github.com/torlangballe/zutil/zgeo"
)

type nativeWebView struct {
	//TODO: general viewptr in NativeView
	webViewPtr unsafe.Pointer
}

func (v *WebView) init(minSize zgeo.Size, isFrame bool) {
	v.webViewPtr = C.NewWKWebView(C.int(minSize.W), C.int(minSize.H))
}

func (v *WebView) SetLogPath(path string) {
	C.WebViewSetLogPath(v.webViewPtr, C.CString(path))
}

func (v *WebView) SetURL(surl string) {
	C.WebViewSetURL(v.webViewPtr, C.CString(surl))
}

func WebViewClearAllCaches() {
	C.WebViewClearAllCaches()
}
