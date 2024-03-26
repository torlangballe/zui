//go:build zui

// zslicegrid is a package for making GridListViews of slices.
//
// It's base type is SliceGridView, which is also used by TableView and SQLTableView
package zslicegrid

import (
	"strings"
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
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zwords"
)

type ChildrenOwner interface {
	GetChildren() any
}

// SliceGridView creates a GridListView for a slice of any struct.
//
// The struct must confirm to zstr.StrIDer, returning a unique string id for each item.
// It has options to add a bar on top, and add edit and delete buttons.
//
// Editing will actually open multiple cells in a dialog with zfields.PresentOKCancelStructSlice then call StoreChangedItemsFunc.
// Deleting will remove cells from the slice, then call DeleteItemsFunc.
// It sets up its GridListViews HandleKeyFunc to edit and delete with return/backspace.
// If its struct type confirms to ChildrenOwner (has GetChildren()), it will show child-slices
// in a hierarchy, setting the GridListViews hierarchy HierarchyLevelFunc, and calculating cell-count,
// id-at-index etc based on open branch toggles.
type SliceGridView[S zstr.StrIDer] struct {
	zcontainer.StackView
	Grid                  *zgridlist.GridListView
	Bar                   *zcontainer.StackView
	EditParameters        zfields.FieldViewParameters
	StructName            string
	ForceUpdateSlice      bool // Set this to make UpdateSlice update, even if slice hash is the same, usefull if other factors cause it to display differently
	NameOfXItemsFunc      func(ids []string, singleSpecial bool) string
	DeleteAskSubTextFunc  func(ids []string) string
	UpdateViewFunc        func()
	SortFunc              func(s []S)                                     // SortFunc is called to sort the slice after any updates.
	FilterFunc            func(s S) bool                                  // FilterFunc is called to decide what cells are shown. Might typically use v.SearchField's text.
	StoreChangedItemsFunc func(items []S)                                 // StoreChangedItemsFunc is called with ids of all cells that have been edited. It must set the items in slicePtr, can use SetItemsInSlice. It ends by calling UpdateViewFunc(). Might call go-routine to push to backend.
	StoreChangedItemFunc  func(item S, last bool) error                   // StoreChangedItemFunc is called by the default StoreChangedItemsFunc with index of item in slicePtr, each in a goroutine which can clear showError to not show more than one error. The items are set in the slicePtr afterwards. last is true if it's the last one in items.
	DeleteItemsFunc       func(ids []string)                              // DeleteItemsFunc is called with ids of all selected cells to be deleted. It must remove them from slicePtr.
	CallDeleteItemFunc    func(id string, showErr *bool, last bool) error // CallDeleteItemFunc is called from default DeleteItemsFunc, with id of each item. They are not removed from slice.

	slicePtr      *[]S
	filteredSlice []S
	options       OptionType
	layedOut      bool

	SearchField *ztext.SearchField
	ActionMenu  *zmenu.MenuedOwner
	Layout      *zimageview.ValuesView
}

type LayoutType string

const (
	LayoutHorizontalFirstType = "hor"
	LayoutVerticalFirstType   = "vert"
	LayoutSingleRowsType      = "single-rows"
)

// OptionType is a set of options for altering a SliceGridView's appearance and behavior
type OptionType int

const (
	AddNone              OptionType = 0
	AddBar               OptionType = 1 << iota // Adds a bar stack above the grid.
	AddSearch                                   // Adds a search field that uses v.FilterFunc to decide if the search matches. Sets AddBar.
	AddMenu                                     // Adds a menu of actions in the bar, with some defaults. Sets AddBar. v.CreateDefaultMenuItems() creates default actions.
	AddChangeLayout                             // If set a button to order horizontal first or vertical first is shown
	AllowNew                                    // There is a Add New item menu
	AllowDuplicate                              // There is a Add Duplicate item menu
	AllowDelete                                 // It is deletable, and allows keyboard/menu delete
	AllowEdit                                   // Allows selected cell(s) to be edited with menu or return key.
	AddDocumentationIcon                        // Adds a icon to press to show doc/name.md file where name is in init
	LastBaseOption
	AllowAllEditing = AllowEdit | AllowNew | AllowDelete | AllowDuplicate
)

