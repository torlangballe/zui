// +build zui

package zui

type NativeView struct {
	baseNativeView
	View      View
	Presented bool
	// allChildrenPresented bool
	doOnRemove []func() // anything that needs to be stopped
	doOnAdd    []func() // anything that needs to be stopped
	doOnReady  []func() // anything that needs to be stopped
}

type DragType string

const (
	DragEnter    DragType = "enter"
	DragLeave    DragType = "leave"
	DragOver     DragType = "over"
	DragDrop     DragType = "drop"
	DragDropFile DragType = "file"
)

func (v *NativeView) AddOnRemoveFunc(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.doOnRemove = append(v.doOnRemove, f)
}

func (v *NativeView) AddOnAddFunc(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.doOnAdd = append(v.doOnAdd, f)
}

func (v *NativeView) AddOnReadyFunc(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.doOnReady = append(v.doOnReady, f)
}

func (v *NativeView) PerformAddRemoveFuncs(add bool) {
	if add {
		for _, f := range v.doOnAdd {
			f()
		}
	} else {
		for _, f := range v.doOnRemove {
			f()
		}
	}
	// ct, _ := v.View.(ContainerType)
	// if ct != nil {
	// 	for _, c := range ct.GetChildren(true) {
	// 		nv := ViewGetNative(c)
	// 		if nv != nil {
	// 			nv.PerformAddRemoveFuncs(add)
	// 		}
	// 	}
	// }
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

