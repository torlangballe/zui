//go:build zui && catalyst

package zweb

// #include <stdlib.h>
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

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type nativeWebView struct {
	//TODO: general viewptr in NativeView
	webViewPtr unsafe.Pointer
}

func init() {
	zwindow.GetViewNativePointerFunc = func(v zview.View) unsafe.Pointer {
		wv, got := v.(*WebView)
		zlog.Assert(got)
		return wv.webViewPtr
	}
}
func (v *WebView) init(minSize zgeo.Size, isFrame bool) {
	v.webViewPtr = C.NewWKWebView(C.int(minSize.W), C.int(minSize.H))
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
