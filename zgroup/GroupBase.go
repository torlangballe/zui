//go:build zui

package zgroup

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmap"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
)

type GroupItem struct {
	ID             string
	Create         func(id string, delete bool) zview.View
	View           zview.View
	ChildAlignment zgeo.Alignment
	Image          *zimage.Image
}

type GroupBase struct {
	zcontainer.StackView
	StoreKey                   string
	ChildView                  zview.View
	CurrentID                  string
	GroupItems                 map[string]*GroupItem
	SetIndicatorSelectionFunc  func(id string, selected bool)
	RemoveIndicatorFunc        func(id string)
	UpdateCurrentIndicatorFunc func(text string)
	HandleAddItemFunc          func()
	HandleDeleteItemFunc       func(id string)
	HasDataChangedInIDsFunc    func() []string
	header                     *zcontainer.StackView
	changedHandlerFunc         func(newID string)
	addButton                  *zimageview.ImageView
	deleteButton               *zimageview.ImageView
	GetAskToDeleteStringFunc   func(id string) string
	Data                       any // Data might be data/slice or something
}

type Grouper interface {
	AddItem(id, title, imagePath string, set bool, view zview.View, create func(id string, delete bool) zview.View)
	RemoveItem(id string)
	SelectItem(id string, done func())
	GetHeader() *zcontainer.StackView // can return nil
	GetCurrentID() string
	GetGroupBase() *GroupBase
	GetStoreKey() string
	// SetChangedHandler(handler func(newID string))
}

func (v *GroupBase) Init() {
	// zlog.Info("GroupBase Init", zlog.GetCallingStackString())
	v.GroupItems = map[string]*GroupItem{}
	v.GetAskToDeleteStringFunc = func(id string) string {
		return "Do you want to delete the selected item?"
	}
	v.HasDataChangedInIDsFunc = func() []string { return nil }
}

func (v *GroupBase) GetStoreKey() string {
	return v.StoreKey
}

func (v *GroupBase) GetHeader() *zcontainer.StackView {
	return v.header
}

func makeStoreKey(store string) string {
	return "zgroup.GroupBase." + store
}

func (v *GroupBase) ReadyToShow(beforeWindow bool) {
	if !beforeWindow {
		return
	}
	var setID string
	if setID == "" && v.CurrentID == "" && len(v.GroupItems) > 0 {
		keys := sort.StringSlice(zmap.GetKeysAsStrings(v.GroupItems))
		keys.Sort()
		setID = v.GroupItems[keys[0]].ID
	}
	if setID != "" {
		g := v.View.(Grouper)
		g.SelectItem(setID, nil)
	}
	v.UpdateButtons()
}

func (v *GroupBase) GetCurrentID() string {
	return v.CurrentID
}

func (v *GroupBase) SetChangedHandler(handler func(newID string)) {
	v.changedHandlerFunc = handler
}

func (v *GroupBase) AddGroupItem(id string, view zview.View, create func(id string, delete bool) zview.View) {
	groupItem := &GroupItem{}
	groupItem.ID = id
	groupItem.ChildAlignment = zgeo.Left | zgeo.Top | zgeo.Expand
	groupItem.View = view
	groupItem.Create = create
	v.GroupItems[id] = groupItem
}

func (v *GroupBase) FindGroupItem(id string) *GroupItem {
	for _, gi := range v.GroupItems {
		if gi.ID == id {
			return gi
		}
	}
	return nil
}

func (v *GroupBase) SetChildAlignment(id string, a zgeo.Alignment) {
	gi := v.FindGroupItem(id)
	gi.ChildAlignment = a
}

