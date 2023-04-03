//go:build zui

package zcontainer

import (
	"math"

	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type ChildFocusNavigator struct {
	children       []zview.View
	CurrentFocused zview.View
	HandleSelect   func(v zview.View, dir zgeo.Alignment)
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
		n.HandleSelect(minView, zgeo.AlignmentNone)
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
	dirAlign := zkeyboard.ArrowKeyToDirection(key)
	dir := dirAlign.Vector()
	if dir.IsNull() {
		return false
	}
	var minView zview.View
	if n.CurrentFocused == nil {
		// zlog.Info("focus")
		n.Focus()
		return true
	} else {
		rect := n.CurrentFocused.Rect().ExpandedD(-2)
		minDist := -1.0
		// zlog.Info(dir, rect, n.CurrentFocused.ObjectName())
		vertical := (dir.Y == 0)
		for _, v := range n.children {
			if v == n.CurrentFocused {
				continue
			}
			r := v.Rect()
			if r.Max().Vertice(vertical) >= rect.Min().Vertice(vertical) && r.Min().Vertice(vertical) <= rect.Max().Vertice(vertical) {
				dist := r.Min().Vertice(!vertical) - rect.Center().Vertice(!vertical)
				if dir.Vertice(!vertical)*dist > 0 {
					// zlog.Info("arrow", r.Max().Y >= rect.Min().Y && r.Min().Y <= rect.Max().Y, dist, "'"+v.ObjectName()+"'", "MM", r.Max().Y, rect.Min().Y, r.Min().Y, rect.Max().Y)
					dist = math.Abs(dist)
					if minDist == -1 || dist < minDist { // minDist only -1 at start, otherwise absolute value
						minDist = dist
						minView = v
					}
				}
			}
		}
	}
	if minView != nil {
		n.HandleSelect(minView, dirAlign)
		n.CurrentFocused = minView // set this after, so we can compare to what was in HandleSelect
		return true
	}
	n.HandleSelect(nil, dirAlign) // send direction even if not moved
	return false
}
