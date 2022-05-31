//go:build zui

package zgridlist

import (
	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zwords"
)

type GridStructType zstr.StrIDer

type SliceGridView[S GridStructType] struct {
	zui.StackView
	Grid         *GridListView
	Bar          *zui.StackView
	slice        *[]S
	hoverLabel   *zui.Label
	editButton   *zui.Button
	deleteButton *zui.Button
	StructName   string

	HandleSelectionChangedFunc func()
	HandleKeyFunc              func(key zkeyboard.Key, mod zkeyboard.Modifier) bool
	NameOfXItemsFunc           func(ids []string, singleSpecial bool) string
	DeleteAskSubTextFunc       func(ids []string) string
	UpdateViewFunc             func()
	SortFunc                   func()
	StoreChangedItemsFunc      func(items *[]S)
	DeleteItemsFunc            func(ids []string)
}

const (
	actionDeleteStream            = "delete"
	actionShowATest               = "show-a-test"
	actionShowStreamURL           = "show-stream-url"
	actionShowRedirectedStreamURL = "show-redirected-stream-url"
	actionShowManifest            = "show-manifest"
	actionDuplicateStream         = "duplicate-stream"
	actionAddTest                 = "add-test"
)

func NewSliceGridView[S GridStructType](slice *[]S) (sv *SliceGridView[S]) {
	sv = &SliceGridView[S]{}
	sv.Init(sv, true, "SliceGridView")
	sv.SetSpacing(0)
	sv.StructName = "item"
	sv.slice = slice

	sv.Bar = zui.StackViewHor("bar")
	sv.Bar.SetSpacing(4)
	sv.Bar.SetMargin(zgeo.RectFromXY2(6, 5, -6, -3))

	sv.Add(sv.Bar, zgeo.TopLeft|zgeo.HorExpand)

	sv.editButton = zui.ButtonNew("")
	sv.editButton.SetMinWidth(130)
	sv.Bar.Add(sv.editButton, zgeo.CenterLeft)

	sv.deleteButton = zui.ButtonNew("")
	sv.deleteButton.SetMinWidth(135)
	sv.Bar.Add(sv.deleteButton, zgeo.CenterLeft)

	sv.hoverLabel = zui.LabelNew("")
	sv.hoverLabel.SetMinWidth(200)
	sv.Bar.Add(sv.hoverLabel, zgeo.CenterLeft|zgeo.HorExpand)

	sv.Grid = New("zgrid")
	// grid.CellCount = func() int {
	// 	return len(AllStreams)
	// }
	// grid.IDAtIndex = func(i int) string {
	// 	return strconv.FormatInt(AllStreams[i].ID, 10)
	// }
	sv.Grid.CellCount = func() int {
		return len(*sv.slice)
	}
	sv.Grid.IDAtIndex = func(i int) string {
		return (*slice)[i].GetStrID()
	}
	sv.HandleKeyFunc = func(key zkeyboard.Key, mod zkeyboard.Modifier) bool {
		if key == zkeyboard.KeyBackspace {
			sv.handleDeleteKey(mod != zkeyboard.ModifierCommand)
			return true
		}
		return false
	}
	sv.HandleSelectionChangedFunc = func() {
		// zlog.Info("HandleSelectionChanged", zlog.GetCallingStackString())
		sv.updateWidgets()
	}
	sv.NameOfXItemsFunc = func(ids []string, singleSpecial bool) string {
		return zwords.PluralWordWithCount(sv.StructName, float64(len(ids)), "", "", 0)
	}
	sv.UpdateViewFunc = func() {
		sv.Grid.LayoutCells(true)
		sv.updateWidgets()
	}

	sv.Grid.SetMargin(zgeo.RectFromXY2(6, 4, -6, -4))
	sv.Grid.Spacing = zgeo.Size{14, 4}
	// sv.Grid.CreateCell = NewStreamCell
	sv.Grid.CellColor = zgeo.ColorNewGray(0.95, 1)
	sv.Grid.MultiplyAlternate = 0.95
	sv.Grid.BorderColor = zgeo.ColorDarkGray
	sv.Grid.MultiSelectable = true
	sv.Grid.SelectColor = zgeo.ColorNew(0.66, 0.86, 0.72, 1)
	sv.Grid.PressedColor = zgeo.ColorNew(0.9, 0.98, 0.88, 1)
	sv.Grid.HoverColor = sv.Grid.PressedColor

	sv.Add(sv.Grid, zgeo.TopLeft|zgeo.Expand, zgeo.Size{}).Margin = zgeo.Size{4, 0}
	sv.updateWidgets()
	return
}

