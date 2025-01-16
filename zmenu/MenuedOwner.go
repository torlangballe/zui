//go:build zui

package zmenu

import (
	"fmt"
	"math/rand"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
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
	"github.com/torlangballe/zutil/ztimer"
	"github.com/torlangballe/zutil/zwords"
)

type MenuedOwner struct {
	View                zview.View
	Name                string   // Name is just used for setting View Object Name for debugging
	PopPos              zgeo.Pos // either View or PopPos
	SelectedHandlerFunc func(edited bool)
	GetTitleFunc        func(itemCount int) string
	ActionHandlerFunc   func(id string)
	CreateItemsFunc     func() []MenuedOItem
	ClosedFunc          func()
	PluralableWord      string // if set, used instead of GetTitle, and pluralized
	TitleIsValueIfOne   bool   // if set and IsMultiple, name of value used as title if only one set
	TitleIsAll          string // if != "", all items are listed in title, separated by TitleIsAll string
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
	DefaultSelected     any
	items               []MenuedOItem
	hasShortcut         bool
	currentPopupStack   *zcontainer.StackView
	needsSave           bool
}

type MenuedOItem struct {
	Name        string
	Value       interface{}
	Selected    bool
	LabelColor  zgeo.Color
	TextColor   zgeo.Color
	Shortcut    zkeyboard.KeyMod
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
	o.Init()
	return o
}

func (o *MenuedOwner) Init() {
	o.Font = zgeo.FontNice(zgeo.FontDefaultSize-1, zgeo.FontStyleNormal)
	o.HoverColor = MenuedOwnerDefaultHightlightColor()
	o.BGColor = MenuedOwnerDefaultBGColor()
	o.TextColor = MenuedOwnerDefaultTextColor()
}

func MenuedAction(name string, val interface{}) MenuedOItem {
	var item MenuedOItem
	item.Name = name
	item.Value = val
	item.IsAction = true
	return item
}

func MenuedFuncAction(name string, f func()) MenuedOItem {
	return MenuedSCFuncAction(name, zkeyboard.KeyNone, zkeyboard.ModifierNone, f)
}

func MenuedSCFuncAction(name string, scKey zkeyboard.Key, mod zkeyboard.Modifier, f func()) MenuedOItem {
	var item MenuedOItem
	item.Name = name
	item.Value = rand.Int31()
	item.IsAction = true
	item.Function = f
	item.Shortcut = zkeyboard.KMod(scKey, mod)
	return item
}

func (m *MenuedOItem) SetShortcut(key zkeyboard.Key, mods zkeyboard.Modifier) {
	s := zkeyboard.KMod(key, mods)
	m.Shortcut = s
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
	var isSet bool
	o.View = view
	if view != nil {
		view.Native().SetPressedHandler("", zkeyboard.ModifierNone, func() { // SetPressedDownHandler doesn't fire for some reaspm. so using SetPressedHandler.
			o.MinWidth = view.Rect().Size.W
			if o.CreateItemsFunc != nil {
				o.items = o.CreateItemsFunc()
			}
			if len(o.items) != 0 {
				o.popup()
			}
		})
	}
	// zlog.Info("Build", o.View.ObjectName(), items == nil)
	isFirst := (items == nil)
	if isFirst {
		if o.CreateItemsFunc != nil {
			items = o.CreateItemsFunc()
			isSet = true
		}
	}
	if !o.needsSave && o.StoreKey != "" {
		dict, got := zkeyvalue.DefaultStore.GetDict(o.StoreKey)
		if got {
			isSet = true
			for i, item := range items {
				str := fmt.Sprint(item.Value)
				// zlog.Info("MO.Build Select:", o.Name, item.Name, dict[str])
				_, items[i].Selected = dict[str]
			}
		}
	}
	// zlog.Info("MO.Build:", o.Name, isFirst, o.DefaultSelected != nil)
	if isFirst && o.DefaultSelected != nil {
		var selected bool
		var def *MenuedOItem
		for i, item := range items {
			if item.Selected {
				selected = true
				break
			}
			// zlog.Info("MO.Build eq:", item.Value, o.DefaultSelected)
			if reflect.DeepEqual(item.Value, o.DefaultSelected) {
				def = &items[i]
			}
		}
		if !selected && def != nil {
			selected = true
			def.Selected = true
		}
	}
	o.UpdateMenuedItems(items)
	if view != nil {
		view.Native().AddOnRemoveFunc(func() {
			delete(menuOwnersMap, view)
		})
		menuOwnersMap[view] = o
	}
	// zlog.Info("MO.Build done:", o.StoreKey, isSet, o.SelectedHandlerFunc != nil)
	if isSet && o.SelectedHandlerFunc != nil {
		o.SelectedHandlerFunc(false)
	}
}

