package zgo

func NativeViewAddToRoot(n *NativeView) {
}

// TextInfo
func (ti *TextInfo) getTextSize(noWidth bool) Size {
	return Size{}
}

// CustomView

func CustomViewInit() *NativeView {
	return nil
}

// Alert

func (a *Alert) Show() {
}

// Screen

func ScreenMain() Screen {
	return Screen{}
}

func zViewAddView(parent View, child View, index int) {
}

// CustomView

func zViewSetRect(view View, rect Rect, layout bool) { // layout only used on android
}
