//go:build zui

package zgroup

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
)

type MenuGroupView struct {
	GroupBase
	Menu *zmenu.MenuView
}

func MenuGroupViewNew(storeName, title string, styling, titleStyling zstyle.Styling) *MenuGroupView {
	v := &MenuGroupView{}
	v.GroupBase.Init()
	v.StackView.Init(v, true, storeName)
	v.SetSpacing(6)

	MakeStackTitledFrame(&v.StackView, title, false, styling, titleStyling)
	v.SetIndicatorSelectionFunc = v.setMenuItem

	h, _ := v.FindViewWithName("header", false)
	// zlog.Info("MenuGroupViewNew:", v.Hierarchy())
	v.header = h.(*zcontainer.StackView)
	v.header.SetSpacing(6)

	a := zgeo.Left
	if title != "" {
		a = zgeo.Right
	}
	v.Menu = zmenu.NewView("menu", nil, nil)
	v.Menu.SetSelectedHandler(v.handleMenuSelected)
	v.header.Add(v.Menu, zgeo.VertCenter|a)
	v.RemoveIndicatorFunc = func(id string) {
		v.Menu.RemoveItemByValue(id)
	}
	return v
}

func (v *MenuGroupView) GetGroupBase() *GroupBase {
	return &v.GroupBase
}

func (v *MenuGroupView) SetRect(rect zgeo.Rect) {
	v.GroupBase.SetRect(rect)
}

func (v *MenuGroupView) GetHeader() *zcontainer.StackView {
	return v.header
}

// AddItem adds a new view to the group.
// id is unique id that identifies it.
// title is what's shown in the menu to select on top
// set makes it the current tab after adding
// align is how to align the content child view
// create is a function to create or delete the content child each time tab is set.
// or view is a View to add directly, one or the other.
func (v *MenuGroupView) AddItem(id, title, imagePath string, set bool, view zview.View, create func(id string, delete bool) zview.View) {
	v.Menu.AddItem(title, id)
	v.AddGroupItem(id, view, create)
	if set {
		v.SelectItem(id, nil)
	}
	v.GroupBase.UpdateButtons()
}

func (v *MenuGroupView) RemoveItem(id string) {
	v.RemoveGroupItem((id))
}

func (v *MenuGroupView) handleMenuSelected() {
	id := v.Menu.CurrentValue().(string)
	v.SetGroupItem(id, nil)
}

func (v *MenuGroupView) UpdateCurrentItemTitle(text string) {
	v.Menu.ChangeNameForValue(text, v.CurrentID)
	a, _ := v.Menu.Parent().View.(zcontainer.Arranger)
	if a != nil {
		a.ArrangeChildren()
	}
}

func (v *MenuGroupView) setMenuItem(id string, selected bool) {
	if selected {
		v.Menu.SelectWithValue(id)
	}
}

func (v *MenuGroupView) SelectItem(id string, done func()) {
	v.SetGroupItem(id, done)
}

func (v *MenuGroupView) Empty() {
	if v.ChildView != nil {
		v.RemoveChild(v.ChildView)
		v.ChildView = nil
	}
	v.Menu.Empty()
	v.GroupItems = map[string]*GroupItem{}
	v.CurrentID = ""
}
