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

func ScreenMain() Screen {
	return Screen{}
}

func zViewAddView(parent View, child View, index int) {
}

// CustomView

func zViewSetRect(view View, rect Rect, layout bool) { // layout only used on android
}

func (v *CustomView) drawIfExposed() {
}

func (c *CustomView) init(view View, name string) {
	vbh := ViewBaseHandler{}
	v := ViewNative{} //!
	vbh.native = &v
	vbh.view = view
	c.ViewBaseHandler = vbh // this must be set after vbh is set up
	view.ObjectName(name)
}
