// +build zui

package zui

import (
	"fmt"
	"math"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type MenuedOItem struct {
	// zdict.Item
	Name       string
	Value      interface{}
	Selected   bool
	LabelColor zgeo.Color
	TextColor  zgeo.Color
	IsAction   bool
}

func MenuedAction(name string, val interface{}) MenuedOItem {
	var item MenuedOItem
	item.Name = name
	item.Value = val
	item.IsAction = true
	return item
}

type MenuedOwner struct {
	View            View
	SelectedHandler func()
	GetTitle        func(itemCount int) string
	ActionHandler   func(id string)
	Font            *Font
	CreateItems     func() []MenuedOItem

	ImagePath      string
	IsStatic       bool // if set, user can't set a different value, but can press and see them. Shows number of items
	IsMultiple     bool
	HasLabelColor  bool
	SetTitle       bool
	StoreKey       string
	BGColor        zgeo.Color
	TextColor      zgeo.Color
	HighlightColor zgeo.Color

	items []MenuedOItem
}

//var MenuedOItemSeparator = MenuedOItem{Item: zdict.Item{Name: MenuSeparatorID}}
var (
	MenuedOItemSeparator              = MenuedOItem{Name: MenuSeparatorID}
	MenuedOwnerDefaultBGColor         = zgeo.ColorNew(0.92, 0.91, 0.90, 1)
	MenuedOwnerDefaultTextColor       = zgeo.ColorNewGray(0.1, 1)
	MenuedOwnerDefaultHightlightColor = zgeo.ColorNewGray(0, 0.7)
)

func MenuedOwnerNew() *MenuedOwner {
	o := &MenuedOwner{}
	o.Font = FontNice(FontDefaultSize-1, FontStyleNormal)
	o.HighlightColor = MenuedOwnerDefaultHightlightColor
	o.BGColor = MenuedOwnerDefaultBGColor
	o.TextColor = MenuedOwnerDefaultTextColor
	return o
}

func (o *MenuedOwner) Build(view View, items []MenuedOItem) {
	if view != nil {
		o.View = view
		nv := ViewGetNative(view)
		nv.AddStopper(o.Stop)
		presser := view.(Pressable)
		presser.SetPressedHandler(func() {
			if o.CreateItems != nil {
				o.items = o.CreateItems()
			}
			if len(o.items) != 0 {
				o.popup()
			}
		})
		presser.SetLongPressedHandler(func() {
			set := true
			if len(o.SelectedItems()) == len(o.items) {
				set = false
			}
			for i := range o.items {
				o.items[i].Selected = set
			}
			o.updateTitleAndImage()
		})
	}
	if len(items) == 0 {
		items = o.getItems()
	}
	if o.StoreKey != "" {
		dict, got := DefaultLocalKeyValueStore.GetDict(o.StoreKey)
		if got {
			for i, item := range items {
				str := fmt.Sprint(item.Value)
				_, items[i].Selected = dict[str]
			}
		}
	}
	o.UpdateMenuedItems(items)
}

func (o *MenuedOwner) Stop() {
	if o.View == nil {
		zlog.Info("Stopping MenuOwner view==nil:", zlog.GetCallingStackString())
	}
	// zlog.Info("Stopping MenuOwner:", o.View != nil)
	// zlog.Info("Stopping MenuOwner for:", o.View.ObjectName())
	presser := o.View.(Pressable)
	// zlog.Info("Stopping MenuOwner for2:", presser != nil)
	presser.SetPressedHandler(nil)
	presser.SetLongPressedHandler(nil)
	*o = MenuedOwner{} // this will
}

func (o *MenuedOwner) SelectedItem() *zdict.Item {
	sitems := o.SelectedItems()
	if len(sitems) == 0 {
		return nil
	}
	si := sitems[0]
	return &si
}

func (o *MenuedOwner) SelectedItems() (sitems zdict.Items) {
	for _, item := range o.items {
		if item.Selected {
			sitems = append(sitems, zdict.Item{Name: item.Name, Value: item.Value})
		}
	}
	return
}

func (o *MenuedOwner) Empty() {
	o.items = o.items[:0]
}

func (o *MenuedOwner) SetText(text string) {
	if o.SetTitle {
		ts, got := o.View.(TextSetter)
		if got {
			ts.SetText(text)
		}
	}
}

func (o *MenuedOwner) AddSeparator() {
	o.items = append(o.items, MenuedOItemSeparator)
}

func (o *MenuedOwner) updateTitleAndImage() {
	// zlog.Info("Menued updateTitleAndImage")
	var nstr string
	if o.IsStatic && o.GetTitle != nil {
		name := o.GetTitle(len(o.items))
		name = strings.Replace(name, `%d`, nstr, -1)
		o.SetText(name)
		return
	}
	if o.IsMultiple {
		var count int
		for _, i := range o.items {
			if i.Selected {
				count++
			}
		}
		nstr = strconv.Itoa(count)
		if count == len(o.items) {
			nstr = "all"
		}
		if o.GetTitle != nil {
			name := o.GetTitle(count)
			name = strings.Replace(name, `%d`, nstr, -1)
			o.SetText(name)
		} else if o.IsMultiple {
			o.SetText(nstr)
		}
	} else {
		var spath, sval string
		item := o.SelectedItem()
		if item != nil {
			sval = fmt.Sprint(item.Value)
			if o.ImagePath != "" {
				spath = path.Join(o.ImagePath, sval+".png")
				// zlog.Info("Menued SetValImage:", str)
			}
		}
		if o.ImagePath != "" {
			zlog.Info("MO SetImagePath:", spath)
			io := o.View.(ImageOwner)
			io.SetImage(nil, spath, nil)
		}
		if item != nil && item.Value != nil {
			o.SetText(sval)
		} else {
			o.SetText("")
		}
	}
}

func (o *MenuedOwner) UpdateItems(items zdict.Items, values []interface{}) {
	var mitems []MenuedOItem
	for _, item := range items {
		var m MenuedOItem
		m.Name = item.Name
		m.Value = item.Value
		for _, v := range values {
			if reflect.DeepEqual(item.Value, v) {
				m.Selected = true
				break
			}
			return
		}
		mitems = append(mitems, m)
	}
	o.UpdateMenuedItems(mitems)
}

func (o *MenuedOwner) UpdateMenuedItems(items []MenuedOItem) {
	o.items = items
	o.updateTitleAndImage()
}

func (o *MenuedOwner) SetSingleValue(val interface{}) {
	found := false
	for i, item := range o.getItems() {
		if !found && reflect.DeepEqual(item.Value, val) {
			o.items[i].Selected = true
			found = true
		} else {
			o.items[i].Selected = false
		}
	}
	o.updateTitleAndImage()
}

func (o *MenuedOwner) getItems() []MenuedOItem {
	if o.CreateItems != nil && len(o.items) == 0 {
		o.items = o.CreateItems()
	}
	return o.items
}

func (o *MenuedOwner) ChangeNameForValue(name string, value interface{}) {
	if zlog.ErrorIf(value == nil, o.View.ObjectName()) {
		return
	}
	for i, item := range o.items {
		if reflect.DeepEqual(item.Value, value) {
			o.items[i].Name = name
			return
		}
	}
	zlog.Error(nil, "no value to change name for")
}

func (o *MenuedOwner) rowDraw(list *ListView, i int, rect zgeo.Rect, canvas *Canvas, allAction bool, imageWidth, imageMarg, rightMarg float64) {
	list.UpdateRowBGColor(i)
	item := o.items[i]
	x := 8.0
	// zlog.Info("DrawMenuedRow:", i, rect, list.HighlightColor)
	if item.Name == MenuSeparatorID {
		canvas.SetColor(zgeo.ColorLightGray)
		canvas.StrokeHorizontal(0, rect.Size.W, math.Floor(rect.Center().Y), 1, zgeo.PathLineSquare)
		return
	}

	ti := TextInfoNew()
	if list.IsRowHighlighted(i) {
		ti.Color = list.HighlightColor.GetContrastingGray()
	} else if item.TextColor.Valid {
		ti.Color = item.TextColor
	} else {
		ti.Color = o.TextColor
	}
	ti.Font = o.Font
	if !item.IsAction {
		if list.IsRowSelected(i) {
			ti.Text = "√"
			ti.Rect = rect.Plus(zgeo.RectFromXY2(8, 0, 0, 0))
			ti.Alignment = zgeo.Left
			ti.Draw(canvas) // we keep black/white hightlighted color
		}
		x += 14
	}

	if o.ImagePath != "" {
		canvas.DownsampleImages = true
		sval := fmt.Sprint(item.Value)
		spath := path.Join(o.ImagePath, sval+".png")
		imageX := x // we need to store this since, ImageFromPath resulting closure uses it later
		ImageFromPath(spath, func(img *Image) {
			if img != nil {
				inRect := zgeo.RectFromXYWH(imageX, 0, imageWidth, rect.Size.H)
				drect := inRect.Align(img.Size(), zgeo.CenterLeft|zgeo.Shrink|zgeo.Proportional, zgeo.Size{2, 4})
				// canvas.SetColor(zgeo.ColorGreen)
				// canvas.FillRect(inRect)
				// zlog.Info("DrawMenuedImage:", item.Name, drect, allAction, x)
				canvas.DrawImage(img, false, true, drect, 1, zgeo.Rect{})
			}
		})
		x += imageWidth + imageMarg
	}

	if item.IsAction {
		ti.Font.Style = FontStyleItalic
	}
	//			zlog.Info("Draw Menu row:", ti.Text, ti.Font.Size)
	ti.Text = item.Name
	ti.Alignment = zgeo.CenterLeft
	ti.Rect = rect.Plus(zgeo.RectFromXY2(x, 0, -rightMarg, 0))
	ti.Draw(canvas)

	if o.HasLabelColor && item.LabelColor.Valid {
		r := rect
		r.SetMinX(rect.Max().X - rightMarg + 6)
		r = r.Expanded(zgeo.Size{-3, -3})
		canvas.SetColor(item.LabelColor)
		path := zgeo.PathNewRect(r, zgeo.Size{2, 2})
		canvas.FillPath(path)
		canvas.SetColor(zgeo.ColorBlack.WithOpacity(item.LabelColor.Opacity()))
		canvas.StrokePath(path, 1, zgeo.PathLineRound)
	}
}

func (o *MenuedOwner) popup() {
	const (
		imageWidth = 40
		imageMarg  = 8
		checkWidth = 14
		topMarg    = 6
		bottomMarg = 6
		rightMarg  = 4
	)

	var selection = map[int]bool{}
	allAction := true
	for i, item := range o.items {
		if item.Selected {
			selection[i] = true
		}
		if !item.IsAction {
			allAction = false
		}
	}
	stack := StackViewVert("menued-pop-stack")
	stack.SetMargin(zgeo.RectFromXY2(0, topMarg, 0, -bottomMarg))
	list := ListViewNew("menu-list", selection)
	list.MinRows = 0
	stack.SetBGColor(o.BGColor)
	list.PressSelectable = true
	list.PressUnselectable = o.IsMultiple
	list.MultiSelect = o.IsMultiple
	list.SelectedColor = zgeo.Color{}
	list.HighlightColor = o.HighlightColor
	list.HoverHighlight = true
	list.ExposeSetsRowBGColor = true
	list.RowColors = []zgeo.Color{o.BGColor}
	stack.Add(list, zgeo.TopLeft|zgeo.Expand)

	lineHeight := o.Font.LineHeight() + 6
	list.GetRowHeight = func(index int) float64 {
		if o.items[index].Name == MenuSeparatorID {
			return lineHeight * 0.5
		}
		return lineHeight
	}
	list.GetRowCount = func() int {
		return len(o.items)
	}
	rm := float64(rightMarg)
	if o.HasLabelColor {
		rm += 24
	}
	list.CreateRow = func(rowSize zgeo.Size, i int) View {
		cv := CustomViewNew("row")
		cv.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
			o.rowDraw(list, i, rect, canvas, allAction, imageWidth, imageMarg, rm)
		})
		return cv
	}

	var max string
	for _, item := range o.items {
		if len(item.Name) > len(max) {
			max = item.Name
		}
	}
	w := TextInfoWidthOfString(max, o.Font) * 1.1
	if !allAction && !o.IsStatic {
		w += checkWidth
	}
	if o.ImagePath != "" {
		w += imageWidth + imageMarg
	}
	w += 18 // test
	stack.SetMinSize(zgeo.Size{w, 0})

	list.HandleRowSelected = func(i int, selected, fromPressed bool) {
		// zlog.Info("list selected", i, selected, o.items[i].IsAction)
		o.items[i].Selected = selected
		o.updateTitleAndImage()
		if o.items[i].IsAction {
			if selected {
				PresentViewClose(stack, false, nil)
			}
			return
		}
		if !o.IsMultiple && fromPressed {
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
	stack.SetStroke(1, zgeo.ColorNewGray(0.5, 1))
	nv := ViewGetNative(o.View)
	pos := nv.AbsoluteRect().Pos
	att.Pos = &pos
	// zlog.Info("menu popup")
	//	list.Focus(true)
	PresentView(stack, att, func(*Window) {
		//		pop.Element.Set("selectedIndex", 0)
	}, func(dismissed bool) {
		// zlog.Info("menued closed", dismissed, o.IsMultiple)
		if !dismissed || o.IsMultiple { // if multiple, we handle any select/deselect done
			for i, item := range o.items {
				if item.IsAction && item.Selected {
					o.items[i].Selected = false
					if o.ActionHandler != nil {
						id := o.items[i].Value.(string)
						o.ActionHandler(id)
						o.updateTitleAndImage()
					}
					return
				}
			}
			if o.SelectedHandler != nil {
				o.SelectedHandler()
			}
			if o.IsStatic {
				for i := range o.items {
					o.items[i].Selected = false
				}
			}
			o.updateTitleAndImage()
		}
	})
}
