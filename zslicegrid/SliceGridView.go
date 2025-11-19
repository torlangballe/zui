//go:build zui

// zslicegrid is a package for making GridListViews of slices.
//
// It's base type is SliceGridView, which is also used by TableView and SQLTableView
package zslicegrid

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zclipboard"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zdocs"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zshortcuts"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zdebug"
	"github.com/torlangballe/zutil/zfloat"
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
// The struct must confirm to zstr.StrIDer, or contain a Field called ID that is a string or in64 id for each row.
// It has options to add a bar on top, and add edit and delete buttons.
//
// Editing will actually open multiple cells in a dialog with zfields.PresentEditOrViewStructSlice then call StoreChangedItemsFunc.
// Deleting will remove cells from the slice, then call DeleteItemsFunc.
// It sets up its GridListViews HandleKeyFunc to edit and delete with return/backspace.
// If its struct type confirms to ChildrenOwner (has GetChildren()), it will show child-slices
// in a hierarchy, setting the GridListViews hierarchy HierarchyLevelFunc, and calculating cell-count,
// id-at-index etc based on open branch toggles.
type SliceGridView[S any] struct {
	zcontainer.StackView
	Grid                            *zgridlist.GridListView
	Bar                             *zcontainer.StackView
	EditParameters                  zfields.FieldViewParameters
	StructName                      string
	ForceUpdateSlice                bool // Set this to make UpdateSlice update, even if slice hash is the same, useful if other factors cause it to display differently
	NameOfXItemsFunc                func(ids []string, singleSpecial bool) string
	DeleteAskSubTextFunc            func(ids []string) string
	UpdateViewFunc                  func(arrange, restoreSelectionScroll bool)             // Filter, sorts, arranges (updates and lays out) and updates widgets. Override to do more. arrange=false if it or parent's ArrangeChildren is going to be called anyway.
	SortFunc                        func(s []S)                                            // SortFunc is called to sort the slice after any updates.
	FilterFunc                      func(s S) bool                                         // FilterFunc is called to decide what cells are shown. Might typically use v.SearchField's text.
	StoreChangedItemsFunc           func(items []S)                                        // StoreChangedItemsFunc is called with ids of all cells that have been edited. It must set the items in slicePtr, can use SetItemsInSlice. It ends by calling UpdateViewFunc(). Might call go-routine to push to backend.
	StoreChangedItemFunc            func(item S, last bool) error                          // StoreChangedItemFunc is called by the default StoreChangedItemsFunc with index of item in slicePtr, each in a goroutine which can clear showError to not show more than one error. The items are set in the slicePtr afterwards. last is true if it's the last one in items.
	DeleteItemsFunc                 func(ids []string)                                     // DeleteItemsFunc is called with ids of all selected cells to be deleted. It must remove them from slicePtr.
	ValidateClipboardPasteItemsFunc func(items *[]S, proceed func())                       // Called on incoming items paste items to zero ID's or something, and validate. Call proceed if any left or user accepts or something
	HandleShortCutInRowFunc         func(rowID string, sc zkeyboard.KeyMod) bool           // Called if key pressed when row selected, and row-cell  or action menu doesn't handle it
	CallDeleteItemFunc              func(id string, showErr *bool, last bool) error        // CallDeleteItemFunc is called from default DeleteItemsFunc, with id of each item. They are not removed from slice.
	CreateActionMenuItemsFunc       func(sids []string, isGlobal bool) []zmenu.MenuedOItem // Used to set ActionMenu and FieldViewParameters.CreateActionMenuItemsFunc
	HandleRowDragOrderFunc          func()
	HandleRowsChangeFunc            func() // Called if rows deleted, added, updated
	CurrentLowerCaseSearchText      string
	EditDialogDocumentationPath     string
	FilterSkipCache                 map[string]bool
	Options                         OptionType

	slicePtr      *[]S
	filteredSlice []S
	laidOut       bool
	SearchField   *ztext.SearchField
	ActionMenu    *zmenu.MenuedOwner
	Layout        *zimageview.ValuesView
}

type LayoutType string

