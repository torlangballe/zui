//go:build zui

package zview

type NativeView struct {
	baseNativeView
	View      View
	Presented bool
	// allChildrenPresented bool
	DoOnRemove []func() // anything that needs to be stopped
	DoOnAdd    []func() // anything that needs to be stopped
	DoOnReady  []func() // anything that needs to be stopped
}

type DragType string

const (
	DragEnter    DragType = "enter"
	DragLeave    DragType = "leave"
	DragOver     DragType = "over"
	DragDrop     DragType = "drop"
	DragDropFile DragType = "file"
)

var (
	ChildOfViewFunc             func(v View, path string) View
	RangeAllVisibleChildrenFunc func(root View, got func(View) bool)
)

func (v *NativeView) AddOnRemoveFunc(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.DoOnRemove = append(v.DoOnRemove, f)
}

func (v *NativeView) AddOnAddFunc(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.DoOnAdd = append(v.DoOnAdd, f)
}

func (v *NativeView) AddOnReadyFunc(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.DoOnReady = append(v.DoOnReady, f)
}

func (v *NativeView) PerformAddRemoveFuncs(add bool) {
	if add {
		for _, f := range v.DoOnAdd {
			f()
		}
	} else {
		for _, f := range v.DoOnRemove {
			f()
		}
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
