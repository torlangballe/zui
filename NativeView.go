package zui

import (
	"github.com/torlangballe/zutil/ztimer"
)

type NativeView struct {
	baseNativeView
	Presented            bool
	allChildrenPresented bool
	stopOnClose          []ztimer.Stopper // anything that needs to be stopped
}

func (v *NativeView) AddStopper(s ztimer.Stopper) {
	// zlog.Info("AddStopper:", v.ObjectName())
	v.stopOnClose = append(v.stopOnClose, s)
}

func (v *NativeView) StopStoppers() {
	// zlog.Info("StopStoppers:", v.View.ObjectName(), len(v.stopOnClose))
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
