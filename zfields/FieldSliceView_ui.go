//go:build zui

package zfields

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zguiutil"
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
	field              *Field
	menu               *zmenu.MenuView
	indicatorFieldName string
	currentIndex       int
	stack              *zcontainer.StackView
	storeKey           string
	isCompositeItems   bool
}

func (fv *FieldView) NewSliceView(slicePtr any, f *Field) *FieldSliceView {
	vert := !f.Vertical.Bool()
	v := &FieldSliceView{}
	v.data = slicePtr
	v.Init(v, vert, f.FieldName)
	v.field = f
	v.ParentFV = fv
	v.params = fv.params
	v.currentIndex = -1
	v.params.Field.MergeInField(f)
	rt := reflect.ValueOf(slicePtr)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	var isInterface bool
	len := rt.Len()
	if rt.Len() == 0 {
		rt = zslice.MakeAnElementOfSliceRValType(rt)
		zlog.Info("Zero:", rt.Kind(), rt.Interface())
	} else {
		rt = rt.Index(0)
		if rt.Kind() == reflect.Interface {
			isInterface = true
			rt = reflect.ValueOf(rt.Interface())

		}
		zlog.Info("NFV:", rt.Kind(), rt.Interface())
	}
	kind := rt.Kind()

	// _, isDict := slice.Rval.
	v.isCompositeItems = (kind == reflect.Struct || kind == reflect.Slice)
	if !v.isCompositeItems {
		v.params.Field.ClearFlag(FlagIsLabelize)
		vert = !vert
	}
	if isInterface {
		vert = true
	}
	// zlog.Info("NewSliceView:", len, v.isCompositeItems, kind, v.params.Field.HasFlag(FlagIsLabelize))
	v.build(f.IsStatic())
	return v
}

