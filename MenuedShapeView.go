// +build zui

package zui

import (
	"fmt"
	"reflect"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
)

type Item struct {
	zdict.Item
	Selected bool
}
type MenuedShapeView struct {
	ShapeView
	maxWidth        float64
	selectedHandler func()
	items           []Item

	ImagePath  string
	IsStatic   bool // if set, user can't set a different value, but can press and see them. Shows number of items
	IsMultiple bool
	GetTitle   func(itemCount int) string
}

func MenuedShapeViewNew(shapeType ShapeViewType, minSize zgeo.Size, name string, items zdict.Items, values []interface{}, isStatic, isMultiple bool) *MenuedShapeView {
	v := &MenuedShapeView{}
	if minSize.IsNull() {
		minSize.Set(20, 26)
	}
	v.ShapeView.init(shapeType, minSize, name)
	v.SetMargin(zgeo.RectFromXY2(10, 3, -3, -3))
	v.IsStatic = isStatic
	v.IsMultiple = isMultiple
	v.SetPressedHandler(func() {
		if len(v.items) != 0 {
			v.popup()
		}
	})

	v.UpdateItems(items, values)
	v.SetTextAlignment(zgeo.CenterLeft)
	v.SetFont(FontNice(14, FontStyleNormal))
	// zlog.Info("MenuedShapeViewNew:", name, v.Color())
	return v
}

func (v *MenuedShapeView) SetPillStyle() {
	// v.SetBGColor(zgeo.ColorClear)
	v.SetColor(zgeo.ColorLightGray)
	v.SetColor(zgeo.ColorRed)
	v.SetTextColor(zgeo.ColorBlack)
	v.SetTextAlignment(zgeo.Center)
	v.SetMargin(zgeo.RectFromXY2(0, 2, 0, -2))
	v.Ratio = 0.3
	v.ImageAlign = zgeo.CenterRight | zgeo.Proportional
	v.ImageMargin.Set(4, 4)
}

func (v *MenuedShapeView) SelectedItem() zdict.Item {
	sitems := v.SelectedItems()
	if len(sitems) == 0 {
		return zdict.Item{}
	}
	return sitems[0]
}

func (v *MenuedShapeView) SelectedItems() (sitems zdict.Items) {
	for _, item := range v.items {
		if item.Selected {
			sitems = append(sitems, item.Item)
		}
	}
	return
}

func (v *MenuedShapeView) Empty() {
	v.items = v.items[:0]
}

func (v *MenuedShapeView) AddSeparator() {
	var item Item

	item.Name = separatorID
	item.Value = nil
	v.items = append(v.items, item)
}

func (v *MenuedShapeView) updateTitle() {
	if v.GetTitle != nil {
		if v.IsStatic || v.IsMultiple {
			name := v.GetTitle(len(v.items))
			v.SetText(name)
		}
	} else if !v.IsMultiple && !v.IsStatic {
		item := v.SelectedItem()
		if item.Value != nil {
			v.SetText(fmt.Sprint(item.Value))
		}
	}
}

func (v *MenuedShapeView) UpdateItems(items zdict.Items, values []interface{}) {
	v.items = v.items[:0]
	for _, item := range items {
		var mi Item
		mi.Item = item
		for _, v := range values {
			if reflect.DeepEqual(item.Value, v) {
				mi.Selected = true
				break
			}
		}
		v.items = append(v.items, mi)
	}
	v.updateTitle()
}

func (v *MenuedShapeView) SetSelectedHandler(handler func()) {
	v.selectedHandler = handler
}

func (v *MenuedShapeView) ChangeNameForValue(name string, value interface{}) {
	if zlog.ErrorIf(value == nil, v.ObjectName()) {
		return
	}
	for i, item := range v.items {
		if reflect.DeepEqual(item.Value, value) {
			v.items[i].Name = name
			return
		}
	}
	zlog.Error(nil, "no value to change name for")
}

func (v *MenuedShapeView) MaxWidth() float64 {
	return v.maxWidth
}

