//go:build zui

package zmenu

import (
	"fmt"
	"math/rand"
	"path"
	"reflect"
	"strconv"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zwords"
)

type MenuedOwner struct {
	View                zview.View
	PopPos              zgeo.Pos // either View or PopPos
	SelectedHandlerFunc func()
	GetTitleFunc        func(itemCount int) string
	ActionHandlerFunc   func(id string)
	CreateItemsFunc     func() []MenuedOItem
	ClosedFunc          func()
	PluralableWord      string // if set, used instead of GetTitle, and pluralized
	TitleIsValueIfOne   bool   // if set and IsMultiple, name of value used as title if only one set
	Font                *zgeo.Font
	ImagePath           string
	IsStatic            bool // if set, user can't set a different value, but can press and see them. Shows number of items
	IsMultiple          bool
	HasLabelColor       bool
	SetTitle            bool
	StoreKey            string
	BGColor             zgeo.Color
	TextColor           zgeo.Color
	HoverColor          zgeo.Color
	MinWidth            float64

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
	Function    func()
}

var (
	MenuedOItemSeparator        = MenuedOItem{IsSeparator: true}
	MenuedOwnerDefaultBGColor   = zstyle.ColF(zgeo.ColorNew(0.92, 0.91, 0.90, 1), zgeo.ColorNew(0.12, 0.11, 0.1, 1))
	MenuedOwnerDefaultTextColor = zstyle.GrayF(0.1, 0.9)
	//	MenuedOwnerDefaultHightlightColor = zstyleColF(zgeo.ColorNewGray(0, 0.7), zgeo.ColorNewGray(1, 0.7))
	MenuedOwnerDefaultHightlightColor = zstyle.ColF(zgeo.ColorNew(0.035, 0.29, 0.85, 1), zgeo.ColorNew(0.8, 0.8, 1, 1))
	menuOwnersMap                     = map[zview.View]*MenuedOwner{}
)

