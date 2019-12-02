package zgo

import (
	"runtime"

	"github.com/torlangballe/zutil/zgeo"
)

func init() {
	DebugPrint("zgo init()")
	runtime.LockOSThread()
}

// func TimerInitBlockingDispatchFromMain() {
// 	runtime.LockOSThread()
// 	for f := range mainfunc {
// 		f()
// 	}
// }

// var mainfunc = make(chan func())

func NativeViewAddToRoot(v View) {
}

// TextInfo
func (ti *TextInfo) getTextSize(noWidth bool) zgeo.Size {
	return zgeo.Size{}
}

// CustomView

func CustomViewInit() *NativeView {
	return nil
}

// Alert

func (a *Alert) Show(handle func(result AlertResult)) {
}

// Screen

func ScreenMain() Screen {
	return Screen{}
}

func zViewAddView(parent View, child View, index int) {
}

// CustomView

func zViewSetRect(view View, rect zgeo.Rect, layout bool) { // layout only used on android
}
