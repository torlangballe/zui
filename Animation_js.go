package zui

import (
	"math/rand"
	"syscall/js"
)

func addAnimationToWindow(win *Window, frameID int, randomID int) {
	if win.animationFrames == nil {
		win.animationFrames = map[int]int{}
		win.animationFrames[randomID] = frameID
	}
}

// Animate calls handler smartly on window Animation frames, until secs has passed, or handler returns false
// TODO: Use zmath.EaseInOut to make rounder animation
func Animate(view View, secs float64, handler func(t float64) bool) {
	var animationFunc js.Func

	if secs == 0 {
		secs = AnimationDefaultSecs
	}
	neg := -1
	var aniFrameID = &neg
	var startJS js.Value

	randomID := int(rand.Int31())
	win := ViewGetNative(view).GetWindow()
	animationFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		posJS := args[0]
		if startJS.IsUndefined() {
			startJS = posJS
		}
		t := (posJS.Float() - startJS.Float()) / 1000 / secs
		// zlog.Info("Ani?", t, posJS.Float(), startJS.Float(), secs)
		if t <= 1 {
			ok := handler(t)
			if !ok {
				if *aniFrameID != -1 {
					delete(win.animationFrames, randomID)
				}
				return nil
			}
			*aniFrameID = win.element.Call("requestAnimationFrame", animationFunc).Int()
			addAnimationToWindow(win, *aniFrameID, randomID)
		}
		if *aniFrameID != -1 {
			delete(win.animationFrames, randomID)
		}
		return nil
	})
	if win != nil {
		*aniFrameID = win.element.Call("requestAnimationFrame", animationFunc).Int()
		addAnimationToWindow(win, *aniFrameID, randomID)
	}
}
