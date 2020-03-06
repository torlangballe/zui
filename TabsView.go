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
	v.SetSpacing(0)
	v.SetMargin(zgeo.RectFromXY2(0, 4, 0, 0))
	v.creators = map[string]func(bool) View{}
	v.Header = StackViewHor("header")
	v.Header.SetSpacing(0)
	v.childAlignmens = map[string]zgeo.Alignment{}

	v.Add(zgeo.Left|zgeo.Top|zgeo.HorExpand, v.Header)
	return v
}

func (v *TabsView) AddTabFunc(id, title string, set bool, align zgeo.Alignment, creator func(del bool) View) {
	if title == "" {
		title = id
	}
	if align == zgeo.AlignmentNone {
		align = zgeo.Left | zgeo.Top | zgeo.Expand
	}
	v.childAlignmens[id] = align
	button := ButtonNew(title, "grayTab", zgeo.Size{20, 28}, zgeo.Size{11, 13})
	button.SetObjectName(id)
	button.SetMarginS(zgeo.Size{10, 0})
	button.TextInfo.Color = zgeo.ColorWhite
	button.TextInfo.Font = FontNice(FontDefaultSize, FontStyleNormal)
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
	// fmt.Println("setButtonOn:", id, on, view != nil)
	if view != nil {
		button := (*view).(*Button)
		str := "grayTab"
		style := FontStyleNormal
		if on {
			str += "Dark"
			style = FontStyleBold
		}
		button.SetImageName(str, zgeo.Size{11, 13})
		button.TextInfo.Font = FontNice(FontDefaultSize, style)
	}
}
func (v *TabsView) SetTab(id string) {
	// fmt.Println("Set Tab", id)
	if v.CurrentID != id {
		if v.CurrentID != "" {
			v.creators[id](true)
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
			// fmt.Println("Set Tab, exit because not presented yet", id)
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
			// fmt.Println("Set Tab container loaded:", waited)
			if waited { // if we waited for some loading, lets re-arrange
				v.ArrangeChildren(nil)
			}
			presentViewPresenting = false
			if et != nil {
				et.drawIfExposed()
			}
		})
	}
	// fmt.Println("Set Tab Done", id)
}
