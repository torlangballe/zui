//go:build zui

package zfields

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
)

type FieldSliceView struct {
	FieldView
	addButton          *zimageview.ImageView
	globalDeleteButton *zimageview.ImageView
	slicePtr           any
	field              *Field
	menu               *zmenu.MenuView
	indicatorFieldName string
	currentIndex       int
	stack              *zcontainer.StackView
	storeKey           string
	isStructItems      bool
}

func (fv *FieldView) NewSliceView(slicePtr any, f *Field) *FieldSliceView {
	vert := !f.Vertical.Bool()
	v := &FieldSliceView{}
	sliceRval := reflect.ValueOf(slicePtr).Elem()
	v.isStructItems = (sliceRval.Type().Elem().Kind() == reflect.Struct)
	if !v.isStructItems {
		vert = !vert
	}
	v.Init(v, vert, f.FieldName)
	v.slicePtr = slicePtr
	v.field = f
	v.parent = fv
	v.params = fv.params
	v.currentIndex = -1
	v.params.Field.MergeInField(f)
	v.build()
	return v
}

func (v *FieldSliceView) build() {
	var header *zcontainer.StackView
	var index int
	sliceRval := reflect.ValueOf(v.slicePtr).Elem()
	if v.field.Flags&FlagHasFrame != 0 {
		var title string
		if v.field.Flags&FlagFrameIsTitled != 0 {
			title = v.field.TitleOrName()
		}
		// zlog.Info("Build:", v.ObjectName(), v.field.Flags&FlagFrameIsTitled != 0, title)
		header = zwidgets.MakeStackATitledFrame(&v.StackView, title, v.field.Flags&FlagFrameTitledOnFrame != 0, v.field.Styling, v.field.Styling)
	}
	if header == nil && !v.field.IsStatic() {
		header = zcontainer.StackViewHor("header")
		v.Add(header, zgeo.TopLeft|zgeo.HorExpand)
	}
	if v.field.IsStatic() {
		v.stack = &v.StackView
	} else {
		v.stack = zcontainer.StackViewNew(v.Vertical, "stack")
		v.Add(v.stack, zgeo.TopLeft|zgeo.Expand)
	}
	var collapse bool
	if v.field.Flags&FlagGroupSingle != 0 && !v.field.IsStatic() {
		collapse = true
		v.createMenu()
		header.Add(v.menu, zgeo.CenterRight)

		if !v.field.IsStatic() {
			v.globalDeleteButton = makeButton("minus", "red")
			header.Add(v.globalDeleteButton, zgeo.CenterRight)
			v.globalDeleteButton.SetPressedHandler(func() {
				v.handleDeleteItem(v.currentIndex)
			})
		}
		v.storeKey = v.CreateStoreKeyForField(v.field, "SliceIndex")
		index, _ = zkeyvalue.DefaultSessionStore.GetInt(v.storeKey, 0)
		zint.Minimize(&index, sliceRval.Len()-1)
		v.menu.SelectWithValue(index)
	}
	if !v.field.IsStatic() {
		v.addButton = makeButton("plus", "gray")
		header.Add(v.addButton, zgeo.CenterRight)
		v.addButton.SetPressedHandler(v.handleAddItem)
	}
	for i := 0; i < sliceRval.Len(); i++ {
		// zlog.Info("SliceFView.Add:", i, v.Hierarchy(), collapse)
		v.addItem(i, sliceRval.Index(i), collapse)
	}
	v.selectItem(index)
}

// func (v *FieldSliceView) ArrangeChildren() {
// zlog.Info("FieldSliceView.ArrangeChildren")
// 	v.FieldView.ArrangeChildren()
// }

func (v *FieldSliceView) createMenu() {
	v.menu = zmenu.NewView("menu", nil, nil)
	v.menu.SetSelectedHandler(v.handleMenuSelected)

	v.indicatorFieldName = FindIndicatorOfSlice(v.slicePtr)
	zlog.Assert(v.indicatorFieldName != "")
	if v.indicatorFieldName != "" && v.params.Flags&FlagSkipIndicator != 0 {
		zstr.AddToSet(&v.params.SkipFieldNames, v.indicatorFieldName)
	}
	v.updateMenu()
	v.menu.SetSelectedHandler(v.handleMenuSelected)
}

func (v *FieldSliceView) updateMenu() {
	if v.menu == nil {
		return
	}
	var menuItems zdict.Items
	rval := reflect.ValueOf(v.slicePtr).Elem()
	for i := 0; i < rval.Len(); i++ {
		a := rval.Index(i).Interface()
		rval, got := zreflect.FindFieldWithNameInStruct(v.indicatorFieldName, a, true)
		zlog.Assert(got)
		str := fmt.Sprint(rval.Interface())
		menuItems.Add(str, i)
	}
	v.menu.UpdateItems(menuItems, v.currentIndex)
}

