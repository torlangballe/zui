//go:build zui

package zcontainer

import (
	"math"

	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type ChildFocusNavigator struct {
	children       []zview.View
	CurrentFocused zview.View
	HandleSelect   func(v zview.View)
}

func (n *ChildFocusNavigator) Focus() {
	var minView zview.View
	var minPos zgeo.Pos
	for i, v := range n.children {
		p := v.Rect().Pos
		if i == 0 || p.Y < minPos.Y || p.Y == minPos.Y && p.X < minPos.X {
			minPos = p
			minView = v
		}
	}
	if minView != nil {
		n.CurrentFocused = minView
		n.HandleSelect(minView)
	}
}

func (n *ChildFocusNavigator) Clear() {
	n.children = n.children[:0]
	n.CurrentFocused = nil
}

func (n *ChildFocusNavigator) AddChild(v zview.View) {
	n.children = append(n.children, v)
}

func (n *ChildFocusNavigator) HandleKey(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
	if mod != zkeyboard.ModifierNone {
		return false
	}
	dir := zkeyboard.ArrowKeyToDirection(key)
	if dir == zgeo.AlignmentNone {
		return false
	}
	var minView zview.View
	if n.CurrentFocused == nil {
		zlog.Info("focus")
		n.Focus()
		return true
	} else {
		rect := n.CurrentFocused.Rect()
		minDist := -1.0
		zlog.Info(dir, rect)
		for _, v := range n.children {
			r := v.Rect()
			if dir&zgeo.Vertical != 0 {
			} else {
				if r.Max().Y >= rect.Min().Y && r.Min().Y <= rect.Max().Y {
					dist := r.Min().X - rect.Center().X
					if dir&zgeo.Right != 0 && dist > 0 || dir&zgeo.Left != 0 && dist < 0 {
						zlog.Info("arrow", r, dist)
						dist = math.Abs(dist)
						if minDist == -1 || dist < minDist { // minDist only -1 at start, otherwise absolute value
							minDist = minDist
							minView = v
						}
					}
				}
			}
		}
	}
	if minView != nil {
		n.HandleSelect(minView)
		n.CurrentFocused = minView // set this after, so we can compare to what was in HandleSelect
		return true
	}
	return false
}