func (v *MenuedShapeView) SetMaxWidth(max float64) View {
	v.maxWidth = max
	return v
}

const (
	topMarg    = 6
	bottomMarg = 6
	leftMarg   = 30
	rightMarg  = 4
)

func (v *MenuedShapeView) popup() {
	stack := StackViewVert("menued-pop-stack")
	stack.SetMargin(zgeo.RectFromXY2(0, topMarg, 0, -bottomMarg))
	list := ListViewNew("menu-list")
	list.MinRows = 1
	stack.SetBGColor(zgeo.ColorWhite)
	//	list.ScrollView.SetBGColor(zgeo.ColorClear)
	list.PressSelectable = true
	list.PressUnselectable = v.IsMultiple
	list.MultiSelect = v.IsMultiple
	list.SelectedColor = zgeo.ColorNew(0, 0.341, 0.847, 1)
	list.HighlightColor = list.SelectedColor
	list.HoverHighlight = true
	list.ExposeSetsRowBGColor = true
	list.RowColors = []zgeo.Color{zgeo.ColorWhite}
	stack.Add(zgeo.TopLeft|zgeo.Expand, list)
	// zlog.Info("POP:", v.Font().Size)
	var bs zgeo.Size
	lineHeight := v.Font().LineHeight() + 4
	list.GetRowHeight = func(index int) float64 {
		return lineHeight
	}
	list.GetRowCount = func() int {
		return len(v.items)
	}
	list.CreateRow = func(rowSize zgeo.Size, i int) View {
		cv := CustomViewNew("row")
		cv.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
			// if i%2 == 1 {
			// 	canvas.SetColor(zgeo.ColorRed, 1)
			// 	canvas.FillPath(zgeo.PathNewRect(rect, zgeo.Size{}))
			// }
			ti := TextInfoNew()
			if list.IsRowHighlighted(i) {
				ti.Color = zgeo.ColorWhite
			}
			ti.Text = v.items[i].Name
			ti.Font = v.Font()
			//			zlog.Info("Draw Menu row:", ti.Text, ti.Font.Size)
			ti.Alignment = zgeo.CenterLeft
			ti.Rect = rect.Plus(zgeo.RectFromXY2(leftMarg, 0, -rightMarg, 0))
			list.UpdateRowBGColor(i)
			ti.Draw(canvas)
		})
		return cv
	}
	win := v.GetWindow()
	win.AddKeypressHandler(stack.View, func(key KeyboardKey, mod KeyboardModifier) {
		if mod == KeyboardModifierNone && key == KeyboardKeyEscape {
			PresentViewClose(stack, false, nil)
		}
	})
	var max string
	for _, item := range v.items {
		if len(item.Name) > len(max) {
			max = item.Name
		}
	}
	bs.W = TextInfoWidthOfString(max, v.Font())
	bs.H = float64(zint.Min(22, len(v.items))) * lineHeight
	stack.SetMinSize(bs.Plus(zgeo.Size{leftMarg + rightMarg, topMarg + bottomMarg}))
	// stack.SetCorner(8)

	list.HandleRowSelected = func(i int, selected bool) {
		// zlog.Info("list selected", i)
		v.items[i].Selected = selected
		if !v.IsMultiple {
			PresentViewClose(stack, false, nil)
		}
	}
	att := PresentViewAttributesNew()
	att.Modal = true
	att.ModalDimBackground = false
	att.ModalCloseOnOutsidePress = true
	att.ModalDropShadow.Delta = zgeo.SizeBoth(1)
	att.ModalDropShadow.Blur = 2
	pos := v.GetAbsoluteRect().Pos
	att.Pos = &pos
	//	list.Focus(true)
	PresentView(stack, att, func(*Window) {
		//		pop.Element.Set("selectedIndex", 0)
	}, func(dismissed bool) {
		// zlog.Info("menu pop closed", dismissed, zlog.GetCallingStackString())
		if !dismissed {
			if v.selectedHandler != nil {
				v.selectedHandler()
			}
		}
	})
}
