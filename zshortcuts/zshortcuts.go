//go:build zui

package zshortcuts

import (
	"time"

	"github.com/torlangballe/zui/zanimation"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/ztimer"
)

// type viewMod struct {
// 	view zview.View
// 	mod  zkeyboard.KeyMod
// }

var (
	helpStacks      = map[*zwindow.Window]*zcontainer.StackView{}
	highlightTimers = map[zview.View]*ztimer.Repeater{}
	showing         bool
)

func StrokeViewToShowHandling(view zview.View, viewKM zkeyboard.KeyMod, scut zkeyboard.KeyMod) bool {
	// zlog.Info("StrokeViewToShowShortcutHandling", scut, viewKM.Matches(scut), viewKM.Key)
	nv := view.Native()
	var col zgeo.Color
	var o float32
	if scut.Key == 0 && scut.Char == "" && scut.Modifier&viewKM.Modifier != 0 {
		col = zgeo.ColorYellow
		o = 0.3
		if scut.Modifier == viewKM.Modifier {
			o = 1
		}
	}
	if viewKM.Matches(scut) {
		col = zgeo.ColorBlue
		o = 1
	} else if o == 0 {
		return false
	}
	nv.SetCorner(3)
	nv.SetBGColor(col.WithOpacity(o))
	// vm := viewMod{nv.View, scut}
	timer := highlightTimers[nv.View]
	if timer == nil {
		timer = ztimer.RepeaterNew()
		highlightTimers[nv.View] = timer
	}
	id := zwindow.FromNativeView(nv).AddFocusHandler(nv, false, func() {
		nv.SetBGColor(zgeo.ColorClear)
	})

	timer.Set(0.7, false, func() bool {
		if viewKM.Modifier == zkeyboard.ModifierNone || zkeyboard.CurrentKeyDown.Modifier == zkeyboard.ModifierNone || zkeyboard.CurrentKeyDown.Modifier != viewKM.Modifier {
			zview.RemoveACallback(id)
			nv.SetBGColor(zgeo.ColorClear)
			return false
		}
		return true
	})
	return false
}

func RegisterShortCutHelperAreaForWindow(win *zwindow.Window, stack *zcontainer.StackView) {
	helpStacks[win] = stack
}

func ShowShortCutHelperForView(view zview.View, scut zkeyboard.KeyMod) {
	win := zwindow.FromNativeView(view.Native())
	stack := helpStacks[win]
	if stack == nil {
		return
	}
	if showing {
		return
	}
	showing = true
	stack.SetAlpha(0.01)
	str := scut.AsString()
	stack.SetSpacing(2)
	for _, s := range str {
		label := zlabel.New(string(s))
		label.SetTextAlignment(zgeo.Center)
		label.SetMinWidth(18)
		label.SetBGColor(zgeo.ColorNew(1, 1, 0.3, 1))
		label.SetCorner(3)
		stack.Add(label, zgeo.CenterLeft)
	}
	a := zcontainer.FindAncestorArranger(stack)
	if a != nil {
		a.ArrangeChildren()
	}
	const dur = 0.2
	ztimer.StartIn(0.2, func() {
		zanimation.SetAlpha(stack, 1, dur, func() {
			time.Sleep(time.Second)
			zanimation.SetAlpha(stack, 0, dur, func() {
				stack.RemoveAllChildren()
				showing = false
			})
		})
	})
}
