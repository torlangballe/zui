package zui

import "github.com/torlangballe/zutil/ztimer"

type NativeView struct {
	baseNativeView
	Presented            bool
	allChildrenPresented bool
	stopOnClose          []ztimer.Stopper // anything that needs to be stopped
}

func (v *NativeView) AddStopper(s ztimer.Stopper) {
	v.stopOnClose = append(v.stopOnClose, s)
}

func (v *NativeView) StopStoppers() {
	for _, s := range v.stopOnClose {
		s.Stop()
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
