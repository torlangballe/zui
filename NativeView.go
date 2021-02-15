package zui

type NativeView struct {
	baseNativeView
	Presented            bool
	allChildrenPresented bool
	stopOnClose          []func() // anything that needs to be stopped
}

func (v *NativeView) AddStopper(f func()) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.stopOnClose = append(v.stopOnClose, f)
}

func (v *NativeView) StopStoppers() {
	// zlog.Info("StopStoppers:", v.View.ObjectName(), len(v.stopOnClose))
	for _, f := range v.stopOnClose {
		f()
	}
	ct, _ := v.View.((ContainerType))
	if ct != nil {
		for _, c := range ct.GetChildren() {
			nv := ViewGetNative(c)
			if nv != nil {
				nv.StopStoppers()
			}
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
