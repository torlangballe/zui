package zgo

//  Created by Tor Langballe on /12/11/15.

type ScreenLayout int

const (
	ScreenPortrait ScreenLayout = iota
	ScreenPortraitUpsideDown
	ScreenLandscapeLeft
	ScreenLandscapeRight
)

type Screen struct {
	isLocked       bool
	MainUsableRect Rect    //= ZRect(UIScreen.main.bounds)
	Scale          float64 //= float64(UIScreen.main.scale)
	SoftScale      float64 // = 1.0
	KeyboardRect   *Rect
}

func ScreenStatusBarHeight() float64 {
	return 0
}

func ScreenIsTall() bool {
	return ScreenMainRect().Size.H > 568
}

func ScreenIsWide() bool {
	return ScreenMainRect().Size.W > 320
}

func ScreenIsPortrait() bool {
	s := ScreenMainRect().Size
	return s.H > s.W
}

func ScreenShowNetworkActivityIndicator(show bool) {
}

func ScreenHasSleepButtonOnSide() bool {
	return false
}

func ScreenStatusBarVisible() bool {
	return false
}

func ScreenSetStatusBarForLightContent(light bool) {
}

func ScreenEnableIdle(on bool) {
}

func ScreenOrientation() ScreenLayout {
	return ScreenLandscapeLeft
}

func ScreenHasNotch() bool {
	return false
}

func ScreenHasSwipeUpAtBottom() bool {
	return false
}
