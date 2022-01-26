//go:build zui
// +build zui

package zui

import (
	"sort"

	"github.com/torlangballe/zutil/zint"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
)

type GridView struct {
	CustomView
	Margin  zgeo.Rect
	Spacing zgeo.Size
	ByRows  bool

	children []View
}

func GridViewNew(name string) *GridView {
	v := &GridView{}
	v.Init(v, name)
	v.SetMinSize(zgeo.SizeBoth(100))
	v.Spacing = zgeo.Size{6, 6}
	return v
}

func (v *GridView) Init(view View, name string) {
	v.CustomView.Init(v, name)
	v.ByRows = true
}

func (v *GridView) CalculatedSize(total zgeo.Size) zgeo.Size {
	count := len(v.children)
	if count == 0 {
		return v.MinSize()
	}
	fit := total.Plus(v.Margin.Size)
	childSize := v.children[0].CalculatedSize(total)
	s := v.Margin.Size.Negative()
	var rows, cols int
	if v.ByRows {
		rows = int((fit.W + v.Spacing.W) / (childSize.W + v.Spacing.W))
		zint.Minimize(&rows, count)
		if rows != 0 {
			cols = (count + rows - 1) / rows
		}
	} else {
		zlog.FatalNotImplemented()
	}
	s.W += float64(rows)*childSize.W + float64(rows-1)*v.Spacing.W
	s.H += float64(cols)*childSize.H + float64(cols-1)*v.Spacing.H
	// zlog.Info("GridView CS:", s, rows, cols, childSize)
	return s
}

func (v *GridView) Add(child View) {
	v.children = append(v.children, child)
	v.AddChild(child, -1)
}

func (v *GridView) Remove(child View) bool {
	for i, c := range v.children {
		if c == child {
			zslice.RemoveAt(&v.children, i)
			v.RemoveChild(child)
			return true
		}
	}
	return false
}

func (v *GridView) GetChildren(collapsed bool) []View {
	return v.children
}

func (v *GridView) SetRect(rect zgeo.Rect) {
	v.CustomView.SetRect(rect)
	ct := v.View.(ContainerType) // in case we are a stack or something inheriting from GridView
	ct.ArrangeChildren()
}

func (v *GridView) ArrangeChildren() {
	if len(v.children) == 0 {
		return
	}
	rect := v.LocalRect().Plus(v.Margin)
	pos := rect.Pos
	childSize := v.children[0].CalculatedSize(rect.Size)
	for _, child := range v.children {
		r := zgeo.Rect{Pos: pos, Size: childSize}
		child.SetRect(r)
		if v.ByRows {
			pos.X += childSize.W + v.Spacing.W
			if pos.X+childSize.W > rect.Max().X {
				pos.X = rect.Pos.X
				pos.Y += childSize.H + v.Spacing.H
			}
		} else {
			zlog.Fatal(nil, "not implemented")
		}
	}
}

func (v *GridView) SortChildren(less func(a, b View) bool) {
	sort.Slice(v.children, func(i, j int) bool {
		return less(v.children[i], v.children[j])
	})
}