func (v *FieldSliceView) build(addItems bool) {
	var header *zcontainer.StackView
	var index int
	// zlog.Info("FieldSliceView build:", v.Hierarchy(), v.data != nil, reflect.ValueOf(v.data).Kind())

	sliceRval := reflect.ValueOf(v.data)
	if sliceRval.Kind() == reflect.Pointer {
		sliceRval = sliceRval.Elem()
	}
	if v.field.Flags&FlagHasFrame != 0 {
		var title string
		if v.field.Flags&FlagFrameIsTitled != 0 {
			title = v.field.TitleOrName()
		}
		header = zguiutil.MakeStackATitledFrame(&v.StackView, title, v.field.Flags&FlagFrameTitledOnFrame != 0, v.field.Styling, v.field.Styling)
	}
	if header == nil && !v.field.IsStatic() {
		header = zcontainer.StackViewHor("header")
		v.Add(header, zgeo.TopLeft|zgeo.HorExpand)
	}
	maybeAddDocToHeader(v.field, header)
	if v.field.IsStatic() {
		v.stack = &v.StackView
	} else {
		if v.field.Description != "" {
			label := zlabel.New("desc")
			label.SetFont(zgeo.FontNice(-3, zgeo.FontStyleNormal))
			label.SetColor(zgeo.ColorDarkGray.WithOpacity(0.7))
			label.SetText(v.field.Description)
			label.SetTextAlignment(zgeo.BottomRight)
			v.Add(label, zgeo.BottomRight|zgeo.HorExpand)
		}
		v.stack = zcontainer.StackViewNew(v.Vertical, "stack")
		v.Add(v.stack, zgeo.TopLeft|zgeo.Expand)
	}
	var collapse bool
	if v.field.Flags&FlagGroupSingle != 0 && !v.field.IsStatic() {
		collapse = true
		v.createMenu()
		header.Add(v.menu, zgeo.CenterRight)

		if !v.field.IsStatic() && !v.field.HasFlag(FlagIsFixed) {
			v.globalDeleteButton = makeButton("minus", "red")
			header.Add(v.globalDeleteButton, zgeo.CenterRight)
			v.globalDeleteButton.SetPressedHandler("", zkeyboard.ModifierNone, func() {
				v.handleDeleteItem(v.currentIndex)
			})
		}
		v.storeKey = v.CreateStoreKeyForField(v.field, "SliceIndex")
		index, _ = zkeyvalue.DefaultSessionStore.GetInt(v.storeKey, 0)
		zint.Minimize(&index, zint.Max(0, sliceRval.Len()-1))
		v.menu.SelectWithValue(index)
	}
	if !v.field.IsStatic() && !v.field.HasFlag(FlagIsFixed) {
		v.addButton = makeButton("plus", "gray")
		header.Add(v.addButton, zgeo.CenterRight)
		v.addButton.SetPressedHandler("", zkeyboard.ModifierNone, v.handleAddItem)
	}
	if addItems {
		for i := 0; i < sliceRval.Len(); i++ {
			// zlog.Info("SliceFView.Add:", i, v.Hierarchy(), collapse)
			v.addItem(i, sliceRval.Index(i), collapse)
		}
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

	v.indicatorFieldName = FindIndicatorOfSlice(v.data)
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
	rval := reflect.ValueOf(v.data).Elem()
	for i := 0; i < rval.Len(); i++ {
		a := rval.Index(i).Interface()
		finfo, found := zreflect.FieldForName(a, FlattenIfAnonymousOrZUITag, v.indicatorFieldName)
		zlog.Assert(found)
		str := fmt.Sprint(finfo.ReflectValue.Interface())
		menuItems.Add(str, i)
	}
	v.menu.UpdateItems(menuItems, v.currentIndex, false)
}

func (v *FieldSliceView) handleAddItem() {
	i := zslice.AddEmptyElementAtEnd(v.data)
	e := reflect.ValueOf(v.data).Elem().Index(i)
	CallStructInitializer(e.Addr().Interface())
	v.addItem(i, e, false)
	v.selectItem(i)
}

func (v *FieldSliceView) addItem(i int, rval reflect.Value, collapse bool) {
	exp := zgeo.VertExpand
	if v.Vertical {
		exp = zgeo.HorExpand
	}
	itemStack := v.stack
	if !v.field.IsStatic() {
		itemStack = zcontainer.StackViewNew(!v.Vertical, fmt.Sprint("zitem-stack", i))
		v.stack.Add(itemStack, zgeo.TopLeft|exp).Collapsed = collapse
		collapse = false
	}
	var add zview.View

	if v.isCompositeItems {
		id := strconv.Itoa(i)
		copyParams := v.params
		// copyParams.Labelize = false
		fv := FieldViewNew(id, rval.Addr().Interface(), copyParams)
		fv.sliceItemIndex = i
		fv.Vertical = !v.Vertical || v.field.HasFlag(FlagIsLabelize)
		// zlog.Info("FieldSliceView:AddItem", v.Hierarchy(), id)
		fv.SetMargin(zgeo.RectFromXY2(4, 4, -4, -4))
		fv.ParentFV = &v.FieldView
		if v.field.Flags&FlagGroupSingle == 0 || v.field.IsStatic() {
			fv.SetCorner(5)
			fv.SetBGColor(zstyle.DefaultFGColor().WithOpacity(0.1))
		}
		add = fv
		fv.Build(true)
		// fv.params.AddTrigger("*", EditedAction, func(fv *FieldView, f *Field, value any, view *zview.View) bool {
		// 	zlog.Info("FVTrigger:", v.Hierarchy(), f.FieldName, zlog.Pointer(v.data))
		// 	return false
		// })
		v.params.AddTrigger("*", EditedAction, func(ap ActionPack) bool {
			// zlog.Info("Trigger:", fv.Hierarchy(), f.FieldName, zlog.Pointer(v.data))
			if ap.Field.FieldName == v.indicatorFieldName {
				v.updateMenu()
			}
			// v.callEditedAction() // we return false below instead
			return false
		})
		itemStack.Add(add, zgeo.TopLeft|exp).Collapsed = collapse
	} else {
		simple := v.buildItem(v.field, rval, -1, zgeo.CenterLeft, zgeo.Size{}, false)
		simple.SetObjectName(fmt.Sprintf("%s.%d", v.ObjectName(), i))
		setter, _ := simple.(zview.AnyValueSetter)
		if setter != nil {
			setter.SetValueWithAny(rval.Interface())
		}
		add = simple
	}
	zlog.Assert(add != nil)
	if v.field.Flags&FlagGroupSingle == 0 && !v.field.IsStatic() && !v.field.HasFlag(FlagIsFixed) {
		deleteButton := makeButton("minus", "red")
		itemStack.Add(deleteButton, zgeo.CenterRight, zgeo.SizeD(3, 0))
		deleteButton.SetPressedHandler("", zkeyboard.ModifierNone, func() {
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
	actionCaller, _ := v.data.(ActionHandler)
	if actionCaller != nil {
		view := v.View
		actionCaller.HandleAction(ActionPack{FieldView: &v.FieldView, Field: v.field, Action: EditedAction, View: &view})
	}
}

func (v *FieldSliceView) handleDeleteItem(i int) {
	zlog.Assert(i >= 0 && i < len(v.stack.Cells), i, len(v.stack.Cells))
	zslice.RemoveAt(v.data, i)
	cell := v.stack.Cells[i]
	v.stack.RemoveChild(cell.View, true)
	v.currentIndex = -1
	v.selectItem(0)
	v.callEditedAction()
}

func (v *FieldSliceView) GetNthSubFieldViewInNonStatic(n int, childName string) zview.View {
	if n >= len(v.stack.Cells) {
		return nil
	}
	itemStack := v.stack.Cells[n].View.(*zcontainer.StackView)
	zlog.Assert(len(itemStack.Cells) == 2)
	view := itemStack.Cells[0].View
	if childName != "" {
		view, _ = zcontainer.ContainerOwnerFindViewWithName(view, childName, false)
	}
	return view
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
	}
	v.updateMenu()
	if v.IsPresented() {
		toModalWindowRootOnly := false
		zcontainer.ArrangeChildrenAtRootContainer(v, toModalWindowRootOnly)
	}
}

func (v *FieldSliceView) UpdateSlice(f *Field, slicePtr any) {
	// zlog.Info("FieldSliceView.UpdateSlice", v.Hierarchy(), reflect.TypeOf(slicePtr))
	var focusedPath string
	focused := v.GetFocusedChildView(false)
	if focused != nil {
		focusedPath = v.GetPathOfChild(focused)
	}
	if slicePtr != nil {
		v.data = slicePtr
	}
	v.RemoveAllChildren()
	v.build(true)
	if focusedPath != "" {
		focused = zview.ChildOfViewFunc(v, focusedPath) // use v.View here to get proper underlying container type in ChildOfViewFunc
		if focused != nil {
			focused.Native().Focus(true)
		}
	}
}

func makeButton(shape, col string) *zimageview.ImageView {
	str := fmt.Sprintf("images/zcore/%s-circled-dark%s.png", shape, col)
	return zimageview.New(nil, true, str, zgeo.SizeD(20, 20))
}
