//go:build zui

package zgroup

import (
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

type TabsView struct {
	GroupBase
	separatorForIDs      []string
	SeparatorLineInset   float64
	ButtonName           string //
	selectedImageBGColor zgeo.Color
	MaxImageSize         zgeo.Size
}

const tabSeparatorID = "tab-separator"

var (
	DefaultButtonName           = "gray-tab"
	DefaultTextColor            = zstyle.GrayF(0.1, 0.9)
	DefaultSelectedImageBGColor = zstyle.ColF(zgeo.ColorNew(0, 0, 1, 0.2), zgeo.ColorNew(0, 0, 9, 0.2))
)

func TabsViewNew(name string, buttons bool) *TabsView {
	v := &TabsView{}
	v.GroupBase.Init()
	v.StackView.Init(v, true, name)
	v.SetBGColor(zstyle.DefaultBGColor())
	v.SetSpacing(0) // note: for vertical stack v
	v.header = zcontainer.StackViewHor("header")
	if buttons {
		v.ButtonName = DefaultButtonName
		v.header.SetMargin(zgeo.RectFromXY2(2, 4, 0, 0))
	} else {
		v.MaxImageSize = zgeo.Size{60, 24}
		v.header.SetMargin(zgeo.RectFromXY2(8, 6, -8, -6))
	}
	v.header.SetSpacing(12) // note: for header
	v.Add(v.header, zgeo.Left|zgeo.Top|zgeo.HorExpand)
	v.selectedImageBGColor = DefaultSelectedImageBGColor()
	v.SetIndicatorSelectionFunc = v.setButtonOn
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
				path := zgeo.PathNewRect(r, zgeo.Size{})
				canvas.FillPath(path)
			}
		})
	}
	return v
}

func (v *TabsView) GetGroupBase() *GroupBase {
	return &v.GroupBase
}

func (v *TabsView) GetHeader() *zcontainer.StackView {
	return v.header
}

func (v *TabsView) AddSeparatorLine(thickness float64, color zgeo.Color, corner float64, forIDs []string) {
	cv := zcustom.NewView(tabSeparatorID)
	cv.SetMinSize(zgeo.Size{10, thickness})
	cv.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
		selectedView, _ := v.header.FindViewWithName(v.CurrentID, false)
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

// AddItem adds a new tab to the row of tabs.
// id is unique id that identifies it.
// title is what's written in the tab, if ButtonName != "".
// ipath is path to image, if ButtonName != "" shown on right, otherwise centered
// set makes it the current tab after adding
// align is how to align the content child view
// create is a function to create or delete the content child each time tab is set.
func (v *TabsView) AddItem(id, title, imagePath string, set bool, view zview.View, create func(id string, delete bool) zview.View) {
	// v.AddGroupItem(id, title, imagePath, set, view, create)
	var button *zshape.ShapeView
	minSize := zgeo.Size{20, 22}
	if v.ButtonName != "" {
		// zlog.Info("Add Tab button:", title, v.ButtonName)
		b := zshape.ImageButtonViewNew(title, v.ButtonName, minSize, zgeo.Size{11, 8})
		button = &b.ShapeView
		button.SetTextColor(DefaultTextColor())
		button.SetMarginS(zgeo.Size{10, 0})
		button.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
		view = b
	} else {
		button = zshape.NewView(zshape.TypeNone, minSize)
		button.MaxSize = v.MaxImageSize
		button.ImageMargin = zgeo.Size{}
		view = button
	}
	button.MaxSize.H = 26
	button.SetObjectName(id)
	if imagePath != "" {
		button.SetImage(nil, imagePath, nil)
	}
	button.SetPressedHandler(func() {
		go v.SelectItem(id, nil)
	})
	v.header.Add(view, zgeo.BottomLeft)
	v.AddGroupItem(id, view, create)
	if set {
		v.SelectItem(id, nil)
	}
}

func (v *TabsView) RemoveItem(id string) {
	v.RemoveGroupItem((id))
}

func (v *TabsView) setButtonOn(id string, selected bool) {
	view, _ := v.header.FindViewWithName(id, false)
	if view != nil {
		button, _ := view.(*zshape.ImageButtonView)
		if button != nil {
			str := DefaultButtonName
			if selected {
				str += "-selected"
			}
			button.SetImageName(str, zgeo.Size{11, 8})
		} else { // image only
			v.header.Expose()
		}
	}
}

func (v *TabsView) SetButtonAlignment(id string, a zgeo.Alignment) {
	cell, _ := v.header.FindCellWithName(id)
	cell.Alignment = a
}

func (v *TabsView) SelectItem(id string, done func()) {
	// zlog.Info("TabsSelect:", id, zlog.GetCallingStackString())
	v.SetGroupItem(id, done)
	hasSeparator := zstr.StringsContain(v.separatorForIDs, id)
	arrange := true // don't arrange on collapse, as it is done below, or on present, and causes problems if done now
	v.CollapseChildWithName(tabSeparatorID, !hasSeparator, arrange)
}
