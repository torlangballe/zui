// SliceGridView creates a GridListView for a slice of any struct.
// The struct must confirm to zstr.StrIDer, returning a unique string id for each item.
// It has options to add a bar on top, and add edit and delete buttons.
//
// Editing will actually open multiple cells in a dialog with zfields.PresentOKCancelStructSlice then call StoreChangedItemsFunc.
// Deleting will remove cells from the slice, then call DeleteItemsFunc.
// It sets up its GridListViews HandleKeyFunc to edit and delete with return/backspace.
// If its struct type confirms to ChildrenOwner (has GetChildren()), it will show child-slices
// in a hierarchy, setting the GridListViews hierarchy HierarchyLevelFunc, and calculating cell-count,
// id-at-index etc based on open branch toggles.

//go:build zui

package zslicegrid

import (
	"strconv"
	"sync"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zwords"
)

type ChildrenOwner interface {
	GetChildren() any
}

type SliceGridView[S zstr.StrIDer] struct {
	zcontainer.StackView
	Grid                  *zgridlist.GridListView
	Bar                   *zcontainer.StackView
	EditParameters        zfields.FieldViewParameters
	StructName            string
	NameOfXItemsFunc      func(ids []string, singleSpecial bool) string
	DeleteAskSubTextFunc  func(ids []string) string
	UpdateViewFunc        func()
	SortFunc              func(s []S)
	FilterFunc            func(s S) bool
	StoreChangedItemsFunc func(items []S)                               // StoreChangedItemsFunc is called with ids of all cells that have been edited. It must set the items in slicePtr, can use SetItemsInSlice.
	StoreChangedItemFunc  func(item S, showErr *bool, last bool) error  // StoreChangedItemFunc is called by the default StoreChangedItemsFunc with index of item in slicePtr, each in a goroutine which can clear showError to not show more than one error. The items are set in the slicePtr afterwards. last is true if it's the last one in items.
	DeleteItemsFunc       func(ids []string)                            // DeleteItemsFunc is called with ids of all selected cells to be deleted. It must remove them from slicePtr.
	DeleteItemFunc        func(item *S, showErr *bool, last bool) error // DeleteItemFunc is called from default DeleteItemsFunc, with index and struct in slicePtr of each item to delete. The items are then removed from the slicePtr. last is true if it's the last one in items.

	slicePtr      *[]S
	filteredSlice []S
	editButton    *zimageview.ImageView
	deleteButton  *zimageview.ImageView
	addButton     *zimageview.ImageView
	SearchField   *ztext.SearchField
}

type OptionType int

const (
	AddNone        OptionType = 0
	AddBar         OptionType = 1
	AddEdit        OptionType = 2
	AddDelete      OptionType = 4
	AddDarkPlus    OptionType = 8
	AddLightPlus   OptionType = 16
	AddSearch      OptionType = 32
	AddEditDelete             = AddEdit | AddDelete
	AddButtonsMask OptionType = AddEdit | AddDelete | AddDarkPlus | AddLightPlus
)

func NewView[S zstr.StrIDer](slice *[]S, storeName string, option OptionType) (sv *SliceGridView[S]) {
	v := &SliceGridView[S]{}
	v.Init(v, slice, storeName, option)
	return v
}

