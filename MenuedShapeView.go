// +build zui

package zui

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
)

type MenuedItem struct {
	zdict.Item
	Selected   bool
	LabelColor zgeo.Color
	TextColor  zgeo.Color
}

type MenuedShapeView struct {
	ShapeView
	maxWidth        float64
	selectedHandler func()
	items           []MenuedItem

	ImagePath     string
	IsStatic      bool // if set, user can't set a different value, but can press and see them. Shows number of items
	IsMultiple    bool
	HasLabelColor bool
	GetTitle      func(itemCount int) string
}

func MenuedShapeViewNew(shapeType ShapeViewType, minSize zgeo.Size, name string, items []MenuedItem, isStatic, isMultiple bool) *MenuedShapeView {
	v := &MenuedShapeView{}
	if minSize.IsNull() {
		minSize.Set(20, 26)
	}
	v.ShapeView.init(shapeType, minSize, name)
	v.IsStatic = isStatic
	v.IsMultiple = isMultiple
	v.ImageMargin = zgeo.Size{}
	v.SetPressedHandler(func() {
		if len(v.items) != 0 {
			v.popup()
		}
	})
	v.SetLongPressedHandler(func() {
		set := true
		if len(v.SelectedItems()) == len(v.items) {
			set = false
		}
		for i := range v.items {
			v.items[i].Selected = set
		}
		v.updateTitle()
	})

	v.UpdateMenuedItems(items)
	v.SetTextAlignment(zgeo.CenterLeft)
	v.SetFont(FontNice(14, FontStyleNormal))
	// zlog.Info("MenuedShapeViewNew:", name, v.Color())
	return v
}

func (v *MenuedShapeView) SetPillStyle() {
	// v.SetBGColor(zgeo.ColorClear)
	v.SetColor(zgeo.ColorLightGray)
	v.SetTextColor(zgeo.ColorBlack)
	v.SetTextAlignment(zgeo.Center)
	v.SetMargin(zgeo.RectFromXY2(0, 2, 0, -2))
	v.Ratio = 0.2
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
	var item MenuedItem

	item.Name = separatorID
	item.Value = nil
	v.items = append(v.items, item)
}

func (v *MenuedShapeView) updateTitle() {
	var nstr string
	if v.IsStatic && v.GetTitle != nil {
		name := v.GetTitle(len(v.items))
		name = strings.Replace(name, `%d`, nstr, -1)
		v.SetText(name)
		return
	}
	if v.IsMultiple {
		var count int
		for _, i := range v.items {
			if i.Selected {
				count++
			}
		}
		nstr = strconv.Itoa(count)
		if count == len(v.items) {
			nstr = "all"
		}
		if v.GetTitle != nil {
			name := v.GetTitle(count)
			name = strings.Replace(name, `%d`, nstr, -1)
			v.SetText(name)
		} else if v.IsMultiple {
			v.SetText(nstr)
		}
	} else {
		item := v.SelectedItem()
		if item.Value != nil {
			v.SetText(fmt.Sprint(item.Value))
		} else {
			v.SetText("")
		}
	}
}

func (v *MenuedShapeView) UpdateItems(items zdict.Items, values []interface{}) {
	var mitems []MenuedItem
	for _, item := range items {
		var m MenuedItem
		m.Item = item
		for _, v := range values {
			if reflect.DeepEqual(item.Value, v) {
				m.Selected = true
				break
			}
			return
		}
		mitems = append(mitems, m)
	}
	v.UpdateMenuedItems(mitems)
}

