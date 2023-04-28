//go:build zui

package zfields

import (
	"reflect"
	"strconv"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zgroup"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
)

func (v *FieldView) updateSliceElementData(slicePtr any, stack *zcontainer.StackView) {
	val := reflect.ValueOf(slicePtr).Elem()
	children := stack.GetChildren(true)
	for i := 0; i < val.Len(); i++ {
		fv, _ := children[i].(*FieldView)
		if fv != nil {
			a := val.Index(i).Addr().Interface()
			fv.data = a
			// fmt.Printf("updateSliceElementData: %d %s %p\n", i, v.Hierarchy(), a)
		} else {
			//	zlog.Error(nil, "updateSliceElementData bad type:", i, val.Index(i).Type()) // it can be some other widget or something
		}
	}
}

func (v *FieldView) updateSliceValue(slicePtr any, stack *zcontainer.StackView, vertical bool, f *Field, sendEdited bool) {
	newStack := v.buildRepeatedStackFromSlice(slicePtr, vertical, f)
	replaceRebuildSliceView(stack, newStack)
	if sendEdited {
		fh, _ := slicePtr.(ActionHandler)
		if fh != nil {
			fh.HandleAction(f, EditedAction, &newStack.View)
		}
	}
	zcontainer.ArrangeChildrenAtRootContainer(stack)
}

func replaceRebuildSliceView(old zview.View, newView zview.View) {
	v := old.Native().Parent().View
	cr := v.(zview.ChildReplacer)
	cr.ReplaceChild(old, newView)
	ca := v.(zcontainer.Arranger)
	zpresent.CallReady(newView, true)
	ca.ArrangeChildren()
	zpresent.CallReady(newView, false)
}

func (v *FieldView) makeNamedSelectionKey(f *Field) string {
	// zlog.Info("makeNamedSelectKey:", v.id, f.FieldName)
	return v.id + "." + f.FieldName + ".NamedSelectionIndex"
}

func (v *FieldView) changeNamedSelectionIndex(i int, f *Field) {
	key := v.makeNamedSelectionKey(f)
	zkeyvalue.DefaultStore.SetInt(i, key, true)
}

func createGroupSliceViewFunc(slicePtr any, params FieldViewParameters, id string, delete bool) zview.View {
	if delete {
		return nil
	}
	// zlog.Info("Create Func:", id, val.Len())
	val := reflect.ValueOf(slicePtr).Elem()
	for i := 0; i < val.Len(); i++ {
		var sid string
		va := val.Index(i)
		a := va.Interface()
		getter, _ := a.(zstr.StrIDer)
		if getter != nil {
			sid = getter.GetStrID()
		} else {
			sid = strconv.Itoa(i)
		}
		if sid == id {
			fview := FieldViewNew(id, va.Addr().Interface(), params)
			update := true
			fview.Build(update)
			return fview
		}
	}
	return nil
}

func CreateSliceGroup(grouper zgroup.Grouper, slicePtr any, parameters FieldViewParameters) {
	params := parameters
	indicatorFieldName := FindIndicatorOfSlice(slicePtr)
	if indicatorFieldName != "" && params.Flags&FlagSkipIndicator != 0 {
		zstr.AddToSet(&params.SkipFieldNames, indicatorFieldName)
	}
	zlog.Assert(indicatorFieldName != "")
	gb := grouper.GetGroupBase()
	gb.UpdateCurrentIndicatorFunc = func(text string) {
		gm, _ := gb.View.(*zgroup.MenuGroupView)
		// zlog.Info("Changes Group parent:", gb.Hierarchy(), text, gm != nil)
		gm.UpdateCurrentItemTitle(text)
	}
	zgroup.CreateSliceGroup(grouper, slicePtr, "", indicatorFieldName, func(id string, delete bool) zview.View {
		return createGroupSliceViewFunc(slicePtr, parameters, id, delete)
	})
	if params.Flags&FlagSkipIndicator == 0 {
		data, _ := gb.Data.(*zgroup.SliceGroupData)
		data.IndicatorID = indicatorFieldName
	}
	old := gb.HandleDeleteItemFunc
	gb.HandleDeleteItemFunc = func(id string) {
		s := zslice.MakeAnElementOfSliceType(slicePtr)
		fh, _ := s.(ActionHandler)
		// zlog.Info("Edit Action on delete:", fh != nil, parameters.Field.Name)
		if fh != nil {
			fh.HandleAction(&parameters.Field, EditedAction, &gb.View)
		}
		// zlog.Info("Delete Item:", id)
		old(id)
	}
}

