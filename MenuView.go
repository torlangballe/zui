package zui

import (
	"reflect"

	"github.com/torlangballe/zutil/zgeo"
)

type MenuView struct {
	NativeView
	maxWidth float64
	changed  func(id, name string, value interface{})
	items    MenuItems
	oldID    string
	oldValue interface{}

	StaticName string // if set, user can't set a different value, but can press and see them. Adds StaticName as first item
}

type MenuItem struct {
	ID    string
	Name  string
	Value interface{}
}
type MenuItems interface {
	GetItem(i int) (id, name string, value interface{}) // name=="" is past end
}

func MenuItemsLength(m MenuItems) int {
	for i := 0; ; i++ {
		id, _, _ := m.GetItem(i)
		if id == "" {
			return i
		}
	}
	return -1
}

func menuItemsIDForValue(m MenuItems, val interface{}) string {
	for i := 0; ; i++ {
		id, _, v := m.GetItem(i)
		if id == "" {
			return ""
		}
		if reflect.DeepEqual(val, v) {
			return id
		}
	}
	return ""
}

func menuItemsAreEqual(a, b MenuItems) bool {
	for i := 0; ; i++ {
		ai, _, av := a.GetItem(i)
		bi, _, bv := b.GetItem(i)
		if ai == "" && bi == "" {
			return true
		}
		if ai == "" || bi == "" {
			return false
		}
		if ai != bi {
			return false
		}
		if !reflect.DeepEqual(av, bv) {
			return false
		}
	}
	return false
}

func (v *MenuView) GetCalculatedSize(total zgeo.Size) zgeo.Size {
	maxString := v.StaticName
	if v.StaticName == "" {
		for i := 0; ; i++ {
			id, name, _ := v.items.GetItem(i)
			if id == "" {
				break
			}
			if len(name) > len(maxString) {
				maxString = name
			}
		}
	}
	maxString += "m"
	s := TextLayoutCalculateSize(zgeo.Left, menuViewGetHackedFontForSize(v.Font()), maxString, 1, v.maxWidth)
	// fmt.Println("MenuView calcedsize:", v.Font().Size, v.GetObjectName(), maxString, s)
	s.Add(zgeo.Size{32, 4})
	return s
}

func (v *MenuView) GetMaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuView) SetMaxWidth(max float64) View {
	v.maxWidth = max
	return v
}

func isSimpleValue(v interface{}) bool {
	rval := reflect.ValueOf(v)
	k := rval.Kind()
	return k != reflect.Slice && k != reflect.Struct && k != reflect.Map

}

func (v *MenuView) SetWithIdOrValue(o interface{}) {
	id, _, val := v.items.GetItem(0)
	if id == "" {
		return
	}
	if isSimpleValue(val) {
		id = menuItemsIDForValue(v.items, o)
	} else {
		id = o.(string)
	}
	// zlog.Debug("set", o, reflect.ValueOf(o).Type(), id)
	v.SetWithID(id)
}

func (v *MenuView) GetCurrentIdOrValue() interface{} {
	if v.oldValue == nil {
		return nil
	}
	// fmt.Println("MenuView GetCurrentIdOrValue", v.oldValue, isSimpleValue(v.oldValue), reflect.ValueOf(v.oldValue).Type())
	if isSimpleValue(v.oldValue) {
		return v.oldValue
	}
	return v.oldID
}
