package zgo

// func imageViewInit(iv *ImageView, path string) {
// 	vbh := ViewBaseHandler{}
// 	e := DocumentJS.Call("createElement", "img")
// 	e.Set("style", "position:absolute")
// 	v := ViewNative(e)
// 	vbh.native = &v
// 	vbh.view = iv
// 	iv.ViewBaseHandler = vbh // this must be set after vbh is set up
// 	vbh.view.ObjectName(path)
// 	e.Set("className", "widget")
// 	if iv.image == nil {
// 		iv.SetImage(nil, path, func() {
// 			iv.Expose()
// 		})
// 	}
// }

// func getImageViewSize(view *ViewNative) Size {
// 	w := view.get("naturalWidth").Float()
// 	h := view.get("naturalHeight").Float()
// 	return Size{w, h}
// }

// func (v *ImageView) PressedHandler(handler func(pos Pos)) {
// 	v.pressed = handler
// 	v.native.set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
// 		if v.pressed != nil {
// 			v.pressed(Pos{})
// 		}
// 		return nil
// 	}))
// }