func (v *FieldSliceView) handleAddItem() {
	i := zslice.AddEmptyElementAtEnd(v.slicePtr)
	e := reflect.ValueOf(v.slicePtr).Elem().Index(i)
	initer, _ := e.Addr().Interface().(StructInitializer)
	if initer != nil {
		initer.InitZFieldStruct()
	}
	v.addItem(i, e, false)
	v.selectItem(i)
}

func (v *FieldSliceView) addItem(i int, rval reflect.Value, collapse bool) {
	var itemStack *zcontainer.StackView
	exp := zgeo.VertExpand
	if v.Vertical {
		exp = zgeo.HorExpand
	}
	itemStack = v.stack
	if !v.field.IsStatic() {
		itemStack = zcontainer.StackViewNew(!v.Vertical, "item-stack")
		v.stack.Add(itemStack, zgeo.TopLeft|exp).Collapsed = collapse
		collapse = false
	}
	var add zview.View
	if v.isStructItems {
		id := strconv.Itoa(i)
		fv := FieldViewNew(id, rval.Addr().Interface(), v.params)
		fv.SetMargin(zgeo.RectFromXY2(4, 4, -4, -4))
		fv.parent = &v.FieldView
		if v.field.Flags&FlagGroupSingle == 0 || v.field.IsStatic() {
			fv.SetCorner(5)
			fv.SetBGColor(zstyle.DefaultFGColor().WithOpacity(0.1))
		}
		add = fv
		fv.Build(true)
		v.params.AddTrigger("*", EditedAction, func(fv *FieldView, f *Field, value any, view *zview.View) bool {
			// zlog.Info("Trigger:", f.FieldName, v.indicatorFieldName)
			if f.FieldName == v.indicatorFieldName {
				v.updateMenu()
			}
			v.callEditedAction()
			return true
		})
	} else {
		special, skip := v.createSpecialView(rval, v.field)
		if !skip {
			setter, _ := special.(zview.AnyValueSetter)
			if setter != nil {
				setter.SetValueWithAny(rval.Interface())
			}
			add = special
		}
	}
	zlog.Assert(add != nil)
	itemStack.Add(add, zgeo.TopLeft|exp).Collapsed = collapse
	if v.field.Flags&FlagGroupSingle == 0 && !v.field.IsStatic() {
		deleteButton := makeButton("minus", "red")
		itemStack.Add(deleteButton, zgeo.CenterRight)
		deleteButton.SetPressedHandler(func() {
			findView := add
			if itemStack != v.stack {
				findView = itemStack
			}
			_, fi := v.stack.FindCellWithView(findView)
			if fi != -1 {
				v.handleDeleteItem(fi)
			}
		})
	}
	v.updateMenu() //
}

func (v *FieldSliceView) callEditedAction() {
	actionCaller, _ := v.slicePtr.(ActionHandler)
	if actionCaller != nil {
		view := v.View
		actionCaller.HandleAction(v.field, EditedAction, &view)
	}
}

func (v *FieldSliceView) handleDeleteItem(i int) {
	zlog.Assert(i >= 0 && i < len(v.stack.Cells), i, len(v.stack.Cells))
	zslice.RemoveAt(v.slicePtr, i)
	cell := v.stack.Cells[i]
	v.stack.RemoveChild(cell.View)
	v.currentIndex = -1
	v.selectItem(0)
	v.callEditedAction()
}

func (v *FieldSliceView) handleMenuSelected() {
	v.selectItem(v.menu.CurrentValue().(int))
}

func (v *FieldSliceView) selectItem(i int) {
	if v.currentIndex < len(v.stack.Cells) && v.currentIndex != -1 {
		cell := v.stack.Cells[v.currentIndex]
		if v.field.Flags&FlagGroupSingle != 0 && !v.field.IsStatic() {
			v.stack.CollapseChild(cell.View, true, false)
		}
	}
	v.currentIndex = i
	if i < len(v.stack.Cells) {
		cell := v.stack.Cells[v.currentIndex]
		if v.field.Flags&FlagGroupSingle != 0 && !v.field.IsStatic() {
			v.stack.CollapseChild(cell.View, false, false)
			zkeyvalue.DefaultSessionStore.SetInt(v.currentIndex, v.storeKey, true)
		}
		zcontainer.FocusNext(cell.View, true, true)
	}
	v.updateMenu()
	zcontainer.ArrangeChildrenAtRootContainer(v)
}

func (v *FieldSliceView) UpdateSlice(slicePtr any) {
	var focusedPath string
	focused := v.GetFocusedChildView(false)
	if focused != nil {
		focusedPath = v.GetPathOfChild(focused)
	}
	if slicePtr != nil {
		v.slicePtr = slicePtr
	}
	v.RemoveAllChildren()
	v.build()
	f := zview.ChildOfViewFunc(v, focusedPath) // use v.View here to get proper underlying container type in ChildOfViewFunc
	if f != nil {
		f.Native().Focus(true)
	}
}

func makeButton(shape, col string) *zimageview.ImageView {
	str := fmt.Sprintf("images/zcore/%s-circled-dark%s.png", shape, col)
	return zimageview.New(nil, str, zgeo.Size{20, 20})
}