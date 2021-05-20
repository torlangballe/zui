// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /12/11/15.

type ScreenLayout int

const (
	ScreenPortrait ScreenLayout = iota
	ScreenPortraitUpsideDown
	ScreenLandscapeLeft
	ScreenLandscapeRight
)

type Screen struct {
	IsMain     bool
	isLocked   bool
	ID         int64
	Rect       zgeo.Rect
	UsableRect zgeo.Rect
	Scale      float64 //= float64(UIScreen.main.scale)
	SoftScale  float64 // = 1.0
	//	KeyboardRect *zgeo.Rect
}

func ScreenMain() Screen {
	// test:
	for _, s := range ScreenGetAll() {
		if s.IsMain {
			return s
		}
	}
	zlog.Error(nil, "No screen!")
	s := Screen{}
	s.SoftScale = 1
	s.Scale = 1
	s.IsMain = true
	return s
}

func ScreenFromID(id int64) *Screen {
	for _, s := range ScreenGetAll() {
		// zlog.Info(id, "ScreenFromID", s)
		if s.ID == id {
			return &s
		}
	}
	return nil
}

// func ScreenStatusBarHeight() float64 {
// 	return 0
// }

// func ScreenIsTall() bool {
// 	return ScreenMain().Rect.Size.H > 568
// }

// func ScreenIsWide() bool {
// 	return ScreenMain().Rect.Size.W > 320
// }

// func ScreenIsPortrait() bool {
// 	s := ScreenMain().Rect.Size
// 	return s.H > s.W
// }

// func ScreenShowNetworkActivityIndicator(show bool) {
// }

// func ScreenHasSleepButtonOnSide() bool {
// 	return false
// }

// func ScreenStatusBarVisible() bool {
// 	return false
// }

// func ScreenSetStatusBarForLightContent(light bool) {
// }

// func ScreenEnableIdle(on bool) {
// }

// func ScreenOrientation() ScreenLayout {
// 	return ScreenLandscapeLeft
// }

// func ScreenHasNotch() bool {
// 	return false
// }

// func ScreenHasSwipeUpAtBottom() bool {
// 	return false
// }
