package zgo

//  Created by Tor Langballe on /9/11/15.

const AnimationDefaultSecs = 0.8
const repeatInfinite = -1

func AnimationDo(durationSecs float64, animations func(), completion func(done bool)) {
}

func AnimationRemoveAllFromView(view *ViewNative) {
}

func AnimationViewHasAnimations(view *ViewNative) bool {
	return false
}

func AnimationPulseView(view *ViewNative, scale float64, durationSecs float64, fromScale float64, repeatCount float64) {
	animateView(view, fromScale, scale, durationSecs, "transform.scale", repeatCount, false)
}

func AnimationScaleView(view *ViewNative, scaleTo float64, durationSecs float64) {
	animateView(view, 1, scaleTo, durationSecs, "transform.scale", 1, false)
}

func AnimationFadeView(view *ViewNative, to float64, durationSecs float64, from float64) {
	animateView(view, from, to, durationSecs, "opacity", 0, false)
}

func AnimationPulseOpacity(view *ViewNative, to float64, durationSecs float64, from float64, repeatCount float64) {
	animateView(view, from, to, durationSecs, "opacity", repeatCount, false)
}

func AnimationRippleWidget(view *ViewNative, durationSecs float64) {
}

func AnimationMoveViewOnPath(view *ViewNative, path *Path, float64, repeatCount float64, begin float64) {
}

func AnimationRotateView(view *ViewNative, degreesClockwise float64, secs float64, repeatCount float64) {
}

type GradientLayer int

func AnimationAddGradientToView(view *ViewNative, colors []Color, locations [][]float64, durationSecs float64, autoReverse bool, speed float64, opacity float64, min, max Pos) *GradientLayer {
	return nil
}

func AnimationSetViewLayerSpeed(view *ViewNative, speed float64, resetTime bool) {
}

func AnimationFlipViewHorizontal(view *ViewNative, durationSecs float64, left bool, animate *func()) {
	//        let  uitrans = left ? *ViewNativeAnimationTransition.flipFromLeft   *ViewNativeAnimationTransition.flipFromRight
}

func animateView(view *ViewNative, from float64, to float64, durationSecs float64, atype string, repeatCount float64, autoreverses bool) {

}
