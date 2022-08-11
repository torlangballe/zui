// Created by Tor Langballe originally on /9/11/15.
// The zanimation package defines transition types, and methods to animate views.

//go:build zui

package zanimation

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
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

const DefaultSecs = 0.8
const repeatInfinite = -1

func AnimationRemoveAllFromView(view *zview.NativeView) {
}

func AnimationViewHasAnimations(view *zview.NativeView) bool {
	return false
}

func AnimationPulseView(view *zview.NativeView, scale float64, durationSecs float64, fromScale float64, repeatCount float64) {
	animateView(view, fromScale, scale, durationSecs, "transform.scale", repeatCount, false)
}

func AnimationScaleView(view *zview.NativeView, scaleTo float64, durationSecs float64) {
	animateView(view, 1, scaleTo, durationSecs, "transform.scale", 1, false)
}

func AnimationFadeView(view *zview.NativeView, to float64, durationSecs float64, from float64) {
	animateView(view, from, to, durationSecs, "opacity", 0, false)
}

func AnimationPulseOpacity(view *zview.NativeView, to float64, durationSecs float64, from float64, repeatCount float64) {
	animateView(view, from, to, durationSecs, "opacity", repeatCount, false)
}

func AnimationRippleWidget(view *zview.NativeView, durationSecs float64) {
}

func AnimationMoveViewOnPath(view *zview.NativeView, path *zgeo.Path, float64, repeatCount float64, begin float64) {
}

func AnimationRotateView(view *zview.NativeView, degreesClockwise float64, secs float64, repeatCount float64) {
}

type GradientLayer int

func AnimationAddGradientToView(view *zview.NativeView, colors []zgeo.Color, locations [][]float64, durationSecs float64, autoReverse bool, speed float64, opacity float64, min, max zgeo.Pos) *GradientLayer {
	return nil
}

func AnimationSetViewLayerSpeed(view *zview.NativeView, speed float64, resetTime bool) {
}

func AnimationFlipViewHorizontal(view *zview.NativeView, durationSecs float64, left bool, animate *func()) {
	//        let  uitrans = left ? *zview.NativeViewAnimationTransition.flipFromLeft   *zview.NativeViewAnimationTransition.flipFromRight
}

func animateView(view *zview.NativeView, from float64, to float64, durationSecs float64, atype string, repeatCount float64, autoreverses bool) {

}