func (sv *SliceGridView[S]) ReadyToShow(beforeWindow bool) {
	if beforeWindow {
		return
	}
	zlog.Info("SGV Ready")
	sv.Grid.HandleSelectionChanged = sv.HandleSelectionChangedFunc
	sv.Grid.HandkeKey = sv.HandleKeyFunc
	sv.editButton.SetPressedHandler(sv.HandleEditButtonPressed)
	sv.deleteButton.SetPressedHandler(func() {
		sv.DeleteItemsAsk(sv.getSelectedItemsIDs())
	})
	// sv.Grid.UpdateCell = sv.UpdateCell
}

func (sv *SliceGridView[S]) HandleEditButtonPressed() {
	var items []S

	for i := 0; i < len(*sv.slice); i++ {
		sid := (*sv.slice)[i].GetStrID()
		if sv.Grid.selectedIDs[sid] {
			items = append(items, (*sv.slice)[i])
		}
	}
	sv.EditItems(&items)
}

func (sv *SliceGridView[S]) EditItems(items *[]S) {
	title := "Edit "
	ids := sv.getSelectedItemsIDs()
	title += sv.NameOfXItemsFunc(ids, true)
	params := zfields.FieldViewParametersDefault()
	params.LabelizeWidth = 120
	zfields.PresentOKCancelStructSlice(items, params, title, zui.PresentViewAttributesNew(), func(ok bool) bool {
		if !ok {
			return true
		}
		for _, item := range *items {
			// zlog.Info("edited:", s.Name, s.On, s.Type, s.AuthenticationID)
			for i, s := range *sv.slice {
				if s.GetStrID() == item.GetStrID() {
					(*sv.slice)[i] = s
				}
			}
		}
		sv.UpdateViewFunc()
		if sv.StoreChangedItemsFunc != nil {
			go sv.StoreChangedItemsFunc(items)
		}
		// go sv.Store
		// func() {
		// 	err := zrpc.ToServerClient.CallRemote("StreamsCalls.UpdateStreams", streams, nil)
		// 	if err != nil {
		// 		zui.AlertShowError(err)
		// 	}
		// }()
		return true
	})
}

func (sv *SliceGridView[S]) getSelectedItemsIDs() (ids []string) {
	for id, _ := range sv.Grid.selectedIDs {
		ids = append(ids, id)
	}
	return
}

func (sv *SliceGridView[S]) setButtonWithCount(verb string, ids []string, button *zui.Button) {
	str := verb + " "
	if len(ids) > 0 {
		str += sv.NameOfXItemsFunc(ids, false)
	}
	button.SetUsable(len(ids) > 0)
	button.SetText(str)
}

func (sv *SliceGridView[S]) updateWidgets() {
	ids := sv.getSelectedItemsIDs()
	sv.setButtonWithCount("edit", ids, sv.editButton)
	sv.setButtonWithCount("delete", ids, sv.deleteButton)
}

func (sv *SliceGridView[S]) handleDeleteKey(ask bool) {
	if len(sv.Grid.SelectedIDs()) == 0 {
		return
	}
	ids := sv.getSelectedItemsIDs()
	if ask {
		sv.DeleteItemsAsk(ids)
	} else {
		sv.DeleteItemsFunc(ids)
	}
}

func (sv *SliceGridView[S]) DeleteItemsAsk(ids []string) {
	title := "Are you sure you want to delete "
	title += sv.NameOfXItemsFunc(ids, true)
	alert := zui.AlertNewWithCancel(title + "?")
	if sv.DeleteAskSubTextFunc != nil {
		sub := sv.DeleteAskSubTextFunc(ids)
		alert.SetSub(sub)
	}
	alert.ShowOK(func() {
		go sv.DeleteItemsFunc(ids)
	})
}
