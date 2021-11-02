package zui

import (
	"fmt"
	"html"
	"reflect"
	"syscall/js"

	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func MenuViewNew(name string, items zdict.Items, value interface{}) *MenuView {
	v := &MenuView{}
	sel := DocumentJS.Call("createElement", "select")
	v.Element = sel
	sel.Set("style", "position:absolute")
	v.View = v
	v.SetFont(zgeo.FontNice(14, zgeo.FontStyleNormal))
	v.SetObjectName(name)
	v.UpdateAndSelect(items, value)

	v.setjs("onchange", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		//			zlog.Info("menuview selected", v.ObjectName())
		index := v.getjs("selectedIndex").Int()
		zlog.Assert(index < len(v.items))
		v.currentValue = v.items[index].Value
		// zlog.Info("Selected:", index, v.items[index].Name, v.items[index].Value)
		if v.selectedHandler != nil {
			v.selectedHandler()
		}
		return nil
	}))
	return v
}

func (v *MenuView) Empty() {
	options := v.getjs("options")
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

func (v *MenuView) ChangeNameForValue(name string, value interface{}) {
	if zlog.ErrorIf(value == nil, v.ObjectName()) {
		return
	}
	for i, item := range v.items {
		if reflect.DeepEqual(item.Value, value) {
			options := v.getjs("options")
			// zlog.Info("MV ChangeNameForValue:", i, v.ObjectName(), name, value)
			v.items[i].Name = name
			o := options.Index(i)
			o.Set("text", name)

			break
		}
	}
}

func (v *MenuView) menuViewAddItem(name string, value interface{}) {
	option := DocumentJS.Call("createElement", "option")
	if name == MenuSeparatorID {
		option.Set("disabled", true)
		option.Set("class", "separator")
	} else {
		id := fmt.Sprint(value)
		option.Set("value", id)
		option.Set("text", id)
		// option.Set("innerHTML", name)
	}
	v.call("appendChild", option)
}

func (v *MenuView) UpdateAndSelect(items zdict.Items, value interface{}) {
	// zlog.Info("MV SetAndSelect:", v, items)
	v.UpdateItems(items, []interface{}{value})
	v.SelectWithValue(value)
}

func (v *MenuView) UpdateItems(items zdict.Items, values []interface{}) {
	// zlog.Info("MV SetValues1", v.ObjectName(), len(items), len(v.items), items.Equal(v.items))
	if !items.Equal(v.items) {
		v.items = items // must be before v.getNumberOfItemsString
		var str string
		for _, item := range v.items {
			str += fmt.Sprintf(`<option value="%s">%s</option>\n`, html.EscapeString(fmt.Sprint(item.Value)), html.EscapeString(item.Name))
		}
		// We use HTML here to add all at once, or slow.
		v.setjs("innerHTML", str)
	}
	if len(values) != 0 {
		v.SelectWithValue(values[0])
	}
	//  zlog.Info("updateVals:", v.ObjectName(), value, setID)
}

func (v *MenuView) SelectWithValue(value interface{}) {
	if zlog.ErrorIf(value == nil, v.ObjectName()) {
		return
	}
	// zlog.Info("MV SelectWithValue:", v.ObjectName(), value)
	for i, item := range v.items {
		if reflect.DeepEqual(item.Value, value) {
			v.currentValue = item.Value
			options := v.getjs("options")
			// zlog.Info("MV SelectWithValue Set:", i, v.ObjectName(), len(v.items), options, value)
			o := options.Index(i)
			o.Set("selected", "true")
			break
		}
	}
}

// func (v *MenuView) IDAndValue() (id string, value interface{}) {
// 	return v.oldID, v.currentValue
// }

func (v *MenuView) SetSelectedHandler(handler func()) {
	v.selectedHandler = handler
}

// https://stackoverflow.com/questions/23718753/javascript-to-create-a-dropdown-list-and-get-the-selected-value
// https://stackoverflow.com/questions/17001961/how-to-add-drop-down-list-select-programmatically

func (v *MenuView) SetFont(font *zgeo.Font) {
	if font.Size != 14 {
		panic("can't set menu view font size to anything except 14 in js")
	}
	f := *font
	if zdevice.WasmBrowser() == "Safari" {
		f.Size = 16
	}
	v.NativeView.SetFont(&f)
}

func (v *MenuView) SetRect(r zgeo.Rect) {
	v.NativeView.SetRect(r.PlusPos(zgeo.Pos{0, 2}))
}