func (v *SliceGridView[S]) Init(view zview.View, slice *[]S, storeName string, options OptionType) {
	v.StackView.Init(view, true, "slice-grid-view")
	v.SetObjectName(storeName)
	v.SetSpacing(0)
	v.StructName = "item"
	v.slicePtr = slice

	v.EditParameters = zfields.FieldViewParametersDefault()
	v.EditParameters.LabelizeWidth = 120
	v.EditParameters.HideStatic = true

	var s S
	var a any = s
	_, hasHierarchy := a.(ChildrenOwner)

	if options&(AddBar) != 0 {
		// zlog.Info("AddBar!!!")
		v.Bar = zcontainer.StackViewHor("bar")
		v.Bar.SetSpacing(8)
		v.Bar.SetMargin(zgeo.RectFromXY2(6, 5, -6, -3))
		v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
	}
	if options&AddEdit != 0 {
		v.editButton = zimageview.New(nil, "images/zcore/edit-dark-gray.png", zgeo.Size{18, 18})
		v.editButton.SetObjectName("edit")
		v.Bar.Add(v.editButton, zgeo.CenterLeft)
	}
	if options&AddDelete != 0 {
		v.deleteButton = zimageview.New(nil, "images/zcore/trash-dark-gray.png", zgeo.Size{18, 18})
		v.deleteButton.SetObjectName("delete")
		v.Bar.Add(v.deleteButton, zgeo.CenterLeft)
	}
	if options&AddSearch != 0 {
		v.SearchField = ztext.SearchFieldNew(ztext.Style{}, 14)
		v.SearchField.TextView.SetChangedHandler(func() {
			v.UpdateViewFunc()
		})
		v.Bar.Add(v.SearchField, zgeo.CenterLeft)
	}
	if options&(AddDarkPlus|AddLightPlus) != 0 {
		str := "white"
		if options&AddDarkPlus != 0 {
			str = "darkgray"
		}
		v.addButton = zimageview.New(nil, "images/plus-circled-"+str+".png", zgeo.Size{16, 16})
		v.Bar.Add(v.addButton, zgeo.CenterLeft)
		v.addButton.SetPressedHandler(v.handlePlusButtonPressed)
	}

	v.Grid = zgridlist.NewView(storeName + "-GridListView")
	v.Grid.CellCountFunc = func() int {
		zlog.Assert(len(v.filteredSlice) <= len(*v.slicePtr), len(v.filteredSlice), len(*v.slicePtr))
		if !hasHierarchy {
			return len(v.filteredSlice)
		}
		return v.countHierarcy(v.filteredSlice)
	}
	v.Grid.IDAtIndexFunc = func(i int) string {
		if !hasHierarchy {
			return v.filteredSlice[i].GetStrID()
		}
		var count int
		return v.getIDForIndex(&v.filteredSlice, i, &count)
	}
	v.Grid.HandleKeyFunc = func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
		if options&AddDelete != 0 && key == zkeyboard.KeyBackspace {
			v.handleDeleteKey(mod != zkeyboard.ModifierCommand)
			return true
		}
		if options&AddEdit != 0 && key == zkeyboard.KeyReturn {
			v.HandleEditButtonPressed()
			return true
		}
		return false
	}
	v.Grid.HandleSelectionChangedFunc = func() {
		v.UpdateWidgets()
	}
	v.NameOfXItemsFunc = func(ids []string, singleSpecial bool) string {
		return zwords.PluralWordWithCount(v.StructName, float64(len(ids)), "", "", 0)
	}
	v.UpdateViewFunc = func() {
		v.doFilterAndSort(*v.slicePtr)
		v.Grid.LayoutCells(true)
		a := v.View.(zcontainer.Arranger)
		a.ArrangeChildren()
		v.UpdateWidgets()
	}
	if hasHierarchy {
		v.Grid.HierarchyLevelFunc = v.calculateHierarchy
	}
	v.Grid.SetMargin(zgeo.RectFromXY2(6, 4, -6, -4))
	v.Grid.MultiSelectable = true

	v.Add(v.Grid, zgeo.TopLeft|zgeo.Expand, zgeo.Size{}) //.Margin = zgeo.Size{4, 0}

	v.StoreChangedItemsFunc = func(items []S) {
		zlog.Info("StoreChangedItemsFunc", len(items), v.StoreChangedItemFunc != nil)
		if v.StoreChangedItemFunc == nil {
			return
		}
		showErr := true
		var storeItems []S
		var wg sync.WaitGroup
		for i, item := range items {
			if len(*v.slicePtr) <= i || zstr.HashAnyToInt64(item, "") != zstr.HashAnyToInt64((*v.slicePtr)[i], "") {
				wg.Add(1)
				go func(i int, item S, showErr *bool) {
					err := v.StoreChangedItemFunc(item, showErr, i == len(items)-1)
					if err == nil {
						storeItems = append(storeItems, item)
					}
					wg.Done()
				}(i, item, &showErr)
			}
		}
		wg.Wait()
		v.SetItemsInSlice(storeItems)
		v.UpdateViewFunc()
	}
	v.DeleteItemsFunc = func(ids []string) {
		if v.DeleteItemFunc == nil {
			return
		}
		showErr := true
		var deleteIDs []string
		var wg sync.WaitGroup
		for i, id := range ids {
			wg.Add(1)
			go func(id string, showErr *bool, i int) {
				s := v.StructForID(id)
				err := v.DeleteItemFunc(s, showErr, i == len(ids)-1)
				if err == nil {
					deleteIDs = append(deleteIDs, id)
				}
				wg.Done()
			}(id, &showErr, i)
		}
		wg.Wait()
		v.RemoveItemsFromSlice(deleteIDs)
		v.UpdateViewFunc()
	}
	return
}