func OwnerForView(view zview.View) *MenuedOwner {
	return menuOwnersMap[view]
}

func (o *MenuedOwner) SelectedItem() *zdict.Item {
	sitems := o.SelectedItems()
	if len(sitems) == 0 {
		return nil
	}
	si := sitems[0]
	return &si
}

func (o *MenuedOwner) SelectedValue() any {
	item := o.SelectedItem()
	if item == nil {
		return nil
	}
	return item.Value
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
	// zlog.Info("MO.SetTitleText3:", text, len(o.items), o.SetTitle, o.TitleIsValueIfOne, o.TitleIsAll, o.PluralableWord, o.GetTitleFunc != nil) //, zlog.CallingStackString())
	if o.View != nil && o.SetTitle || o.TitleIsValueIfOne || o.TitleIsAll != "" || o.PluralableWord != "" || o.GetTitleFunc != nil {
		ts, got := o.View.(ztextinfo.TextSetter)
		if got {
			ts.SetText(text)
		}
	}
}

func (o *MenuedOwner) UpdateTitleAndImage() {
	var nstr string
	if o.IsMultiple && !o.IsStatic {
		if o.TitleIsAll != "" {
			var s []string
			for _, i := range o.items {
				// zlog.Info("UpdateTitleAndImage?", i.Name, i.IsAction, i.IsSeparator, i.Selected)
				if !i.IsAction && !i.IsSeparator && i.Selected {
					s = append(s, i.Name)
				}
			}
			o.SetTitleText(strings.Join(s, o.TitleIsAll))
			return
		}
		var count, total int
		for _, i := range o.items {
			if !i.IsAction && !i.IsSeparator {
				if i.Selected {
					// zlog.Info("Menued UpdateTitleAndImage add", i.Name)
					count++
				}
				total++
			}
		}
		// zlog.Info("MO UpdateTitleAndImage:", len(o.items), o.TitleIsValueIfOne, count)
		if !(o.TitleIsValueIfOne && count == 1) {
			if o.PluralableWord != "" {
				// if count > 0 {
				nstr = zwords.Pluralize(o.PluralableWord, count)
				// }
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
		nstr = o.GetTitleFunc(len(o.items))
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
		nstr = item.Name
	}
	o.SetTitleText(nstr)
	if o.ImagePath != "" {
		// zlog.Info("MO SetImagePath:", spath)
		io := o.View.(zimage.Owner)
		io.SetImage(nil, spath, nil)
	}
}

// MOItemsFromZDictItemsAndValues creates MenuedOItem slice from zdict Items and a slice or single value of anything
func MOItemsFromZDictItemsAndValues(enum zdict.Items, values any, isActions bool) []MenuedOItem {
	// zlog.Info("FV.MOItemsFromZDictItemsAndValues:", zlog.Full(enum))
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
		// zlog.Info("MOItemsFromZDictItemsAndValues:", item, vals)
		sitem := fmt.Sprint(item.Value)
		for _, v := range vals {
			// zlog.Info("EQ:", item.Value, v, reflect.DeepEqual(item.Value, v), reflect.TypeOf(item.Value), reflect.TypeOf(v))
			if reflect.DeepEqual(item.Value, v) || sitem == fmt.Sprint(v) {
				m.Selected = true
				break
			}
		}
		// zlog.Info("MOItemsFromZDictItemsAndValues:", m)
		m.IsAction = isActions
		m.Name = item.Name
		m.Value = item.Value
		mItems = append(mItems, m)
	}
	return mItems
}

func (o *MenuedOwner) UpdateItems(items zdict.Items, values any, isAction bool) {
	// zlog.Info("MO UpdateItems", values, zlog.CallingStackString())
	mitems := MOItemsFromZDictItemsAndValues(items, values, isAction)
	o.UpdateMenuedItems(mitems)
}

func (o *MenuedOwner) AddMOItem(item MenuedOItem) {
	o.items = append(o.items, item)
}

func (o *MenuedOwner) AddSeparator(item MenuedOItem) {
	var i MenuedOItem
	i.IsSeparator = true
	o.AddMOItem(i)
}

func (o *MenuedOwner) UpdateMenuedItems(items []MenuedOItem) {
	// zlog.Info("Update:", zlog.Full(items), zlog.CallingStackString())
	o.items = items
	o.UpdateTitleAndImage()
}

func (o *MenuedOwner) SetSelectedValuesAndEdited(vals []interface{}, edited bool) {
	// zlog.Info("SetSelectedValues1", vals, o.Name)
outer:
	for i, item := range o.getItems() {
		for j, v := range vals {
			// zlog.Info("SetSelectedValues?", o.Name, item.Name, item.Value, "==", v, reflect.DeepEqual(item.Value, v))
			if reflect.DeepEqual(item.Value, v) {
				o.items[i].Selected = true
				zslice.RemoveAt(&vals, j)
				continue outer
			}
		}
		o.items[i].Selected = false
	}
	// zlog.Info("SetSelectedValues", vals, o.Name, zlog.Full(o.items), o.SelectedItem())
	o.UpdateTitleAndImage()
	if o.SelectedHandlerFunc != nil {
		o.SelectedHandlerFunc(edited)
	}
}

func (o *MenuedOwner) SetSelectedValues(vals []interface{}) {
	o.SetSelectedValuesAndEdited(vals, false)
}

func (o *MenuedOwner) SetSelectedValue(val interface{}) {
	o.SetSelectedValues([]interface{}{val})
}

func (o *MenuedOwner) RegenerateItems() {
	o.getItems()
}

func (o *MenuedOwner) getItems() []MenuedOItem {
	if o.CreateItemsFunc != nil {
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
	zlog.Error("no value to change name for")
}

const (
	gap           = 4
	colorWidth    = 20
	imageWidth    = 40
	checkWidth    = 14
	shortcutWidth = 28
	topMarg       = 6
	bottomMarg    = 2
	rightMarg     = 4
)

func (o *MenuedOwner) popup() {
	allAction := true
	o.hasShortcut = false
	for _, item := range o.items {
		if !item.IsAction {
			allAction = false
		}
		if item.Shortcut.Key != 0 {
			o.hasShortcut = true
		}
	}
	o.getItems()
	stack := zcontainer.StackViewVert("menued-pop-stack")
	o.currentPopupStack = stack
	stack.SetMargin(zgeo.RectFromXY2(0, topMarg, 0, -bottomMarg))
	list := zgridlist.NewView("menu-list")
	list.SetMargin(zgeo.RectFromXY2(0, 0, -8, 0))
	list.FocusWidth = 0
	list.JSSet("className", "znofocus")
	stack.SetBGColor(o.BGColor)
	list.MultiSelectable = o.IsMultiple
	list.SetStroke(0, zgeo.ColorClear, false)
	list.Selectable = !o.IsMultiple
	list.HoverColor = o.HoverColor
	list.BorderColor.Valid = false
	list.SelectColor = zgeo.ColorClear
	list.PressedColor = zgeo.ColorClear
	list.CellColor = o.BGColor
	list.MakeFullSize = true
	list.MaxColumns = 1
	list.MultiplyColorAlternate = 1
	list.CurrentHoverID = "0"
	list.BorderColor = zgeo.Color{}
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
	list.UpdateCellSelectionFunc = o.updateCellSelection
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
		w += checkWidth + gap
	}
	if o.ImagePath != "" {
		w += imageWidth + gap
	}
	if o.hasShortcut {
		w += shortcutWidth + gap
	}
	w += 40 // test
	zfloat.Maximize(&w, o.MinWidth)
	list.SetMinSize(zgeo.SizeD(w, 0))
	if len(o.items) < 20 {
		list.ShowBar = false
	}

	list.HandleKeyFunc = func(km zkeyboard.KeyMod, down bool) bool {
		if list.CurrentHoverID != "" && km.Key.IsReturnish() {
			list.SelectCell(list.CurrentHoverID, true, false)
			return true
		}
		for i, item := range o.items {
			if item.Shortcut.Key == km.Key && item.Shortcut.Modifier == km.Modifier {
				list.SelectCell(strconv.Itoa(i), true, false)
				return true
			}
		}
		return false
	}
	list.HandleSelectionChangedFunc = func() {
		if o.IsStatic {
			return
		}
		ids := list.SelectedIDs()
		if len(ids) == 1 {
			i, _ := strconv.Atoi(ids[0])
			item := o.items[i]
			if item.IsSeparator {
				zpresent.Close(stack, false, nil)
				return
			}
			if item.IsAction {
				// o.items[i].Selected = false
				if item.Function != nil {
					zpresent.Close(stack, false, nil)
					item.Function()
					// o.getItems() // getItems+UpdateTitleAndImage assumes things; Entire view owning menu might be gone... removing to see what will happen
					// o.UpdateTitleAndImage()
					o.getItems()
					o.UpdateTitleAndImage()
				} else if o.ActionHandlerFunc != nil {
					id := item.Value.(string)
					zpresent.Close(stack, false, nil)
					o.ActionHandlerFunc(id)
					o.getItems()
					o.UpdateTitleAndImage()
				}
				return
			}
		}
		oldSelected := map[string]bool{}
		if !o.IsMultiple {
			for i := range o.items {
				if o.items[i].Selected {
					oldSelected[strconv.Itoa(i)] = true
				}
				o.items[i].Selected = false
			}
		}
		if len(ids) == 1 {
			i, _ := strconv.Atoi(ids[0])
			item := o.items[i]
			o.items[i].Selected = !item.Selected
			list.LayoutCells(true)
			if !o.IsMultiple {
				zpresent.Close(stack, false, nil)
			}
		}
		o.UpdateTitleAndImage()
		for _, id := range ids {
			oldSelected[id] = true
		}
		for id := range oldSelected {
			list.UpdateCellSelectionFunc(list, id)
		}
		//!!! if !o.IsMultiple { // && fromPressed {
		// 	// list.HandleSelectionChangedFunc = nil // we do this so we don't get any mouse-up extra events
		// 	// zlog.Info("MenuPopup close", zlog.CallingStackString())
		// zpresent.Close(stack, false, nil)
		// }
		if o.SelectedHandlerFunc != nil {
			// zlog.Info("Menu edited???")
			o.SelectedHandlerFunc(true)
		}
		list.UnselectAll(false)
		o.saveToStore()
	}
	att := zpresent.AttributesNew()
	att.Modal = true
	att.ModalDimBackground = false
	att.ModalCloseOnOutsidePress = true
	att.ModalDropShadows[0].Delta = zgeo.SizeBoth(1)
	att.ModalDropShadows[0].Blur = 2
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
	att.PresentedFunc = func(win *zwindow.Window) {
		if win == nil {
			return
		}
		ztimer.StartIn(0.1, func() {
			list.Focus(true)
		})
	}
	att.ClosedFunc = func(dismissed bool) {
		// zlog.Info("menued closed", dismissed, o.IsMultiple)
		if !dismissed || o.IsMultiple { // if multiple, we handle any select/deselect done
			o.UpdateTitleAndImage()
			if o.ClosedFunc != nil {
				o.ClosedFunc()
			}
			if o.IsStatic {
				// for i := range o.items {
				// 	o.items[i].Selected = false
				// }
			}
			o.UpdateTitleAndImage()
		}
	}
	zpresent.PresentView(stack, att)
}

func (o *MenuedOwner) ClosePopup() {
	if o.currentPopupStack != nil {
		zpresent.Close(o.currentPopupStack, false, nil)
	}
}

func (o *MenuedOwner) HandleOutsideShortcut(sc zkeyboard.KeyMod) bool {
	for i, item := range o.getItems() {
		if item.Shortcut.Matches(sc) {
			if o.CreateItemsFunc != nil {
				// o.items = o.CreateItemsFunc() // we need to re-generate menu items -- done in getItems above
				item = o.items[i]
				if item.Function != nil {
					go item.Function()
					return true
				}
				zlog.Assert(o.ActionHandlerFunc != nil)
				id := item.Value.(string)
				o.ActionHandlerFunc(id)
				return true
			}
		}
	}
	return false
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
	// zlog.Info("updateRow", id, item.Name, item.Selected, item.IsDisabled || o.IsStatic)
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
	marg := zgeo.SizeD(4, 0)

	// zlog.Info("CreateRow:", i, item.Name, item.Selected, item.IsAction, item.Value)
	if !item.IsAction {
		status := zlabel.New("")
		status.SetObjectName("status")
		status.SetFont(o.Font)
		if item.Selected {
			status.SetText("√")
		}
		v.Add(status, zgeo.CenterLeft, marg)
	}
	marg.W += 20
	title := zlabel.New(item.Name)
	title.SetText(item.Name)
	title.SetObjectName("title")
	font := *o.Font
	if item.IsAction {
		font.Style = zgeo.FontStyleItalic
	}
	title.SetFont(&font)
	v.Add(title, zgeo.CenterLeft, marg)

	marg.W = 8

	if o.ImagePath != "" {
		sval := fmt.Sprint(item.Value)
		spath := path.Join(o.ImagePath, sval+".png")
		iv := zimageview.NewWithCachedPath(spath, zgeo.SizeD(32, 20))
		iv.DownsampleImages = true
		v.Add(iv, zgeo.CenterRight, marg)
		marg.W += gap + imageWidth
	}
	if o.HasLabelColor {
		if item.LabelColor.Valid {
			cv := zcustom.NewView(id)
			col := item.LabelColor
			if item.IsDisabled {
				col.SetOpacity(0.5)
			}
			cv.SetMinSize(zgeo.SizeD(colorWidth, 0))
			cv.SetBGColor(col)
			cv.SetCorner(3)
			cv.SetObjectName("color-label")
			v.Add(cv, zgeo.CenterRight, marg)
		}
		marg.W += gap + colorWidth
	}
	if o.hasShortcut {
		singleLetterKey := false
		str := item.Shortcut.AsString(singleLetterKey)
		keyLabel := zlabel.New(str)
		title.SetObjectName("shortcut")
		font := o.Font
		font.Style = zgeo.FontStyleBold
		keyLabel.SetFont(font)
		v.Add(keyLabel, zgeo.CenterRight, marg)
		marg.W += gap + shortcutWidth
	}

	return v
}

func (o *MenuedOwner) saveToStore() {
	// zlog.Info("MO SAVE", o.StoreKey, o.items)
	if o.StoreKey != "" {
		dict := zdict.Dict{}
		for _, item := range o.items {
			if item.Selected {
				str := fmt.Sprint(item.Value)
				dict[str] = true
			}
		}
		// zlog.Info("MO SAVE2", dict)
		zkeyvalue.DefaultStore.SetDict(dict, o.StoreKey, true)
	}
	o.needsSave = false
}

func MenuOwningButtonCreate(menu *MenuedOwner, items []MenuedOItem, shape zshape.Type) *zshape.ShapeView {
	v := zshape.NewView(shape, zgeo.SizeD(60, 20))
	name := menu.Name
	if name == "" {
		name = "menu"
	}
	v.SetObjectName(name)
	v.ImageMargin = zgeo.RectFromXY2(0, 0, -4, 0)
	v.ImageAlign = zgeo.CenterRight | zgeo.Proportional // both must be before SetImage
	v.Ratio = 0.3
	v.SetImage(nil, true, zgeo.SizeBoth(12), "images/zcore/zmenu-arrows.png", zgeo.SizeNull, nil)
	v.SetTextAlignment(zgeo.CenterLeft)
	v.SetSpacing(4)
	v.SetMinSize(zgeo.SizeD(20, 22))
	v.TextMargin = zgeo.RectFromXY2(8, 0, -20, 0)
	v.SetColor(zstyle.DefaultBGColor().Mixed(zgeo.ColorWhite, 0.2))
	v.StrokeColor = zgeo.ColorNewGray(0, 0.5)
	v.StrokeWidth = 1
	v.SetTextWrap(ztextinfo.WrapTailTruncate)
	v.SetTextColor(zstyle.DefaultFGColor())
	menu.Build(v, items)
	return v
}

func (o *MenuedOwner) Dump() {
	for _, item := range o.items {
		zlog.Info("MDump:", item.Name, item.Selected)
	}
}
