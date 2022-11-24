//go:build zui

package zview

import "github.com/torlangballe/zutil/zgeo"

type NativeView struct {
	baseNativeView
	View      View
	Presented bool
	// allChildrenPresented bool
	DoOnRemove []func() // anything that needs to be stopped
	DoOnAdd    []func() // anything that needs to be stopped
}

type DragType string

const (
	DragEnter         DragType = "enter"
	DragLeave         DragType = "leave"
	DragOver          DragType = "over"
	DragDrop          DragType = "drop"
	DragDropFilePreflight DragType = "filepre" // if a handle returns true to DragDropPreflight, don't process file(s)
	DragDropFile      DragType = "file"
)

var (
	ChildOfViewFunc             func(v View, path string) View
	RangeAllVisibleChildrenFunc func(root View, got func(View) bool)
	LastPressedPos              zgeo.Pos
)

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
	if add {
		for _, f := range v.DoOnAdd {
			f()
		}
		return
	}
	RangeAllVisibleChildrenFunc(v.View, func(child View) bool {
		child.Native().PerformAddRemoveFuncs(false)
		return true
	})
	for _, f := range v.DoOnRemove {
		f()
	}
}

func (v *NativeView) RootParent() *NativeView {
	all := v.AllParents()
	if len(all) == 0 {
		return v
	}
	return all[0]
}

func (v *NativeView) IsParentOf(c *NativeView) bool {
	for _, p := range c.AllParents() {
		if p == v {
			return true
		}
	}
	return false
}

func (v *NativeView) FocusNext(forward bool) {
	// TODO
}
