package zgo

func AddViewToRoot(child AnyView) {
}

// TextInfo
func (ti *TextInfo) getTextSize(noWidth bool) Size {
	return Size{}
}

// CustomView

func CustomViewInit() *ViewNative {
	return nil
}

// Alert

func (a *Alert) Show() {
}

// Screen

func ScreenMainRect() Rect {
	return Rect{}
}

func zViewAddView(parent View, child View, index int) {
}
