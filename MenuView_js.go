package zui

import (
	"reflect"
	"syscall/js"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zlog"
)

const separatorID = "$sep"

func MenuViewNew(name string, items zdict.NamedValues, value interface{}, isStatic bool) *MenuView {
	v := &MenuView{}
	v.items = items

	v.IsStatic = isStatic
	sel := DocumentJS.Call("createElement", "select")
	v.Element = sel
	sel.Set("style", "position:absolute")
	v.View = v
	v.SetFont(FontNice(14, FontStyleNormal))
	// v.style().Set("webkitAppearance", "none") // to set to non-system look
	v.SetObjectName(name)
	v.SetValues(items, value)

	v.set("onchange", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		//			zlog.Info("menuview changed", v.ObjectName())
		if v.IsStatic && v.oldValue != nil {
			v.SetWithID(v.oldID)
			return nil
		}
		sid := args[0].Get("currentTarget").Get("value").String()
		for i := 0; i < v.items.Count(); i++ {
			id, in, iv := v.items.GetItem(i)
			if id == sid {
				v.oldID = id
				v.oldValue = iv
				if v.changed != nil {
					v.changed(id, in, iv)
				}
				return nil
			}
		}
		for _, oi := range v.otherItems {
			if oi.ID == sid {
				v.oldID = oi.ID
				v.oldValue = nil
				if v.changed != nil {
					v.changed(oi.ID, oi.Title, nil)
				}
			}
		}
		return nil
	}))
	return v
}

func (v *MenuView) Empty() {
	options := v.get("options")
	options.Set("length", 0)
	v.items = nil
	v.otherItems = v.otherItems[:0]
	v.oldID = ""
	v.oldValue = nil
}

func (v *MenuView) UpdateValues(items zdict.NamedValues) {
	// zlog.Info("MV UpdateValues", v.ObjectName(), v.items, items)
	if !zdict.NamedValuesAreEqual(v.items, items) {
		// zlog.Info("MV UpdateValues2")
		options := v.get("options")
		options.Set("length", 0)
		v.SetValues(items, nil)
	}
}

func (v *MenuView) AddSeparator() {
	v.AddAction(separatorID, "")
}

func (v *MenuView) AddAction(id, name string) {
	v.otherItems = append(v.otherItems, otherItem{ID: id, Title: name})
	v.menuViewAddItem(id, name)
}

func (v *MenuView) menuViewAddItem(id, name string) {
	option := DocumentJS.Call("createElement", "option")
	if id == separatorID {
		option.Set("disabled", true)
		option.Set("class", "separator")
	} else {
		option.Set("value", id)
		option.Set("title", id)
		option.Set("innerHTML", name)
	}
	v.call("appendChild", option)
}

func (v *MenuView) SetValues(items zdict.NamedValues, value interface{}) {
	var setID string

	v.items = items // must be before v.getNumberOfItemsString
	if v.IsStatic {
		// zlog.Info("Items:", v.getNumberOfItemsString())
		v.AddAction("$STATICNAME", v.getNumberOfItemsString())
		v.SetWithID("$STATICNAME")
	}
	if items == nil {
		return
	}
	for i := 0; i < v.items.Count(); i++ {
		in, id, iv := items.GetItem(i)
		if i == 0 {
			setID = id
		}
		if reflect.DeepEqual(value, iv) {
			setID = id
		}
		v.menuViewAddItem(in, id)
	}
	//  fmt.Println("updateVals:", v.ObjectName(), value, setID)
	if setID != "" {
		v.SetWithID(setID)
	}
}

func (v *MenuView) SetWithID(setID string) *MenuView {
	if zlog.ErrorIf(setID == "", v.ObjectName()) {
		zlog.Info(zlog.GetCallingStackString())
		return v
	}
	// fmt.Println("mv:setwithid:", setID, v.ObjectName())
	for i := 0; i < v.items.Count(); i++ {
		id, _, iv := v.items.GetItem(i)
		if id == setID {
			// fmt.Println("mv set:", id, i, len(v.otherItems))
			v.oldValue = iv
			v.oldID = id
			options := v.get("options")
			o := options.Index(i + len(v.otherItems))
			o.Set("selected", "true")
			break
		}
	}
	return v
}

func (v *MenuView) IDAndValue() (id string, value interface{}) {
	return v.oldID, v.oldValue
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
