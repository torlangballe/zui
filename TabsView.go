// +build zui

package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
)

const tabSeparatorID = "tab-separator"

type tab struct {
	id             string
	create         func(delete bool) View
	view           View
	childAlignment zgeo.Alignment
	image          *Image
}

type TabsView struct {
	StackView
	Header              *StackView
	ChildView           View
	CurrentID           string
	tabs                map[string]*tab
	separatorForIDs     []string
	SeparatorLineInset  float64
	ChangedHandler      func(newID string)
	ButtonName          string //
	selectedButtonColor zgeo.Color
	MaxImageSize        zgeo.Size
}

func TabsViewNew(name string, buttons bool) *TabsView {
	v := &TabsView{}
	v.StackView.Init(v, true, name)
	v.SetSpacing(0)
	v.Header = StackViewHor("header")
	if buttons {
		v.ButtonName = TabsDefaultButtonName
		v.SetMargin(zgeo.RectFromXY2(0, 4, 0, 0))
		v.Header.SetMargin(zgeo.RectFromXY2(2, 0, 0, 0))
	} else {
		v.MaxImageSize = zgeo.Size{60, 20}
		v.Header.SetMargin(zgeo.RectFromXY2(6, 3, -6, -3))
	}
	v.tabs = map[string]*tab{}
	v.Header.SetSpacing(12)
	v.Add(v.Header, zgeo.Left|zgeo.Top|zgeo.HorExpand)
	return v
}

func (v *TabsView) AddSeparatorLine(thickness float64, color zgeo.Color, corner float64, forIDs []string) {
	cv := CustomViewNew(tabSeparatorID)
	cv.SetMinSize(zgeo.Size{10, thickness})
	cv.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
		selectedView := v.Header.FindViewWithName(v.CurrentID, false)
		canvas.SetColor(color, 1)
		if selectedView != nil {
			r := selectedView.Rect()
			x0 := r.Pos.X + v.SeparatorLineInset
			x1 := r.Max().X - v.SeparatorLineInset
			r = rect
			r.SetMaxX(x0)
			path := zgeo.PathNewRect(r, zgeo.Size{})
			canvas.FillPath(path)
			r = rect
			r.SetMinX(x1)
			path = zgeo.PathNewRect(r, zgeo.Size{})
			canvas.FillPath(path)
		} else {
			path := zgeo.PathNewRect(rect, zgeo.Size{})
			canvas.FillPath(path)
		}
	})
	v.Add(cv, zgeo.TopLeft|zgeo.HorExpand)
	v.separatorForIDs = forIDs
}

var TabsDefaultButtonName = "gray-tab"
var TabsDefaultTextColor = zgeo.ColorWhite

// AddTab adds a new tab to the row of tabs.
// id is unique id that identifies it.
// title is what's written in the tab, if ButtonName != "".
// ipath is path to image, if ButtonName != "" shown on right, otherwise centered
// set makes it the current tab after adding
// align is how to align the content child view
// create is a function to create or delete the content child each time tab is set.
func (v *TabsView) AddTab(id, title, ipath string, set bool, create func(delete bool) View) {
	if title == "" {
		title = id
	}
	var button *ShapeView
	var view View
	tab := &tab{}
	tab.id = id
	minSize := zgeo.Size{20, 26}
	tab.childAlignment = zgeo.Left | zgeo.Top | zgeo.Expand
	if v.ButtonName != "" {
		// zlog.Info("Add Tab button:", title, v.ButtonName)
		b := ButtonViewNew(title, v.ButtonName, minSize, zgeo.Size{11, 12})
		button = &b.ShapeView
		button.SetTextColor(TabsDefaultTextColor)
		button.SetMarginS(zgeo.Size{10, 0})
		button.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
		view = b
	} else {
		button = ShapeViewNew(ShapeViewTypeRoundRect, minSize)
		button.SetColor(v.selectedButtonColor)
		button.MaxSize = v.MaxImageSize
		button.ImageMargin = zgeo.Size{}
		view = button
	}
	button.MaxSize.H = 26
	button.SetObjectName(id)
	if ipath != "" {
		button.SetImage(nil, ipath, nil)
	}
	tab.create = create
	v.tabs[id] = tab
	button.SetPressedHandler(func() {
		v.SetTab(id)
	})
	v.Header.Add(view, zgeo.BottomLeft)
	if set {
		v.SetTab(id)
	}
}