func (v *SliceGridView[S]) SetItemsInSlice(items []S) {
	for _, item := range items {
		for i, s := range *v.slicePtr {
			if s.GetStrID() == item.GetStrID() {
				// fmt.Printf("edited: %+v %v %d\n", (*v.slicePtr)[i], item, i)
				(*v.slicePtr)[i] = item
			}
		}
	}
	v.doFilterAndSort(*v.slicePtr)
}

func (v *SliceGridView[S]) RemoveItemsFromSlice(ids []string) {
	for _, id := range ids {
		i := v.Grid.IndexOfID(id)
		zslice.RemoveAt(v.slicePtr, i)
	}
	v.doFilterAndSort(*v.slicePtr)
}

func (v *SliceGridView[S]) handlePlusButtonPressed() {
	var s S
	*v.slicePtr = append(*v.slicePtr, s)
	index := len(*v.slicePtr) - 1
	var a any = &(*v.slicePtr)[index]
	id := strconv.Itoa(index)
	st, _ := a.(zstr.StrIDer)
	ct, _ := a.(zstr.CreateStrIDer)
	if st != nil && (ct != nil) { // it has GetStrID and CreateStrID...
		ct.CreateStrID()
		id = st.GetStrID()
		// zlog.Info("CREATE:", id, s)
	}
	v.UpdateViewFunc()
	v.Grid.SelectCell(id, false)
	v.HandleEditButtonPressed()
}

func (v *SliceGridView[S]) getIDForIndex(slice *[]S, index int, count *int) string {
	for i, s := range *slice {
		// if index != -1 {
		// 	zlog.Info("getIDForIndex", index, i, s)
		// }
		id := s.GetStrID()
		if index == *count {
			return id
		}
		(*count)++
		if v.Grid.OpenBranches[id] {
			children := v.getChildren(slice, i)
			// if co .GetChildren()
			// if a != nil {
			// 	children := a.(*[]S)
			id := v.getIDForIndex(children, index, count)
			if id != "" {
				return id
			}
		}
	}
	return ""
}

func (v *SliceGridView[S]) getChildren(slice *[]S, i int) *[]S {
	var a any = (*slice)[i]
	return a.(ChildrenOwner).GetChildren().(*[]S)
}

func (v *SliceGridView[S]) countHierarcy(slice []S) int {
	var count int
	v.getIDForIndex(&slice, -1, &count)
	// zlog.Info("Count:", v.ObjectName(), count, v.openBranches)
	return count
}

func (v *SliceGridView[S]) structForID(slice *[]S, id string) *S {
	for i, s := range *slice {
		sid := s.GetStrID()
		if id == sid {
			return &(*slice)[i]
		}
		if v.Grid.HierarchyLevelFunc != nil {
			if !v.Grid.OpenBranches[sid] {
				continue
			}
			children := v.getChildren(slice, i)
			cs := v.structForID(children, id)
			if cs != nil {
				return cs
			}
		}
	}
	return nil
}

func (v *SliceGridView[S]) StructForID(id string) *S {
	return v.structForID(v.slicePtr, id)
}

