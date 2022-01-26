//go:build zui
// +build zui

package zui

import (
	"fmt"
	"math"
	"path"
	"reflect"
	"strconv"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zwords"
)

type MenuedOwner struct {
	View            View
	PopPos          zgeo.Pos // either View or PopPos
	SelectedHandler func()
	GetTitle        func(itemCount int) string
	ActionHandler   func(id string)
	CreateItems     func() []MenuedOItem
	PluralableWord  string // if set, used instead of GetTitle, and pluralized
	Font            *zgeo.Font
	ImagePath       string
	IsStatic        bool // if set, user can't set a different value, but can press and see them. Shows number of items
	IsMultiple      bool
	HasLabelColor   bool
	SetTitle        bool
	StoreKey        string
	BGColor         zgeo.Color
	TextColor       zgeo.Color
	HighlightColor  zgeo.Color

	items []MenuedOItem
}

type MenuedOItem struct {
	Name        string
	Value       interface{}
	Selected    bool
	LabelColor  zgeo.Color
	TextColor   zgeo.Color
	IsDisabled  bool
	IsAction    bool
	IsSeparator bool
}

var (
	MenuedOItemSeparator              = MenuedOItem{IsSeparator: true}
	MenuedOwnerDefaultBGColor         = StyleColF(zgeo.ColorNew(0.92, 0.91, 0.90, 1), zgeo.ColorNew(0.12, 0.11, 0.1, 1))
	MenuedOwnerDefaultTextColor       = StyleGrayF(0.1, 0.9)
	MenuedOwnerDefaultHightlightColor = StyleColF(zgeo.ColorNewGray(0, 0.7), zgeo.ColorNewGray(1, 0.7))
)

func MenuedOwnerNew() *MenuedOwner {
	o := &MenuedOwner{}
	o.Font = zgeo.FontNice(zgeo.FontDefaultSize-1, zgeo.FontStyleNormal)
	o.HighlightColor = MenuedOwnerDefaultHightlightColor()
	o.BGColor = MenuedOwnerDefaultBGColor()
	o.TextColor = MenuedOwnerDefaultTextColor()
	return o
}

func MenuedAction(name string, val interface{}) MenuedOItem {
	var item MenuedOItem
	item.Name = name
	item.Value = val
	item.IsAction = true
	return item
}

func (o *MenuedOwner) PopInPos(pos zgeo.Pos, items []MenuedOItem) {
	o.Build(nil, items)
	o.PopPos = pos
	o.popup()
}

func (o *MenuedOwner) IsKeyStored() bool {
	if o.StoreKey == "" {
		return false
	}
	_, got := DefaultLocalKeyValueStore.GetDict(o.StoreKey)
	return got
}

