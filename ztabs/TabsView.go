//go:build zui

package ztabs

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zshortcuts"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

type item struct {
	id     string
	view   zview.View
	create func(id string, delete bool) zview.View
}

type TabsView struct {
	zcontainer.StackView
	separatorForIDs       []string
	SeparatorLineInset    float64
	ButtonName            string //
	selectedImageBGColor  zgeo.Color
	MaxImageSize          zgeo.Size
	InvertSelectedTabText bool
	CurrentID             string
	DefaultID             string

	storeKey           string
	items              []item
	currentChild       zview.View
	header             *zcontainer.StackView
	ChangedHandlerFunc func(newID string)
	Dark               bool
}

const (
	tabSeparatorID = "tab-separator"
	storeKeyPrefix = "zui.TabsView.CurrentID."
)

var (
	DefaultButtonName           = "gray-tab"
	DefaultTextColor            = zstyle.GrayF(0.1, 0.9)
	DefaultSelectedImageBGColor = zstyle.ColF(zgeo.ColorNew(0, 0, 1, 0.2), zgeo.ColorNew(0, 0, 9, 0.2))
)

func TabsViewNew(storeName string, buttons bool) *TabsView {
	v := &TabsView{}
	v.StackView.Init(v, true, storeName)
	v.SetBGColor(zstyle.DefaultBGColor())
	v.SetSpacing(0) // note: for vertical stack v
	v.Dark = zstyle.Dark
	v.header = zcontainer.StackViewHor("header")
	if buttons {
		v.ButtonName = DefaultButtonName
		v.header.SetMargin(zgeo.RectFromXY2(2, 4, 0, 0))
	} else {
		v.MaxImageSize = zgeo.SizeD(60, 24)
		v.header.SetMargin(zgeo.RectFromXY2(8, 6, -8, -6))
	}
	v.storeKey = storeName
	if storeName != "" {
		str, got := zkeyvalue.DefaultStore.GetString(storeKeyPrefix + storeName)
		if got {
			v.CurrentID = str
		}
	}
	v.header.SetSpacing(12) // note: for header
	v.Add(v.header, zgeo.Left|zgeo.Top|zgeo.HorExpand)
	v.selectedImageBGColor = DefaultSelectedImageBGColor()
	// v.SetIndicatorSelectionFunc = v.setButtonOn
	if !buttons {
		v.header.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
			sv, i := v.header.FindViewWithName(v.CurrentID, false)
			if sv != nil {
				r := sv.Rect()
				r.Pos.Y = 0
				r.Size.H = rect.Size.H
				r.SetMinX(r.Min().X - v.header.Spacing()/2)
				r.SetMaxX(r.Max().X + v.header.Spacing()/2)
				if i == 0 {
					r.SetMinX(rect.Min().X)
				}
				if i == v.header.CountChildren() && r.Max().X > rect.Max().X-8 {
					r.SetMaxX(rect.Max().X)
				}
				canvas.SetColor(v.selectedImageBGColor)
				path := zgeo.PathNewRect(r, zgeo.SizeNull)
				canvas.FillPath(path)
			}
		})
	}
	v.SetKeyHandler(func(km zkeyboard.KeyMod, down bool) bool {
		if !down {
			return false
		}
		return zshortcuts.HandleOutsideShortcutRecursively(v, km)
	})
	return v
}

func (v *TabsView) ReadyToShow(beforeWindow bool) {
	if !beforeWindow {
		return
	}
	defID := v.DefaultID
	if defID == "" && len(v.items) > 0 {
		defID = v.items[0].id
	}
	if v.CurrentID == "" && defID != "" {
		v.SelectItem(defID, nil)
	}
	for _, item := range v.items {
		v.setButtonOn(item.id, item.id == v.CurrentID)
	}
}

// AddItem adds a new tab to the row of tabs.
// id is unique id that identifies it.
// title is what's written in the tab, if ButtonName != "".
// ipath is path to image, if ButtonName != "" shown on right, otherwise centered
// set makes it the current tab after adding
// view/create are either the view to show for this tab, or how to make/delete it dynamically.
// It is added/removed from view hierarchy by this method.
// create is a function to create or delete the content child each time tab is set.
func (v *TabsView) AddItem(id, title, imagePath string, view zview.View, create func(id string, delete bool) zview.View) {
	// v.AddGroupItem(id, title, imagePath, set, view, create)
	var button *zshape.ShapeView
	minSize := zgeo.SizeD(20, 22)
	if title == "" {
		title = id
	}
	if v.ButtonName != "" {
		// zlog.Info("Add Tab button:", title, v.ButtonName)
		b := zshape.ImageButtonViewNew(title, v.ButtonName, minSize, zgeo.SizeD(11, 8))
		button = &b.ShapeView
		button.SetTextColor(DefaultTextColor())
		button.SetMarginS(zgeo.SizeD(10, 0))
		button.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
		view = b
	} else {
		button = zshape.NewView(zshape.TypeNone, minSize)
		button.MaxSize = v.MaxImageSize
		button.ImageMargin = zgeo.RectNull
		view = button
	}
	button.MaxSize.H = 26
	button.SetObjectName(id)
	if imagePath != "" {
		button.SetImage(nil, true, zgeo.SizeNull, imagePath, zgeo.SizeNull, nil)
	}
	button.SetPressedHandler("", zkeyboard.ModifierNone, func() {
		go v.SelectItem(id, nil)
	})
	v.header.Add(view, zgeo.BottomLeft)
	v.items = append(v.items, item{id: id, view: view, create: create})
	ilen := len(v.items)
	if ilen < 10 {
		key := zkeyboard.Key('0' + ilen)
		button.KeyboardShortcut = zkeyboard.KMod(key, 0)
	}
	if v.CurrentID == id {
		v.SelectItem(id, nil)
	}
}

