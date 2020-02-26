package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

type TabsView struct {
	StackView
	header    *StackView
	ChildView View
	creators  map[string]func(bool) View
	CurrentID string
}

func TabsViewNew(name string) *TabsView {
	v := &TabsView{}
	v.StackView.init(v, name)
	v.Vertical = true
	v.SetSpacing(0)
	v.SetMargin(zgeo.RectFromXY2(0, 4, 0, 0))
	v.creators = map[string]func(bool) View{}
	v.header = StackViewHor("header")
	v.header.SetSpacing(0)

	v.Add(zgeo.Left|zgeo.Top, v.header)
	return v
}

func (v *TabsView) AddTabFunc(id, title string, set bool, creator func(del bool) View) {
	if title == "" {
		title = id
	}
	button := ButtonNew(title, "grayTab", zgeo.Size{20, 28}, zgeo.Size{11, 13})
	button.SetObjectName(id)
	button.SetMarginS(zgeo.Size{10, 0})
	button.TextInfo.Color = zgeo.ColorWhite
	button.TextInfo.Font = FontNice(FontDefaultSize, FontStyleNormal)
	v.creators[id] = creator
	button.SetPressedHandler(func() {
		v.SetTab(id)
	})
	v.header.Add(zgeo.Left|zgeo.VertCenter, button)
	if set {
		v.SetTab(id)
	}
}

func (v *TabsView) AddTab(title, id string, set bool, view View) {
	v.AddTabFunc(title, id, set, func(del bool) View {
		if del {
			return nil
		}
		return view
	})
}

func (v *TabsView) setButtonOn(id string, on bool) {
	view := v.header.FindViewWithName(id, false)
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
	if v.CurrentID != id {
		if v.CurrentID != "" {
			v.creators[id](true)
			v.setButtonOn(v.CurrentID, false)
		}
		if v.ChildView != nil {
			v.RemoveChild(v.ChildView)
		}
		v.ChildView = v.creators[id](false)
		v.Add(zgeo.Left|zgeo.Top|zgeo.Expand, v.ChildView)
		v.CurrentID = id
		v.setButtonOn(id, true)
		o := v.View.(NativeViewOwner)
		if o != nil {
			if !o.GetNative().presented {
				return
			}
		}
		presentViewCallReady(v.ChildView)
		if v.presented { // don't do if not first set up yet
			v.header.ArrangeChildren(nil)
			v.ArrangeChildren(&v.ChildView)
		}
		et, _ := v.ChildView.(ExposableType)
		if et != nil {
			et.Expose()
		}
	}
}
