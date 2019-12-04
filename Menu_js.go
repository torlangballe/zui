package zgo

import (
	"fmt"
	"reflect"
	"syscall/js"

	"github.com/torlangballe/zutil/zmap"
)

func MenuViewNew(vals Dictionary, value interface{}) *MenuView {
	m := &MenuView{}
	m.keyVals = vals
	sel := DocumentJS.Call("createElement", "select")
	m.Element = sel
	sel.Set("style", "position:absolute")
	m.View = m
	f := FontNice(18, FontStyleNormal)
	m.Font(f)
	//	m.style().Set("webkitAppearance", "none") -- to set to non-system look
	m.updateVals(vals, false, value)

	return m
}

func (v *MenuView) UpdateValues(vals Dictionary) {
	if !reflect.DeepEqual(v.keyVals, vals) {
		options := v.get("options")
		options.Set("length", 0)
		setFirst := (len(v.keyVals) == 0)
		v.updateVals(vals, setFirst, nil)
	}
}

func (v *MenuView) updateVals(vals Dictionary, setFirst bool, value interface{}) {
	for _, k := range zmap.GetSortedKeysFromSIMap(vals) {
		val := vals[k]
		option := DocumentJS.Call("createElement", "option")
		sval := fmt.Sprint(val)
		option.Set("value", sval)
		if setFirst || fmt.Sprint(value) == sval {
			option.Set("selected", "true")
		}
		option.Set("innerHTML", k)
		v.call("appendChild", option)
	}
}

func (v *MenuView) NameAndValue() (string, interface{}) {
	index := v.get("selectedIndex").Int()
	options := v.get("options")
	o := options.Index(index)
	name := o.Get("innerHTML").String()
	val := v.keyVals[name]
	return name, val
}

func (v *MenuView) ChangedHandler(handler func(key string, val interface{})) {
	v.changed = handler
	v.set("onchange", js.FuncOf(func(js.Value, []js.Value) interface{} {
		if v.changed != nil {
			str, val := v.NameAndValue()
			v.changed(str, val)
		}
		return nil
	}))
}

// https://stackoverflow.com/questions/23718753/javascript-to-create-a-dropdown-list-and-get-the-selected-value
// https://stackoverflow.com/questions/17001961/how-to-add-drop-down-list-select-programmatically
