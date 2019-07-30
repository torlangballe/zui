package zgo

func LabelNew(text string) *Label {
	label := &Label{}
	bvh := &ViewBaseHandler{}
	tbh := &TextBaseHandler{}
	label.TextBase = tbh
	e := DocumentJS.Call("createElement", "label")
	e.Set("style", "position:absolute")
	v := ViewNative(e)
	//	label.native = &v
	bvh.native = &v
	label.View = bvh
	tbh.view = label
	textNode := DocumentJS.Call("createTextNode", text)
	e.Call("appendChild", textNode)
	f := FontNice(18, FontNormal)
	label.Font(f)
	return label
}
