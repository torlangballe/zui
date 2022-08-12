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
	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zbutton"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zgridlist"
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
	StructName            string
	NameOfXItemsFunc      func(ids []string, singleSpecial bool) string
	DeleteAskSubTextFunc  func(ids []string) string
	UpdateViewFunc        func()
	SortFunc              func(s []S)
	StoreChangedItemsFunc func(items []S)
	DeleteItemsFunc       func(ids []string)

	slice        *[]S
	editButton   *zbutton.Button
	deleteButton *zbutton.Button
}

type OptionType int

const (
	AddNone    OptionType = 0
	AddBar     OptionType = 1
	AddEdit    OptionType = 2
	AddDelete  OptionType = 4
	AddButtons OptionType = AddEdit | AddDelete
)

func NewView[S zstr.StrIDer](slice *[]S, storeName string, option OptionType) (sv *SliceGridView[S]) {
	v := &SliceGridView[S]{}
	v.Init(v, slice, storeName, option)
	return v
}

func (v *SliceGridView[S]) Init(view zview.View, slice *[]S, storeName string, option OptionType) {
	v.StackView.Init(view, true, "slice-grid-view")
	v.SetObjectName(storeName)
	v.SetSpacing(0)
	v.StructName = "item"
	v.slice = slice

	var s S
	var a any = s
	_, hasHierarchy := a.(ChildrenOwner)

	if option&(AddBar|AddButtons) != 0 {
		v.Bar = zcontainer.StackViewHor("bar")
		v.Bar.SetSpacing(4)
		v.Bar.SetMargin(zgeo.RectFromXY2(6, 5, -6, -3))
		v.Add(v.Bar, zgeo.TopLeft|zgeo.HorExpand)
	}
	if option&AddEdit != 0 {
		v.editButton = zbutton.New("")
		v.editButton.SetObjectName("edit")
		v.editButton.SetMinWidth(130)
		v.Bar.Add(v.editButton, zgeo.CenterLeft)
	}
	if option&AddDelete != 0 {
		v.deleteButton = zbutton.New("")
		v.deleteButton.SetObjectName("delete")
		v.deleteButton.SetMinWidth(135)
		v.Bar.Add(v.deleteButton, zgeo.CenterLeft)
	}
	v.Grid = zgridlist.NewView(storeName + "-GridListView")

	v.Grid.CellCountFunc = func() int {
		if !hasHierarchy {
			return len(*v.slice)
		}
		return v.countHierarcy(v.slice)
	}
	v.Grid.IDAtIndexFunc = func(i int) string {
		if !hasHierarchy {
			return (*slice)[i].GetStrID()
		}
		var count int
		return v.getIDForIndex(v.slice, i, &count)
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

func (v *SliceGridView[S]) countHierarcy(slice *[]S) int {
	var count int
	v.getIDForIndex(slice, -1, &count)
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
	return v.structForID(v.slice, id)
}

func (v *SliceGridView[S]) ReadyToShow(beforeWindow bool) {
	if beforeWindow {
		if v.SortFunc != nil {
			v.SortFunc(*v.slice) // Do this beforeWindow shown, as the sorted cells get placed correctly then
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
	update := (len(s) != len(*v.slice) || zstr.HashAnyToInt64(s) != zstr.HashAnyToInt64(*v.slice))
	// zlog.Info("UpdateSlice!", update)
	if update {
		if v.SortFunc != nil {
			v.SortFunc(s)
		}
		*v.slice = s
		if v.UpdateViewFunc != nil {
			v.UpdateViewFunc()
		}
	}
}

func (v *SliceGridView[S]) HandleEditButtonPressed() {
	ids := v.Grid.SelectedIDs()
	v.EditItems(ids)
}

func (v *SliceGridView[S]) EditItems(ids []string) {
	title := "Edit "
	var items []S

	for i := 0; i < len(*v.slice); i++ {
		sid := (*v.slice)[i].GetStrID()
		if zstr.StringsContain(ids, sid) {
			items = append(items, (*v.slice)[i])
		}
	}
	if len(items) == 0 {
		zlog.Fatal(nil, "SGV EditItems: no items. ids:", len(ids))
	}
	title += v.NameOfXItemsFunc(ids, true)
	params := zfields.FieldViewParametersDefault()
	params.LabelizeWidth = 120
	params.HideStatic = true
	zfields.PresentOKCancelStructSlice(&items, params, title, zpresent.AttributesNew(), func(ok bool) bool {
		zlog.Info("Edited items:", ok)
		if !ok {
			return true
		}
		for _, item := range items {
			for i, s := range *v.slice {
				if s.GetStrID() == item.GetStrID() {
					(*v.slice)[i] = item
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
	for i, s := range *slice {
		children := v.getChildren(slice, i)
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
	level, leaf, _ = v.getHierarchy(v.slice, 1, id)
	return
}