func (v *GroupBase) SetGroupItem(id string, done func()) {
	if v.CurrentID == id {
		if done != nil {
			done()
		}
		return
	}
	if v.CurrentID != "" {
		v.GroupItems[v.CurrentID].Create(v.CurrentID, true)
		v.SetIndicatorSelectionFunc(v.CurrentID, false)
	}
	if v.StoreKey != "" {
		zkeyvalue.DefaultStore.SetString(id, makeStoreKey(v.StoreKey), true)
	}
	if v.ChildView != nil {
		v.RemoveChild(v.ChildView)
	}
	if id == "" {
		v.ChildView = nil
		v.CurrentID = ""
		return
	}
	groupItem := v.GroupItems[id]
	v.ChildView = groupItem.Create(id, false)
	// zlog.Info("SetGroupItem:", id, v.ChildView != nil)
	v.Add(v.ChildView, groupItem.ChildAlignment)
	v.CurrentID = id
	v.SetIndicatorSelectionFunc(id, true)
	if !v.Presented {
		return
	}
	zview.ExposeView(v.View)
	v.ArrangeChildren() // This can create table rows and do all kinds of things that load images etc.
	zpresent.CallReady(v.ChildView, false)
	if v.changedHandlerFunc != nil {
		v.changedHandlerFunc(id)
	}
	ct := v.View.(zcontainer.ContainerType)
	zcontainer.WhenContainerLoaded(ct, func(waited bool) {
		// if waited { // if we waited for some loading, caused by above arranging, lets re-arrange
		// v.ArrangeChildren()
		// zlog.Info("setGroupItem loaded")
		zcontainer.ArrangeChildrenAtRootContainer(v)
		// }
		if done != nil {
			done()
		}
	})
}

func (v *GroupBase) RemoveGroupItem(id string) {
	var newID string
	for k := range v.GroupItems {
		if k != id {
			newID = k
			break
		}
	}
	v.SetGroupItem(newID, nil)
	delete(v.GroupItems, id)
	v.UpdateButtons()
	if v.RemoveIndicatorFunc != nil {
		v.RemoveIndicatorFunc(id)
	}
}

func (v *GroupBase) handleDeletePressed() {
	if v.HandleDeleteItemFunc != nil {
		v.HandleDeleteItemFunc(v.CurrentID)
	}
	g := v.View.(Grouper)
	g.RemoveItem(v.CurrentID)
}

func (v *GroupBase) UpdateButtons() {
	if v.deleteButton != nil {
		v.deleteButton.SetUsable(len(v.GroupItems) > 0)
	}
}

func (v *GroupBase) AddEditing() {
	v.deleteButton = zimageview.New(nil, "images/minus-circled-darkgray.png", zgeo.Size{16, 16})
	v.header.Add(v.deleteButton, zgeo.CenterRight, zgeo.Size{-6, 0})
	v.deleteButton.SetPressedHandler(func() {
		if v.GetAskToDeleteStringFunc != nil {
			text := v.GetAskToDeleteStringFunc(v.GetCurrentID())
			zalert.Ask(text, func(ok bool) {
				if ok {
					v.handleDeletePressed()
				}
			})
			return
		}
		v.handleDeletePressed()
	})
	v.addButton = zimageview.New(nil, "images/plus-circled-darkgray.png", zgeo.Size{16, 16})
	v.header.Add(v.addButton, zgeo.CenterRight)
	v.addButton.SetPressedHandler(func() {
		if v.HandleAddItemFunc != nil {
			v.HandleAddItemFunc()
		}
	})
}

func IndexForIDFromSlice(slicePtr any, id string) int {
	val := reflect.ValueOf(slicePtr).Elem()
	for i := 0; i < val.Len(); i++ {
		a := val.Index(i).Interface()
		indexID := zstr.GetIDFromAnySliceItemWithIndex(a, i)
		if indexID == id {
			return i
		}
	}
	return -1
}

type SliceGroupData struct {
	IndicatorID           string // IndicatorID is info about what part of subview is used as indicator
	SlicePtr              any
	SliceElementCheckSums []int64
}