// NewView creates a new SliceGridView using v.Init()
func NewView[S zstr.StrIDer](slice *[]S, storeName string, options OptionType) (sv *SliceGridView[S]) {
	v := &SliceGridView[S]{}
	v.Init(v, slice, storeName, options)
	return v
}

// Init sets up an allocated SliceGridView with a slice, storeName for hierarchy state, and options for setup
func (v *SliceGridView[S]) Init(view zview.View, slice *[]S, storeName string, options OptionType) {
	zlog.Assert(slice != nil)
	v.StackView.Init(view, true, "slice-grid-view")
	v.SetObjectName(storeName)
	v.SetSpacing(0)
	v.StructName = "item"
	v.slicePtr = slice

	v.EditParameters = zfields.FieldViewParametersDefault()
	v.EditParameters.Field.Flags |= zfields.FlagIsLabelize
	v.EditParameters.EditWithoutCallbacks = true

	var s S
	var a any = s
	_, hasHierarchy := a.(ChildrenOwner)

	if options&AllowAllEditing != 0 {
		options |= AddMenu
	}
	if options&AddBarInHeader != 0 {
		options |= AddHeader
	}
	if options&(AddSearch|AddMenu|AddChangeLayout|AddDocumentationIcon|AddBarInHeader) != 0 {
		options |= AddBar
	}
	v.options = options
	if options&AddBar != 0 {
		v.Bar = zcontainer.StackViewHor("bar")
		v.Bar.SetSpacing(8)
		v.Bar.SetMargin(zgeo.RectFromXY2(6, 5, -6, -3))
		v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
	}
	if options&AddSearch != 0 {
		v.SearchField = ztext.SearchFieldNew(ztext.Style{}, 14)
		v.SearchField.TextView.SetChangedHandler(func() {
			v.updateView()
		})
		v.Bar.Add(v.SearchField, zgeo.CenterLeft)
	}

	horFirst := true
	if options&AddChangeLayout != 0 {
		var key string
		if storeName != "" {
			key = storeName + ".layout"
		}
		v.Layout = zimageview.NewValuesView(zgeo.Size{31, 19}, key)
		v.Layout.SetObjectName("layout")
		v.Layout.AddVariant(LayoutHorizontalFirstType, "images/zcore/order-hor-first.png")
		v.Layout.AddVariant(LayoutVerticalFirstType, "images/zcore/order-vert-first.png")
		v.Layout.AddVariant(LayoutSingleRowsType, "images/zcore/order-single-rows.png")
		v.Layout.ValueChangedHandlerFunc = func() {
			v.handleLayoutButton(v.Layout.Value())
		}
		v.Bar.Add(v.Layout, zgeo.CenterLeft)
	}
	if options&AddMenu != 0 {
		actions := zimageview.NewWithCachedPath("images/zcore/gear.png", zgeo.Size{18, 18})
		actions.SetObjectName("action-menu")
		actions.DownsampleImages = true
		v.ActionMenu = zmenu.NewMenuedOwner()
		v.ActionMenu.Build(actions, nil)
		v.Bar.Add(actions, zgeo.CenterRight, zgeo.Size{})
	}
	if options&AddDocumentationIcon != 0 {
		doc := zwidgets.DocumentationIconViewNew(storeName)
		doc.SetZIndex(200)
		v.Bar.Add(doc, zgeo.CenterRight, zgeo.Size{})
	}

	v.Grid = zgridlist.NewView(storeName + "-GridListView")
	v.Grid.HorizontalFirst = horFirst
	v.Grid.CellCountFunc = func() int {
		zlog.Assert(len(v.filteredSlice) <= len(*v.slicePtr), len(v.filteredSlice), len(*v.slicePtr))
		if !hasHierarchy {
			return len(v.filteredSlice)
		}
		return v.countHierarcy(v.filteredSlice)
	}
	v.Grid.IDAtIndexFunc = func(i int) string {
		if !hasHierarchy {
			if i >= len(v.filteredSlice) {
				return ""
			}
			return v.filteredSlice[i].GetStrID()
		}
		var count int
		return v.getIDForIndex(&v.filteredSlice, i, &count)
	}
	v.Grid.HandleKeyFunc = func(km zkeyboard.KeyMod, down bool) bool {
		var oneID string
		if len(v.Grid.SelectedIDs()) == 1 {
			oneID = v.Grid.SelectedIDs()[0]
		}
		if oneID == "" && len(v.Grid.SelectedIDs()) == 0 {
			oneID = v.Grid.CurrentHoverID
		}
		if oneID != "" {
			cell := v.Grid.CellView(oneID)
			if zcontainer.HandleOutsideShortcutRecursively(cell, km) {
				return true
			}
		}
		if v.ActionMenu != nil {
			return v.ActionMenu.HandleOutsideShortcut(km)
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
	v.Grid.SetMargin(zgeo.RectFromXY2(6, 0, -6, -0))
	v.Grid.MultiSelectable = true

	v.Add(v.Grid, zgeo.TopCenter|zgeo.Expand) //, zgeo.Size{4, 4}) //.Margin = zgeo.Size{4, 0}

	v.StoreChangedItemsFunc = func(items []S) {
		// zlog.Info("StoreChangedItemsFunc", len(items), v.StoreChangedItemFunc != nil)
		zlog.Assert(v.StoreChangedItemFunc != nil)
		showErr := true
		var storeItems []S
		var wg sync.WaitGroup
		for i, item := range items {
			// if true { //len(*v.slicePtr) <= i || zreflect.HashAnyToInt64(item, "") != zreflect.HashAnyToInt64((*v.slicePtr)[i], "") {
			wg.Add(1)
			go func(i int, item S) {
				err := v.StoreChangedItemFunc(item, i == len(items)-1)
				if err == nil {
					storeItems = append(storeItems, item)
				} else {
					if showErr {
						showErr = false
						zalert.ShowError(err)
					}
				}
				wg.Done()
			}(i, item)
		}
		wg.Wait()
		v.SetItemsInSlice(storeItems)
		v.UpdateViewFunc() // here we call UpdateViewFunc and not updateView, as just sorted in line above
		// zlog.Info("StoreChangedItemsFunc done")
	}

	v.DeleteItemsFunc = func(ids []string) {
		if v.CallDeleteItemFunc == nil {
			return
		}
		showErr := true
		var deleteIDs []string
		var wg sync.WaitGroup
		for i, id := range ids {
			wg.Add(1)
			go func(id string, showErr *bool, i int) {
				err := v.CallDeleteItemFunc(id, showErr, i == len(ids)-1)
				if err == nil {
					deleteIDs = append(deleteIDs, id)
				}
				wg.Done()
			}(id, &showErr, i)
		}
		wg.Wait()
		v.RemoveItemsFromSlice(deleteIDs)
		v.updateView()
	}
	return
}

func (v *SliceGridView[S]) handleLayoutButton(value string) {
	if v.Grid == nil {
		return
	}
	switch LayoutType(value) {
	case LayoutHorizontalFirstType:
		v.Grid.HorizontalFirst = true
		v.Grid.MaxColumns = 0
	case LayoutVerticalFirstType:
		v.Grid.HorizontalFirst = false
		v.Grid.MaxColumns = 0
	case LayoutSingleRowsType:
		v.Grid.HorizontalFirst = true
		v.Grid.MaxColumns = 1
	}
	v.Grid.RecreateCells = v.layedOut
	v.layedOut = true
	// zlog.Info("handleLayoutButton:", v.Grid.RecreateCells)
	v.ArrangeChildren()
}

func (v *SliceGridView[S]) updateView() {
	// v.doFilterAndSort(*v.slicePtr)
	v.UpdateViewFunc()
}

func setItemsInSlice[S zstr.StrIDer](items []S, slicePtr *[]S) int {
	var added int
	found := false
	for _, item := range items {
		for i, s := range *slicePtr {
			if s.GetStrID() == item.GetStrID() {
				found = true
				// fmt.Printf("edited: %+v %v %d\n", (*v.slicePtr)[i], item, i)
				(*slicePtr)[i] = item
				break
			}
		}
		if !found {
			*slicePtr = append(*slicePtr, item)
			added++
		}
	}
	return added
}

func (v *SliceGridView[S]) SetItemsInSlice(items []S) (added int) {
	added = setItemsInSlice(items, v.slicePtr)
	v.doFilterAndSort(*v.slicePtr)
	return added
}

func (v *SliceGridView[S]) RemoveItemsFromSlice(ids []string) {
	for i := 0; i < len(*v.slicePtr); i++ {
		id := (*v.slicePtr)[i].GetStrID()
		if zstr.StringsContain(ids, id) {
			zslice.RemoveAt(v.slicePtr, i)
			i--
		}
	}
	v.doFilterAndSort(*v.slicePtr)
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
			// zlog.Info("doFilter", v.Hierarchy(), len(slice), len(v.filteredSlice), v.FilterFunc(s))
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
	v.UpdateWidgets() // we do this here, so user can set up other widgets etc

	// v.Grid.UpdateCell = v.UpdateCell
}

func UpdateRows[S zstr.StrIDer](rows []S, onGrid any, orSlice *[]S) {
	sgv, _ := onGrid.(*SliceGridView[S])
	if sgv == nil {
		table, _ := onGrid.(*TableView[S])
		if table != nil {
			sgv = &table.SliceGridView
		}
	}
	if sgv != nil {
		orSlice = sgv.slicePtr
	}
	setItemsInSlice(rows, orSlice)
	if sgv != nil {
		sgv.updateView()
	}
}

func (v *SliceGridView[S]) UpdateSlice(s []S) {
	update := v.ForceUpdateSlice
	if !update {
		update = (len(s) != len(*v.slicePtr) || zreflect.HashAnyToInt64(s, "") != zreflect.HashAnyToInt64(*v.slicePtr, ""))
	}
	if update {
		v.ForceUpdateSlice = false
		*v.slicePtr = s
		//remove non-selected :
		v.updateView()
	}
}

func (v *SliceGridView[S]) UpdateSliceWithSelf() {
	v.UpdateSlice(*v.slicePtr)
}

func (v *SliceGridView[S]) getItemsFromIDs(ids []string) []S {
	var items []S
	for i := 0; i < len(*v.slicePtr); i++ {
		sid := (*v.slicePtr)[i].GetStrID()
		if zstr.StringsContain(ids, sid) {
			items = append(items, (*v.slicePtr)[i])
		}
	}
	return items
}

func (v *SliceGridView[S]) EditItemIDs(ids []string, after func(ok bool)) {
	items := v.getItemsFromIDs(ids)
	if len(items) == 0 {
		zlog.Fatal("SGV EditItemIDs: no items. ids:", ids, v.Hierarchy())
	}
	title := "Edit " + v.NameOfXItemsFunc(ids, true)
	// zlog.Info("EditItemIDs", zlog.Pointer(&items), zlog.Pointer(v.slicePtr))
	v.EditItems(items, title, false, false, after)
}

func (v *SliceGridView[S]) UpdateWidgets() {
	// ids := v.Grid.SelectedIDs()
}

func (v *SliceGridView[S]) EditItems(ns []S, title string, isEditOnNewStruct, selectAfterEditing bool, after func(ok bool)) {
	params := v.EditParameters
	params.Field.Flags |= zfields.FlagIsLabelize
	params.Styling.Spacing = 10
	params.IsEditOnNewStruct = isEditOnNewStruct
	zfields.PresentOKCancelStructSlice(&ns, params, title, zpresent.AttributesNew(), func(ok bool) bool {
		if !ok {
			if after != nil {
				after(false)
			}
			return true
		}
		if v.StoreChangedItemsFunc != nil { // if we do this before setting the slice below, StoreChangedItemsFunc func can compare with original items
			v.StoreChangedItemsFunc(ns)
		}
		if selectAfterEditing {
			var ids []string
			for _, n := range ns {
				ids = append(ids, n.GetStrID())
			}
			v.Grid.SelectCells(ids, true)
		}
		if after != nil {
			after(true)
		}
		return true
	})
}

func (v *SliceGridView[S]) addNewItem() {
	var ns S
	var a any
	title := "Add New " + v.StructName + ":"
	a = &ns
	zfields.CallStructInitializer(a)
	v.EditItems([]S{ns}, title, true, true, nil)
}

func (v *SliceGridView[S]) duplicateItems(nitems string, ids []string) {
	var newItems []S
	title := "Duplicate " + nitems + ":"
	for _, o := range v.getItemsFromIDs(ids) {
		var n = o
		zfields.CallStructInitializer(&n)
		newItems = append(newItems, n)
	}
	v.EditItems(newItems, title, false, true, nil)
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
		alert.SubText = sub
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
	v := TableViewNew[Row](&stuff, "hierarchy", AddHeader)
	stack.Add(v, zgeo.TopLeft|zgeo.Expand)
}
*/

func (v *SliceGridView[S]) CreateDefaultMenuItems(forSingleCell bool) []zmenu.MenuedOItem {
	var items []zmenu.MenuedOItem
	// zlog.Info("CreateDefaultMenuItems", forSingleCell, zlog.CallingStackString())
	ids := v.Grid.SelectedIDs()
	if v.options&AllowNew != 0 && !forSingleCell {
		del := zmenu.MenuedSCFuncAction("Add New "+v.StructName+"…", 'N', 0, v.addNewItem)
		items = append(items, del)
	}
	if v.Grid.CellCountFunc() > 0 {
		if v.Grid.MultiSelectable && !forSingleCell {
			all := zmenu.MenuedSCFuncAction("Select All", 'A', 0, func() {
				v.Grid.SelectAll(true)
			})
			items = append(items, all)
		}
		if len(ids) > 0 {
			nitems := v.NameOfXItemsFunc(ids, true)
			if v.options&AllowDuplicate != 0 {
				del := zmenu.MenuedSCFuncAction("Duplicate "+nitems+"…", 'D', 0, func() {
					v.duplicateItems(nitems, ids)
				})
				items = append(items, del)
			}
			if v.options&AllowDelete != 0 {
				del := zmenu.MenuedSCFuncAction("Delete "+nitems+"…", zkeyboard.KeyBackspace, 0, func() {
					v.handleDeleteKey(true)
				})
				items = append(items, del)
			}
			if v.options&AllowEdit != 0 {
				edit := zmenu.MenuedSCFuncAction("Edit "+nitems, 'E', 0, func() {
					v.EditItemIDs(ids, nil)
				})
				items = append(items, edit)
			}
		}
	}
	return items
}

func (o OptionType) String() string {
	str := ""
	if o&AddBar != 0 {
		str += "bar "
	}
	if o&AddSearch != 0 {
		str += "search "
	}
	if o&AddMenu != 0 {
		str += "menu "
	}
	if o&AddChangeLayout != 0 {
		str += "chlayout "
	}
	if o&AllowNew != 0 {
		str += "new "
	}
	if o&AllowDuplicate != 0 {
		str += "dup "
	}
	if o&AllowDelete != 0 {
		str += "del "
	}
	if o&AllowEdit != 0 {
		str += "edit "
	}
	if o&AddBarInHeader != 0 {
		str += "hbar "
	}
	if o&AddDocumentationIcon != 0 {
		str += "doc "
	}
	if o&AddHeader != 0 {
		str += "head "
	}
	return strings.TrimRight(str, " ")
}
