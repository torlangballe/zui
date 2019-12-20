package zui

import (
	"fmt"
	"reflect"
	"syscall/js"

	"github.com/torlangballe/zutil/zdict"
)

func MenuViewNew(vals zdict.Items, value interface{}) *MenuView {
	v := &MenuView{}
	v.keyVals = vals
	sel := DocumentJS.Call("createElement", "select")
	v.Element = sel
	sel.Set("style", "position:absolute")
	v.View = v
	f := FontNice(18, FontStyleNormal)
	v.SetFont(f)
	//	v.style().Set("webkitAppearance", "none") -- to set to non-system look
	v.updateVals(vals, value)
	// fmt.Println("NVAL:", v.oldValue)

	v.set("onchange", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if v.IsStatic && v.oldValue != nil {
			v.SetValue(v.oldValue)
			return nil
		}
		sval := args[0].Get("currentTarget").Get("value").String()
		for i, kv := range v.keyVals {
			if fmt.Sprint(kv.Value) == sval {
				okv := v.keyVals[i]
				v.oldValue = &okv
				if v.changed != nil {
					v.changed(*v.oldValue)
				}
				break
			}
		}
		return nil
	}))
	return v
}

func (v *MenuView) UpdateValues(vals zdict.Items) {
	if !reflect.DeepEqual(v.keyVals, vals) {
		options := v.get("options")
		options.Set("length", 0)
		v.updateVals(vals, nil)
	}
}

func (v *MenuView) updateVals(vals zdict.Items, value interface{}) {
	// fmt.Println("updateVals:", value)
	for _, di := range vals {
		sval := fmt.Sprint(di.Value)
		option := DocumentJS.Call("createElement", "option")
		option.Set("value", sval)
		option.Set("innerHTML", di.Name)
		v.call("appendChild", option)
	}
	v.SetValue(value)
}

func (v *MenuView) SetValue(val interface{}) *MenuView {
	if len(v.keyVals) == 0 {
		return v
	}
	sval := fmt.Sprint(val)
	index := -1
	for i, kv := range v.keyVals {
		if fmt.Sprint(kv.Value) == sval {
			index = i
			break
		}
	}
	if index == -1 {
		index = 0
	}
	okv := v.keyVals[index]
	v.oldValue = &okv
	options := v.get("options")
	o := options.Index(index)
	o.Set("selected", "true")

	return v
}

func (v *MenuView) NameAndValue() *zdict.Item {
	return v.oldValue
	// index := v.get("selectedIndex").Int()
	// if index == -1 {
	// 	return nil
	// }
	// options := v.get("options")
	// o := options.Index(index)
	// name := o.Get("innerHTML").String()
	// return v.keyVals.FindName(name)
}

func (v *MenuView) ChangedHandler(handler func(item zdict.Item)) {
	v.changed = handler
}

// https://stackoverflow.com/questions/23718753/javascript-to-create-a-dropdown-list-and-get-the-selected-value
// https://stackoverflow.com/questions/17001961/how-to-add-drop-down-list-select-programmatically