func buildMenuGroup(slicePtr any, storeKey string, params FieldViewParameters) *zgroup.MenuGroupView {
	// zlog.Info("FV:buildMenuGroup storeKey:", storeKey)
	mv := zgroup.MenuGroupViewNew("menugroup", params.Field.TitleOrName(), params.Field.Styling, params.Field.Styling)
	mv.AddEditing()
	mv.StoreKey = storeKey
	CreateSliceGroup(mv, slicePtr, params)
	mv.StoreKey = storeKey
	return mv
}

func repopulateMenuGroup(mg *zgroup.MenuGroupView, slicePtr any, params FieldViewParameters) {
	data := mg.Data.(*zgroup.SliceGroupData)
	// zlog.Info("Repop:", mg.GetCurrentID())
	currentID := mg.GetCurrentID() // we need to store currentID, as mg.Empty() below clears it
	mg.Empty()
	// mg.SetBGColor(zgeo.ColorRandom())
	zgroup.AddSliceItems(mg, slicePtr, currentID, data.IndicatorID, func(id string, delete bool) zview.View {
		return createGroupSliceViewFunc(slicePtr, params, id, delete)
	})
	mg.ArrangeChildren()
	zpresent.CallReady(mg, false)
}

func (v *FieldView) buildRepeatedStackFromSlice(slicePtr any, vertical bool, f *Field) *zcontainer.StackView {
	stack := zcontainer.StackViewNew(vertical, f.ID+"-stack")
	stack.SetSpacing(f.Styling.Spacing)
	var fieldView *FieldView
	val := reflect.ValueOf(slicePtr).Elem()
	// zlog.Info("buildStackFromSlice:", vertical, f.ID, val.Len())
	for i := 0; i < val.Len(); i++ {
		nval := val.Index(i)
		var view zview.View
		a := zgeo.Center
		if vertical {
			a = zgeo.TopLeft
		}
		// zlog.Info("buildStackFromSlice:", vertical, f.ID, f.WidgetName)
		if f.WidgetName != "" {
			w := widgeters[f.WidgetName]
			if w != nil {
				view = w.Create(f)
				stack.Add(view, a)
				w.SetValue(view, nval.Interface())
			}
		}
		h, _ := nval.Interface().(ActionHandler)
		if view == nil && h != nil {
			if h.HandleAction(f, CreateFieldViewAction, &view) {
				stack.Add(view, a)
				// fmt.Println("buildStackFromSlice element:", f.FieldName)
			}
		}
		if view == nil {
			childStruct := nval.Addr().Interface()
			// fmt.Printf("Build sub slice element: %s %p\n", v.id, childStruct)
			vert := !vertical
			if !f.Vertical.IsUnknown() {
				vert = f.Vertical.Bool()
			}
			if f.LabelizeWidth != 0 {
				vert = true
			}
			// fmt.Printf("buildStackFromSlice element: %s %p\n", f.FieldName, childStruct)
			params := FieldViewParametersDefault()
			fieldView = fieldViewNew(strconv.Itoa(i), vert, childStruct, params, zgeo.Size{}, v)
			view = fieldView
			// fieldView.parentField = f
			a := zgeo.Left //| zgeo.HorExpand
			if fieldView.Vertical {
				a |= zgeo.Top
			} else {
				a |= zgeo.VertCenter
			}
			fieldView.BuildStack(f.ID, a, zgeo.Size{}, true)
			if !f.IsStatic() {
				trash := zimageview.New(nil, "images/minus-circled-darkgray.png", zgeo.Size{16, 16})
				fieldView.Add(trash, zgeo.CenterRight)
				index := i // remember i so it's correct in below function
				trash.SetPressedHandler(func() {
					zslice.RemoveAt(v.data, index)
					// zlog.Info("newlen:", index, val.Len())
					v.updateSliceValue(v.data, stack, vertical, f, true)
				})
			}
			stack.Add(fieldView, zgeo.TopLeft|zgeo.HorExpand)
		}
		zlog.Assert(view != nil)
	}
	if f.MinWidth == 0 && f.Size.W != 0 {
		flen := float64(val.Len())
		f.MinWidth = f.Size.W*flen + f.Styling.Spacing*(flen-1)
	}
	if !f.IsStatic() {
		plus := zimageview.New(nil, "images/plus-circled-darkgray.png", zgeo.Size{16, 16})
		plus.SetPressedHandler(func() {
			a := reflect.New(val.Type().Elem()).Elem()
			nv := reflect.Append(val, a)
			if fieldView != nil {
				fieldView.data = nv.Interface()
			}
			val.Set(nv)
			a = val.Index(val.Len() - 1) // we re-set a, as it is now a new value at the end of slice
			// fmt.Printf("SLICER + Pressed: %p %p\n", val.Index(val.Len()-1).Addr().Interface(), a.Addr().Interface())
			fhItem, _ := a.Addr().Interface().(ActionHandler)
			if fhItem != nil {
				fhItem.HandleAction(f, NewStructAction, nil)
			}
			v.updateSliceValue(v.data, stack, vertical, f, true)
		})
		stack.Add(plus, zgeo.TopLeft)
	}
	return stack
}

