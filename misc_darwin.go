package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func init() {
	zlog.Info("zui init()")
	// runtime.LockOSThread() // ! skip for now
}

// func TimerInitBlockingDispatchFromMain() {
// 	runtime.LockOSThread()
// 	for f := range mainfunc {
// 		f()
// 	}
// }

// var mainfunc = make(chan func())

// var backgroundIndex = 0
// var backgroundFuncs = map[int]func(){}

// //export runFuncFromQue
// func runFuncFromQue(i int32) {

// }

// func MainQueAsync(do func()) {
// 	i := rand.Int31()
// 	backgroundFuncs[int(i)] = do
// 	//C.runOnMain(i)
// 	// go func() {
// 	// 	mainfunc <- do
// 	// }()
// }

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
	s := Screen{}
	s.SoftScale = 1
	s.Scale = 1
	return s
}

func zViewAddView(parent View, child View, index int) {
}

// CustomView

func zViewSetRect(view View, rect zgeo.Rect, layout bool) { // layout only used on android
}
