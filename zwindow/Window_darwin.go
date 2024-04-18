//go:build zui && catalyst

package zwindow

// #include <stdlib.h>
// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa
// void *NewWindow(int x, int y, int width, int height);
// void MakeKeyAndOrderFront(void *);
// void SetTitle(void *, char* title);
// void AddView(void *win, void *view);
import "C"
import (
	"strconv"
	"unsafe"

	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
)

// https://github.com/alediaferia/gogoa/blob/master/window.go

type windowNative struct {
	windowPtr unsafe.Pointer
	rect      zgeo.Rect
}

var GetViewNativePointerFunc func(v zview.View) unsafe.Pointer

func Current() *Window {
	return nil
}

func FromNativeView(v *zview.NativeView) *Window {
	return GetMain()
}

func (w *Window) Rect() zgeo.Rect {
	return w.rect
}

func (w *Window) ContentRect() zgeo.Rect {
	return w.rect
}

func (win *Window) SetOnKeyEvents() {
}

func Open(o Options) *Window {
	w := &Window{}
	sm := zscreen.GetMain()
	var r zgeo.Rect
	zlog.Assert(o.FullScreenID != 0, o.FullScreenID)
	zlog.Assert(!o.Size.IsNull())
	zlog.Info("Win Open")
	if o.FullScreenID == 0 {
		a := o.Alignment
		if a == zgeo.AlignmentNone {
			a = zgeo.Center
		}
		r = sm.UsableRect.Align(o.Size, a, zgeo.SizeNull)
	} else if o.FullScreenID == -1 {
		r = sm.Rect
	} else {
		sid := strconv.FormatInt(o.FullScreenID, 10)
		screen := zscreen.FindForID(sid)
		if screen == nil {
			zlog.Error("ScreenFromID is nil:", o.FullScreenID)
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
	ctitle := C.CString(title)
	C.SetTitle(w.windowPtr, ctitle)
	C.free(unsafe.Pointer(ctitle))
}

func (w *Window) AddView(v zview.View) {
	p := GetViewNativePointerFunc(v)
	zlog.Assert(p != nil)
	C.AddView(w.windowPtr, p)
}

func (win *Window) AddKeypressHandler(v zview.View, handler func(km zkeyboard.KeyMod, down bool) bool) {
}

func (w *Window) SetScrollHandler(handler func(pos zgeo.Pos)) {}
func (win *Window) SetAddressBarURL(surl string)              {}
func (win *Window) SetLocation(surl string)                   {}
func (w *Window) Close()                                      {}
func (win *Window) SetOnResizeHandling()                      {}
func (win *Window) AddStyle()                                 {}
func (w *Window) Reload()                                     {}
