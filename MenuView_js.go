package zui

import (
	"github.com/torlangballe/zutil/zlog"
	"reflect"
	"syscall/js"
)

func MenuViewNew(name string, items MenuItems, value interface{}, staticName string) *MenuView {
	v := &MenuView{}
	v.items = items
	v.StaticName = staticName
	sel := DocumentJS.Call("createElement", "select")
	v.Element = sel
	sel.Set("style", "position:absolute")
	v.View = v
	v.SetFont(FontNice(14, FontStyleNormal))
	// v.style().Set("webkitAppearance", "none") // to set to non-system look
	v.SetObjectName(name)
	v.updateVals(items, value)

	v.set("onchange", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if v.StaticName != "" && v.oldValue != nil {
			v.SetWithID(v.oldID)
			return nil
		}
		sid := args[0].Get("currentTarget").Get("value").String()
		for i := 0; ; i++ {
			id, in, iv := v.items.GetItem(i)
			if id == "" {
				break
			}
			zlog.Info("menuview changed", v.GetObjectName(), id, sid)
			if id == sid {
				v.oldID = id
				v.oldValue = iv
				if v.changed != nil {
					v.changed(id, in, iv)
				}
				break
			}
		}
		return nil
	}))
	return v
}

func (v *MenuView) UpdateValues(items MenuItems) {
	if !menuItemsAreEqual(v.items, items) {
		options := v.get("options")
		options.Set("length", 0)
		v.updateVals(items, nil)
	}
}

func (v *MenuView) menuViewAddItem(id, name string) {
	option := DocumentJS.Call("createElement", "option")
	option.Set("value", id)
	option.Set("innerHTML", name)
	v.call("appendChild", option)
}

func (v *MenuView) updateVals(items MenuItems, value interface{}) {
	var setID string
	if v.StaticName != "" {
		v.menuViewAddItem("$STATICNAME", v.StaticName)
	}
	if items == nil {
		return
	}
	for i := 0; ; i++ {
		in, id, iv := items.GetItem(i)
		if i == 0 {
			setID = id
		}
		if id == "" {
			break
		}
		if reflect.DeepEqual(value, iv) {
			setID = id
		}
		v.menuViewAddItem(in, id)
	}
	v.items = items
	//  fmt.Println("updateVals:", v.GetObjectName(), value, setID)
	if setID != "" {
		v.SetWithID(setID)
	}
}

func (v *MenuView) SetWithID(setID string) *MenuView {
	if zlog.ErrorIf(setID == "", v.GetObjectName()) {
		zlog.Info(zlog.GetCallingStackString())
		return v
	}
	//  fmt.Println("mv:setwithid:", setID, v.GetObjectName())
	for i := 0; ; i++ {
		id, _, iv := v.items.GetItem(i)
		if id == "" {
			break
		}
		if id == setID {
			v.oldValue = iv
			v.oldID = id
			options := v.get("options")
			o := options.Index(i)
			o.Set("selected", "true")
			break
		}
	}
	return v
}

// func (v *MenuView) SetValue(val interface{}) *MenuView {
// 	oi, on, ov := v.items.GetItem(0)
// 	if oi == "" {
// 		return v
// 	}
// 	index := 0
// 	for i := 0; ; i++ {
// 		id, _, iv := v.items.GetItem(i)
// 		if id == "" {
// 			break
// 		}
// 		if oi == id {
// 			ov = iv
// 			index = i
// 			break
// 		}
// 	}
// 	v.oldValue = ov
// 	v.oldID = oi
// 	options := v.get("options")
// 	o := options.Index(index)
// 	o.Set("selected", "true")

// 	return v
// }

func (v *MenuView) IDAndValue() (id string, value interface{}) {
	return v.oldID, v.oldValue
	// index := v.get("selectedIndex").Int()
	// if index == -1 {
	// 	return nil
	// }
	// options := v.get("options")
	// o := options.Index(index)
	// name := o.Get("innerHTML").String()
	// return v.keyVals.FindName(name)
}

func (v *MenuView) ChangedHandler(handler func(id, name string, value interface{})) {
	v.changed = handler
}

// https://stackoverflow.com/questions/23718753/javascript-to-create-a-dropdown-list-and-get-the-selected-value
// https://stackoverflow.com/questions/17001961/how-to-add-drop-down-list-select-programmatically

func (v *MenuView) SetFont(font *Font) View {
	if font.Size != 14 {
		panic("can't set menu view font size to anything except 14 in js")
	}
	f := *font
	if DeviceWasmBrowser() == "Safari" {
		f.Size = 16
	}
	v.NativeView.SetFont(&f)
	return v
}
