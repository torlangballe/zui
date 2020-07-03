package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type TabsView struct {
	StackView
	Header         *StackView
	ChildView      View
	creators       map[string]func(bool) View
	CurrentID      string
	childAlignmens map[string]zgeo.Alignment
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
	v.childAlignmens = map[string]zgeo.Alignment{}

	v.Add(zgeo.Left|zgeo.Top|zgeo.HorExpand, v.Header)
	return v
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
	v.childAlignmens[id] = align
	button := ButtonNew(title, TabsDefaultButtonName, zgeo.Size{20, 24}, zgeo.Size{11, 13})
	button.SetObjectName(id)
	button.SetMarginS(zgeo.Size{10, 0})
	button.SetColor(TabsDefaultTextColor)
	button.SetFont(FontNice(FontDefaultSize, FontStyleNormal))
	v.creators[id] = creator
	button.SetPressedHandler(func() {
		v.SetTab(id)
	})
	v.Header.Add(zgeo.Left|zgeo.Bottom, button)
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
		button.SetImageName(str, zgeo.Size{11, 13})
		button.SetFont(FontNice(FontDefaultSize, style))
	}
}
func (v *TabsView) SetTab(id string) {
	// zlog.Info("Set Tab", id)
	if v.CurrentID != id {
		if v.CurrentID != "" {
			v.creators[v.CurrentID](true)
			v.setButtonOn(v.CurrentID, false)
		}
		if v.ChildView != nil {
			v.RemoveChild(v.ChildView)
		}
		v.ChildView = v.creators[id](false)
		v.Add(v.childAlignmens[id], v.ChildView)
		v.CurrentID = id
		v.setButtonOn(id, true)
		if !v.presented {
			// zlog.Info("Set Tab, exit because not presented yet", id)
			return
		}
		ct := v.ChildView.(ContainerType)
		//		et, _ := v.ChildView.(ExposableType)
		et, _ := v.View.(ExposableType)
		// if !v.presented {
		// 	return
		// }
		presentViewCallReady(v.ChildView)
		presentViewPresenting = true
		v.ArrangeChildren(nil)
		WhenContainerLoaded(ct, func(waited bool) {
			// zlog.Info("Set Tab container loaded:", waited)
			if waited { // if we waited for some loading, lets re-arrange
				v.ArrangeChildren(nil)
			}
			presentViewPresenting = false
			if et != nil {
				et.drawIfExposed()
			}
		})
	}
	// zlog.Info("Set Tab Done", id)
}
