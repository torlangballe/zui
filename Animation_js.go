package zui

import (
	"syscall/js"
)

var animationFunc js.Func

// Animate calls handler smartly on window Animation frames, until secs has passed, or handler returns false
// TODO: Use zmath.EaseInOut to make rounder animation
func Animate(secs float64, handler func(secsPos float64) bool) {
	if secs == 0 {
		secs = AnimationDefaultSecs
	}
	var startJS js.Value
	animationFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		posJS := args[0]
		if startJS.IsUndefined() {
			startJS = posJS
		}
		diff := (posJS.Float() - startJS.Float()) / 1000
		if diff < secs {
			ok := handler(diff)
			if !ok {
				return nil
			}
			WindowJS.Call("requestAnimationFrame", animationFunc)
		}
		return nil
	})
	WindowJS.Call("requestAnimationFrame", animationFunc)
}
