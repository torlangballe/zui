package zmenu

import (
	"fmt"
	"html"
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
)

func NewView(name string, items zdict.Items, value any) *MenuView {
	// zlog.Info("NewView:", name, value, reflect.TypeOf(value))
	v := &MenuView{}
	sel := zdom.DocumentJS.Call("createElement", "select")
	v.Element = sel
	sel.Set("style", "position:absolute")
	v.View = v
	// v.SetNativeMargin(zgeo.RectFromXY2(0, 0, 0, -12))
	// v.SetFont(zgeo.FontNice(14, zgeo.FontStyleNormal)) // Has to be 14!!!
	v.SetObjectName(name)
	if len(items) > 0 {
		v.UpdateItems(items, value, false)
	}
	v.JSSet("onchange", js.FuncOf(func(_ js.Value, args []js.Value) any {
		index := v.JSGet("selectedIndex").Int()
		zlog.Assert(index < len(v.items), "index too big", index, len(v.items))
		v.currentValue = v.items[index].Value
		if v.storeKey != "" {
			// zlog.Info("Selected:", v.storeKey, v.currentValue)
			zkeyvalue.DefaultStore.SetItem(v.storeKey, v.currentValue, true)
		}
		if v.selectedHandler != nil {
			v.selectedHandler()
		}
		return nil
	}))
	return v
}

func (v *MenuView) ReadyToShow(beforeWindow bool) {
	if beforeWindow && v.storeKey != "" {
		val, got := zkeyvalue.DefaultStore.GetItemAsAny(v.storeKey)
		if got {
			v.SelectWithValue(val)
		}
	}
}

func (v *MenuView) Empty() {
	options := v.JSGet("options")
	options.Set("length", 0)
	v.items = v.items[:0]
	v.currentValue = nil
}

func (v *MenuView) AddSeparator() {
	var item zdict.Item

	item.Name = MenuSeparatorID
	item.Value = nil
	v.items = append(v.items, item)
}

func (v *MenuView) ChangeNameForValue(name string, value any) {
	if zlog.ErrorIf(value == nil, v.ObjectName()) {
		return
	}
	for i, item := range v.items {
		if fmt.Sprint(item.Value) == fmt.Sprint(value) {
			options := v.JSGet("options")
			v.items[i].Name = name
			o := options.Index(i)
			o.Set("text", name)

			break
		}
	}
}

func (v *MenuView) AddItem(name string, value any) {
	option := zdom.DocumentJS.Call("createElement", "option")
	if name == MenuSeparatorID {
		option.Set("disabled", true)
		option.Set("class", "separator")
	} else {
		option.Set("value", fmt.Sprint(value))
		option.Set("text", name)
		// option.Set("innerHTML", name)
	}
	var item zdict.Item
	item.Name = name
	item.Value = value
	v.items = append(v.items, item)
	v.JSCall("appendChild", option)
}

func (v *MenuView) RemoveItemByValue(value any) {
	sval := fmt.Sprint(value)
	options := v.JSGet("options")
	for i, item := range v.items {
		if fmt.Sprint(item.Value) == sval {
			zslice.RemoveAt(&v.items, i)
			break
		}
	}
	for i := 0; i < options.Length(); i++ {
		v := options.Index(i).Get("value")
		if v.String() == sval {
			options.Call("remove", i)
			break
		}
	}
	if fmt.Sprint(v.currentValue) == sval {
		v.currentValue = nil
	}
}

func (v *MenuView) UpdateItems(items zdict.Items, value any, isAction bool) {
	// zlog.Info("MV UpdateItems", v.ObjectName(), len(items), v.RowFormat)
	v.items = items // must be before v.getNumberOfItemsString
	var str string
	for _, item := range v.items {
		rowStr := getRowString(v.RowFormat, item)
		str += fmt.Sprintf(`<option value="%s">%s</option>\n`, html.EscapeString(fmt.Sprint(item.Value)), html.EscapeString(rowStr))
	}
	// We use HTML here to add all at once, or slow.
	// zlog.Info(v.ObjectName(), "menu updateitems:", str)
	v.JSSet("innerHTML", str)
	v.SelectWithValue(value)
	//  zlog.Info("updateVals:", v.ObjectName(), value, setID)
}

func (v *MenuView) SelectWithValue(value any) bool {
	// zlog.Info("MV SelectWithValue:", value, zlog.CallingStackString())
	if value == nil {
		return false
		// if len(v.items) != 0 {
		// 	v.SelectWithValue(v.items[0].Value)
		// }
		// return true
	}
	// if zlog.ErrorIf(value == nil, v.ObjectName(), zlog.CallingStackString()) {
	// 	return
	// }
	// zlog.Info("MV SelectWithValue:", v.ObjectName(), value)
	for i, item := range v.items {
		// zlog.Info("MV SelectWithValue Set?:", item.Value, value, reflect.TypeOf(item.Value), reflect.TypeOf(value))
		if fmt.Sprint(item.Value) == fmt.Sprint(value) {
			v.currentValue = item.Value
			options := v.JSGet("options")
			// zlog.Info("MV SelectWithValue Set:", i, v.ObjectName(), len(v.items), options, value)
			o := options.Index(i)
			o.Set("selected", "true")
			return true
		}
	}
	return false
}

// func (v *MenuView) IDAndValue() (id string, value any) {
// 	return v.oldID, v.currentValue
// }

func (v *MenuView) SetSelectedHandler(handler func()) {
	v.selectedHandler = handler
}

// https://stackoverflow.com/questions/23718753/javascript-to-create-a-dropdown-list-and-get-the-selected-value
// https://stackoverflow.com/questions/17001961/how-to-add-drop-down-list-select-programmatically

func (v *MenuView) SetFont(font *zgeo.Font) {
	if font.Size != 14 {
		zlog.Fatal("can't set menu view font size to anything except 14 in js")
	}
	f := *font
	if zdevice.WasmBrowser() == zdevice.Safari {
		f.Size = 16
	}
	v.NativeView.SetFont(&f)
}

func (v *MenuView) SetRect(r zgeo.Rect) {
	if zdevice.WasmBrowser() == zdevice.Safari {
		r.Add(zgeo.RectFromXY2(2, 2, -2, -5))
	}
	v.NativeView.SetRect(r)
}
