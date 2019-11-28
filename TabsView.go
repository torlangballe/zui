package zgo

type TabsView struct {
	StackView
	header    *StackView
	childView View
	creators  map[string]func() View
	currentID string
}

func TabsViewNew(name string) *TabsView {
	v := &TabsView{}
	v.StackView.init(v, name)
	v.Vertical = true
	v.Spacing(0)
	v.Margin(RectFromXY2(0, 4, 0, 0))
	v.creators = map[string]func() View{}
	v.header = StackViewNew(false, AlignmentNone, "header")
	v.header.Spacing(0)

	v.Add(v.header, AlignmentLeft|AlignmentTop)
	return v
}

func (v *TabsView) AddTabFunc(id, title string, set bool, creator func() View) {
	if title == "" {
		title = id
	}
	button := ButtonNew(title, "grayTab", Size{20, 28}, Size{11, 13})
	button.ObjectName(id)
	button.MarginS(Size{10, 0})
	button.TextInfo.Color = ColorWhite
	button.TextInfo.Font = FontNice(FontDefaultSize, FontStyleNormal)
	v.creators[id] = creator
	button.PressedHandler(func() {
		v.SetTab(id)
	})
	v.header.Add(button, AlignmentLeft|AlignmentVertCenter)
	if set {
		v.SetTab(id)
	}
}

func (v *TabsView) AddTab(title, id string, set bool, view View) {
	v.AddTabFunc(title, id, set, func() View {
		return view
	})
}

func (v *TabsView) setButtonOn(id string, on bool) {
	view := v.header.FindViewWithName(id, false)
	if view != nil {
		button := (*view).(*Button)
		str := "grayTab"
		style := FontStyleNormal
		if on {
			str += "Dark"
			style = FontStyleBold
		}
		button.SetNamedColor(str, Size{11, 13})
		button.TextInfo.Font = FontNice(FontDefaultSize, style)
	}
}
func (v *TabsView) SetTab(id string) {
	if v.currentID != id {
		if v.currentID != "" {
			v.setButtonOn(v.currentID, false)
		}
		if v.childView != nil {
			v.RemoveChild(v.childView)
		}
		v.childView = v.creators[id]()
		v.Add(v.childView, AlignmentLeft|AlignmentTop|AlignmentExpand|AlignmentNonProp)
		v.currentID = id
		v.setButtonOn(id, true)
		o := v.View.(NativeViewOwner)
		if o != nil {
			if !o.GetNative().presented {
				return
			}
		}
		presentViewCallReady(v.childView)
		if v.presented { // don't do if not first set up yet
			v.header.ArrangeChildren(nil)
			v.ArrangeChildren(nil)
		}
		et, _ := v.childView.(ExposableType)
		if et != nil {
			et.Expose()
		}
	}
}
