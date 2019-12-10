package zgo

import (
	"fmt"
	"reflect"
	"syscall/js"

	"github.com/torlangballe/zutil/zdict"
)

func MenuViewNew(vals zdict.Items, value interface{}) *MenuView {
	m := &MenuView{}
	m.keyVals = vals
	sel := DocumentJS.Call("createElement", "select")
	m.Element = sel
	sel.Set("style", "position:absolute")
	m.View = m
	f := FontNice(18, FontStyleNormal)
	m.Font(f)
	setFirst := (value == nil)
	//	m.style().Set("webkitAppearance", "none") -- to set to non-system look
	m.updateVals(vals, setFirst, value)
	fmt.Println("NVAL:", m.oldValue)

	return m
}

func (v *MenuView) UpdateValues(vals zdict.Items) {
	if !reflect.DeepEqual(v.keyVals, vals) {
		options := v.get("options")
		options.Set("length", 0)
		setFirst := (len(v.keyVals) == 0)
		v.updateVals(vals, setFirst, nil)
	}
}

func (v *MenuView) updateVals(vals zdict.Items, setFirst bool, value interface{}) {
	for _, di := range vals {
		option := DocumentJS.Call("createElement", "option")
		sval := fmt.Sprint(di.Value)
		option.Set("value", sval)
		if setFirst || fmt.Sprint(value) == sval {
			option.Set("selected", "true")
			o := di
			v.oldValue = &o
		}
		option.Set("innerHTML", di.Name)
		v.call("appendChild", option)
	}
	fmt.Println("NVAL:", v.oldValue)
}

func (v *MenuView) SetValue(val interface{}) *MenuView {
	sval := fmt.Sprint(val)
	for i, kv := range v.keyVals {
		if fmt.Sprint(kv.Value) == sval {
			okv := kv
			v.oldValue = &okv
			options := v.get("options")
			o := options.Index(i)
			fmt.Println("MV Set:", i, o, val)
			o.Set("selected", "true")
			break
		}
	}

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
	v.set("onchange", js.FuncOf(func(js.Value, []js.Value) interface{} {
		old := v.NameAndValue()
		fmt.Println("Change:", old, v.IsStatic)
		if v.IsStatic && old != nil {
			v.SetValue(old.Value)
			return nil
		}
		if v.changed != nil {
			nv := v.NameAndValue()
			if nv != nil {
				v.changed(*nv)
			}
		}
		return nil
	}))
}

// https://stackoverflow.com/questions/23718753/javascript-to-create-a-dropdown-list-and-get-the-selected-value
// https://stackoverflow.com/questions/17001961/how-to-add-drop-down-list-select-programmatically