const (
	LayoutSingleRowsType      = "single-rows"
	LayoutHorizontalFirstType = "hor"
	LayoutVerticalFirstType   = "vert"
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
	AllowEdit                                   // Allows selected cell(s) to be edited with menu or E key.
	AllowView                                   // Allows selected cell(s) to be viewed (like edit) with menu.
	AllowCopyPaste                              // Allows copy of rows to paste into same table on another server.
	AddDocumentationIcon                        // Adds a icon to press to show doc/name.md file where name is in init
	AddHeader                                   // Adds a Header to the top of table. Currenty used by TableView
	AddBarInHeader                              // Sets the bar inside the Header, in right-most column. Sets  AddHeader and AddBar
	AddShortHelperArea                          // Adds an view where what you could have pressed to invoke shortcuts is shown
	RowsGUISearchable                           // Allows rows to be part of gui search
	AddNameAsSearchItem                         // If true, this table adds ObjectName() to currentPath for searching
	LastBaseOption
	AllowAllEditing = AllowEdit | AllowNew | AllowDelete | AllowDuplicate
)

// NewView creates a new SliceGridView using v.Init()
func NewView[S any](slice *[]S, storeName string, options OptionType) (sv *SliceGridView[S]) {
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
	v.FilterSkipCache = map[string]bool{}
	v.NoCalculatedMaxSize.W = true

	v.EditParameters = zfields.DefaultFieldViewParameters
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
	v.Options = options
	if options&AddBar != 0 {
		v.Bar = zcontainer.StackViewHor("bar")
		// v.Bar.SetBGColor(zstyle.Gray(0.9, 0.2))
		if options&AddBarInHeader == 0 {
			v.Bar.SetBGColor(zstyle.DefaultBGColor())
		}
		v.Bar.NoCalculatedMaxSize.W = true
		v.Bar.SetMargin(zgeo.RectFromXY2(6, 3, 0, -3))
		if options&AddBarInHeader == 0 {
			v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
		}
	}
	if options&AddSearch != 0 {
		v.SearchField = ztext.SearchFieldNew(ztext.Style{}, 12)
		v.SearchField.TextView.SetKeepFocusOnOutsideClick()
		v.SearchField.SetValueHandler("zslicegrid.Search", func(edited bool) {
			v.CurrentLowerCaseSearchText = strings.ToLower(v.SearchField.Text())
			v.ClearFilterSkipCache()
			v.UpdateViewFunc(true, false)
		})
		v.Bar.Add(v.SearchField, zgeo.CenterRight)
	}
	horFirst := true
	if options&AddChangeLayout != 0 {
		var key string
		if storeName != "" {
			key = storeName + ".layout"
		}
		v.Layout = zimageview.NewValuesView(zgeo.SizeD(31, 19), key)
		v.Layout.MixColorForDarkMode = zgeo.ColorNewGray(0.5, 1)
		v.Layout.SetObjectName("layout")
		v.Layout.AddVariant(LayoutSingleRowsType, "images/zcore/order-single-rows.png")
		v.Layout.AddVariant(LayoutHorizontalFirstType, "images/zcore/order-hor-first.png")
		v.Layout.AddVariant(LayoutVerticalFirstType, "images/zcore/order-vert-first.png")
		v.Layout.ValueChangedHandlerFunc = func() {
			v.handleLayoutButton(v.Layout.Value())
		}
		v.Bar.Add(v.Layout, zgeo.CenterLeft)
	}
	if options&AddMenu != 0 {
		actions := zimageview.NewWithCachedPath("images/zcore/gear.png", zgeo.SizeD(18, 18))
		if options&AddBarInHeader == 0 {
			actions.MixColorForDarkMode = zgeo.ColorNewGray(0.5, 1)
		}
		actions.SetObjectName("action-menu")
		actions.DownsampleImages = true
		v.ActionMenu = zmenu.NewMenuedOwner()
		v.ActionMenu.Build(actions, nil)
		v.ActionMenu.Name = v.ObjectName()
		v.Bar.Add(actions, zgeo.CenterRight, zgeo.SizeNull)
	}
	if options&AddDocumentationIcon != 0 {
		doc := zwidgets.DocumentationIconViewNew(storeName)
		doc.SetZIndex(200)
		v.Bar.Add(doc, zgeo.CenterRight, zgeo.SizeD(zstyle.DefaultRowRightMargin, 0))
	}
	if options&AddShortHelperArea != 0 {
		scStack := zcontainer.StackViewHor("shortcut-stack")
		zshortcuts.RegisterShortCutHelperAreaForWindow(zwindow.GetMain(), scStack)
		v.Bar.Add(scStack, zgeo.CenterRight, zgeo.SizeD(zstyle.DefaultRowRightMargin, 0))
	}

	v.Grid = zgridlist.NewView(storeName + "-GridListView")
	v.Grid.HorizontalFirst = horFirst
	v.Grid.SetSearchable(options&RowsGUISearchable == 0)
	v.Grid.CellCountFunc = func() int {
		zlog.Assert(len(v.filteredSlice) <= len(*v.slicePtr), len(v.filteredSlice), len(*v.slicePtr))
		if !hasHierarchy {
			return len(v.filteredSlice)
		}
		return v.countHierarcy(v.filteredSlice)
	}
	if v.Layout != nil {
		v.Layout.Update()
	}
	v.Grid.IDAtIndexFunc = func(i int) string {
		if !hasHierarchy {
			if i >= len(v.filteredSlice) {
				return ""
			}
			return GetIDForItem(&v.filteredSlice[i])
		}
		var count int
		return v.getIDForIndex(&v.filteredSlice, i, &count)
	}
	old := v.Grid.HandleKeyFunc
	v.Grid.HandleKeyFunc = func(km zkeyboard.KeyMod, down bool) bool {
		if !down {
			if old != nil {
				return old(km, down)
			}
			return false
		}
		// isInFocus := v.IsInAFocusedView()
		var oneID string
		slen := len(v.Grid.SelectedIDs())
		if slen == 1 {
			oneID = v.Grid.SelectedID()
		} else if slen == 0 {
			oneID = v.Grid.CurrentHoverID
		}
		// if oneID != "" {
		// 	cell := v.Grid.CellView(oneID)
		// 	if zshortcuts.HandleOutsideShortcutRecursively(cell, km, zbool.FromBool(isInFocus)) {
		// 		return true
		// 	}
		// }
		if v.ActionMenu != nil {
			// focused := isInFocus
			// if !focused {
			// 	focused = v.Grid.IsFocused()
			// }
			if v.ActionMenu.HandleShortcut(km, v.IsInAFocusedView()) {
				return true
			}
		}
		if oneID != "" && v.HandleShortCutInRowFunc != nil {
			if v.HandleShortCutInRowFunc(oneID, km) {
				return true
			}
		}
		if old != nil {
			return old(km, down)
		}
		return false
	}
	v.Grid.HandleSelectionChangedFunc = func() {
		v.UpdateWidgets()
	}
	v.NameOfXItemsFunc = func(ids []string, singleSpecial bool) string {
		ilen := len(ids)
		if singleSpecial && ilen == 1 {
			s := v.StructForID(ids[0])
			if s == nil {
				zlog.Error("NameOfXItemsFunc: s is nil:", v.Hierarchy(), ids[0], zdebug.CallingStackString())
				return v.StructName
			}
			//s is zero sometimes!!!!
			var a any = s
			ng, got := a.(zstr.NameGetter)
			if got && ng != nil {
				return `"` + ng.GetName() + `"`
			}
		}
		if ilen > 1 && ilen == v.Grid.CellCountFunc() {
			ilen := len(ids)
			word := zwords.PluralizeWord(v.StructName, float64(ilen), "", "")
			return fmt.Sprintf("all %d %s", ilen, word)
		}
		return zwords.PluralWordWithCount(v.StructName, float64(len(ids)), "", "", 0)
	}
	v.UpdateViewFunc = func(arrange, restoreSelectionScroll bool) {
		// prof := zlog.NewProfile(0.05, "UpdateViewFunc", v.ObjectName())
		newSIDs := v.doFilterAndSort(*v.slicePtr)
		a := v.View.(zcontainer.Arranger)
		// v.Grid.UpdateOnceOnSetRect = updateAllRows
		// prof.Log("After Filter")
		if arrange {
			a.ArrangeChildren() // We might be a table or other derivative, so need to do it with an arranger
		}
		// prof.Log("After Arrange")
		v.UpdateWidgets()
		if len(newSIDs) != 0 {
			v.Grid.SelectCells(newSIDs, restoreSelectionScroll, false)
		}
		if v.HandleRowsChangeFunc != nil {
			v.HandleRowsChangeFunc()
		}
		// prof.End("After UpdateWidgets")
	}
	if hasHierarchy {
		v.Grid.HierarchyLevelFunc = v.calculateHierarchy
	}
	v.Grid.SetMargin(zgeo.RectFromXY2(6, 0, -6, -0))
	v.Grid.MultiSelectable = true

	v.Grid.HandleGripDragRowFunc = func(id, ontoID string, dir zgeo.Alignment) {
		si := v.Grid.IndexOfID(id)
		ei := v.Grid.IndexOfID(ontoID)
		if dir == zgeo.Bottom {
			ei++
		}
		if ei == si+1 {
			return
		}
		// zlog.Info("DropMove:", si, ei)
		zslice.MoveElement(*v.slicePtr, si, ei)
		v.UpdateViewFunc(true, true)
		if v.HandleRowDragOrderFunc != nil {
			v.HandleRowDragOrderFunc()
		}
		if v.HandleRowsChangeFunc != nil {
			v.HandleRowsChangeFunc()
		}
	}
	v.StoreChangedItemsFunc = func(items []S) {
		// zlog.Info("StoreChangedItemsFunc", len(items), v.ObjectName(), zdebug.CallingStackString())
		if v.StoreChangedItemFunc == nil {
			return
		}
		showErr := true
		var storeItems []S
		var wg sync.WaitGroup
		for i, item := range items {
			// if true { //len(*v.slicePtr) <= i || zreflect.HashAnyToInt64(item, "") != zreflect.HashAnyToInt64((*v.slicePtr)[i], "") {
			wg.Add(1)
			v.Grid.SetDirtyRow(GetIDForItem(&item))
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
		v.UpdateViewFunc(true, false)
		// zlog.Info("StoreChangedItemsFunc done")
	}

	v.DeleteItemsFunc = func(ids []string) {
		var deleteIDs []string
		if v.CallDeleteItemFunc == nil {
			deleteIDs = ids
		} else {
			var wg sync.WaitGroup
			showErr := true
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
		}
		v.RemoveItemsFromSlice(deleteIDs)
		v.UpdateViewFunc(true, false)
	}
	v.Add(v.Grid, zgeo.TopLeft|zgeo.Expand)

	return
}

func GetIDForItem[S any](item *S) string {
	var a any
	a = item
	g, _ := a.(zstr.StrIDer)
	if g != nil {
		return g.GetStrID()
	}
	zlog.Fatal("Not a StrIDer!!!", reflect.TypeOf(a))
	var sid string
	zreflect.ForEachField(item, zreflect.FlattenIfAnonymous, func(each zreflect.FieldInfo) bool {
		if each.StructField.Name == "ID" {
			if each.StructField.Type.Kind() == reflect.String {
				sid = reflect.ValueOf(item).Elem().Field(each.FieldIndex).String()
				return false
			}
			sid = fmt.Sprint(reflect.ValueOf(item).Elem().Field(each.FieldIndex).Interface())
			return false
		}
		return true
	})
	zlog.Assert(sid != "", *item)
	return sid
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
	v.Grid.RecreateCells = v.laidOut
	v.laidOut = true
	// zlog.Info("handleLayoutButton:", v.Grid.MaxColumns, value, v.Grid.RecreateCells, v.IsPresented())
	if v.IsPresented() {
		a := v.View.(zcontainer.Arranger)
		a.ArrangeChildren()
	}
}

func (v *SliceGridView[S]) ClearFilterSkipCache() {
	v.FilterSkipCache = map[string]bool{}
}

func (v *SliceGridView[S]) insertItemsIntoASlice(items []S, slicePtr *[]S) int {
	var added int
	found := false
	if v != nil {
		v.Grid.DirtyIDs = map[string]bool{}
	}
	for _, item := range items {
		isid := GetIDForItem(&item)
		if v != nil {
			delete(v.FilterSkipCache, isid)
			v.Grid.DirtyIDs[isid] = true
		}
		for i, s := range *slicePtr {
			if GetIDForItem(&s) == isid {
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

// SetItemsInSlice adds or changes existing items with items.
func (v *SliceGridView[S]) SetItemsInSlice(items []S) (added int) {
	for _, s := range items {
		v.Grid.SetDirtyRow(GetIDForItem(&s))
	}
	added = v.insertItemsIntoASlice(items, v.slicePtr)
	v.doFilterAndSort(*v.slicePtr)
	return added
}

func (v *SliceGridView[S]) RemoveItemsFromSlice(ids []string) {
	for i := 0; i < len(*v.slicePtr); i++ {
		id := GetIDForItem(&(*v.slicePtr)[i])
		delete(v.FilterSkipCache, id)
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
		id := GetIDForItem(&s)
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

func (v *SliceGridView[S]) structForID(fromSlice *[]S, id string) *S {
	for i, s := range *fromSlice {
		sid := GetIDForItem(&s)
		if id == sid {
			return &(*fromSlice)[i]
		}
		if v.Grid.HierarchyLevelFunc != nil {
			if !v.Grid.OpenBranches[sid] {
				continue
			}
			children := v.getChildren(fromSlice, i)
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

func (v *SliceGridView[S]) doFilter(slice []S) (selectedAfter []string) {
	// start := time.Now()
	sids := v.Grid.SelectedIDs()
	// length := len(sids)
	if v.FilterFunc != nil {
		var f []S
		var skipCount, keepCount int
		for _, s := range slice {
			// zlog.Info("doFilter", v.Hierarchy(), len(slice), len(v.filteredSlice), v.FilterFunc(s))
			sid := GetIDForItem(&s)
			skip, got := v.FilterSkipCache[sid]
			if !got {
				skip = !v.FilterFunc(s)
				v.FilterSkipCache[sid] = skip
			}
			if skip {
				sids = zstr.RemovedFromSet(sids, sid)
				skipCount++
			} else {
				keepCount++
				f = append(f, s)
			}
		}
		v.filteredSlice = f
		// if len(sids) != length {
		// 	v.Grid.SelectCells(sids, false)
		// }
		// zlog.Info("doFilter filterd:", v.ObjectName(), time.Since(start), skipCount, keepCount)
	} else {
		v.filteredSlice = slice
	}
	return sids
}

func (v *SliceGridView[S]) doFilterAndSort(slice []S) (newSelected []string) {
	sids := v.doFilter(slice)
	if v.SortFunc != nil {
		v.SortFunc(v.filteredSlice) // Do this beforeWindow shown, as the sorted cells get placed correctly then
	}
	return sids
}

func (v *SliceGridView[S]) ReadyToShow(beforeWindow bool) {
	v.StackView.ReadyToShow(beforeWindow)
	if beforeWindow {
		v.doFilterAndSort(*v.slicePtr)
		return
	}
	v.UpdateWidgets() // we do this here, so user can set up other widgets etc
	if v.Options&(AllowEdit|AllowView) != 0 {
		v.Grid.HandleRowDoubleTappedFunc = func(id string) {
			v.editOrViewItemIDs([]string{id}, false, v.Options&AllowView != 0, nil)
		}
	}
	if v.CreateActionMenuItemsFunc != nil {
		if v.ActionMenu != nil {
			v.ActionMenu.CreateItemsFunc = func() []zmenu.MenuedOItem {
				selected := v.Grid.SelectedIDs()
				return v.CreateActionMenuItemsFunc(selected, true)
			}
		}
	}
}

func UpdateRows[S any](rows []S, onGrid any, orSlice *[]S) {
	sgv, _ := onGrid.(*SliceGridView[S])
	// zlog.Info("UpdateRows:", len(rows), onGrid != nil, len(*orSlice))
	if sgv == nil {
		table, _ := onGrid.(*TableView[S])
		if table != nil {
			sgv = &table.SliceGridView
		}
	}
	if sgv != nil {
		orSlice = sgv.slicePtr
	}
	sgv.insertItemsIntoASlice(rows, orSlice)
	if sgv != nil && len(sgv.Grid.DirtyIDs) != 0 {
		sgv.UpdateViewFunc(true, false)
	}
}

func (v *SliceGridView[S]) InsertRows(rows []S, selectRows bool) {
	var ids []string
	v.insertItemsIntoASlice(rows, v.slicePtr)
	v.UpdateViewFunc(true, !selectRows)
	if selectRows {
		for _, r := range rows {
			id := GetIDForItem(&r)
			ids = append(ids, id)
		}
		v.Grid.SelectCells(ids, true, true)
	}
}

func (v *SliceGridView[S]) UpdateSlice(s []S, arrange bool) {
	update := v.ForceUpdateSlice
	if !update {
		update = (len(s) != len(*v.slicePtr) || zreflect.HashAnyToInt64(s, "") != zreflect.HashAnyToInt64(*v.slicePtr, ""))
	}
	for _, si := range s {
		sid := GetIDForItem(&si)
		v.Grid.SetDirtyRow(sid)
		delete(v.FilterSkipCache, sid)
	}
	if update {
		v.ForceUpdateSlice = false
		*v.slicePtr = s
		v.UpdateViewFunc(arrange, false)
	}
}

// func (v *SliceGridView[S]) UpdateSliceWithSelf() {
// 	v.UpdateSlice(*v.slicePtr)
// }

func (v *SliceGridView[S]) getItemsFromIDs(ids []string) []S {
	var items []S
	for i := 0; i < len(*v.slicePtr); i++ {
		sid := GetIDForItem(&(*v.slicePtr)[i])
		if zstr.StringsContain(ids, sid) {
			items = append(items, (*v.slicePtr)[i])
		}
	}
	return items
}

func (v *SliceGridView[S]) EditItemIDs(ids []string, isEditOnNewStruct bool, after func(ok bool)) {
	v.editOrViewItemIDs(ids, v.EditParameters.IsEditOnNewStruct, false, after)
}

func (v *SliceGridView[S]) ViewItemIDs(ids []string, isEditOnNewStruct bool, after func(ok bool)) {
	v.editOrViewItemIDs(ids, v.EditParameters.IsEditOnNewStruct, true, after)
}

func (v *SliceGridView[S]) editOrViewItemIDs(ids []string, isEditOnNewStruct, isReadOnly bool, after func(ok bool)) {
	items := v.getItemsFromIDs(ids)
	if len(items) == 0 {
		zlog.Fatal("SGV EditItemIDs: no items. ids:", ids, v.Hierarchy())
	}
	title := "Edit"
	if isReadOnly {
		title = "View"
	}
	title += " " + v.NameOfXItemsFunc(ids, true)
	// zlog.Info("editOrViewItemIDs", title, isReadOnly, v.Hierarchy(), zdebug.CallingStackString())
	v.editOrViewItems(items, isReadOnly, title, isEditOnNewStruct, false, after)
}

func (v *SliceGridView[S]) UpdateWidgets() {
	// ids := v.Grid.SelectedIDs()
}

func (v *SliceGridView[S]) EditItems(ns []S, title string, isEditOnNewStruct, selectAfterEditing bool, after func(ok bool)) {
	v.editOrViewItems(ns, false, title, isEditOnNewStruct, selectAfterEditing, after)
}

func (v *SliceGridView[S]) ViewItems(ns []S, title string, isEditOnNewStruct, selectAfterEditing bool, after func(ok bool)) {
	v.editOrViewItems(ns, true, title, isEditOnNewStruct, selectAfterEditing, after)
}

func (v *SliceGridView[S]) editOrViewItems(ns []S, isReadOnly bool, title string, isEditOnNewStruct, selectAfterEditing bool, after func(ok bool)) {
	params := v.EditParameters
	params.Field.Flags |= zfields.FlagIsLabelize
	if params.Styling.Spacing == zfloat.Undefined {
		params.Styling.Spacing = 10
	}
	params.IsEditOnNewStruct = isEditOnNewStruct
	if isEditOnNewStruct {
		params.HideStatic = true
	}
	att := zpresent.ModalConfirmAttributes()
	if isReadOnly {
		att = zpresent.ModalPopupAttributes()
		att.ModalDimBackground = true
	}
	att.DocumentationIconPath = v.EditDialogDocumentationPath
	att.TitledMargin = zgeo.RectFromXY2(zstyle.DefaultRowLeftMargin, 0, 0, 0)
	zfields.EditOrViewStructSlice(&ns, isReadOnly, params, title, att, func(ok bool) bool {
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
				ids = append(ids, GetIDForItem(&n))
			}
			v.Grid.SelectCells(ids, true, true)
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

func (v *SliceGridView[S]) HandleDeleteKey(ask bool, overrideIDs []string) {
	ids := v.Grid.SelectedIDsOrHoverID()
	if len(ids) == 0 {
		ids = overrideIDs
	}
	if len(ids) == 0 {
		return
	}
	if ask {
		v.DeleteItemsAsk(ids)
	} else {
		go v.DeleteItemsFunc(ids)
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
		sid := GetIDForItem(&s)
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

func (v *SliceGridView[S]) CreateDefaultMenuItems(ids []string, forSingleCell bool) []zmenu.MenuedOItem {
	var items []zmenu.MenuedOItem
	if zdocs.IsGettingSearchItems || v.Options&AllowNew != 0 && !forSingleCell {
		add := zmenu.MenuedSCFuncAction("Add New "+v.StructName+"…", 'N', 0, v.addNewItem)
		items = append(items, add)
	}
	if v.Grid.CellCountFunc() > 0 || zdocs.IsGettingSearchItems {
		if v.Grid.MultiSelectable && !forSingleCell {
			all := zmenu.MenuedSCFuncAction("Select All", 'A', 0, func() {
				v.Grid.SelectAll(true)
			})
			items = append(items, all)
		}
		if len(ids) > 0 || zdocs.IsGettingSearchItems {
			nitems := v.NameOfXItemsFunc(ids, true)
			if v.Options&AllowDuplicate != 0 || zdocs.IsGettingSearchItems {
				del := zmenu.MenuedSCFuncAction("Duplicate "+nitems+"…", 'D', 0, func() {
					v.duplicateItems(nitems, ids)
				})
				items = append(items, del)
			}
			if v.Options&AllowDelete != 0 || zdocs.IsGettingSearchItems {
				del := zmenu.MenuedSCFuncAction("Delete "+nitems+"…", zkeyboard.KeyBackspace, 0, func() {
					v.HandleDeleteKey(true, ids)
				})
				items = append(items, del)
			}
			if v.Options&AllowEdit != 0 || zdocs.IsGettingSearchItems {
				edit := zmenu.MenuedSCFuncAction("Edit "+nitems, ' ', 0, func() {
					// zlog.Info("SGV.Edit")
					v.EditItemIDs(ids, false, nil)
				})
				items = append(items, edit)
			}
			if v.Options&AllowView != 0 || zdocs.IsGettingSearchItems {
				edit := zmenu.MenuedSCFuncAction("View "+nitems, ' ', 0, func() {
					v.ViewItemIDs(ids, false, nil)
				})
				items = append(items, edit)
			}
		}
		if v.Options&AllowCopyPaste != 0 || zdocs.IsGettingSearchItems {
			if len(ids) > 0 || zdocs.IsGettingSearchItems {
				nitems := v.NameOfXItemsFunc(ids, true)
				copy := zmenu.MenuedFuncAction("Copy "+nitems+" to Clipboard", func() {
					v.copyItemsToClipboard(ids)
				})
				copy.Shortcut = zkeyboard.CopyKeyMod
				items = append(items, copy)
			}
			if !forSingleCell || zdocs.IsGettingSearchItems {
				name := "items"
				if v.StructName != "" {
					name = zwords.PluralizeEnglishWord(v.StructName)
				}
				paste := zmenu.MenuedFuncAction("Paste from Clipboard to add "+name+"…", func() {
					v.pasteItemsFromClipboard()
				})
				paste.Shortcut = zkeyboard.PasteKeyMod
				items = append(items, paste)
			}
		}
	}
	return items
}

func (v *SliceGridView[S]) pasteItemsFromClipboard() {
	zclipboard.GetString(func(str string) {
		line, body := zstr.SplitInTwo(str, "\n")
		var stype string
		if !zstr.HasPrefix(line, "zcopyitem: ", &stype) {
			zalert.ShowError(nil, "Paste buffer doesn't contain items to paste")
			return
		}
		var s S
		rtype := reflect.TypeOf(s)
		stype2 := zreflect.MakeTypeNameWithPackage(rtype)
		if stype != stype2 {
			zalert.ShowError(nil, "Paste type is not the same as wanted receive type\n", stype, "\n", stype2)
			return
		}
		var slice []S
		err := json.Unmarshal([]byte(body), &slice)
		if err != nil {
			zalert.ShowError(nil, "Couldn't unpack paste data")
			return
		}
		v.ValidateClipboardPasteItemsFunc(&slice, func() {
			go v.StoreChangedItemsFunc(slice)
			var ids []string
			for _, s := range slice {
				ids = append(ids, GetIDForItem(&s))
			}
			v.Grid.SelectCells(ids, true, false)
		})
	})
}

func (v *SliceGridView[S]) copyItemsToClipboard(ids []string) {
	var items []S
	for _, id := range ids {
		s := v.StructForID(id)
		items = append(items, *s)
	}
	djson, err := json.Marshal(items)
	if err != nil {
		zalert.ShowError(err)
		return
	}
	var s S
	rtype := reflect.TypeOf(s)
	str := "zcopyitem: " + zreflect.MakeTypeNameWithPackage(rtype) + "\n"
	str += string(djson)
	zclipboard.SetString(str)
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

func (v *SliceGridView[S]) GetSearchableItems(currentPath []zdocs.PathPart) []zdocs.SearchableItem {
	var parts []zdocs.SearchableItem
	if !v.IsSearchable() {
		return nil
	}
	basePath := currentPath
	if v.Options&AddNameAsSearchItem != 0 {
		basePath = zdocs.AddedPath(currentPath, zdocs.StaticField, v.ObjectName(), v.ObjectName())
	}
	tablePath := zdocs.AddedPath(basePath, zdocs.StaticField, "Table", "Table")
	columnsPath := zdocs.AddedPath(tablePath, zdocs.StaticField, "Columns", "Columns")
	var s S
	v.ReadyToShow(false)
	zfields.ForEachField(&s, zfields.FieldParameters{IgnoreUseInAndINTags: true}, nil, func(each zfields.FieldInfo) bool {
		if each.Field.HasFlag(zfields.FlagIsNotGUISearchable) {
			return true
		}
		items := zfields.MakeSearchableFieldItems(columnsPath, each.Field)
		parts = append(parts, items...)
		return true
	})
	if v.ActionMenu != nil {
		path := zdocs.AddedPath(tablePath, zdocs.StaticField, "Action Menu", "menu")
		items := v.ActionMenu.GetSearchableItems(path)
		parts = append(parts, items...)
	}
	if v.Options&RowsGUISearchable != 0 {
		v.doFilter(*v.slicePtr) // we need to get filteredSlice set since it's what's acrtually used
		path := zdocs.AddedPath(tablePath, zdocs.StaticField, "Rows", "Rows")
		items := v.Grid.GetSearchableItems(path)
		parts = append(parts, items...)
	}
	if v.Bar != nil {
		path := zdocs.AddedPath(tablePath, zdocs.StaticField, "Bar", "Bar")
		items := v.Bar.GetSearchableItems(path)
		parts = append(parts, items...)
	}
	return parts
}
