package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /9/11/15.

const AnimationDefaultSecs = 0.8
const repeatInfinite = -1

func AnimationRemoveAllFromView(view *NativeView) {
}

func AnimationViewHasAnimations(view *NativeView) bool {
	return false
}

func AnimationPulseView(view *NativeView, scale float64, durationSecs float64, fromScale float64, repeatCount float64) {
	animateView(view, fromScale, scale, durationSecs, "transform.scale", repeatCount, false)
}

func AnimationScaleView(view *NativeView, scaleTo float64, durationSecs float64) {
	animateView(view, 1, scaleTo, durationSecs, "transform.scale", 1, false)
}

func AnimationFadeView(view *NativeView, to float64, durationSecs float64, from float64) {
	animateView(view, from, to, durationSecs, "opacity", 0, false)
}

func AnimationPulseOpacity(view *NativeView, to float64, durationSecs float64, from float64, repeatCount float64) {
	animateView(view, from, to, durationSecs, "opacity", repeatCount, false)
}

func AnimationRippleWidget(view *NativeView, durationSecs float64) {
}

func AnimationMoveViewOnPath(view *NativeView, path *zgeo.Path, float64, repeatCount float64, begin float64) {
}

func AnimationRotateView(view *NativeView, degreesClockwise float64, secs float64, repeatCount float64) {
}

type GradientLayer int

func AnimationAddGradientToView(view *NativeView, colors []zgeo.Color, locations [][]float64, durationSecs float64, autoReverse bool, speed float64, opacity float64, min, max zgeo.Pos) *GradientLayer {
	return nil
}

func AnimationSetViewLayerSpeed(view *NativeView, speed float64, resetTime bool) {
}

func AnimationFlipViewHorizontal(view *NativeView, durationSecs float64, left bool, animate *func()) {
	//        let  uitrans = left ? *NativeViewAnimationTransition.flipFromLeft   *NativeViewAnimationTransition.flipFromRight
}

func animateView(view *NativeView, from float64, to float64, durationSecs float64, atype string, repeatCount float64, autoreverses bool) {

}
