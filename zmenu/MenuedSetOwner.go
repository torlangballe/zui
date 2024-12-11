//go:build zui

package zmenu

import (
	"math/rand"
	"strings"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zmap"
	"github.com/torlangballe/zutil/zwords"
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
	customName  = "Custom"
)

var CustomID int64 = -1 // must be int64 not const, as set to any which assumes in64

func NewSetOwner[A any](storeKey string) *MenuedSetOwner[A] {
	o := &MenuedSetOwner[A]{}
	o.Init(storeKey)
	return o
}

func (o *MenuedSetOwner[A]) SetAndSelectStoredValue(id int64, vals A) {
	o.values[id] = storage[A]{Name: customName, Value: vals}
	o.needsSave = true
	// zlog.Info("MSO.SetStoredValue:", id)
	o.SetSelectedValue(id)
}

func (o *MenuedSetOwner[A]) GetStoredValue(id int64) (A, bool) {
	s, got := o.values[id]
	return s.Value, got
}

func (o *MenuedSetOwner[A]) Init(storeKey string) {
	// zlog.Info("MSO.Init:")
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
	o.MenuedOwner.saveToStore()
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
	const (
		addID = iota
		delID
		renameID
		dupID
		saveID
		delAllID
	)
	var items []MenuedOItem
	var selID int64

	if o.SelectedValue() != nil {
		selID = o.SelectedValue().(int64)
	}
	isCustom := (selID == CustomID)
	// if o.DefaultSelected == CustomID {
	// zlog.Info("MSO items:", isCustom, selID)
	items = append(items, MenuedOItem{Name: customName, Value: CustomID, Selected: isCustom})
	items = append(items, MenuedOItemSeparator)
	// }
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
		// zlog.Info("MSO add:", name, id, id == selID)
	}
	if len(o.values) > 1 {
		items = append(items, MenuedOItemSeparator)
	}
	add := MenuedAction("Add…", addID)
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
			o.saveToStore()
			// zlog.Info("MSO Added:", len(o.values), id)
		})
	}
	items = append(items, add)
	selectedItem := o.SelectedItem()
	var qname string
	if selectedItem != nil {
		qname = "'" + selectedItem.Name + "'"
	}
	if isCustom {
		qname = customName
	}
	if selectedItem != nil && !isCustom {
		del := MenuedAction("Delete "+qname+"…", delID)
		del.Function = func() {
			id := selectedItem.Value.(int64)
			if id != 0 {
				str := "Are you sure you want to delete " + qname + "?"
				zalert.Ask(str, func(ok bool) {
					if !ok {
						return
					}
					delete(o.values, id)
					o.SetSelectedValue(CustomID)
					o.saveToStore()
				})
			}
		}
		items = append(items, del)
		rename := MenuedAction("Rename "+qname+"…", renameID)
		rename.Function = func() {
			id := selectedItem.Value.(int64)
			if id != 0 {
				title := "Rename from " + qname + " to:"
				zalert.PromptForText(title, selectedItem.Name, func(got string) {
					iv, has := o.values[id]
					if has {
						iv.Name = got
						o.values[id] = iv
						// o.SetSelectedValue(id)
						o.saveToStore()
					}
				})
			}
		}
		items = append(items, rename)
	}
	if selectedItem != nil {
		dup := MenuedAction("Duplicate "+qname+"…", dupID)
		dup.Function = func() {
			id := selectedItem.Value.(int64)
			if id != 0 {
				title := "New name for duplicate of " + qname + ":"
				zalert.PromptForText(title, selectedItem.Name, func(got string) {
					result := o.values[id]
					newID := rand.Int63()
					result.Name = got
					o.values[newID] = result
					o.SetSelectedValue(newID)
					o.saveToStore()
				})
			}
		}
		items = append(items, dup)
	}
	if len(o.values) > 6 {
		delAll := MenuedAction("Delete All…", delAllID)
		delAll.Function = func() {
			plural := zwords.Pluralize("preset", len(o.values)-1)
			str := "Are you sure you want to delete all " + plural + "?"
			zalert.Ask(str, func(ok bool) {
				if !ok {
					return
				}
				for k := range o.values {
					if k != CustomID {
						delete(o.values, k)
					}
				}
				o.SetSelectedValue(CustomID)
				o.saveToStore()
			})
		}
		items = append(items, delAll)

	}
	if o.needsSave {
		save := MenuedAction("Save Changes", saveID)
		save.Function = func() {
			o.saveToStore()
		}
		items = append(items, save)
	}
	return items
}