func (o *MenuedOwner) Build(view View, items []MenuedOItem) {
	if view != nil {
		o.View = view
		nv := ViewGetNative(view)
		// zlog.Info("MO ADDStopper:", nv.Hierarchy(), zlog.GetCallingStackString())
		nv.AddOnRemoveFunc(o.Stop)
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
	if o.StoreKey != "" {
		dict, got := DefaultLocalKeyValueStore.GetDict(o.StoreKey)
		if got {
			for i, item := range items {
				str := fmt.Sprint(item.Value)
				_, items[i].Selected = dict[str]
			}
		} else {
			if o.CreateItems != nil {
				items = o.CreateItems()
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
	if o.SetTitle || o.GetTitle != nil {
		zlog.Assert(o.View != nil)
		ts, got := o.View.(TextSetter)
		if got {
			ts.SetText(text)
		}
	}
}

func (o *MenuedOwner) updateTitleAndImage() {
	var nstr string
	if o.IsMultiple {
		var count, total int
		for _, i := range o.items {
			if !i.IsAction && !i.IsSeparator {
				if i.Selected {
					// zlog.Info("Menued updateTitleAndImage add", i.Name)
					count++
				}
				total++
			}
		}
		if o.PluralableWord != "" {
			nstr = zwords.PluralizeWordAndCountWords(o.PluralableWord, float64(count), "", "", map[int]string{0: "no", total: "all"})
		} else if o.GetTitle != nil {
			nstr = o.GetTitle(count)
		} else {
			nstr = strconv.Itoa(count)
		}
		return
	}
	if o.GetTitle != nil {
		nstr = o.GetTitle(1)
		o.SetText(nstr)
		return
	}
	var spath, sval string
	item := o.SelectedItem()
	if item != nil && item.Value != nil {
		sval = fmt.Sprint(item.Value)
		if o.ImagePath != "" {
			spath = path.Join(o.ImagePath, sval+".png")
			// zlog.Info("Menued SetValImage:", str)
		}
		nstr = sval
	}
	o.SetText(nstr)
	if o.ImagePath != "" {
		// zlog.Info("MO SetImagePath:", spath)
		io := o.View.(ImageOwner)
		io.SetImage(nil, spath, nil)
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

func (o *MenuedOwner) SetSelectedValues(vals []interface{}) {
outer:
	for i, item := range o.getItems() {
		for j, v := range vals {
			if reflect.DeepEqual(item.Value, v) {
				o.items[i].Selected = true
				zslice.RemoveAt(&vals, j)
			}
			continue outer
		}
		o.items[i].Selected = false
	}
	o.updateTitleAndImage()
}

func (o *MenuedOwner) SetSelectedValue(val interface{}) {
	o.SetSelectedValues([]interface{}{val})
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
	if item.IsSeparator {
		canvas.SetColor(zgeo.ColorLightGray)
		canvas.StrokeHorizontal(0, rect.Size.W, math.Floor(rect.Center().Y), 1, zgeo.PathLineSquare)
		return
	}

	ti := TextInfoNew()
	if list.IsRowHighlighted(i) {
		ti.Color = list.HighlightColor.ContrastingGray()
	} else if item.TextColor.Valid {
		ti.Color = item.TextColor
	} else {
		ti.Color = o.TextColor
	}
	// zlog.Info("DrawMenuedRow:", i, item.Name, item.IsAction, o.Font.Style)
	ti.Font = &zgeo.Font{}
	*ti.Font = *o.Font
	if item.IsDisabled {
		ti.Color.SetOpacity(0.5)
	}
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
				canvas.DrawImage(img, true, drect, 1, zgeo.Rect{})
			}
		})
		x += imageWidth + imageMarg
	}

	if item.IsAction {
		ti.Font.Style = zgeo.FontStyleItalic
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
		col := item.LabelColor
		if item.IsDisabled {
			col.SetOpacity(0.5)
		}
		canvas.SetColor(col)
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
	list.MinRows = 10
	stack.SetBGColor(o.BGColor)
	list.PressSelectable = true
	list.PressUnselectable = o.IsMultiple
	list.MultiSelect = o.IsMultiple
	list.SelectedColor = zgeo.Color{}
	list.HighlightColor = o.HighlightColor
	list.HoverHighlight = true
	list.ExposeSetsRowBGColor = true
	list.GetRowColor = func(i int) zgeo.Color {
		return o.BGColor
	}
	stack.Add(list, zgeo.TopLeft|zgeo.Expand)

	lineHeight := o.Font.LineHeight() + 6
	list.GetRowHeight = func(index int) float64 {
		if o.items[index].IsSeparator {
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
	pos := o.PopPos
	if o.View != nil {
		nv := ViewGetNative(o.View)
		pos = nv.AbsoluteRect().Pos
	}
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
						o.getItems()
						o.updateTitleAndImage()
					}
					return
				}
			}
			if o.SelectedHandler != nil {
				o.SelectedHandler()
			}
			o.saveToStore()
			if o.IsStatic {
				for i := range o.items {
					o.items[i].Selected = false
				}
			}
			o.updateTitleAndImage()
		}
	})
}

func (o *MenuedOwner) saveToStore() {
	if o.StoreKey != "" {
		dict := zdict.Dict{}
		for _, item := range o.items {
			if item.Selected {
				str := fmt.Sprint(item.Value)
				dict[str] = true
			}
		}
		DefaultLocalKeyValueStore.SetDict(dict, o.StoreKey, true)
	}
}

func MenuOwningButtonCreate(menu *MenuedOwner, items []MenuedOItem) *ShapeView {
	v := ShapeViewNew(ShapeViewTypeRoundRect, zgeo.Size{20, 24})
	v.SetImage(nil, "images/zmenu-arrows.png", nil)
	v.ImageMargin = zgeo.Size{3, 3}
	v.ImageGap = 4
	v.SetColor(StyleDefaultFGColor().Mixed(zgeo.ColorGray, 0.2))
	v.StrokeColor = zgeo.ColorNewGray(0, 0.3)
	v.StrokeWidth = 1
	v.SetTextColor(StyleDefaultBGColor())
	menu.Build(v, items)
	return v
}
