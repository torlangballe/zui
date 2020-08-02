package zui

import (
	"fmt"
	"reflect"
	"syscall/js"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zlog"
)

const separatorID = "$sep"

func MenuViewNew(name string, items zdict.Items, value interface{}, isStatic bool) *MenuView {
	v := &MenuView{}
	v.IsStatic = isStatic
	sel := DocumentJS.Call("createElement", "select")
	v.Element = sel
	sel.Set("style", "position:absolute")
	v.View = v
	v.SetFont(FontNice(14, FontStyleNormal))
	// v.style().Set("webkitAppearance", "none") // to set to non-system look
	v.SetObjectName(name)
	v.SetAndSelect(items, value)

	v.set("onchange", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		//			zlog.Info("menuview changed", v.ObjectName())
		if v.IsStatic && v.currentValue != nil {
			v.SelectWithValue(v.currentValue)
			return nil
		}
		index := v.get("selectedIndex").Int()
		zlog.Info("Selected:", index)
		zlog.Assert(index < len(v.items))
		v.currentValue = v.items[index].Value
		if v.changed != nil {
			v.changed(v.items[index].Name, v.items[index].Value)
		}
		return nil
	}))
	return v
}

func (v *MenuView) Empty() {
	options := v.get("options")
	options.Set("length", 0)
	v.items = v.items[:0]
	v.currentValue = nil
}

func (v *MenuView) AddSeparator() {
	var item zdict.Item

	item.Name = separatorID
	item.Value = nil
	v.items = append(v.items, item)
}

// func (v *MenuView) AddAction(id, name string) {
// 	v.otherItems = append(v.otherItems, otherItem{ID: id, Title: name})
// 	v.menuViewAddItem(id, name)
// }

func (v *MenuView) menuViewAddItem(name string, value interface{}) {
	option := DocumentJS.Call("createElement", "option")
	if name == separatorID {
		option.Set("disabled", true)
		option.Set("class", "separator")
	} else {
		id := fmt.Sprint(value)
		option.Set("value", id)
		option.Set("title", id)
		option.Set("innerHTML", name)
	}
	v.call("appendChild", option)
}

func (v *MenuView) SetAndSelect(items zdict.Items, value interface{}) {
	// zlog.Info("MV SetAndSelect:", v.ObjectName(), value)
	v.SetValues(items)
	v.SelectWithValue(value)
}

func (v *MenuView) SetValues(items zdict.Items) {
	// zlog.Info("MV SetValues1", v.ObjectName(), len(items), len(v.items), items.Equal(v.items))
	if !items.Equal(v.items) {
		old := v.currentValue // we need to remember current value to re-set as Empty() below clears it
		v.Empty()
		v.items = items // must be before v.getNumberOfItemsString
		if v.IsStatic {
			// zlog.Info("Items:", v.getNumberOfItemsString())
			// v.AddAction("$STATICNAME", v.getNumberOfItemsString())
			// v.SetWithID("$STATICNAME")
		}
		for _, item := range v.items {
			// zlog.Info("MV SetValues:", v.ObjectName(), i, len(items), item.Name)
			v.menuViewAddItem(item.Name, item.Value)
		}
		if old != nil {
			v.SelectWithValue(old)
		}
	}
	//  zlog.Info("updateVals:", v.ObjectName(), value, setID)
}

func (v *MenuView) SelectWithValue(value interface{}) *MenuView {
	if zlog.ErrorIf(value == nil, v.ObjectName(), zlog.GetCallingStackString()) {
		return v
	}
	// zlog.Info("MV SelectWithValue:", v.ObjectName(), value)
	for i, item := range v.items {
		if reflect.DeepEqual(item.Value, value) {
			v.currentValue = item.Value
			options := v.get("options")
			// zlog.Info("MV SelectWithValue Set:", i, v.ObjectName(), len(v.items), options, value)
			o := options.Index(i)
			o.Set("selected", "true")
			break
		}
	}
	return v
}

// func (v *MenuView) IDAndValue() (id string, value interface{}) {
// 	return v.oldID, v.currentValue
// }

func (v *MenuView) ChangedHandler(handler func(name string, value interface{})) {
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
