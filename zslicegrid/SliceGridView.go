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

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zbutton"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
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
	StoreChangedItemsFunc func(items []S)
	DeleteItemsFunc       func(ids []string)

	slicePtr     *[]S
	editButton   *zbutton.Button
	deleteButton *zbutton.Button
	addButton    *zimageview.ImageView
}

type OptionType int

const (
	AddNone        OptionType = 0
	AddBar         OptionType = 1
	AddEdit        OptionType = 2
	AddDelete      OptionType = 4
	AddDarkPlus    OptionType = 8
	AddLightPlus   OptionType = 16
	AddEditDelete             = AddEdit | AddDelete
	AddButtonsMask OptionType = AddEdit | AddDelete | AddDarkPlus | AddLightPlus
)

func NewView[S zstr.StrIDer](slicePtr *[]S, storeName string, option OptionType) (sv *SliceGridView[S]) {
	v := &SliceGridView[S]{}
	v.Init(v, slicePtr, storeName, option)
	return v
}

func (v *SliceGridView[S]) Init(view zview.View, slicePtr *[]S, storeName string, options OptionType) {
	v.StackView.Init(view, true, "slice-grid-view")
	v.SetObjectName(storeName)
	v.SetSpacing(0)
	v.StructName = "item"
	v.slicePtr = slicePtr

	v.EditParameters = zfields.FieldViewParametersDefault()
	v.EditParameters.LabelizeWidth = 120
	v.EditParameters.HideStatic = true

	var s S
	var a any = s
	_, hasHierarchy := a.(ChildrenOwner)

	if options&(AddBar) != 0 {
		// zlog.Info("AddBar!!!")
		v.Bar = zcontainer.StackViewHor("bar")
		v.Bar.SetSpacing(4)
		v.Bar.SetMargin(zgeo.RectFromXY2(6, 5, -6, -3))
		v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
	}
	if options&AddEdit != 0 {
		v.editButton = zbutton.New("")
		v.editButton.SetObjectName("edit")
		v.editButton.SetMinWidth(130)
		v.Bar.Add(v.editButton, zgeo.CenterLeft)
	}
	if options&AddDelete != 0 {
		v.deleteButton = zbutton.New("")
		v.deleteButton.SetObjectName("delete")
		v.deleteButton.SetMinWidth(135)
		v.Bar.Add(v.deleteButton, zgeo.CenterLeft)
	}
	if options&(AddDarkPlus|AddLightPlus) != 0 {
		str := "white"
		if options&AddDarkPlus != 0 {
			str = "dark"
		}
		v.addButton = zimageview.New(nil, "images/plus-circled-"+str+".png", zgeo.Size{16, 16})
		v.Bar.Add(v.addButton, zgeo.CenterLeft)
		v.addButton.SetPressedHandler(v.handlePlusButtonPressed)
	}

	v.Grid = zgridlist.NewView(storeName + "-GridListView")
	v.Grid.CellCountFunc = func() int {
		if !hasHierarchy {
			return len(*v.slicePtr)
		}
		return v.countHierarcy(v.slicePtr)
	}
	v.Grid.IDAtIndexFunc = func(i int) string {
		if !hasHierarchy {
			return (*slicePtr)[i].GetStrID()
		}
		var count int
		return v.getIDForIndex(v.slicePtr, i, &count)
	}
	v.Grid.HandleKeyFunc = func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
		if key == zkeyboard.KeyBackspace {
			v.handleDeleteKey(mod != zkeyboard.ModifierCommand)
			return true
		}
		if key == zkeyboard.KeyReturn {
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
		v.Grid.LayoutCells(true)
		v.UpdateWidgets()
	}
	if hasHierarchy {
		v.Grid.HierarchyLevelFunc = v.calculateHierarchy
	}
	v.Grid.SetMargin(zgeo.RectFromXY2(6, 4, -6, -4))
	v.Grid.Spacing = zgeo.Size{14, 4}
	v.Grid.CellColor = zgeo.ColorNewGray(0.95, 1)
	v.Grid.MultiplyAlternate = 0.95
	v.Grid.MultiSelectable = true

	v.Add(v.Grid, zgeo.TopLeft|zgeo.Expand, zgeo.Size{}) //.Margin = zgeo.Size{4, 0}
	return
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

func (v *SliceGridView[S]) getIDForIndex(slicePtr *[]S, index int, count *int) string {
	for i, s := range *slicePtr {
		// if index != -1 {
		// 	zlog.Info("getIDForIndex", index, i, s)
		// }
		id := s.GetStrID()
		if index == *count {
			return id
		}
		(*count)++
		if v.Grid.OpenBranches[id] {
			children := v.getChildren(slicePtr, i)
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

func (v *SliceGridView[S]) getChildren(slicePtr *[]S, i int) *[]S {
	var a any = (*slicePtr)[i]
	return a.(ChildrenOwner).GetChildren().(*[]S)
}

func (v *SliceGridView[S]) countHierarcy(slicePtr *[]S) int {
	var count int
	v.getIDForIndex(slicePtr, -1, &count)
	// zlog.Info("Count:", v.ObjectName(), count, v.openBranches)
	return count
}

func (v *SliceGridView[S]) structForID(slicePtr *[]S, id string) *S {
	for i, s := range *slicePtr {
		sid := s.GetStrID()
		if id == sid {
			return &(*slicePtr)[i]
		}
		if v.Grid.HierarchyLevelFunc != nil {
			if !v.Grid.OpenBranches[sid] {
				continue
			}
			children := v.getChildren(slicePtr, i)
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

func (v *SliceGridView[S]) ReadyToShow(beforeWindow bool) {
	if beforeWindow {
		if v.SortFunc != nil {
			v.SortFunc(*v.slicePtr) // Do this beforeWindow shown, as the sorted cells get placed correctly then
			// v.UpdateViewFunc()
		}
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
	update := (len(s) != len(*v.slicePtr) || zstr.HashAnyToInt64(s) != zstr.HashAnyToInt64(*v.slicePtr))
	// zlog.Info("UpdateSlice!", update)
	if update {
		if v.SortFunc != nil {
			v.SortFunc(s)
		}
		*v.slicePtr = s
		v.UpdateViewFunc()
	}
}

func (v *SliceGridView[S]) UpdateSliceWithSelf() {
	v.UpdateSlice(*v.slicePtr)
}

func (v *SliceGridView[S]) HandleEditButtonPressed() {
	ids := v.Grid.SelectedIDs()
	if len(ids) == 0 && v.Grid.CurrentHoverID != "" {
		ids = []string{v.Grid.CurrentHoverID}
	}
	v.EditItems(ids)
}

func (v *SliceGridView[S]) EditItems(ids []string) {
	title := "Edit "
	var items []S

	for i := 0; i < len(*v.slicePtr); i++ {
		sid := (*v.slicePtr)[i].GetStrID()
		if zstr.StringsContain(ids, sid) {
			items = append(items, (*v.slicePtr)[i])
		}
	}
	if len(items) == 0 {
		zlog.Fatal(nil, "SGV EditItems: no items. ids:", len(ids))
	}
	title += v.NameOfXItemsFunc(ids, true)
	zfields.PresentOKCancelStructSlice(&items, v.EditParameters, title, zpresent.AttributesNew(), func(ok bool) bool {
		// zlog.Info("Edited items:", ok)
		if !ok {
			return true
		}
		for _, item := range items {
			for i, s := range *v.slicePtr {
				if s.GetStrID() == item.GetStrID() {
					(*v.slicePtr)[i] = item
					// fmt.Printf("edited: %+v %d\n", (*v.slice)[i], i)
				}
			}
		}
		v.UpdateViewFunc()
		if v.StoreChangedItemsFunc != nil {
			ztimer.StartIn(0.1, func() {
				go v.StoreChangedItemsFunc(items)
			})
		}
		return true
	})
}

func (v *SliceGridView[S]) setButtonWithCount(ids []string, button *zbutton.Button) {
	// zlog.Info("setButtonWithCount", button.ObjectName())
	str := button.ObjectName() + " "
	if len(ids) > 0 {
		str += v.NameOfXItemsFunc(ids, false)
	}
	button.SetUsable(len(ids) > 0)
	button.SetText(str)
}

func (v *SliceGridView[S]) UpdateWidgets() {
	ids := v.Grid.SelectedIDs()
	if v.Bar != nil {
		for _, c := range v.Bar.GetChildren(false) {
			b, _ := c.(*zbutton.Button)
			if b != nil {
				v.setButtonWithCount(ids, b)
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
	for i, s := range *v.slicePtr {
		children := v.getChildren(v.slicePtr, i)
		sid := s.GetStrID()
		kids := (v.Grid.OpenBranches[sid] && len(*children) > 0)
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
	level, leaf, _ = v.getHierarchy(v.slicePtr, 1, id)
	return
}