func removeNotUsedEverywhereSliceItems() {
	// if val.Kind() == reflect.Slice {
	// 	for i := 0; i < val.Len(); i++ {
	// 		for j := 0; j < sliceVal.Len(); j++ {
	// 			var has bool
	// 			sliceField := sliceVal.Index(j).Field(index)
	// 			for k := 0; k < sliceField.Len(); k++ {
	// 				if reflect.DeepEqual(val.Interface(), sliceField.Index(k).Interface()) {
	// 					has = true
	// 					break
	// 				}
	// 			}
	// 			if has {
	// 				i++
	// 			} else {
	// 				zslice.RemoveAt(sliceVal.Addr().Interface(), j)
	// 			}
	// 		}
	// 	}
	// }
	// zslice.AddEmptyElementAtEnd(sliceVal.Addr().Interface())
}

func accumilateSlice(accSlice, fromSlice reflect.Value) {
	for i := 0; i < fromSlice.Len(); i++ {
		var has bool
		av := fromSlice.Index(i).Interface()
		for j := 0; j < accSlice.Len(); j++ {
			if reflect.DeepEqual(av, accSlice.Index(j).Interface()) {
				has = true
				break
			}
		}
		if !has {
			zslice.AddAtEnd(accSlice.Addr().Interface(), av)
		}
	}
}

func reduceLocalEnumField[S any](editStruct *S, enumField reflect.Value, index int, fromStruct reflect.Value, f *Field) {
	ei, findex := FindLocalFieldWithID(editStruct, f.LocalEnum)
	if zlog.ErrorIf(findex == -1, f.Name, f.LocalEnum) {
		return
	}
	zlog.Assert(ei.Kind() == reflect.Slice)
	fromVal, _ := zreflect.FieldForIndex(fromStruct.Interface(), true, findex)
	// zlog.Info("reduceLocalEnumField", f.Name, findex, ei.Len(), ei.Type(), fromStruct.Type(), fromStruct.Kind())
	var reduce, hasZero bool
	for i := 0; i < ei.Len(); {
		eval := ei.Index(i)
		if eval.IsZero() {
			hasZero = true
		}
		var has bool
		for j := 0; j < fromVal.Len(); j++ {
			if reflect.DeepEqual(eval.Interface(), fromVal.Index(j).Interface()) {
				has = true
				break
			}
		}
		if has {
			i++
		} else {
			zslice.RemoveAt(ei.Addr().Interface(), i)
			reduce = true
		}
	}
	if ei.Len() == 0 && fromVal.Len() > 0 {
		reduce = true
	}
	// zlog.Info("REDUCE?", f.Name, reduce, hasZero)
	if reduce && !hasZero {
		zslice.AddEmptyElementAtEnd(ei.Addr().Interface())
		// enumField.Set(reflect.Zero(enumField.Type()))
	}
}

// var tempFieldEnums = map[string]zdict.Items{}

// func reduceEnumField(editVal, indexVal reflect.Value, enumName string) {
// 	tempEnum, _ := tempFieldEnums[enumName]
// 	if tempEnum == nil {
// 		tempEnum = zdict.Items{}
// 		tempEnum = fieldEnums[enumName]
// 		tempFieldEnums[enumName] = tempEnum
// 	}
// 	zlog.Info("reduceEnumField", editVal, indexVal, enumName)
// 	var reduced bool
// 	var isZero bool
// 	for i := 0; i < len(tempEnum); i++ {
// 		ival := tempEnum[i].Value
// 		if ival != nil && reflect.ValueOf(ival).IsZero() {
// 			isZero = true
// 		}
// 		if !reflect.DeepEqual(ival, indexVal.Interface()) {
// 			reduced = true
// 			zslice.RemoveAt(&tempEnum, i)
// 			i--
// 		}
// 	}
// 	if reduced && !isZero {
// 		zslice.AddEmptyElementAtEnd(&tempEnum)
// 		editVal.Set(reflect.Zero(editVal.Type()))
// 	}
// }

