//go:build zui

package zview

import (
	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zutil/zgeo"
)

type FocusType int

type ViewFlags int64

type NativeView struct {
	baseNativeView
	View       View
	Flags      ViewFlags
	DoOnRemove []func() // anything that needs to be stopped // these could be in a global map if added/removed properly?
	DoOnAdd    []func() // anything that needs to be stopped
}

type DragType string

const (
	DragEnter             DragType = "enter"
	DragLeave             DragType = "leave"
	DragOver              DragType = "over"
	DragDrop              DragType = "drop"
	DragDropFilePreflight DragType = "filepre" // if a handle returns false to DragDropPreflight, don't process file(s)
	DragDropFile          DragType = "file"

	BaseZIndex = 100

	ViewPresentedFlag = 1
	ViewUsableFlag    = 2
)

var (
	ChildOfViewFunc   func(v View, path string) View
	RangeChildrenFunc func(root View, recursive, includeCollapsed bool, got func(View) bool)
	LastPressedPos    zgeo.Pos
	SkipEnterHandler  bool
)

func (v *NativeView) IsPresented() bool {
	return v.Flags&ViewPresentedFlag != 0
}

func (v *NativeView) IsUsable() bool {
	return v.Flags&ViewUsableFlag != 0
}

// AddOnRemoveFunc adds a function to call when the v is removed from it's parent.
func (v *NativeView) AddOnRemoveFunc(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.DoOnRemove = append(v.DoOnRemove, f)
}

// AddOnAddFunc adds a function to call when v is added to it's parent
// Some views don't get these functions called until visible, so always create timers etc on a view
// with AddOnAddFunc and use timer.Stop with AddOnRemoveFunc to ensure they are added/removed correctly
func (v *NativeView) AddOnAddFunc(f func()) {
	v.DoOnAdd = append(v.DoOnAdd, f)
}

func (v *NativeView) PerformAddRemoveFuncs(add bool) {
	var count int
	if add {
		for _, f := range v.DoOnAdd {
			f()
		}
		return
	}
	// zlog.Info("PerformAddRemoveFuncs START:", zlog.Pointer(v), v.Hierarchy(), len(v.DoOnRemove))
	RangeChildrenFunc(v.View, false, true, func(child View) bool {
		child.Native().PerformAddRemoveFuncs(false)
		count++
		return true
	})
	for _, f := range v.DoOnRemove {
		f()
	}
	// zlog.Info(indent+"PerformAddRemoveFuncs  DONE:", zlog.Pointer(v), v.Hierarchy(), len(v.DoOnRemove), count)
}

func (v *NativeView) IsParentOf(c *NativeView) bool {
	for _, p := range c.AllParents() {
		if p == v {
			return true
		}
	}
	return false
}

func (v *NativeView) SetResizeCursorFromAlignment(a zgeo.Alignment) bool {
	var cursor zcursor.Type
	switch a {
	case zgeo.Top:
		cursor = zcursor.ResizeTop
	case zgeo.Bottom:
		cursor = zcursor.ResizeBottom
	case zgeo.Left:
		cursor = zcursor.ResizeLeft
	case zgeo.Right:
		cursor = zcursor.ResizeRight
	case zgeo.TopLeft:
		cursor = zcursor.ResizeTopLeft
	case zgeo.TopRight:
		cursor = zcursor.ResizeTopRight
	case zgeo.BottomLeft:
		cursor = zcursor.ResizeBottomLeft
	case zgeo.BottomRight:
		cursor = zcursor.ResizeBottomRight
	case zgeo.Center:
		cursor = zcursor.Grab
	default:
		return false
	}
	v.SetCursor(cursor)
	return true
}
