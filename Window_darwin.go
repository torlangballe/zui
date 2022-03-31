//go:build zui && catalyst
// +build zui,catalyst

package zui

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa
// void *NewWindow(int x, int y, int width, int height);
// void MakeKeyAndOrderFront(void *);
// void SetTitle(void *, char* title);
// void AddView(void *win, void *view);
import "C"
import (
	"unsafe"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyboard"
)

// https://github.com/alediaferia/gogoa/blob/master/window.go

type windowNative struct {
	windowPtr unsafe.Pointer
	rect      zgeo.Rect
}

func WindowGetMain() *Window {
	return nil
}

func (w *Window) Rect() zgeo.Rect {
	return w.rect
}

func (w *Window) ContentRect() zgeo.Rect {
	return w.rect
}

func WindowOpen(o WindowOptions) *Window {
	w := &Window{}
	sm := zscreen.GetMain()
	var r zgeo.Rect
	zlog.Assert(o.FullScreenID != 0 || !o.Size.IsNull())
	if o.FullScreenID == 0 {
		a := o.Alignment
		if a == zgeo.AlignmentNone {
			a = zgeo.Center
		}
		r = sm.UsableRect.Align(o.Size, a, zgeo.Size{})
	} else if o.FullScreenID == -1 {
		r = sm.Rect
	} else {
		screen := ScreenFromID(o.FullScreenID)
		if screen == nil {
			zlog.Error(nil, "ScreenFromID is nil:", o.FullScreenID)
			return nil
		}
		r = screen.Rect
	}
	w.rect = r
	w.windowPtr = C.NewWindow(C.int(r.Pos.X), C.int(r.Pos.Y), C.int(r.Size.W), C.int(r.Size.H))
	return w
}

func (win *Window) GetURL() string {
	return ""
}

func (w *Window) Activate() {
	C.MakeKeyAndOrderFront(w.windowPtr)
}

func (w *Window) SetTitle(title string) {
	C.SetTitle(w.windowPtr, C.CString(title))
}

func (w *Window) AddView(v View) {
	wv, got := v.(*WebView)
	zlog.Assert(got)
	C.AddView(w.windowPtr, wv.webViewPtr)
}

func (win *Window) AddKeypressHandler(v View, handler func(zkeyboard.Key, zkeyboard.Modifier)) {
}

func (w *Window) SetScrollHandler(handler func(pos zgeo.Pos)) {}
func (win *Window) SetAddressBarURL(surl string)              {}
func (win *Window) SetLocation(surl string)                   {}
func (w *Window) Close()                                      {}
func (win *Window) setOnResizeHandling()                      {}