func reduceSliceField(reduceSlice, fromSlice reflect.Value) {
	var reduced bool
	for i := 0; i < reduceSlice.Len(); {
		rval := reduceSlice.Index(i).Interface()
		var has bool
		for j := 0; i < fromSlice.Len(); j++ {
			fval := fromSlice.Index(j).Interface()
			if reflect.DeepEqual(rval, fval) {
				has = true
				break
			}
		}
		if has {
			i++
		} else {
			zslice.RemoveAt(reduceSlice.Addr().Interface(), i)
			reduced = true
		}
	}
	if reduced {
		zslice.AddEmptyElementAtEnd(reduceSlice.Addr().Interface())
	}
}

func PresentOKCancelStructSlice[S any](structSlicePtr *[]S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) bool) {
	sliceVal := reflect.ValueOf(structSlicePtr).Elem()
	first := (*structSlicePtr)[0] // we want a copy, so do in two stages
	editStruct := &first
	length := len(*structSlicePtr)
	unknownBoolViewIDs := map[string]bool{}

	ForEachField(editStruct, params.FieldParameters, nil, func(index int, f *Field, val reflect.Value, sf reflect.StructField) {
		var notEqual bool
		for i := 0; i < length; i++ {
			sliceField := sliceVal.Index(i).Field(index)
			if !sliceField.CanInterface() || !val.CanInterface() {
				continue
			}
			zlog.Info(f.Name, val.Interface())
			if !reflect.DeepEqual(sliceField.Interface(), val.Interface()) {
				// zlog.Info(f.Name, i, index, "not-equal", sliceField.Interface(), val.Interface())
				if f.IsStatic() {
					if val.Kind() == reflect.Slice {
						accumilateSlice(val, sliceField)
					}
				} else {
					// if f.Enum != "" {
					// 	reduceEnumField(val, sliceField, f.Enum)
					// } else
					if f.LocalEnum != "" {
						reduceLocalEnumField(editStruct, val, index, sliceVal.Index(i), f)
					} else if val.Kind() == reflect.Slice {
						reduceSliceField(val, sliceField)
					} else {
						val.Set(reflect.Zero(val.Type()))
					}
				}
				notEqual = true
				break
			}
		}
		// zlog.Info("ForEach:", f.Name, val.Type(), val.Interface(), notEqual, f.Enum)
		if notEqual {
			if val.Kind() == reflect.Bool {
				unknownBoolViewIDs[sf.Name] = true
				// zslice.AddEmptyElementAtEnd(val.Addr().Interface())
			}
		}
		return
	})
	params.EditWithoutCallbacks = true
	params.MultiSliceEditInProgress = (len(*structSlicePtr) > 1)
	params.UseInValues = []string{"$dialog"}
	fview := FieldViewNew("OkCancel", editStruct, params)
	update := true
	fview.Build(update)
	for bid := range unknownBoolViewIDs {
		view, _ := fview.findNamedViewOrInLabelized(bid)
		check, _ := view.(*zcheckbox.CheckBox)
		if check != nil {
			check.SetValue(zbool.Unknown)
		}
	}
	zalert.PresentOKCanceledView(fview, title, att, func(ok bool) bool {
		if ok {
			err := fview.ToData(true)
			if err != nil {
				return false
			}
			// zlog.Info("EDITAfter2data:", zlog.Full(editStruct))
			zreflect.ForEachField(editStruct, true, func(index int, val reflect.Value, sf reflect.StructField) bool {
				if sf.Tag.Get("zui") == "-" {
					return true // skip to next
				}
				// zlog.Info("PresentOKCanceledView foreach:", sf.Name)
				bid := sf.Name
				view, _ := fview.findNamedViewOrInLabelized(bid)
				check, _ := view.(*zcheckbox.CheckBox)
				isCheck := (check != nil)
				if isCheck && check.Value().IsUnknown() {
					return true // skip to next
				}
				for i := 0; i < length; i++ {
					sliceField := sliceVal.Index(i).Field(index)
					// zlog.Info("SetSliceVal?:", i, sf.Name, val.Interface(), sliceField.Interface(), val.IsZero())
					if !val.IsZero() || length == 1 || isCheck {
						sliceField.Set(val)
					}
				}
				return true
			})
		}
		return done(ok)
	})
}