func (v *TabsView) SelectItem(id string, done func()) {
	if v.CurrentID == id && v.findItem(id) != -1 && v.currentChild != nil {
		if done != nil {
			done()
		}
		return
	}
	i := v.findItem(v.CurrentID)
	if i != -1 {
		item := v.items[i]
		item.create(v.CurrentID, true)
	}
	if v.currentChild != nil {
		v.RemoveChild(v.currentChild, true)
		v.currentChild = nil
	}
	if v.CurrentID != "" {
		v.setButtonOn(v.CurrentID, false)
	}
	v.CurrentID = id
	if v.storeKey != "" {
		zkeyvalue.DefaultStore.SetString(v.CurrentID, storeKeyPrefix+v.storeKey, true)
	}
	v.setButtonOn(v.CurrentID, true)
	item := v.items[v.findItem(id)]
	v.currentChild = item.view
	if item.create != nil {
		v.currentChild = item.create(id, false)
	}
	//	v.Add(v.currentChild, zgeo.TopLeft|zgeo.Expand)
	marg := zgeo.Size{}
	if v.currentChild.Native().CanTabFocus() {
		marg = zgeo.SizeBoth(3)
	}
	v.Add(v.currentChild, zgeo.Center|zgeo.Expand, marg)
	hasSeparator := zstr.StringsContain(v.separatorForIDs, id)
	arrange := true // don't arrange on collapse, as it is done below, or on present, and causes problems if done now
	v.CollapseChildWithName(tabSeparatorID, !hasSeparator, arrange)
	if v.IsPresented() {
		zcontainer.ArrangeAncestorContainer(v)
	}
	if v.ChangedHandlerFunc != nil {
		v.ChangedHandlerFunc(id)
	}
	if done != nil {
		done()
	}
	ztimer.StartIn(0.02, func() {
		v.SetRootYContentOffset(0) // in case old tab's view caused scroll offset, set back to 0
	})
}

func (v *TabsView) RemoveItem(id string) {
	i := v.findItem(id)
	zslice.RemoveAt(&v.items, i)
	item := v.items[i]
	if item.create != nil {
		item.create(id, true)
	}
	if v.currentChild != nil {
		v.RemoveChild(v.currentChild, true)
		v.currentChild = nil
	}
}

func (v *TabsView) GetHeader() *zcontainer.StackView {
	return v.header
}

func (v *TabsView) AddSeparatorLine(thickness float64, color zgeo.Color, corner float64, forIDs []string) {
	cv := zcustom.NewView(tabSeparatorID)
	cv.SetMinSize(zgeo.SizeD(10, thickness))
	cv.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
		selectedView, _ := v.header.FindViewWithName(v.CurrentID, false)
		canvas.SetColor(color)
		if selectedView != nil {
			r := selectedView.Rect()
			x0 := r.Pos.X + v.SeparatorLineInset
			x1 := r.Max().X - v.SeparatorLineInset
			r = rect
			r.SetMaxX(x0)
			path := zgeo.PathNewRect(r, zgeo.SizeNull)
			canvas.FillPath(path)
			r = rect
			r.SetMinX(x1)
			path = zgeo.PathNewRect(r, zgeo.SizeNull)
			canvas.FillPath(path)
		} else {
			path := zgeo.PathNewRect(rect, zgeo.SizeNull)
			canvas.FillPath(path)
		}
	})
	v.Add(cv, zgeo.TopLeft|zgeo.HorExpand)
	v.separatorForIDs = forIDs
}

func (v *TabsView) findItem(id string) int {
	for i := range v.items {
		if v.items[i].id == id {
			return i
		}
	}
	return -1
}

func (v *TabsView) setButtonOn(id string, selected bool) {
	view, _ := v.header.FindViewWithName(id, false)
	if view != nil {
		button, _ := view.(*zshape.ImageButtonView)
		if button != nil {
			str := DefaultButtonName
			if selected != v.Dark {
				str += "-selected"
			}
			button.SetImageName(str, zgeo.SizeD(11, 8))
			if v.InvertSelectedTabText {
				col := DefaultTextColor()
				if selected != v.Dark {
					col = col.ContrastingGray()
				}
				button.SetTextColor(col)
			}
		} else { // image only
			v.header.Expose()
		}
	}
}

func (v *TabsView) SetButtonAlignment(id string, a zgeo.Alignment) {
	cell, _ := v.header.FindCellWithName(id)
	cell.Alignment = a
}