func NewMenuedOwner() *MenuedOwner {
	o := &MenuedOwner{}
	o.Font = zgeo.FontNice(zgeo.FontDefaultSize-1, zgeo.FontStyleNormal)
	o.HoverColor = MenuedOwnerDefaultHightlightColor()
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

func MenuedFuncAction(name string, f func()) MenuedOItem {
	var item MenuedOItem
	item.Name = name
	item.Value = rand.Int31()
	item.IsAction = true
	item.Function = f
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
	_, got := zkeyvalue.DefaultStore.GetDict(o.StoreKey)
	return got
}

func (o *MenuedOwner) Build(view zview.View, items []MenuedOItem) {
	if view == nil {
		zlog.Fatal(nil, "MO Build with view==nil")
	} else {
		o.View = view
		// nv := view.Native()
		// zlog.Info("MO ADDStopper:", nv.Hierarchy(), zlog.GetCallingStackString())
		// nv.AddOnRemoveFunc(o.Stop)
		presser := view.(zview.DownPressable)
		presser.SetPressedDownHandler(func() {
			o.MinWidth = view.Rect().Size.W
			if o.CreateItemsFunc != nil {
				o.items = o.CreateItemsFunc()
			}
			if len(o.items) != 0 {
				o.popup()
			}
		})
	}
	if o.StoreKey != "" {
		dict, got := zkeyvalue.DefaultStore.GetDict(o.StoreKey)
		if got {
			for i, item := range items {
				str := fmt.Sprint(item.Value)
				_, items[i].Selected = dict[str]
			}
		} else {
			if o.CreateItemsFunc != nil {
				items = o.CreateItemsFunc()
			}
		}
	}
	o.UpdateMenuedItems(items)
	view.Native().AddOnRemoveFunc(func() {
		delete(menuOwnersMap, view)
	})
	menuOwnersMap[view] = o
}

func OwnerForView(view zview.View) *MenuedOwner {
	return menuOwnersMap[view]
}

// func (o *MenuedOwner) Stop() {
// 	if o.View == nil {
// 		zlog.Info("Stopping MenuOwner view==nil:", zlog.CallingStackString())
// 	}
// 	// zlog.Info("Stopping MenuOwner:", o.View != nil)
// 	// zlog.Info("Stopping MenuOwner for:", o.View.ObjectName())
// 	presser := o.View.(zview.Pressable)
// 	// zlog.Info("Stopping MenuOwner for2:", presser != nil)
// 	presser.SetPressedHandler(nil)
// 	presser.SetLongPressedHandler(nil)
// 	*o = MenuedOwner{} // this will
// }

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

func (o *MenuedOwner) SetTitleText(text string) {
	if o.SetTitle || o.TitleIsValueIfOne || o.PluralableWord != "" || o.GetTitleFunc != nil {
		zlog.Assert(o.View != nil)
		ts, got := o.View.(ztextinfo.TextSetter)
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
		// zlog.Info("MO updateTitleAndImage:", len(o.items), o.TitleIsValueIfOne, count)
		if !(o.TitleIsValueIfOne && count == 1) {
			if o.PluralableWord != "" {
				if count > 0 {
					nstr = zwords.PluralizeWordAndCountWords(o.PluralableWord, float64(count), "", "", nil) //  map[int]string{0: "no", total: "all"})
				}
			} else if o.GetTitleFunc != nil {
				nstr = o.GetTitleFunc(count)
			} else {
				nstr = strconv.Itoa(count)
			}
			o.SetTitleText(nstr)
			return
		}
	}
	if o.GetTitleFunc != nil {
		nstr = o.GetTitleFunc(1)
		o.SetTitleText(nstr)
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
	o.SetTitleText(nstr)
	if o.ImagePath != "" {
		// zlog.Info("MO SetImagePath:", spath)
		io := o.View.(zimage.Owner)
		io.SetImage(nil, spath, nil)
	}
}

// func (o *MenuedOwner) UpdateItems(items zdict.Items, values []interface{}) {
// 	var mitems []MenuedOItem
// 	for _, item := range items {
// 		var m MenuedOItem
// 		m.Name = item.Name
// 		m.Value = item.Value
// 		for _, v := range values {
// 			if reflect.DeepEqual(item.Value, v) {
// 				m.Selected = true
// 				break
// 			}
// 		}
// 		mitems = append(mitems, m)
// 	}
// 	zlog.Info("MO UpdateItems", values)
// 	o.UpdateMenuedItems(mitems)
// }

// MOItemsFromZDictItemsAndValues creates MenuedOItem slice from zdict Items and a slice or single value of anything
func MOItemsFromZDictItemsAndValues(enum zdict.Items, values any, isActions bool) []MenuedOItem {
	var mItems []MenuedOItem
	var vals []any
	rval := reflect.ValueOf(values)
	if rval.Kind() != reflect.Slice {
		vals = []any{values}
	} else {
		for j := 0; j < rval.Len(); j++ {
			vals = append(vals, rval.Index(j).Interface())
		}
	}
	for _, item := range enum {
		var m MenuedOItem
		for _, v := range vals {
			// zlog.Info("EQ:", item.Value, rval.Index(j).Interface())
			if reflect.DeepEqual(item.Value, v) {
				m.Selected = true
				break
			}
		}
		m.IsAction = isActions
		m.Name = item.Name
		m.Value = item.Value
		mItems = append(mItems, m)
	}
	return mItems
}

func (o *MenuedOwner) UpdateItems(items zdict.Items, values any, isAction bool) {
	mitems := MOItemsFromZDictItemsAndValues(items, values, isAction)
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
	if o.CreateItemsFunc != nil && len(o.items) == 0 {
		o.items = o.CreateItemsFunc()
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

func (o *MenuedOwner) popup() {
	const (
		imageWidth = 40
		imageMarg  = 8
		checkWidth = 14
		topMarg    = 6
		bottomMarg = 6
		rightMarg  = 4
	)
	allAction := true
	for _, item := range o.items {
		// zlog.Info("popitem:", i, item.Selected, o.IsMultiple)
		if !item.IsAction {
			allAction = false
		}
	}
	stack := zcontainer.StackViewVert("menued-pop-stack")
	stack.SetMargin(zgeo.RectFromXY2(0, topMarg, 0, -bottomMarg))
	list := zgridlist.NewView("menu-list")
	stack.SetBGColor(o.BGColor)
	list.MultiSelectable = o.IsMultiple
	list.Selectable = !o.IsMultiple
	list.HoverColor = o.HoverColor
	list.BorderColor.Valid = false
	list.SelectColor = zgeo.ColorClear
	list.PressedColor = zgeo.ColorClear
	list.CellColor = o.BGColor
	list.MakeFullSize = true
	list.MaxColumns = 1
	list.MultiplyColorAlternate = 1
	// list.CellColorFunc = func(id string) zgeo.Color {
	// 	i := list.IndexOfID(id)
	// 	if o.items[i].IsSeparator {
	// 		return zgeo.ColorLightGray
	// 	}
	// 	return o.BGColor
	// }
	stack.Add(list, zgeo.TopLeft|zgeo.Expand)

	lineHeight := o.Font.LineHeight() + 8
	list.CellHeightFunc = func(id string) float64 {
		i, _ := strconv.Atoi(id)
		if o.items[i].IsSeparator {
			return 1
		}
		return lineHeight
	}
	list.CellCountFunc = func() int {
		return len(o.items)
	}
	list.IDAtIndexFunc = func(i int) string {
		return strconv.Itoa(i)
	}
	list.UpdateSelectionFunc = o.updateCellSelection
	rm := float64(rightMarg)
	if o.HasLabelColor {
		rm += 24
	}
	list.CreateCellFunc = o.createRow

	var max string
	for _, item := range o.items {
		if len(item.Name) > len(max) {
			max = item.Name
		}
	}
	w := ztextinfo.WidthOfString(max, o.Font) * 1.1
	if !allAction && !o.IsStatic {
		w += checkWidth
	}
	if o.ImagePath != "" {
		w += imageWidth + imageMarg
	}
	w += 18 // test
	zfloat.Maximize(&w, o.MinWidth)
	stack.SetMinSize(zgeo.Size{w, 0})

	list.HandleSelectionChangedFunc = func() {
		if o.IsStatic {
			return
		}
		old := map[string]bool{}
		if !o.IsMultiple {
			for i := range o.items {
				if o.items[i].Selected {
					old[strconv.Itoa(i)] = true
				}
				o.items[i].Selected = false
			}
		}
		ids := list.SelectedIDs()
		if len(ids) == 1 {
			i, _ := strconv.Atoi(ids[0])
			o.items[i].Selected = !o.items[i].Selected
			if o.items[i].IsAction {
				o.items[i].Selected = false
				if o.items[i].Function != nil {
					o.items[i].Function()
					o.getItems()
					o.updateTitleAndImage()
				} else if o.ActionHandlerFunc != nil {
					id := o.items[i].Value.(string)
					o.ActionHandlerFunc(id)
					o.getItems()
					o.updateTitleAndImage()
				}
				zpresent.Close(stack, false, nil)
				return
			}
		}
		o.updateTitleAndImage()
		for _, id := range ids {
			old[id] = true
		}
		for id := range old {
			list.UpdateSelectionFunc(list, id)
		}
		if !o.IsMultiple { // && fromPressed {
			// list.HandleSelectionChangedFunc = nil // we do this so we don't get any mouse-up extra events
			// zlog.Info("MenuPopup close", zlog.CallingStackString())
			zpresent.Close(stack, false, nil)
		}
		if o.SelectedHandlerFunc != nil {
			o.SelectedHandlerFunc()
		}
		list.UnselectAll(false)
	}
	att := zpresent.AttributesNew()
	att.Modal = true
	att.ModalDimBackground = false
	att.ModalCloseOnOutsidePress = true
	att.ModalDropShadow.Delta = zgeo.SizeBoth(1)
	att.ModalDropShadow.Blur = 2
	att.ModalDismissOnEscapeKey = true
	stack.SetStroke(1, zgeo.ColorNewGray(0.5, 1), false)
	pos := o.PopPos
	if o.View != nil {
		nv := o.View.Native()
		pos = nv.AbsoluteRect().Pos
	}
	att.Pos = &pos
	// zlog.Info("menu popup")
	//	list.Focus(true)
	zpresent.PresentView(stack, att, func(*zwindow.Window) {
		//		pop.Element.Set("selectedIndex", 0)
	}, func(dismissed bool) {
		// zlog.Info("menued closed", dismissed, o.IsMultiple)
		if !dismissed || o.IsMultiple { // if multiple, we handle any select/deselect done
			o.updateTitleAndImage()
			if o.ClosedFunc != nil {
				o.ClosedFunc()
			}
			if o.IsStatic {
				// for i := range o.items {
				// 	o.items[i].Selected = false
				// }
			}
			o.updateTitleAndImage()
		}
	})
}

func (o *MenuedOwner) updateCellSelection(grid *zgridlist.GridListView, id string) {
	i, _ := strconv.Atoi(id)
	item := o.items[i]
	col := o.TextColor
	if grid.IsHoverCell(id) {
		col = grid.HoverColor.ContrastingGray()
	} else if item.TextColor.Valid {
		col = item.TextColor
	}
	if item.IsDisabled || o.IsStatic {
		col.SetOpacity(0.5)
	}
	// zlog.Info("updateRow", id, item.Name, item.Selected)
	v := grid.CellView(id)
	zcontainer.ViewRangeChildren(v, false, false, func(view zview.View) bool {
		label, _ := view.(*zlabel.Label)
		if label != nil {
			label.SetColor(col)
			if label.ObjectName() == "status" {
				str := ""
				if item.Selected {
					str = "√"
				}
				label.SetText(str)
			}
		}
		return true
	})
}

func (o *MenuedOwner) createRow(grid *zgridlist.GridListView, id string) zview.View {
	v := zcontainer.New("id")
	i, _ := strconv.Atoi(id)
	item := o.items[i]
	if item.IsSeparator {
		v.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
			canvas.SetColor(zgeo.ColorWhite)
			canvas.StrokeHorizontal(rect.Min().X, rect.Max().X+1, rect.Center().Y+1, 1, zgeo.PathLineButt)
			canvas.SetColor(zgeo.ColorGray)
			canvas.StrokeHorizontal(rect.Min().X, rect.Max().X+1, rect.Center().Y, 1, zgeo.PathLineButt)
		})
		return v
	}
	marg := zgeo.Size{8, 0}

	if !item.IsAction {
		status := zlabel.New("")
		status.SetObjectName("status")
		status.SetFont(o.Font)
		if item.Selected {
			status.SetText("√")
		}
		v.Add(status, zgeo.CenterLeft, marg)
		marg.W += 20
	}
	title := zlabel.New(item.Name)
	title.SetText(item.Name)
	title.SetObjectName("title")
	font := o.Font
	if item.IsAction {
		font.Style = zgeo.FontStyleItalic
	}
	title.SetFont(font)
	v.Add(title, zgeo.CenterLeft, marg)

	//	marg.W = 8

	if o.ImagePath != "" {
		sval := fmt.Sprint(item.Value)
		spath := path.Join(o.ImagePath, sval+".png")
		iv := zimageview.New(nil, spath, zgeo.Size{32, 20})
		iv.DownsampleImages = true
		v.Add(iv, zgeo.CenterRight, marg)
		marg.W += 22
	}
	if o.HasLabelColor && item.LabelColor.Valid {
		cv := zcustom.NewView(id)
		col := item.LabelColor
		if item.IsDisabled {
			col.SetOpacity(0.5)
		}
		cv.SetBGColor(col)
		cv.SetCorner(3)
		cv.SetObjectName("color-label")
		v.Add(cv, zgeo.CenterRight, marg)
	}
	return v
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
		zkeyvalue.DefaultStore.SetDict(dict, o.StoreKey, true)
	}
}

func MenuOwningButtonCreate(menu *MenuedOwner, items []MenuedOItem, shape zshape.Type) *zshape.ShapeView {
	v := zshape.NewView(shape, zgeo.Size{20, 20})
	v.SetImage(nil, "images/zmenu-arrows.png", nil)
	v.ImageMargin = zgeo.Size{4, 4}
	v.ImageAlign = zgeo.CenterRight | zgeo.Proportional
	v.SetTextAlignment(zgeo.CenterLeft)
	v.ImageGap = 4
	v.SetColor(zstyle.DefaultBGColor().Mixed(zgeo.ColorGray, 0.2))
	v.StrokeColor = zgeo.ColorNewGray(0, 0.3)
	v.StrokeWidth = 1
	v.SetTextColor(zstyle.DefaultFGColor())
	menu.Build(v, items)

	return v
}
