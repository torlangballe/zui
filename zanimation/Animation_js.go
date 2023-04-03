//go:build zui

package zanimation

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"

	"fmt"
	"math/rand"
	"syscall/js"
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

func doDone(done func(), nv *zview.NativeView, secs float64) {
	if done != nil {
		ztimer.StartIn(secs+0.01, func() {
			done()
		})
	}
}

func transform(v zview.View, secs float64, css string, done func()) {
	doDone(done, v.Native(), secs)
	v.Native().SetJSStyle("transition", fmt.Sprintf("transform %fs linear", secs))
	v.Native().SetJSStyle("transform", css)
}

func Translate(v zview.View, dir zgeo.Pos, secs float64, done func()) {
	transform(v, secs, fmt.Sprintf("translate(%fpx,%fpx)", dir.X, dir.Y), done)
}

func SetAlpha(v zview.View, alpha, secs float64, done func()) {
	v.Native().SetJSStyle("transition", fmt.Sprintf("opacity, %fs ease-in-out", secs))
	v.Native().SetJSStyle("opacity", fmt.Sprint(alpha))
}

func FlipHorizontal(v zview.View, secs float64, done func()) {
	doDone(done, v.Native(), secs)
	v.Native().SetJSStyle("transition", fmt.Sprintf("transform %fs linear", secs)) // remove linear
	// v.Native().SetJSStyle("transformStyle", "preserve3d")
	v.Native().SetJSStyle("transform", "rotateY(360deg)")
}

type Swapper struct {
	OriginalRect  zgeo.Rect
	LastTransform zgeo.Pos
}

func (s *Swapper) TranslateSwapViews(parent, oldView, newView zview.View, dir zgeo.Alignment, secs float64, done func()) {
	newView.Native().SetAlpha(0.1)
	r := s.OriginalRect
	parent.Native().AddChild(newView, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	dirPos := dir.Vector()
	move := dirPos.Times(r.Size.Pos())
	newView.SetRect(r.Translated(move))
	r.Pos.Subtract(move)
	newView.SetRect(r)
	newView.Native().SetAlpha(1)
	Translate(newView, move, secs, nil)
	delta := move.Plus(s.LastTransform)
	Translate(oldView, delta, secs, done)
	s.LastTransform = move
}

func (s *Swapper) FlipSwapViews(parent, oldView, newView zview.View, dir zgeo.Alignment, secs float64, done func()) {
	zlog.Info("FlipViews!")
	// newView.Native().SetAlpha(0)
	secs = 2
	r := s.OriginalRect
	newView.Native().SetJSStyle("transform", "rotateY(180deg)")
	newView.Native().ShowBackface(false)
	parent.Native().AddChild(newView, -1) // needs to preserve index, which isn't really supported in AddChild yet anyway
	newView.SetRect(r)
	// newView.Native().SetAlpha(1)
	// oldView.Native().ShowBackface(false)
	// FlipHorizontal(oldView, secs, nil)
	FlipHorizontal(newView, secs, done)
}