// AddTabWithView calls AddTabFunc, but with a fixed view instead of dynamically created/deleted one
func (v *TabsView) AddTabWithView(id, title, ipath string, set bool, view View) {
	v.AddTab(id, title, ipath, set, func(delete bool) View {
		if delete {
			return nil // AddTab deletes
		}
		return view
	})
}

func (v *TabsView) setButtonOn(id string, on bool) {
	view := v.Header.FindViewWithName(id, false)
	// zlog.Info("setButtonOn:", id, on, view != nil)
	if view != nil {
		button, _ := view.(*ButtonView)
		if button != nil {
			str := TabsDefaultButtonName
			style := FontStyleNormal
			if on {
				str += "-selected"
				style = FontStyleBold
			}
			button.SetImageName(str, zgeo.Size{11, 12})
			button.SetFont(FontNice(FontDefaultSize, style))
		} else { // image only
			col := zgeo.ColorClear
			if on {
				col = v.selectedButtonColor
			}
			v.SetBGColor(col)
		}
	}
}

func (v *TabsView) findTab(id string) *tab {
	for _, t := range v.tabs {
		zlog.Info("Find:", id, t.id)
		if t.id == id {
			return t
		}
	}
	return nil
}

func (v *TabsView) SetChildAlignment(id string, a zgeo.Alignment) {
	t := v.findTab(id)
	t.childAlignment = a
}

func (v *TabsView) SetButtonAlignment(id string, a zgeo.Alignment) {
	cell, _ := v.Header.FindCellWithName(id)
	// zlog.Info("FIND:", id, cell)
	cell.Alignment = a
}

func (v *TabsView) SetTab(id string) {
	if v.CurrentID != id {
		// zlog.Info("SetTab!:", v.CurrentID, id, len(v.cells))
		if v.CurrentID != "" {
			v.tabs[v.CurrentID].create(true)
			v.setButtonOn(v.CurrentID, false)
		}
		if v.ChildView != nil {
			// zlog.Info("Remove Child!:", v.ChildView.ObjectName())
			v.RemoveChild(v.ChildView)
		}
		tab := v.tabs[id]
		v.ChildView = tab.create(false)
		v.Add(v.ChildView, tab.childAlignment)
		v.CurrentID = id
		v.setButtonOn(id, true)
		hasSeparator := zstr.StringsContain(v.separatorForIDs, id)
		arrange := false // don't arrange on collapse, as it is done below, or on present, and causes problems if done now
		// zlog.Info("Call collapse:", id, len(v.cells))
		v.CollapseChildWithName(tabSeparatorID, !hasSeparator, arrange)
		// zlog.Info("TV SetTab", v.Presented)
		if !v.Presented {
			// zlog.Info("Set Tab, exit because not presented yet", id)
			return
		}
		et, _ := v.View.(ExposableType)
		// if !v.Presented {
		// 	return
		// }
		presentViewCallReady(v.ChildView, true)
		presentViewPresenting = true
		v.ArrangeChildren(nil) // This can create table rows and do all kinds of things that load images etc.

		ct := v.View.(ContainerType)
		WhenContainerLoaded(ct, func(waited bool) {
			// zlog.Info("SetTab Loaded")
			// zlog.Info("Set Tab container loaded:", waited)
			if waited { // if we waited for some loading, caused by above arranging, lets re-arrange
				v.ArrangeChildren(nil)
			}
		})

		// zlog.Info("bt-banner tab arranged.", tab.childAlignment, v.ChildView.Rect())
		presentViewPresenting = false
		presentViewCallReady(v.ChildView, false)
		if et != nil {
			et.drawIfExposed()
		}
		/*
			})
		*/
		if v.ChangedHandler != nil {
			v.ChangedHandler(id)
		}
	}
}

// func (v *TabsView) GetChildren() []View {
// 	return v.StackView.GetChildren()
// }

func (v *TabsView) ArrangeChildren(onlyChild *View) {
	// zlog.Info("TabView ArrangeChildren")
	v.StackView.ArrangeChildren(onlyChild)
}

func GetParentTabsCurrentID(child View) string {
	n := ViewGetNative(child)
	for _, p := range n.AllParents() {
		t, _ := p.View.(*TabsView)
		if t != nil {
			return t.CurrentID
		}
	}
	return ""
}
