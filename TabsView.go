//go:build zui
// +build zui

package zui

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

type tab struct {
	id             string
	create         func(delete bool) View
	view           View
	childAlignment zgeo.Alignment
	image          *zimage.Image
}

type TabsView struct {
	StackView
	Header               *StackView
	ChildView            View
	CurrentID            string
	tabs                 map[string]*tab
	separatorForIDs      []string
	SeparatorLineInset   float64
	ChangedHandler       func(newID string)
	ButtonName           string //
	selectedImageBGColor zgeo.Color
	MaxImageSize         zgeo.Size
}

const tabSeparatorID = "tab-separator"

var (
	TabsDefaultButtonName           = "gray-tab"
	TabsDefaultTextColor            = StyleGrayF(0.1, 0.9)
	TabsDefaultSelectedImageBGColor = StyleColF(zgeo.ColorNew(0, 0, 1, 0.2), zgeo.ColorNew(0, 0, 9, 0.2))
)

func TabsViewNew(name string, buttons bool) *TabsView {
	v := &TabsView{}
	v.StackView.Init(v, true, name)
	v.SetSpacing(0)
	v.SetBGColor(ListViewDefaultBGColor())
	v.Header = StackViewHor("header")
	if buttons {
		v.ButtonName = TabsDefaultButtonName
		v.Header.SetMargin(zgeo.RectFromXY2(2, 4, 0, 0))
	} else {
		v.MaxImageSize = zgeo.Size{60, 24}
		v.Header.SetMargin(zgeo.RectFromXY2(8, 6, -8, -6))
	}
	v.tabs = map[string]*tab{}
	v.Header.SetSpacing(12)
	v.Add(v.Header, zgeo.Left|zgeo.Top|zgeo.HorExpand)
	v.selectedImageBGColor = TabsDefaultSelectedImageBGColor()
	if !buttons {
		v.Header.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view View) {
			sv, i := v.Header.FindViewWithName(v.CurrentID, false)
			if sv != nil {
				r := sv.Rect()
				r.Pos.Y = 0
				r.Size.H = rect.Size.H
				r.SetMinX(r.Min().X - v.Header.Spacing()/2)
				r.SetMaxX(r.Max().X + v.Header.Spacing()/2)
				if i == 0 {
					r.SetMinX(rect.Min().X)
				}
				if i == v.Header.CountChildren() && r.Max().X > rect.Max().X-8 {
					r.SetMaxX(rect.Max().X)
				}
				canvas.SetColor(v.selectedImageBGColor)
				path := zgeo.PathNewRect(r, zgeo.Size{})
				canvas.FillPath(path)
			}
		})
	}
	return v
}

func (v *TabsView) AddSeparatorLine(thickness float64, color zgeo.Color, corner float64, forIDs []string) {
	cv := CustomViewNew(tabSeparatorID)
	cv.SetMinSize(zgeo.Size{10, thickness})
	cv.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view View) {
		selectedView, _ := v.Header.FindViewWithName(v.CurrentID, false)
		// zlog.Info("SepDRAW:", selectedView.Rect())
		canvas.SetColor(color)
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
	minSize := zgeo.Size{20, 22}
	tab.childAlignment = zgeo.Left | zgeo.Top | zgeo.Expand
	if v.ButtonName != "" {
		// zlog.Info("Add Tab button:", title, v.ButtonName)
		b := ImageButtonViewNew(title, v.ButtonName, minSize, zgeo.Size{11, 8})
		button = &b.ShapeView
		button.SetTextColor(TabsDefaultTextColor())
		button.SetMarginS(zgeo.Size{10, 0})
		button.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
		view = b
	} else {
		button = ShapeViewNew(ShapeViewTypeNone, minSize)
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
		go v.SetTab(id, nil)
	})
	v.Header.Add(view, zgeo.BottomLeft)
	if set {
		v.SetTab(id, nil)
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
	view, _ := v.Header.FindViewWithName(id, false)
	// zlog.Info("setButtonOn:", id, on, view != nil)
	if view != nil {
		button, _ := view.(*ImageButtonView)
		if button != nil {
			str := TabsDefaultButtonName
			if on {
				str += "-selected"
			}
			button.SetImageName(str, zgeo.Size{11, 8})
		} else { // image only
			v.Header.Expose()
		}
	}
}

func (v *TabsView) findTab(id string) *tab {
	for _, t := range v.tabs {
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
	cell.Alignment = a
}

func (v *TabsView) SetTab(id string, done func()) {
	if v.CurrentID == id {
		if done != nil {
			done()
		}
		return
	}
	if v.CurrentID != "" {
		v.tabs[v.CurrentID].create(true)
		v.setButtonOn(v.CurrentID, false)
	}
	if v.ChildView != nil {
		v.RemoveChild(v.ChildView)
	}
	tab := v.tabs[id]
	v.ChildView = tab.create(false)
	v.Add(v.ChildView, tab.childAlignment)
	v.CurrentID = id
	v.setButtonOn(id, true)
	hasSeparator := zstr.StringsContain(v.separatorForIDs, id)
	arrange := false // don't arrange on collapse, as it is done below, or on present, and causes problems if done now
	v.CollapseChildWithName(tabSeparatorID, !hasSeparator, arrange)
	if !v.Presented {
		return
	}
	// zlog.Info("SetTab!:", v.CurrentID, id, len(v.cells), "collapse:", !hasSeparator)

	ExposeView(v.View)
	//!		et, _ := v.View.(ExposableType)
	// if !v.Presented {
	// 	return
	// }
	//		PresentViewCallReady(v.ChildView, true)
	// presentViewPresenting = true
	v.ArrangeChildren() // This can create table rows and do all kinds of things that load images etc.
	ct := v.View.(ContainerType)
	// presentViewPresenting = false
	PresentViewCallReady(v.ChildView, false)
	//! if et != nil {
	// 	et.drawIfExposed()
	// }
	if v.ChangedHandler != nil {
		v.ChangedHandler(id)
	}
	WhenContainerLoaded(ct, func(waited bool) {
		// zlog.Info("Set Tab container loaded:", waited)
		if waited { // if we waited for some loading, caused by above arranging, lets re-arrange
			v.ArrangeChildren()
		}
		if done != nil {
			done()
		}
	})
}

func (v *TabsView) ArrangeChildren() {
	// zlog.Info("TabView ArrangeChildren")
	v.StackView.ArrangeChildren()
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
