package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

const tabSeparatorID = "tab-separator"

type TabsView struct {
	StackView
	Header             *StackView
	ChildView          View
	CurrentID          string
	creators           map[string]func(bool) View
	childAlignments    map[string]zgeo.Alignment
	separatorForIDs    []string
	SeparatorLineInset float64
	ChangedHandler     func(newID string)
}

func TabsViewNew(name string) *TabsView {
	v := &TabsView{}
	v.StackView.init(v, name)
	v.Vertical = true
	v.SetMargin(zgeo.RectFromXY2(0, 4, 0, 0))
	v.creators = map[string]func(bool) View{}
	v.Header = StackViewHor("header")
	v.Header.SetMargin(zgeo.RectFromXY2(5, 0, 0, 0))
	v.Header.SetSpacing(10)
	v.childAlignments = map[string]zgeo.Alignment{}
	v.Add(zgeo.Left|zgeo.Top|zgeo.HorExpand, v.Header)
	return v
}

func (v *TabsView) AddSeparatorLine(thickness float64, color zgeo.Color, corner float64, forIDs []string) {
	cv := CustomViewNew(tabSeparatorID)
	cv.SetMinSize(zgeo.Size{10, thickness})
	cv.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
		selectedView := v.Header.FindViewWithName(v.CurrentID, false)
		canvas.SetColor(color, 1)
		if selectedView != nil {
			r := (*selectedView).Rect()
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
	v.Add(zgeo.TopLeft|zgeo.HorExpand, cv)
	v.separatorForIDs = forIDs
}

var TabsDefaultButtonName = "gray-tab"
var TabsDefaultTextColor = zgeo.ColorWhite

func (v *TabsView) AddTabFunc(id, title string, set bool, align zgeo.Alignment, creator func(del bool) View) {
	if title == "" {
		title = id
	}
	if align == zgeo.AlignmentNone {
		align = zgeo.Left | zgeo.Top | zgeo.Expand
	}
	v.childAlignments[id] = align
	button := ButtonNew(title, TabsDefaultButtonName, zgeo.Size{20, 26}, zgeo.Size{11, 12})
	button.SetObjectName(id)
	button.SetMarginS(zgeo.Size{10, 0})
	button.SetColor(TabsDefaultTextColor)
	button.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	v.creators[id] = creator
	button.SetPressedHandler(func() {
		v.SetTab(id)
	})
	v.Header.Add(zgeo.BottomLeft, button)
	if set {
		v.SetTab(id)
	}
}

func (v *TabsView) AddTab(title, id string, set bool, align zgeo.Alignment, view View) {
	v.AddTabFunc(title, id, set, align, func(del bool) View {
		if del {
			return nil
		}
		return view
	})
}

func (v *TabsView) setButtonOn(id string, on bool) {
	view := v.Header.FindViewWithName(id, false)
	// zlog.Info("setButtonOn:", id, on, view != nil)
	if view != nil {
		button := (*view).(*Button)
		str := TabsDefaultButtonName
		style := FontStyleNormal
		if on {
			str += "-selected"
			style = FontStyleBold
		}
		button.SetImageName(str, zgeo.Size{11, 12})
		button.SetFont(FontNice(FontDefaultSize, style))
	}
}
func (v *TabsView) SetTab(id string) {
	// zlog.Info("SetTab!:", v.CurrentID, id, len(v.cells))
	if v.CurrentID != id {
		if v.CurrentID != "" {
			v.creators[v.CurrentID](true)
			v.setButtonOn(v.CurrentID, false)
		}
		if v.ChildView != nil {
			// zlog.Info("Remove Child!:", v.ChildView.ObjectName())
			v.RemoveChild(v.ChildView)
		}
		v.ChildView = v.creators[id](false)
		v.Add(v.childAlignments[id], v.ChildView)
		v.CurrentID = id
		v.setButtonOn(id, true)
		hasSeparator := zstr.StringsContain(v.separatorForIDs, id)
		arrange := false // don't arrange on collapse, as it is done below, or on present, and causes problems if done now
		// zlog.Info("Call collapse:", id, len(v.cells))
		v.CollapseChildWithName(tabSeparatorID, !hasSeparator, arrange)
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
		v.ArrangeChildren(nil) // This can cerate table rows and do all kinds of things that load images etc.
		/*
			ct := v.View.(ContainerType)
				WhenContainerLoaded(ct, func(waited bool) {
					zlog.Info("SetTab Loaded")
					// zlog.Info("Set Tab container loaded:", waited)
					if waited { // if we waited for some loading, caused by above arranging, lets re-arrange
						v.ArrangeChildren(nil)
					}
					zlog.Info("SetTab Loaded re-arranged")
		*/
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
