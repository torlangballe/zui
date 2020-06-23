package zui

import (
	"fmt"
	"reflect"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type otherItem struct {
	ID    string
	Title string
}

type MenuView struct {
	NativeView
	maxWidth   float64
	changed    func(id, name string, value interface{})
	items      zdict.NamedValues
	oldID      string
	oldValue   interface{}
	otherItems []otherItem

	IsStatic bool // if set, user can't set a different value, but can press and see them. Shows number of items
}

var menuViewHeight = 22.0

// type MenuItem struct {
// 	ID    string
// 	Name  string
// 	Value interface{}
// }

func (v *MenuView) Dump() {
	zlog.Info("DumpMenu:", v.ObjectName())
	zdict.DumpNamedValues(v.items)
}

func (v *MenuView) getNumberOfItemsString() string {
	return WordsPluralizeString("%d %s", "en", float64(v.items.Count()), "item")
}

func (v *MenuView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var maxString string
	if v.IsStatic {
		maxString = "658 items" // make it big enough to not need to resize much
	} else {
		// zlog.Info("MV Calc size:", v)
		for i := 0; i < v.items.Count(); i++ {
			_, name, _ := v.items.GetItem(i)
			if len(name) > len(maxString) {
				maxString = name
			}
		}
		for _, oi := range v.otherItems {
			if len(oi.Title) > len(maxString) {
				maxString = oi.Title
			}
		}
	}
	maxString += "m"
	s := TextLayoutCalculateSize(zgeo.Left, v.Font(), maxString, 1, v.maxWidth, true)
	// zlog.Info("MenuView calcedsize:", v.Font().Size, v.ObjectName(), maxString, s)
	s.W += 32
	s.H = menuViewHeight
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	return s
}

func (v *MenuView) MaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuView) SetMaxWidth(max float64) View {
	v.maxWidth = max
	return v
}

func isSimpleValue(v interface{}) bool {
	rval := reflect.ValueOf(v)
	k := rval.Kind()
	_, isui := v.(UIStringer)
	_, isnv := v.(zdict.NVStringer)
	if isui || isnv {
		return true
	}
	return k != reflect.Slice && k != reflect.Struct && k != reflect.Map
}

func (v *MenuView) SetWithIdOrValue(o interface{}) {
	if v.items.Count() == 0 {
		return
	}
	_, _, val := v.items.GetItem(0)
	var id string
	if isSimpleValue(val) {
		id = zdict.NamedValuesIDForValue(v.items, o)
	} else {
		id = fmt.Sprint(o)
	}
	// zlog.Debug("set", o, reflect.ValueOf(o).Type(), id)
	if id != "" {
		v.SetWithID(id)
	}
}

func (v *MenuView) GetCurrentIdOrValue() interface{} {
	if v.oldValue == nil {
		return nil
	}
	// zlog.Info("MenuView GetCurrentIdOrValue", v.oldValue, p(v.oldValue), reflect.ValueOf(v.oldValue).Type())
	if isSimpleValue(v.oldValue) {
		return v.oldValue
	}
	return v.oldID
}