func (v *SliceGridView[S]) doFilter(slice []S) {
	if v.FilterFunc != nil {
		sids := v.Grid.SelectedIDs()
		length := len(sids)
		var f []S
		for _, s := range slice {
			if v.FilterFunc(s) {
				f = append(f, s)
			} else {
				sid := s.GetStrID()
				sids = zstr.RemovedFromSlice(sids, sid)
			}
		}
		v.filteredSlice = f
		if len(sids) != length {
			v.Grid.SelectCells(sids, false)
		}
	} else {
		v.filteredSlice = slice
	}
}

func (v *SliceGridView[S]) doFilterAndSort(slice []S) {
	v.doFilter(slice)
	if v.SortFunc != nil {
		v.SortFunc(v.filteredSlice) // Do this beforeWindow shown, as the sorted cells get placed correctly then
		// v.UpdateViewFunc()
	}
}

func (v *SliceGridView[S]) ReadyToShow(beforeWindow bool) {
	if beforeWindow {
		v.doFilterAndSort(*v.slicePtr)
		return
	}
	if v.editButton != nil {
		v.editButton.SetPressedHandler(v.HandleEditButtonPressed)
	}
	if v.deleteButton != nil {
		v.deleteButton.SetPressedHandler(func() {
			v.DeleteItemsAsk(v.Grid.SelectedIDs())
		})
	}
	v.UpdateWidgets() // we do this here, so user can set up other widgets etc

	// v.Grid.UpdateCell = v.UpdateCell
}

func (v *SliceGridView[S]) UpdateSlice(s []S) {
	update := (len(s) != len(*v.slicePtr) || zstr.HashAnyToInt64(s, "") != zstr.HashAnyToInt64(*v.slicePtr, ""))
	if update {
		// v.doFilterAndSort(s)
		*v.slicePtr = s
		//remove non-selected :
		v.UpdateViewFunc()
		var selected []string
		oldSelected := v.Grid.SelectedIDs()
		var hoverOK bool
		for _, sl := range s {
			sid := sl.GetStrID()
			if v.Grid.CurrentHoverID == sid {
				hoverOK = true
			}
			if zstr.StringsContain(oldSelected, sid) {
				selected = append(selected, sid)
			}
		}
		if !hoverOK {
			v.Grid.SetHoverID("")
		}
		v.Grid.SelectCells(selected, false)
	}
}

func (v *SliceGridView[S]) UpdateSliceWithSelf() {
	v.UpdateSlice(*v.slicePtr)
}

func (v *SliceGridView[S]) HandleEditButtonPressed() {
	ids := v.Grid.SelectedIDs()
	if len(ids) == 0 {
		if v.Grid.CurrentHoverID == "" {
			return
		}
		ids = []string{v.Grid.CurrentHoverID}
	}
	v.EditItems(ids)
}

func (v *SliceGridView[S]) EditItems(ids []string) {
	title := "Edit "
	var items []S

	// zlog.Info("Edite items:", ids)

	for i := 0; i < len(*v.slicePtr); i++ {
		sid := (*v.slicePtr)[i].GetStrID()
		if zstr.StringsContain(ids, sid) {
			// zlog.Info("EDIT:", sid, (*v.slicePtr)[i])
			items = append(items, (*v.slicePtr)[i])
		}
	}
	if len(items) == 0 {
		zlog.Fatal(nil, "SGV EditItems: no items. ids:", ids, v.Hierarchy())
	}
	title += v.NameOfXItemsFunc(ids, true)
	zfields.PresentOKCancelStructSlice(&items, v.EditParameters, title, zpresent.AttributesNew(), func(ok bool) bool {
		// zlog.Info("Edited items:", ok, v.StoreChangedItemsFunc != nil)
		if !ok {
			return true
		}
		if v.StoreChangedItemsFunc != nil { // if we do this before setting the slice below, StoreChangedItemsFunc func can compare with original items
			// ztimer.StartIn(0.1, func() {
			v.StoreChangedItemsFunc(items)
			// })
		}
		v.UpdateViewFunc()
		return true
	})
}

func (v *SliceGridView[S]) UpdateWidgets() {
	ids := v.Grid.SelectedIDs()
	if v.Bar != nil {
		for _, c := range v.Bar.GetChildren(false) {
			if c == v.SearchField {
				continue
			}
			_, stack := c.(*zcontainer.StackView)
			if !stack {
				c.SetUsable(len(ids) > 0)
			}
		}
	}
	// if v.editButton != nil {
	// 	v.setButtonWithCount("edit", ids, v.editButton)
	// }
	// if v.deleteButton != nil {
	// 	v.setButtonWithCount("delete", ids, v.deleteButton)
	// }
}

