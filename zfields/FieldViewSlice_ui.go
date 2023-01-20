//go:build zui
// +build zui

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
	ct := old.Native().Parent().View.(zcontainer.ContainerType)
	ct.ReplaceChild(old, newView)
	ctp := old.Native().Parent().View.(zcontainer.ContainerType)
	zpresent.CallReady(newView, true)
	ctp.ArrangeChildren()
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
	var indicatorFieldName string
	val := reflect.ValueOf(slicePtr).Elem()
	s := reflect.New(reflect.TypeOf(val.Interface()).Elem()).Interface()
	// fmt.Printf("CreateSliceGroupOwner %s %+v\n", grouper.GetGroupBase().Hierarchy(), s)
	zreflect.ForEachField(s, func(index int, val reflect.Value, sf reflect.StructField) {
		for _, part := range zreflect.GetTagAsMap(string(sf.Tag))["zui"] {
			if part == "-" {
				return
			}
			if part == "indicator" {
				indicatorFieldName = sf.Name
				if params.Flags&FlagSkipIndicator != 0 {
					zstr.AddToSet(&params.SkipFieldNames, indicatorFieldName)
				}
				return
			}
		}
	})
	// fmt.Printf("CreateSliceGroupOwner2 %s\n", indicatorFieldName)
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
	stack.SetSpacing(v.params.Styling.Spacing)
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

func PresentOKCancelStructSlice[S any](structSlicePtr *[]S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) bool) {
	sliceVal := reflect.ValueOf(structSlicePtr).Elem()
	first := sliceVal.Index(0)
	editStruct := reflect.New(first.Type())
	// zlog.Info("editStruct:", editStruct.Type(), editStruct.Kind(), editStruct)
	editStruct.Elem().Set(first)
	len := len(*structSlicePtr)
	unknownBoolViewIDs := map[string]bool{}
	zreflect.ForEachField(editStruct.Interface(), func(index int, val reflect.Value, sf reflect.StructField) {
		if sf.Tag.Get("zui") == "-" {
			return
		}
		var notEqual bool
		for i := 0; i < len; i++ {
			sliceField := sliceVal.Index(i).Field(index)
			if !reflect.DeepEqual(sliceField.Interface(), val.Interface()) {
				// zlog.Info(i, index, "not-equal", sliceField.Interface(), val.Interface())
				notEqual = true
				val.Set(reflect.Zero(val.Type()))
				break
			}
		}
		if notEqual && val.Kind() == reflect.Bool {
			unknownBoolViewIDs[fieldNameToID(sf.Name)] = true
		}
	})
	params.EditWithoutCallbacks = true
	params.UseInValues = []string{"$dialog"}
	fview := FieldViewNew("OkCancel", editStruct.Interface(), params)
	update := true
	fview.Build(update)
	for bid := range unknownBoolViewIDs {
		view, _ := fview.findNamedViewOrInLabelized(bid)
		check := view.(*zcheckbox.CheckBox)
		check.SetValue(zbool.Unknown)
	}
	zalert.PresentOKCanceledView(fview, title, att, func(ok bool) bool {
		if ok {
			err := fview.ToData(true)
			if err != nil {
				return false
			}
			zreflect.ForEachField(editStruct.Interface(), func(index int, val reflect.Value, sf reflect.StructField) {
				if sf.Tag.Get("zui") == "-" {
					return // skip to next
				}
				bid := fieldNameToID(sf.Name)
				view, _ := fview.findNamedViewOrInLabelized(bid)
				check, _ := view.(*zcheckbox.CheckBox)
				isCheck := (check != nil)
				if isCheck && check.Value().IsUnknown() {
					return // skip to next
				}
				for i := 0; i < len; i++ {
					sliceField := sliceVal.Index(i).Field(index)
					// zlog.Info("SetSliceVal?:", isCheck, unknownBoolViewIDs, i, sf.Name, val.Interface(), sliceField.Addr())
					if !val.IsZero() || isCheck {
						sliceField.Set(val)
					}
				}
			})
		}
		return done(ok)
	})
}
