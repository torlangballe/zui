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

func (n *ChildFocusNavigator) FocusNext() {
	var minView zview.View
	var minPos zgeo.Pos
	for i, v := range n.children {
		p := v.Rect().Pos
		if i == 0 || p.Y < minPos.Y || p.Y == minPos.Y && p.X < minPos.X {
			minPos = p
			minView = v
		}
	}
	// zlog.Info("Nav Foc:", minView != nil)
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

func (n *ChildFocusNavigator) HasChild(v zview.View) bool {
	for _, c := range n.children {
		if c == v {
			return true
		}
	}
	return false
}

func (n *ChildFocusNavigator) HandleKey(km zkeyboard.KeyMod, down bool) bool {
	if km.Modifier != zkeyboard.ModifierNone {
		return false
	}
	dirAlign := zkeyboard.ArrowKeyToDirection(km.Key)
	dir := dirAlign.Vector()
	if dir.IsNull() {
		return false
	}
	var minView zview.View
	if n.CurrentFocused == nil {
		// zlog.Info("focus")
		n.FocusNext()
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
			if r.Max().Element(vertical) >= rect.Min().Element(vertical) && r.Min().Element(vertical) <= rect.Max().Element(vertical) {
				dist := r.Min().Element(!vertical) - rect.Center().Element(!vertical)
				if dir.Element(!vertical)*dist > 0 {
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