func CreateSliceGroup(grouper Grouper, slicePtr any, setID string, indicatorFieldName string, create func(id string, delete bool) zview.View) {
	if setID == "" && grouper.GetStoreKey() != "" {
		setID, _ = zkeyvalue.DefaultStore.GetString(makeStoreKey(grouper.GetStoreKey()))
	}
	AddSliceItems(grouper, slicePtr, setID, indicatorFieldName, create)
	gb := grouper.GetGroupBase()
	data := new(SliceGroupData)
	data.SlicePtr = slicePtr
	gb.Data = data
	gb.HandleDeleteItemFunc = func(id string) {
		i := IndexForIDFromSlice(slicePtr, id)
		if i != -1 {
			zslice.RemoveAt(slicePtr, i)
		}
	}
	gb.HasDataChangedInIDsFunc = func() (changed []string) {
		val := reflect.ValueOf(slicePtr).Elem()
		var newSums []int64
		for i := 0; i < val.Len(); i++ {
			a := val.Index(i).Interface()
			cc := zstr.HashAnyToInt64(a, "")
			if val.Len() != len(data.SliceElementCheckSums) || cc != data.SliceElementCheckSums[i] {
				// zlog.Info("Changed:", i, len(data.SliceElementCheckSums))
				id := zstr.GetIDFromAnySliceItemWithIndex(a, i)
				changed = append(changed, id)
			}
			newSums = append(newSums, cc)
		}
		data.SliceElementCheckSums = newSums
		return changed
	}
	gb.HasDataChangedInIDsFunc() // we call this to set initial check sums

	val := reflect.ValueOf(slicePtr).Elem()
	sliceElementType := val.Type().Elem()
	var strIDer zstr.StrIDer
	st := reflect.TypeOf(&strIDer).Elem()
	var createIDer zstr.CreateStrIDer
	ct := reflect.TypeOf(&createIDer).Elem()
	// zlog.Info("sliceElementType:", sliceElementType.Implements(st), reflect.PointerTo(sliceElementType).Implements(ct))
	if sliceElementType.Implements(st) && (sliceElementType.Implements(ct) || reflect.PointerTo(sliceElementType).Implements(ct)) { // it has GetStrID and CreateStrID...
		gb.HandleAddItemFunc = func() {
			index := zslice.AddEmptyElementAtEnd(slicePtr)
			// fmt.Printf("AddItem: v:%p sliceptr:%p len:%d\n", gb, reflect.ValueOf(slicePtr).Interface(), reflect.ValueOf(slicePtr).Elem().Len())
			e := reflect.ValueOf(slicePtr).Elem().Index(index)
			if e.Kind() != reflect.Pointer {
				e = e.Addr()
			}
			ct := e.Interface().(zstr.CreateStrIDer)
			ct.CreateStrID()
			st := e.Interface().(zstr.StrIDer)
			id := st.GetStrID()
			// zlog.Info("AddID:", id)
			grouper.AddItem(id, "", "", true, nil, create)
		}
	} else {
		gb.HandleAddItemFunc = func() {
			index := zslice.AddEmptyElementAtEnd(slicePtr)
			id := strconv.Itoa(index)
			// zlog.Info("AddIndex:", id, index)
			grouper.AddItem(id, "", "", true, nil, create)
		}
	}
}

func AddSliceItems(g Grouper, slicePtr any, setID string, indicatorFieldName string, create func(id string, delete bool) zview.View) {
	val := reflect.ValueOf(slicePtr).Elem()
	var firstID string
	for i := 0; i < val.Len(); i++ {
		a := val.Index(i).Interface()
		fval, _, got := zreflect.FieldForName(a, true, indicatorFieldName)
		zlog.Assert(got, i, indicatorFieldName)
		title := fmt.Sprint(fval)
		id := zstr.GetIDFromAnySliceItemWithIndex(a, i)
		if firstID == "" {
			firstID = id
		}
		g.AddItem(id, title, "", id == setID, nil, create)
	}
	// zlog.Info("GroupBase.ReadyToShow key:", g.GetStoreKey(), setID, g.GetCurrentID())
	if g.GetCurrentID() == "" && firstID != "" {
		g.SelectItem(firstID, nil)
	}
}

func FindGroupBaseChild(parent zview.View) (gb *GroupBase) {
	// zlog.Info("FindGroupBaseChild:", parent.ObjectName(), ct != nil)
	zcontainer.ViewRangeChildren(parent, true, true, func(view zview.View) bool {
		// zlog.Info("FindGroupBaseChild:", view.ObjectName())
		g, _ := view.(Grouper)
		if g != nil {
			gb = g.GetGroupBase()
			return false
		}
		return true
	})
	return
}

// GetAncestorGroupBase goes up the parent stack looking for a *Grouper to get GroupBase from
func GetAncestorGroupBase(child zview.View) *GroupBase {
	all := child.Native().AllParents()
	all = append(all, child.Native())
	zslice.Reverse(all)
	for i, p := range all {
		g, _ := p.View.(Grouper)
		if g != nil {
			return g.GetGroupBase()
		}
		if i >= 2 {
			break
		}
	}
	return nil
}
