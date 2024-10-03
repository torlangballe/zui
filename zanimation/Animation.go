// Created by Tor Langballe originally on /9/11/15.
// The zanimation package defines transition types, and methods to animate views.

//go:build zui

package zanimation

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

type Transition int

const (
	TransitionNone Transition = iota
	TransitionFromLeft
	TransitionFromRight
	TransitionFromTop
	TransitionFromBottom
	TransitionFade
	TransitionReverse
	TransitionSame
)

type Swapper struct {
	OriginalRect  zgeo.Rect
	LastTransform zgeo.Pos
}

const DefaultSecs = 0.8
const repeatInfinite = -1

func RemoveAllFromView(view *zview.NativeView) {
}

func ViewHasAnimations(view *zview.NativeView) bool {
	return false
}

func PulseView(view *zview.NativeView, scale float64, durationSecs float64, fromScale float64, repeatCount float64) {
	animateView(view, fromScale, scale, durationSecs, "transform.scale", repeatCount, false)
}

func ScaleView(view *zview.NativeView, scaleTo float64, durationSecs float64) {
	animateView(view, 1, scaleTo, durationSecs, "transform.scale", 1, false)
}

func FadeView(view *zview.NativeView, to float64, durationSecs float64, from float64) {
	animateView(view, from, to, durationSecs, "opacity", 0, false)
}

func PulseOpacity(view *zview.NativeView, to float64, durationSecs float64, from float64, repeatCount float64) {
	animateView(view, from, to, durationSecs, "opacity", repeatCount, false)
}

func RippleWidget(view *zview.NativeView, durationSecs float64) {
}

func MoveViewOnPath(view *zview.NativeView, path *zgeo.Path, float64, repeatCount float64, begin float64) {
}

func RotateView(view *zview.NativeView, degreesClockwise float64, secs float64, repeatCount float64) {
}

type GradientLayer int

func AddGradientToView(view *zview.NativeView, colors []zgeo.Color, locations [][]float64, durationSecs float64, autoReverse bool, speed float64, opacity float64, min, max zgeo.Pos) *GradientLayer {
	return nil
}

func SetViewLayerSpeed(view *zview.NativeView, speed float64, resetTime bool) {
}

func FlipViewHorizontal(view *zview.NativeView, durationSecs float64, left bool, animate *func()) {
	//        let  uitrans = left ? *zview.NativeViewAnimationTransition.flipFromLeft   *zview.NativeViewAnimationTransition.flipFromRight
}

func animateView(view *zview.NativeView, from float64, to float64, durationSecs float64, atype string, repeatCount float64, autoreverses bool) {

}

func (s *Swapper) SlideViewInOverOld(parent, oldView, newView zview.View, dir zgeo.Alignment, secs float64, done func()) {
	translateViews(s, false, parent, oldView, newView, dir, secs, done)
}

func (s *Swapper) SlideViewInOldOut(parent, oldView, newView zview.View, dir zgeo.Alignment, secs float64, done func()) {
	translateViews(s, true, parent, oldView, newView, dir, secs, done)
}

func translateViews(s *Swapper, moveOld bool, parent, oldView, newView zview.View, dir zgeo.Alignment, secs float64, done func()) {
	// newView.Native().SetAlpha(0.1)
	r := s.OriginalRect
	parent.Native().AddChild(newView, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	dirPos := dir.Vector()
	move := dirPos.Times(r.Size.Pos())
	r.Pos.Subtract(move)
	newView.SetRect(r)
	newView.Native().SetAlpha(1)
	ztimer.StartIn(0.01, func() { // without this it does the first translate in 0 time.
		Translate(newView, move, secs, nil)
		if moveOld {
			delta := move.Plus(s.LastTransform)
			Translate(oldView, delta, secs, done)
		}
		s.LastTransform = move
	})
}
