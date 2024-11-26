//go:build zui

package zmenu

import (
	"math/rand"
	"strings"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zmap"
)

type MenuedSetOwner[A any] struct {
	MenuedOwner
	MakeItemFunc   func() A
	DeleteItemFunc func(id int64, a A)
	AddTitle       string
	AddPromptText  string
	values         map[int64]storage[A]
	itemStoreKey   string
}

type storage[A any] struct {
	Name  string
	Value A
}

const (
	deleteValue = "$delete"
	addValue    = "$add"
)

var CustomID int64 = -1 // must be int64 not const, as set to any which assumes in64

func NewSetOwner[A any](storeKey string) *MenuedSetOwner[A] {
	o := &MenuedSetOwner[A]{}
	o.Init(storeKey)
	return o
}

func (o *MenuedSetOwner[A]) Init(storeKey string) {
	o.MenuedOwner.Init()
	o.StoreKey = storeKey
	o.AddPromptText = "New item name:"
	o.itemStoreKey = storeKey + ".items"
	o.values = map[int64]storage[A]{}
	zkeyvalue.DefaultStore.GetObject(o.itemStoreKey, &o.values)
	o.CreateItemsFunc = o.createItems
	o.DefaultSelected = CustomID
}

func (o *MenuedSetOwner[A]) saveToStore() {
	// zlog.Info("MSO: save", o.itemStoreKey, zlog.Full(o.values))
	zkeyvalue.DefaultStore.SetObject(o.values, o.itemStoreKey, true)
}

func (o *MenuedSetOwner[A]) Set(id int64, value A) {
	s := o.values[id]
	s.Value = value
	o.values[id] = s
	o.saveToStore()
}

func (o *MenuedSetOwner[A]) Get(id int64) A {
	return o.values[id].Value
}

func (o *MenuedSetOwner[A]) CurrentID() int64 {
	item := o.SelectedItem()
	if item != nil {
		return item.Value.(int64)
	}
	return 0
}

func (o *MenuedSetOwner[A]) GetSelected() A {
	return o.Get(o.CurrentID())
}

func (o *MenuedSetOwner[A]) createItems() []MenuedOItem {
	var items []MenuedOItem
	var selID int64

	if o.SelectedValue() != nil {
		selID = o.SelectedValue().(int64)
	}
	isCustom := (selID == CustomID)
	if o.DefaultSelected == CustomID {
		items = append(items, MenuedOItem{Name: "Custom", Value: CustomID, Selected: isCustom})
		items = append(items, MenuedOItemSeparator)
	}
	ids, store := zmap.KeySortedKeyValues(o.values, func(a, b storage[A]) bool {
		return strings.Compare(a.Name, b.Name) < 0
	})
	// zlog.Info("MSO items1:", len(o.values), o.values, selID, isCustom, ids, store)
	for i, id := range ids {
		if id == -1 {
			continue
		}
		name := store[i].Name
		items = append(items, MenuedOItem{Name: name, Value: id, Selected: id == selID})
		// zlog.Info("MSO add:", name, id)
	}
	empty := (len(o.values) == 0)
	if !empty {
		items = append(items, MenuedOItemSeparator)
	}
	add := MenuedAction("Add", -2)
	add.Function = func() {
		zalert.PromptForText(o.AddPromptText, "", func(str string) {
			var result storage[A]
			id := rand.Int63()
			result.Name = str
			if o.MakeItemFunc != nil {
				result.Value = o.MakeItemFunc()
			}
			o.values[id] = result
			// zlog.Info("MSO Added1:", len(o.values), id)
			o.SetSelectedValue(id)
			// zlog.Info("MSO Added:", len(o.values), id)
		})
	}
	items = append(items, add)
	if !empty {
		del := MenuedAction("Remove…", -3)
		del.IsDisabled = isCustom || len(ids) == 0
		del.Function = func() {
			item := o.SelectedItem()
			if item != nil {
				id := item.Value.(int64)
				if id != 0 {
					str := "Are you sure you want to delete item '" + item.Name + "'?"
					zalert.Ask(str, func(ok bool) {

					})
				}
			}
		}
		items = append(items, del)
	}
	return items
}
