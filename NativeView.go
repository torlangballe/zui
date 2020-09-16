package zui

type NativeView struct {
	baseNativeView
	Presented            bool
	allChildrenPresented bool
}
