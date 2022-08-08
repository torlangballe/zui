//go:build zui

package zanimation

import (
	"math/rand"
	"syscall/js"

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
)

func addAnimationToWindow(win *zwindow.Window, frameID int, randomID int) {
	if win.AnimationFrames == nil {
		win.AnimationFrames = map[int]int{}
		win.AnimationFrames[randomID] = frameID
	}
}

// Animate calls handler smartly on window Animation frames, until secs has passed, or handler returns false
// TODO: Use zmath.EaseInOut to make rounder animation
func Animate(view zview.View, secs float64, handler func(t float64) bool) {
	var animationFunc js.Func

	if secs == 0 {
		secs = DefaultSecs
	}
	neg := -1
	var aniFrameID = &neg
	var startJS js.Value

	randomID := int(rand.Int31())
	win := zwindow.GetFromNativeView(view.Native())
	animationFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// zlog.Info("Ani")
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
					delete(win.AnimationFrames, randomID)
				}
				return nil
			}
			*aniFrameID = win.Element.Call("requestAnimationFrame", animationFunc).Int()
			addAnimationToWindow(win, *aniFrameID, randomID)
		}
		handler(-1)
		if *aniFrameID != -1 {
			delete(win.AnimationFrames, randomID)
		}
		return nil
	})
	if win != nil {
		*aniFrameID = win.Element.Call("requestAnimationFrame", animationFunc).Int()
		addAnimationToWindow(win, *aniFrameID, randomID)
	}
}