func (v *SliceGridView[S]) handleDeleteKey(ask bool) {
	if len(v.Grid.SelectedIDs()) == 0 {
		return
	}
	ids := v.Grid.SelectedIDs()
	if ask {
		v.DeleteItemsAsk(ids)
	} else {
		v.DeleteItemsFunc(ids)
	}
}

func (v *SliceGridView[S]) DeleteItemsAsk(ids []string) {
	title := "Are you sure you want to delete "
	title += v.NameOfXItemsFunc(ids, true)
	alert := zalert.NewWithCancel(title + "?")
	if v.DeleteAskSubTextFunc != nil {
		sub := v.DeleteAskSubTextFunc(ids)
		alert.SetSub(sub)
	}
	alert.ShowOK(func() {
		go v.DeleteItemsFunc(ids)
	})
}

func (v *SliceGridView[S]) getHierarchy(slice *[]S, level int, id string) (hlevel int, leaf, got bool) {
	for i, s := range *slice {
		children := v.getChildren(&v.filteredSlice, i)
		sid := s.GetStrID()
		kids := (v.Grid.OpenBranches[sid] && len(*children) > 0)
		zlog.Info("getHier:", id, sid, level, kids, len(*children))
		if id == sid {
			leaf = !kids
			if leaf && *children != nil {
				leaf = false
			}
			return level, leaf, true
		}
		if kids {
			hlevel, leaf, got = v.getHierarchy(children, level+1, id)
			if got {
				return
			}
		}
	}
	return level, false, false
}

func (v *SliceGridView[S]) calculateHierarchy(id string) (level int, leaf bool) {
	level, leaf, _ = v.getHierarchy(&v.filteredSlice, 1, id)
	return
}

/* Slice hierarch example:
type Row struct {
	Name        string `zui:"static,width:120,indicator"`
	IsFruit     bool   `zui:"width:50"`
	IsVegetable bool
	Children    []Row `zui:"-"`
}

func (r Row) GetChildren() any {
	return &r.Children
}

func (r Row) GetStrID() string {
	return r.Name
}

var stuff = []Row{
	Row{"Tomato", true, false, nil},
	Row{"Carrot", false, true, nil},
	Row{"Steak", false, false, nil},
	Row{"Fruit", true, false, []Row{
		Row{"Banana", true, false, nil},
		Row{"Apple", true, false, nil},
	},
	},
	Row{"Animal", true, false, []Row{
		Row{"Dog", true, false, []Row{
			Row{"Poodle", true, false, nil},
			Row{"Setter", true, false, nil},
		},
		},
		Row{"Cat", true, false, nil},
	},
	},
}

func addHierarchy(stack *zcontainer.StackView) {
	v := zslicegrid.TableViewNew[Row](&stuff, "hierarchy", zslicegrid.AddHeader)
	stack.Add(v, zgeo.TopLeft|zgeo.Expand)
}
*/

func (v *SliceGridView[S]) CreateDefaultMenuItems() []zmenu.MenuedOItem {
	var items []zmenu.MenuedOItem
	if v.Grid.CellCountFunc() > 0 {
		if v.Grid.MultiSelectable {
			all := zmenu.MenuedShortcutFuncAction("Select All", zkeyboard.SCut('A', 0), func() {
				v.Grid.SelectAll(true)
			})
			items = append(items, all)
		}
		ids := v.Grid.SelectedIDs()
		if len(ids) > 0 {
			nitems := v.NameOfXItemsFunc(ids, true)
			del := zmenu.MenuedFuncAction("Delete "+nitems+"…", func() {
				v.handleDeleteKey(true)
			})
			items = append(items, del)
			edit := zmenu.MenuedFuncAction("Edit "+nitems, func() {
				v.EditItems(ids)
			})
			items = append(items, edit)
		}
	}
	return items
}