func (v *MenuedShapeView) UpdateMenuedItems(items []MenuedItem) {
	v.items = items
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

func (v *MenuedShapeView) SetMaxWidth(max float64) {
	v.maxWidth = max
}

const (
	topMarg    = 6
	bottomMarg = 6
	leftMarg   = 30
	rightMarg  = 4
)

func (v *MenuedShapeView) popup() {
	var selection = map[int]bool{}
	for i, item := range v.items {
		if item.Selected {
			selection[i] = true
		}
	}

	stack := StackViewVert("menued-pop-stack")
	stack.SetMargin(zgeo.RectFromXY2(0, topMarg, 0, -bottomMarg))
	list := ListViewNew("menu-list", selection)
	list.MinRows = 1
	stack.SetBGColor(zgeo.ColorWhite)
	//	list.ScrollView.SetBGColor(zgeo.ColorClear)
	list.PressSelectable = true
	list.PressUnselectable = v.IsMultiple
	list.MultiSelect = v.IsMultiple
	list.SelectedColor = zgeo.ColorWhite
	list.HighlightColor = zgeo.ColorNew(0, 0.341, 0.847, 1)
	list.HoverHighlight = true
	list.ExposeSetsRowBGColor = true
	list.RowColors = []zgeo.Color{zgeo.ColorWhite}
	stack.Add(list, zgeo.TopLeft|zgeo.Expand)
	// zlog.Info("POP:", v.Font().Size)
	var bs zgeo.Size
	lineHeight := v.Font().LineHeight() + 4
	list.GetRowHeight = func(index int) float64 {
		return lineHeight
	}
	list.GetRowCount = func() int {
		return len(v.items)
	}
	rm := float64(rightMarg)
	if v.HasLabelColor {
		rm += 24
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
			} else if v.items[i].TextColor.Valid {
				ti.Color = v.items[i].TextColor
			}
			ti.Text = v.items[i].Name
			ti.Font = v.Font()
			//			zlog.Info("Draw Menu row:", ti.Text, ti.Font.Size)
			ti.Alignment = zgeo.CenterLeft
			ti.Rect = rect.Plus(zgeo.RectFromXY2(leftMarg, 0, -rm, 0))
			list.UpdateRowBGColor(i)
			ti.Draw(canvas)
			if list.IsRowSelected(i) {
				ti.Text = "√"
				ti.Rect = rect.Plus(zgeo.RectFromXY2(8, 0, 0, 0))
				ti.Alignment = zgeo.Left
				ti.Draw(canvas) // we keep black/white hightlighted color
			}
			if v.HasLabelColor && v.items[i].LabelColor.Valid {
				r := rect
				r.SetMinX(rect.Max().X - rm + 6)
				r = r.Expanded(zgeo.Size{-3, -3})
				canvas.SetColor(v.items[i].LabelColor)
				path := zgeo.PathNewRect(r, zgeo.Size{2, 2})
				canvas.FillPath(path)
				canvas.SetColor(zgeo.ColorBlack.WithOpacity(v.items[i].LabelColor.Opacity()))
				canvas.StrokePath(path, 1, zgeo.PathLineRound)
			}
		})
		return cv
	}
	var max string
	for _, item := range v.items {
		if len(item.Name) > len(max) {
			max = item.Name
		}
	}
	bs.W = TextInfoWidthOfString(max, v.Font())
	bs.H = float64(zint.Min(22, len(v.items))) * lineHeight
	stack.SetMinSize(bs.Plus(zgeo.Size{leftMarg + rm, topMarg + bottomMarg}))
	// stack.SetCorner(8)

	list.HandleRowSelected = func(i int, selected bool) {
		// zlog.Info("list selected", i, selected)
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
	att.ModalDismissOnEscapeKey = true
	pos := v.GetAbsoluteRect().Pos
	att.Pos = &pos
	//	list.Focus(true)
	PresentView(stack, att, func(*Window) {
		//		pop.Element.Set("selectedIndex", 0)
	}, func(dismissed bool) {
		// zlog.Info("menu pop closed", dismissed)
		if !dismissed || v.IsMultiple { // if multiple, we handle any select/deselect done
			if v.selectedHandler != nil {
				v.selectedHandler()
			}
			v.updateTitle()
		}
	})
}
