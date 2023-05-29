//go:build zui

package zfields

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zclipboard"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zgroup"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbits"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zguiutil"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
	"github.com/torlangballe/zutil/zwords"
)

type FieldView struct {
	zcontainer.StackView
	parent *FieldView
	Fields []Field
	// parentField  *Field
	data     interface{}
	dataHash int64
	id       string
	params   FieldViewParameters

	grouper zgroup.Grouper
}

type FieldViewParameters struct {
	Field
	FieldParameters
	ImmediateEdit            bool
	MultiSliceEditInProgress bool
	EditWithoutCallbacks     bool // Set so not get edit/changed callbacks when editing. Example: Dialog box editing edits a copy, so no callbacks needed.
	triggerHandlers          map[trigger]func(fv *FieldView, f *Field, value any, view *zview.View) bool
}

type ActionHandler interface {
	HandleAction(f *Field, action ActionType, view *zview.View) bool
}

var fieldViewEdited = map[string]time.Time{}

func FieldViewParametersDefault() (f FieldViewParameters) {
	f.ImmediateEdit = true
	f.Styling = zstyle.EmptyStyling
	f.Styling.Spacing = 4
	return f
}

func (v *FieldView) ID() string {
	return v.id
}

func (v *FieldView) Data() any {
	return v.data
}

func (v *FieldView) IsSlice() bool {
	return reflect.ValueOf(v.data).Elem().Kind() == reflect.Slice
}

func setFieldViewEdited(fv *FieldView) {
	// zlog.Info("FV Edited:", fv.Hierarchy())
	fieldViewEdited[fv.Hierarchy()] = time.Now()
}

func IsFieldViewEditedRecently(fv *FieldView) bool {
	h := fv.Hierarchy()
	t, got := fieldViewEdited[h]
	if got {
		if time.Since(t) < time.Second*3 {
			return true
		}
		delete(fieldViewEdited, h)
	}
	return false
}

func makeFrameIfFlag(f *Field, child zview.View) zview.View {
	if f.Flags&FlagHasFrame == 0 {
		return nil
	}
	var title string
	if f.Flags&FlagFrameIsTitled != 0 {
		title = f.TitleOrName()
	}
	frame := zcontainer.StackViewVert("frame")
	zgroup.MakeStackTitledFrame(frame, title, f.Flags&FlagFrameTitledOnFrame != 0, f.Styling, f.Styling)
	frame.Add(child, zgeo.TopLeft)
	return frame
}

func fieldViewNew(id string, vertical bool, data any, params FieldViewParameters, marg zgeo.Size, parent *FieldView) *FieldView {
	// start := time.Now()
	v := &FieldView{}
	v.StackView.Init(v, vertical, id)
	v.SetSpacing(params.Styling.Spacing)

	// zlog.Info("fieldViewNew", id, parent != nil, zlog.CallingStackString())
	v.data = data
	zreflect.ForEachField(v.data, true, func(index int, val reflect.Value, sf reflect.StructField) bool {
		f := EmptyField
		if !f.SetFromReflectValue(val, sf, index, params.ImmediateEdit) {
			return true
		}
		// zlog.Info("fieldViewNew f:", f.Name, f.UpdateSecs)
		callActionHandlerFunc(v, &f, SetupFieldAction, val.Addr().Interface(), nil)
		v.Fields = append(v.Fields, f)
		return true
	})
	if params.Field.Flags&FlagHasFrame == 0 {
		v.SetMargin(zgeo.RectFromMinMax(marg.Pos(), marg.Pos().Negative()))
	}
	v.params = params
	v.id = id
	v.parent = parent

	return v
}

func (v *FieldView) Build(update bool) {
	a := zgeo.Left //| zgeo.HorExpand
	if v.Vertical {
		a |= zgeo.Top
	} else {
		a |= zgeo.VertCenter
	}
	v.BuildStack(v.ObjectName(), a, zgeo.Size{}, true)
	if update {
		dontOverwriteEdited := false
		v.Update(nil, dontOverwriteEdited)
	}
}

func (v *FieldView) findNamedViewOrInLabelized(name string) (view, maybeLabel zview.View) {
	for _, c := range (v.View.(zcontainer.ChildrenOwner)).GetChildren(false) {
		n := c.ObjectName()
		if n == name {
			return c, c
		}
		if strings.HasPrefix(n, "$labelize.") {
			s, _ := c.(*zcontainer.StackView)
			if s != nil {
				v, _ := s.FindViewWithName(name, true)
				if v != nil {
					return v, c
				}
			}
		}
	}
	return nil, nil
}

func (v *FieldView) updateShowEnableFromZeroer(isZero, isShow bool, toID string) {
	for _, f := range v.Fields {
		var id string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		// zlog.Info("updateShowEnableFromZeroer:", v.Hierarchy(), f.FieldName, isZero, isShow, id, toID, "==", local)
		if zstr.HasPrefix(local, "./", &id) && id == toID {
			_, foundView := v.findNamedViewOrInLabelized(f.FieldName)
			if foundView == nil {
				continue
			}
			// zlog.Assert(foundView != nil, v.Hierarchy(), f.FieldName)
			if neg {
				isShow = !isShow
			}
			if isShow {
				foundView.Show(!isZero)
			} else {
				foundView.SetUsable(!isZero)
			}
			continue
		}
	}
	//TODO: handle ../ and substruct/id style
}

func doShowEnableItem(rval reflect.Value, isShow bool, view zview.View, not bool) {
	// zlog.Info("doShowEnableItem:", item.FieldName, isShow, not, view.Native().Hierarchy())
	zero := rval.IsZero()
	if not {
		zero = !zero
	}
	if isShow {
		view.Show(!zero)
	} else {
		view.SetUsable(!zero)
	}
}

func getLocalFromShowOrEnable(isShow bool, f *Field) (local string, neg bool) {
	if isShow {
		if f.LocalShow != "" {
			local = f.LocalShow
		} else {
			local = f.LocalHide
			neg = true
		}
	} else {
		if f.LocalEnable != "" {
			local = f.LocalEnable
		} else {
			local = f.LocalDisable
			neg = true
		}
	}
	return
}

func (v *FieldView) updateShowEnableOnView(view zview.View, isShow bool, toFieldName string) {
	// zlog.Info("updateShowOrEnable:", isShow, toID, len(v.Fields))
	for _, f := range v.Fields {
		if f.FieldName != toFieldName {
			continue
		}
		var prefix, fname string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		if zstr.HasPrefix(local, "./", &fname) {
			// zlog.Info("local:", toID, f.FieldName, id)
			rval, _, findex := zreflect.FieldForName(v.data, true, fname)
			if findex != -1 {
				doShowEnableItem(rval, isShow, view, neg)
			}
			continue
		}
		if zstr.SplitN(local, "/", &prefix, &fname) && prefix == f.FieldName {
			fv := view.(*FieldView)
			if fv == nil {
				zlog.Error(nil, "updateShowOrEnable: not field view:", f.FieldName, local, v.ObjectName)
				return
			}
			rval, _, findex := zreflect.FieldForName(v.data, true, fname)
			if findex != -1 {
				doShowEnableItem(rval, isShow, view, neg)
			}
			continue
		}
		if zstr.HasPrefix(local, "../", &fname) && v.parent != nil {
			rval, _, findex := zreflect.FieldForName(v.parent.Data(), true, fname)
			if findex != -1 {
				doShowEnableItem(rval, isShow, view, neg)
			}
			continue
		}
	}
}

func (v *FieldView) Update(data any, dontOverwriteEdited bool) {
	if data != nil { // must be after IsFieldViewEditedRecently, or we set new data without update slice pointers and maybe more
		v.data = data
	}
	// zlog.Info("fv.Update:", v.Hierarchy(), dontOverwriteEdited)
	if dontOverwriteEdited && IsFieldViewEditedRecently(v) {
		zreflect.ForEachField(v.data, true, func(index int, rval reflect.Value, sf reflect.StructField) bool {
			if sf.Type.Kind() == reflect.Slice {
				v.updateField(index, rval, sf, dontOverwriteEdited) // even if edited recently, we call v.updateField on slices, to set new address of each slice in FieldView.data
			}
			return true
		})
		zlog.Info("FV No Update because recently edited", v.Hierarchy())
		return
	}
	fh, _ := v.data.(ActionHandler)
	sview := v.View
	if fh != nil {
		fh.HandleAction(nil, DataChangedActionPre, &sview)
	}
	// fmt.Println("FV Update", v.id, len(children))
	// fmt.Printf("FV Update: %s %d %+v\n", v.id, len(children), v.data)
	zreflect.ForEachField(v.data, true, func(index int, rval reflect.Value, sf reflect.StructField) bool {
		v.updateField(index, rval, sf, dontOverwriteEdited)
		return true
	})
	// call general one with no id. Needs to be after above loop, so values set
	if fh != nil {
		fh.HandleAction(nil, DataChangedAction, &sview)
	}
}

func (v *FieldView) updateField(index int, rval reflect.Value, sf reflect.StructField, dontOverwriteEdited bool) bool {
	var valStr string
	f := findFieldWithIndex(&v.Fields, index)
	if f == nil {
		// zlog.Info("FV Update no index found:", index, v.id)
		return true
	}
	foundView, flabelized := v.findNamedViewOrInLabelized(f.FieldName)
	// zlog.Info("fv.UpdateF:", v.Hierarchy(), index, f.ID, f.FieldName, foundView != nil)
	if foundView == nil {
		// zlog.Info("FV Update no view found:", i, v.id, f.ID)
		return true
	}
	v.updateShowEnableOnView(flabelized, true, foundView.ObjectName())
	v.updateShowEnableOnView(flabelized, false, foundView.ObjectName())
	var called bool
	tri, _ := rval.Interface().(TriggerDataChangedTriggerer)
	if tri != nil {
		called = tri.HandleDataChange(v, f, rval.Addr().Interface(), &foundView)
	}
	if !called {
		called = v.callTriggerHandler(f, DataChangedAction, rval.Addr().Interface(), &foundView)
	}
	if !called {
		called = callActionHandlerFunc(v, f, DataChangedAction, rval.Addr().Interface(), &foundView)
	}
	// zlog.Info("fv.Update:", v.ObjectName(), f.ID, called)
	if called {
		// fmt.Println("FV Update called", v.id, f.Kind, f.ID)
		return true
	}
	if f.WidgetName != "" && f.Kind != zreflect.KindSlice {
		w := widgeters[f.WidgetName]
		// zlog.Info("fv update !slice:", f.Name, reflect.ValueOf(foundView).Type(), w != nil, rval.Interface())
		if w != nil {
			w.SetValue(foundView, rval.Interface())
			return true
		}
	}
	menuType, _ := foundView.(zmenu.MenuType)
	if menuType == nil {
		o := zmenu.OwnerForView(foundView)
		if o != nil {
			menuType = o
		}
	}
	// zlog.Info("fv.Update2 menu?:", v.ObjectName(), menuType != nil)
	if menuType != nil && ((f.Enum != "") || f.LocalEnum != "") { // && f.Kind != zreflect.KindSlice
		var enum zdict.Items
		// zlog.Info("Update FV: Menu:", f.Name, f.Enum, f.LocalEnum)
		if f.Enum != "" {
			enum, _ = fieldEnums[f.Enum]
		} else {
			ei, findex := FindLocalFieldWithID(v.data, f.LocalEnum)
			zlog.Assert(findex != -1, f.Name, f.LocalEnum)
			enum = ei.Interface().(zdict.ItemsGetter).GetItems()
		}
		if v.params.ForceZeroOption {
			var rtype reflect.Type
			var hasZero bool
			for _, item := range enum {
				if item.Value != nil {
					rval := reflect.ValueOf(item)
					rtype = rval.Type()
					if rval.IsZero() {
						hasZero = true
						break
					}
				}
			}
			if !hasZero && rtype.Kind() != reflect.Invalid {
				item := zdict.Item{Name: "", Value: reflect.Zero(rtype)}
				enum = append(zdict.Items{item}, enum...)
			}
		}
		// zlog.Info("Update FV: Menu3:", f.Name, reflect.ValueOf(rval.Addr().Interface()).Elem(), reflect.TypeOf(rval.Addr().Interface()).Elem()) // enum
		menuType.UpdateItems(enum, rval.Interface(), f.Flags&FlagIsActions != 0)
		return true
	}
	updateItemLocalToolTip(f, v.data, foundView)
	if f.IsStatic() || v.params.AllStatic {
		zuistringer, _ := rval.Interface().(UIStringer)
		if zuistringer != nil {
			label, _ := foundView.(*zlabel.Label)
			if label != nil {
				label.SetText(zuistringer.ZUIString())
				return true
			}
		}
	}
	switch f.Kind {
	case zreflect.KindSlice:
		// zlog.Info("updateSliceFieldView:", v.Hierarchy())
		// val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
		// fmt.Printf("updateSliceFieldView: %s %p %p %v %p\n", v.id, item.Interface, val.Interface(), found, foundView)
		// if f.WidgetName != "" && rval.Type().Kind() != reflect.Struct {
		// 	w := widgeters[f.WidgetName]
		// 	if w != nil {
		// 		i := 0
		// 		stack := foundView.Native().Parent()
		// 		zlog.Info("fv update widget slice:", f.Name, rval.Len(), reflect.TypeOf(stack), foundView.Native().Hierarchy(), stack.Hierarchy(), zcontainer.CountChildren(stack))
		// 		zcontainer.ViewRangeChildren(stack, false, false, func(view zview.View) bool {
		// 			if i >= rval.Len() {
		// 				return false
		// 			}
		// 			w.SetValue(view, rval.Index(i).Interface())
		// 			i++
		// 			return true
		// 		})
		// 		return true
		// 	}
		// }
		fv, _ := foundView.(*FieldView)
		if fv == nil {
			return false
		}
		// zlog.Info("updateSliceFieldView2:", v.Hierarchy(), zlog.Pointer(fv.data), zlog.Pointer(rval.Addr().Interface()))
		fv.data = rval.Addr().Interface()

		hash := zstr.HashAnyToInt64(reflect.ValueOf(fv.data).Elem(), "")
		// zlog.Info("update any SliceValue:", f.Name, hash, fv.dataHash, reflect.ValueOf(fv.data).Elem())
		sameHash := (fv.dataHash == hash)
		fv.dataHash = hash
		// getter, _ := item.Interface.(zdict.ItemsGetter)
		// zlog.Info("fv update slice:", f.Name, reflect.ValueOf(foundView).Type(), getter != nil)
		// if !sameHash && getter != nil {
		// 	items := getter.GetItems()
		// 	mt := foundView.(zmenu.MenuType)
		// 	// zlog.Info("fv update slice:", f.Name, len(items), mt != nil, reflect.ValueOf(foundView).Type())
		// 	if mt != nil {
		// 		// assert menu is static...
		// 		mt.UpdateItems(items, nil)
		// 	}
		// }
		if f.Flags&FlagIsGroup == 0 {
			stack := fv.GetChildren(true)[0].(*zcontainer.StackView)
			// zlog.Info("updateSliceElementData?:", sameHash)
			if sameHash {
				fv.updateSliceElementData(rval.Addr().Interface(), stack)
				break
			}
			if menuType != nil {
				break
			}
			vert := v.Vertical
			if v.params.LabelizeWidth != 0 {
				vert = false
			}
			fv.updateSliceValue(rval.Addr().Interface(), stack, vert, f, false)
		} else {
			// zlog.Info("UpdateGSlice:", fv.Hierarchy(), item.Interface)
			mg := fv.grouper.(*zgroup.MenuGroupView) // we have to convert it to a MenuGroupView for replace to work, for comparison
			if !sameHash {
				changedIDs := mg.HasDataChangedInIDsFunc()
				if len(changedIDs) != 0 && !(len(changedIDs) == 1 && changedIDs[0] == mg.GetCurrentID()) { // if it's not the id of current slice element, we di a full rebuild
					// zlog.Info("UpdateSliceChanged in other item:", changedIDs)
					repopulateMenuGroup(mg, rval.Addr().Interface(), fv.params)
					break
				}
			}
			updateMenuGroupSlice(mg, rval.Addr().Interface(), dontOverwriteEdited)
		}
	case zreflect.KindTime:
		tv, _ := foundView.(*ztext.TextView)
		if tv != nil && tv.IsEditing() {
			break
		}
		if f.Flags&FlagIsDuration != 0 {
			// val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
			// if found {
			// t := val.Interface().(time.Time)
			// fmt.Println("FV Update Time Dur", v.id, time.Since(t))
			// }
			v.updateSinceTime(foundView.(*zlabel.Label), f)
			break
		}
		valStr = getTimeString(rval, f)
		to := foundView.(ztext.LayoutOwner)
		to.SetText(valStr)

	case zreflect.KindStruct:
		// zlog.Info("Update Struct:", item.FieldName)
		fv, _ := foundView.(*FieldView)
		if fv == nil {
			fv = findSubFieldView(foundView, "")
		}
		if fv == nil {
			break
		}
		// zlog.Info("Update Struct2:", item.FieldName, fv.Hierarchy())
		fv.Update(rval.Addr().Interface(), dontOverwriteEdited)

	case zreflect.KindBool:
		if f.Flags&FlagIsImage != 0 && f.IsImageToggle() && rval.Kind() == reflect.Bool {
			iv, _ := foundView.(*zimageview.ImageView) // it might be a button or something instead
			path := f.OffImagePath
			val, _ := v.fieldToDataItem(f, iv)
			on := val.Bool()
			if on {
				path = f.ImageFixedPath
			}
			iv.SetImage(nil, path, nil)
			break
		}
		cv, _ := foundView.(*zcheckbox.CheckBox) // it might be a button or something instead
		if cv != nil {
			b := zbool.ToBoolInd(rval.Interface().(bool))
			v := cv.Value()
			if v != b {
				cv.SetValue(b)
			}
		}

	case zreflect.KindInt, zreflect.KindFloat:
		_, got := rval.Interface().(zbits.BitsetItemsOwner)
		if got {
			updateFlagStack(rval, f, foundView)
		}
		// zlog.Info("FV Update Int:", v.Hierarchy(), f.Name, item)
		valStr = getTextFromNumberishItem(rval, f)
		if f.IsStatic() || v.params.AllStatic {
			label, _ := foundView.(*zlabel.Label)
			if label != nil {
				label.SetText(valStr)
			}
			break
		}
		tv, _ := foundView.(*ztext.TextView)
		if tv != nil {
			if tv.IsEditing() {
				break
			}
			tv.SetText(valStr)
		}

	case zreflect.KindString, zreflect.KindFunc:
		valStr = rval.String()
		if f.Flags&FlagIsImage != 0 {
			// zlog.Info("FVUpdate SETIMAGE:", f.Name, str)
			path := ""
			if f.Kind == zreflect.KindString {
				path = valStr
			}
			if path != "" && strings.Contains(f.ImageFixedPath, "*") {
				path = strings.Replace(f.ImageFixedPath, "*", path, 1)
			} else if path == "" || f.Flags&FlagIsFixed != 0 {
				path = f.ImageFixedPath
			}
			io := foundView.(zimage.Owner)
			io.SetImage(nil, path, nil)
		} else {
			if f.IsStatic() || v.params.AllStatic {
				label, _ := foundView.(*zlabel.Label)
				// zlog.Info("UpdateText:", item.FieldName, item.Interface, label != nil)
				if label != nil {
					if f.Flags&FlagIsFixed != 0 {
						valStr = f.Name
					}
					label.SetText(valStr)
				}
			} else {
				tv, _ := foundView.(*ztext.TextView)
				if tv != nil {
					if tv.IsEditing() {
						break
					}
					tv.SetText(valStr)
				}
			}
		}
	}
	gb := zgroup.GetAncestorGroupBase(foundView)
	if gb != nil {
		// zlog.Info("UpdateIndicator:", foundView.Native().Hierarchy(), gb.Hierarchy())
		data := gb.Data.(*zgroup.SliceGroupData)
		if data.IndicatorID == f.FieldName {
			if gb.UpdateCurrentIndicatorFunc != nil {
				gb.UpdateCurrentIndicatorFunc(fmt.Sprint(valStr))
			}
		}
	}
	return true
}
func updateMenuGroupSlice(mg *zgroup.MenuGroupView, slicePtr any, dontOverwriteEdited bool) {
	data := mg.Data.(*zgroup.SliceGroupData)
	data.SlicePtr = slicePtr
	id := mg.GetCurrentID()
	if id != "" {
		fv := mg.ChildView.(*FieldView)
		i := zgroup.IndexForIDFromSlice(slicePtr, id)
		if i != -1 {
			//	zlog.Assert(i != -1, id, reflect.ValueOf(slicePtr).Elem().Len())
			a := reflect.ValueOf(slicePtr).Elem().Index(i).Addr().Interface()
			fv.Update(a, dontOverwriteEdited)
		} else {
			str := zstr.Concat(" ", "updateMenuGroupSlice unknown item id", id, reflect.ValueOf(slicePtr).Elem().Len())
			if zui.DebugOwnerMode {
				zlog.Fatal(nil, str)
			} else {
				zlog.Error(nil, str)
			}
		}
	}
}

func findSubFieldView(view zview.View, optionalID string) (fv *FieldView) {
	zcontainer.ViewRangeChildren(view, true, true, func(view zview.View) bool {
		f, _ := view.(*FieldView)
		if f != nil {
			if optionalID == "" || f.ID() == optionalID {
				fv = f
				return false
			}
		}
		return true
	})
	return
}

func FieldViewNew(id string, data any, params FieldViewParameters) *FieldView {
	v := fieldViewNew(id, true, data, params, zgeo.Size{10, 10}, nil)
	// zlog.Info("FieldViewNew:", id)
	return v
}

func (v *FieldView) Rebuild() {
	fview := FieldViewNew(v.id, v.data, v.params)
	fview.Build(true)
	rep, _ := v.Parent().View.(zview.ChildReplacer)
	zlog.Info("REP:", rep != nil, v.Parent().Hierarchy())
	if rep != nil {
		rep.ReplaceChild(v, fview)
	}
	zcontainer.ArrangeChildrenAtRootContainer(v)
}

func (v *FieldView) CallFieldAction(fieldID string, action ActionType, fieldValue interface{}) {
	view, _ := v.FindViewWithName(fieldID, false)
	if view == nil {
		zlog.Error(nil, "CallFieldAction find view", fieldID)
		return
	}
	f := v.findFieldWithID(fieldID)
	if f == nil {
		zlog.Error(nil, "CallFieldAction find field", fieldID)
		return
	}
	callActionHandlerFunc(v, f, action, fieldValue, &view)
}

// callActionHandlerFunc can have v with only data set for SetupFieldAction
func callActionHandlerFunc(v *FieldView, f *Field, action ActionType, fieldValue interface{}, view *zview.View) bool {
	if action == EditedAction {
		if f.SetEdited {
			setFieldViewEdited(v)
		}
		gb := zgroup.GetAncestorGroupBase(v.View)
		if gb != nil && gb.UpdateCurrentIndicatorFunc != nil {
			data := gb.Data.(*zgroup.SliceGroupData)
			if data.IndicatorID == f.FieldName {
				gb.UpdateCurrentIndicatorFunc(fmt.Sprint(fieldValue))
			}
		}
		if v.params.EditWithoutCallbacks {
			return true
		}
		// fmt.Printf("call edit ActionHandlerFunc2: %s %p\n", f.Name, v.data)
	}
	// zlog.Info("callActionHandlerFunc:", f.FieldName, f.Name, action)
	direct := (action == CreateFieldViewAction || action == SetupFieldAction)
	// zlog.Info("callActionHandlerFunc  get sub:", f.FieldName, f.Name, action)
	// data := v.getSubStruct(structID, direct)
	fh, _ := v.data.(ActionHandler)
	// zlog.Info("callFieldHandler1", action, f.Name, fh != nil, reflect.TypeOf(v.data))
	var result bool
	if fh != nil {
		result = fh.HandleAction(f, action, view)
	}
	// zlog.Info("callActionHandlerFunc2:", f.FieldName, f.Name, action)

	if view != nil && *view != nil {
		first := true
		n := (*view).Native()
		for n != nil {
			parent := n.Parent()
			if parent != nil {
				fv, _ := parent.View.(*FieldView)
				// var sss string
				// if fv != nil {
				// 	sss = fmt.Sprint(reflect.TypeOf(fv.data))
				// }
				// zlog.Info("callFieldHandler parent", action, f.Name, parent.ObjectName(), fv != nil, reflect.TypeOf(parent.View), sss)
				if fv != nil {
					if fv.IsSlice() {
						fv.dataHash = zstr.HashAnyToInt64(reflect.ValueOf(fv.data).Elem(), "")
						// zlog.Info("Edit Set Hash", fv.Hierarchy(), fv.dataHash)
					}
					if !first {
						fh2, _ := fv.data.(ActionHandler)
						if fh2 != nil {
							fh2.HandleAction(nil, action, &parent.View)
						}
					}
					first = false
				}
			}
			n = parent
		}
	}

	if !result {
		var fieldAddress interface{}
		if !direct {
			changed := false
			sv := reflect.ValueOf(v.data)
			// zlog.Info("\n\nNew struct search for children?:", f.FieldName, sv.Kind(), sv.CanAddr(), data != nil)
			if sv.Kind() == reflect.Ptr || sv.CanAddr() {
				// Here we run thru the possiblly new struct again, and find the item with same id as field
				// s := data
				// if sv.Kind() != reflect.Ptr {
				// 	s = sv.Addr().Interface()
				// }
				fieldVal, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
				if found {
					changed = true
					fieldValue = fieldVal.Interface()
					if fieldVal.CanAddr() {
						fieldAddress = fieldVal.Addr().Interface()
					}
				}
				// options := zreflect.Options{UnnestAnonymous: true, Recursive: false}
				// items, err := zreflect.ItterateStruct(s, options)
				// // zlog.Info("New struct search for children:", f.FieldName, len(items.Children), err)
				// if err != nil {
				// 	zlog.Fatal(err, "children action")
				// }
				// for _, c := range items.Children {
				// 	// zlog.Info("New struct search for children find:", f.FieldName, c.FieldName)
				// 	if c.FieldName == f.FieldName {
				// 		// zlog.Info("New struct search for children got:", f.FieldName)
				// 		fieldValue = c.Interface
				// 		changed = true
				// 	}
				// }
			}
			if !changed {
				zlog.Info("NOOT!!!", f.Name, action, v.data != nil)
				zlog.Fatal(nil, "Not CHANGED!", f.Name)
			}
		}
		aih, _ := fieldValue.(ActionHandler)
		// vvv := reflect.ValueOf(fieldValue)
		// if aih == nil {
		// 	rval := reflect.ValueOf(fieldValue)
		// 	// zlog.Info("callActionHandlerFunc", f.Name, rval.Kind(), rval.Type(), rval.CanAddr())
		// 	if rval.Kind() != reflect.Ptr && rval.CanAddr() {
		// 		inter := rval.Addr().Interface()
		// 		aih, _ = inter.(ActionFieldHandler)
		// 	}
		// }
		if aih == nil && fieldAddress != nil {
			aih, _ = fieldAddress.(ActionHandler)
		}
		// zlog.Info("callActionHandlerFunc bottom:", f.Name, action, aih != nil, reflect.ValueOf(fieldValue).Type(), reflect.ValueOf(fieldValue).Kind())
		if aih != nil {
			result = aih.HandleAction(f, action, view)
			// zlog.Info("callActionHandlerFunc bottom:", f.Name, action, result, view, aih)
		}
	}
	// zlog.Info("callActionHandlerFunc top done:", f.ID, f.Name, action)
	return result
}

func (v *FieldView) findFieldWithID(id string) *Field {
	for i, f := range v.Fields {
		if f.ID == id {
			return &v.Fields[i]
		}
	}
	return nil
}

func (fv *FieldView) makeButton(rval reflect.Value, f *Field) *zshape.ImageButtonView {
	// zlog.Info("makeButton:", f.Name, f.Height)
	format := f.Format
	if format == "" {
		format = "%v"
	}
	color := "gray"
	if len(f.Colors) > 0 {
		color = f.Colors[0]
	}
	name := f.Name
	if f.Title != "" && fv.params.LabelizeWidth == 0 {
		name = f.Title
	}
	s := zgeo.Size{20, 24}
	if f.Height != 0 {
		s.H = f.Height
	}
	button := zshape.ImageButtonViewNew(name, color, s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
	button.SetTextColor(zgeo.ColorBlack)
	button.TextXMargin = 0
	return button
}

func (v *FieldView) makeMenu(rval reflect.Value, f *Field, items zdict.Items) zview.View {
	var view zview.View
	static := f.IsStatic() || v.params.AllStatic
	isSlice := rval.Kind() == reflect.Slice
	if static || isSlice {
		multi := isSlice
		// zlog.Info("FV Menu Make static:", f.ID, f.Format, f.Name)
		isImage := (f.ImageFixedPath != "")
		shape := zshape.TypeRoundRect
		if isImage {
			shape = zshape.TypeNone
		}
		menuOwner := zmenu.NewMenuedOwner()
		menuOwner.IsStatic = static
		menuOwner.IsMultiple = multi
		menuOwner.StoreKey = f.ValueStoreKey
		for _, format := range strings.Split(f.Format, "|") {
			if menuOwner.TitleIsAll == " " {
				menuOwner.TitleIsAll = format
			}
			switch format {
			case "all":
				menuOwner.TitleIsAll = " "
				// zlog.Info("MakeMenu All:", f.Name, f.Format)
			case "%d":
				menuOwner.GetTitleFunc = func(icount int) string { return strconv.Itoa(icount) }
			case `%n`:
				menuOwner.TitleIsValueIfOne = true
			default:
				menuOwner.PluralableWord = format
			}
		}
		mItems := zmenu.MOItemsFromZDictItemsAndValues(items, nil, f.Flags&FlagIsActions != 0)
		menu := zmenu.MenuOwningButtonCreate(menuOwner, mItems, shape)
		if isImage {
			menu.SetImage(nil, f.ImageFixedPath, nil)
			menu.ImageMaxSize = f.Size
		} else {
			// menu.SetPillStyle()
			if len(f.Colors) != 0 {
				// zlog.Info("SETMENIBG:", f.Colors[0])
				menu.SetColor(zgeo.ColorFromString(f.Colors[0]))
			}
		}
		menu.Ratio = 0.3
		view = menu
		// zlog.Info("Make Menu Format", f.Name, f.Format)
		menuOwner.SelectedHandlerFunc = func() {
			// v.fieldToDataItem(f, menu)
			if menuOwner.IsStatic {
				sel := menuOwner.SelectedItem()
				if sel != nil {
					kind := reflect.ValueOf(sel.Value).Kind()
					// zlog.Info("action pressed", kind, sel.Name, "val:", sel.Value)
					if kind != reflect.Ptr && kind != reflect.Struct {
						nf := *f
						nf.ActionValue = sel.Value
						callActionHandlerFunc(v, &nf, PressedAction, rval.Interface(), &view)
					}
				}
			}
			v.callTriggerHandler(f, EditedAction, rval.Interface(), &view)
			callActionHandlerFunc(v, f, EditedAction, rval.Interface(), &view)
		}
		menuOwner.ClosedFunc = func() {
			if menuOwner.IsMultiple {
				zlog.Assert(isSlice)
				zslice.Empty(rval.Addr().Interface())
				// zlog.Info("CopyBack menu1:", reflect.TypeOf(item.Interface))
				for _, mi := range menuOwner.SelectedItems() {
					// zlog.Info("CopyBack menu:", mi.Name, mi.Value)
					zslice.AddAtEnd(rval.Addr().Interface(), mi.Value)
				}
			}
		}
	} else {
		menu := zmenu.NewView(f.Name+"Menu", items, rval.Interface())
		menu.SetMaxWidth(f.MaxWidth)
		view = menu
		menu.SetSelectedHandler(func() {
			// valInterface, _ := v.fieldToDataItem(f, menu, false)
			v.fieldToDataItem(f, menu)
			v.callTriggerHandler(f, EditedAction, rval.Interface(), &view)
			callActionHandlerFunc(v, f, EditedAction, rval.Interface(), &view)
		})
	}
	return view
}

func getTimeString(rval reflect.Value, f *Field) string {
	var str string
	t := rval.Interface().(time.Time)
	if t.IsZero() {
		return ""
	}
	format := f.Format
	secs := (f.Flags&FlagHasSeconds != 0)
	if format == "" {
		format = "15:04"
		if secs {
			format = "15:04:03"
		}
		if zlocale.DisplayServerTime.Get() {
			format += "-07"
		}
		format += " 02-Jan-06"
	}
	if format == "nice" {
		str = ztime.GetNice(t, f.Flags&FlagHasSeconds != 0)
		// zlog.Info("Nice:", t, f.Name, ztime.DefaultDisplayServerTime)
	} else {
		t = ztime.GetTimeWithServerLocation(t)
		str = t.Format(format)
	}
	return str
}

func getTextFromNumberishItem(rval reflect.Value, f *Field) string {
	// zlog.Info("getTextFromNumberishItem", f.Name, f.Flags&FlagZeroIsEmpty != 0, rval.IsZero())
	if f.Flags&FlagZeroIsEmpty != 0 {
		if rval.IsZero() {
			return ""
		}
	}
	stringer, got := rval.Interface().(UIStringer)
	if got {
		return stringer.ZUIString()
	}
	zkind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	isDurTime := zkind == zreflect.KindTime && f.Flags&FlagIsDuration != 0
	// zlog.Info("makeTextTime:", f.Name, isDurTime, f.Format)
	if zkind == zreflect.KindTime && !isDurTime {
		return getTimeString(rval, f)
	}
	if isDurTime || f.PackageName == "time" && rval.Type().Name() == "Duration" {
		var t float64
		if isDurTime {
			t = ztime.Since(rval.Interface().(time.Time))
		} else {
			t = ztime.DurSeconds(time.Duration(rval.Int()))
		}
		return ztime.GetSecsAsHMSString(t, f.Flags&FlagHasSeconds != 0, 0)
		// zlog.Info("makeTextTime:", str, f.Name)
	}
	format := f.Format
	significant := f.Columns
	switch format {
	case "memory":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			return zwords.GetMemoryString(b, "", significant)
		}
	case "storage":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			return zwords.GetStorageSizeString(b, "", significant)
		}
	case "bps":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			return zwords.GetBandwidthString(b, "", significant)
		}
	case "":
		format = "%v"
	}
	return fmt.Sprintf(format, rval.Interface())
}

func (v *FieldView) makeText(rval reflect.Value, f *Field, noUpdate bool) zview.View {
	// fmt.Printf("make Text %s %s %s %p\n:", item.FieldName, f.Title, f.Name, v.data)
	str := getTextFromNumberishItem(rval, f)
	if f.IsStatic() || v.params.AllStatic {
		// zlog.Info("makeText:", f.FieldName, str)
		label := zlabel.New(str)
		label.SetMaxLines(f.Rows)
		if f.Flags&FlagIsDuration != 0 {
			v.updateSinceTime(label, f) // we should really not do getTextFromNumberishItem above if we do this
		}
		j := f.Justify
		if j == zgeo.AlignmentNone {
			j = f.Alignment & (zgeo.Left | zgeo.HorCenter | zgeo.Right)
			if j == zgeo.AlignmentNone {
				j = zgeo.Left
			}
		}
		// label.SetMaxLines(strings.Count(str, "\n") + 1)
		f.SetFont(label, nil)
		label.SetTextAlignment(j)
		label.SetWrap(ztextinfo.WrapTailTruncate)
		if f.Flags&FlagToClipboard != 0 {
			label.SetPressedHandler(func() {
				text := label.Text()
				zclipboard.SetString(text)
				label.SetText("📋 " + text)
				ztimer.StartIn(0.6, func() {
					label.SetText(text)
				})
			})
		}
		return label
	}
	var style ztext.Style
	cols := f.Columns
	if cols == 0 {
		cols = 20
	}
	zkind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	if f.Flags&FlagIsPassword != 0 {
		style.KeyboardType = zkeyboard.TypePassword
	} else if zkind == zreflect.KindInt {
		style.KeyboardType = zkeyboard.TypeInteger
	} else if zkind == zreflect.KindFloat {
		style.KeyboardType = zkeyboard.TypeFloat
	}
	if f.Flags&FlagDisableAutofill != 0 {
		style.DisableAutoComplete = true
	}
	if f.Format == "email" {
		style.KeyboardType = zkeyboard.TypeEmailAddress
	}
	tv := ztext.NewView(str, style, cols, f.Rows)
	tv.SetObjectName(f.FieldName)
	f.SetFont(tv, nil)
	tv.UpdateSecs = f.UpdateSecs
	if !noUpdate && tv.UpdateSecs == -1 {
		tv.UpdateSecs = 4
	}
	tv.SetPlaceholder(f.Placeholder)
	tv.SetChangedHandler(func() {
		v.fieldToDataItem(f, tv)
		// fmt.Printf("Changed text1: %s %p %+v\n", f.FieldName, v.data, reflect.ValueOf(v.data).Elem().Interface())
		view := zview.View(tv)
		callActionHandlerFunc(v, f, EditedAction, tv.Text(), &view)
	})
	// tv.SetKeyHandler(func(key zkeyboard.Key, mods zkeyboard.Modifier) bool {
	// zlog.Info("keyup!")
	// })
	// zlog.Info("FV makeText:", f.FieldName, tv.MinWidth, tv.Columns)
	return tv
}

func (v *FieldView) makeCheckbox(f *Field, b zbool.BoolInd) zview.View {
	cv := zcheckbox.New(b)
	cv.SetObjectName(f.FieldName)
	cv.SetValueHandler(func() {
		val, _ := v.fieldToDataItem(f, cv)
		v.updateShowEnableFromZeroer(val.IsZero(), true, cv.ObjectName())
		v.updateShowEnableFromZeroer(val.IsZero(), false, cv.ObjectName())
		view := zview.View(cv)
		if v.callTriggerHandler(f, EditedAction, cv.On(), &cv.View) {
			return
		}
		callActionHandlerFunc(v, f, EditedAction, val.Interface(), &view)
	})
	if v.params.LabelizeWidth == 0 && !zstr.StringsContain(v.params.UseInValues, "$row") {
		_, stack := zcheckbox.Labelize(cv, f.TitleOrName())
		return stack
	}
	return cv
}

func (v *FieldView) makeImage(rval reflect.Value, f *Field) zview.View {
	iv := zimageview.New(nil, "", f.Size)
	iv.DownsampleImages = true
	iv.SetMinSize(f.Size)
	iv.SetObjectName(f.FieldName)
	iv.OpaqueDraw = (f.Flags&FlagIsOpaque != 0)
	if f.Styling.FGColor.Valid {
		iv.EmptyColor = f.Styling.FGColor
	}
	if f.IsImageToggle() && rval.Kind() == reflect.Bool {
		iv.SetPressedHandler(func() {
			val, _ := v.fieldToDataItem(f, iv)
			on := val.Bool()
			on = !on
			val.SetBool(on)
			v.Update(nil, false)
			// zlog.Info("CallBoolImageTrigger")
			v.callTriggerHandler(f, EditedAction, on, &iv.View)
			callActionHandlerFunc(v, f, EditedAction, val.Interface(), &iv.View)

			// zlog.Info("CallBoolImageTrigger done")
		})
	} else {
		iv.SetPressedHandler(func() {
			v.callTriggerHandler(f, PressedAction, v.id, &iv.View)
		})
	}
	return iv
}

func setColorFromField(view zview.View, f *Field) {
	col := zstyle.DefaultFGColor()
	if f.Styling.FGColor.Valid {
		col = f.Styling.FGColor
	}
	view.SetColor(col)
}

func (v *FieldView) updateOldTime(label *zlabel.Label, f *Field) {
	val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
	if found {
		t := val.Interface().(time.Time)
		if ztime.Since(t) > float64(f.OldSecs) {
			label.SetColor(zgeo.ColorRed)
		} else {
			setColorFromField(label, f)
		}
	}
}

func (v *FieldView) updateSinceTime(label *zlabel.Label, f *Field) {
	if zlog.IsInTests { // if in unit-tests, we don't show anything as it would change
		label.SetText("")
		return
	}
	sv := reflect.ValueOf(v.data)
	// zlog.Info("\n\nNew struct search for children?:", f.FieldName, sv.Kind(), sv.CanAddr(), v.data != nil)
	zlog.Assert(sv.Kind() == reflect.Ptr || sv.CanAddr())
	// Here we run thru the possiblly new struct again, and find the item with same id as field
	// s := data
	// if sv.Kind() != reflect.Ptr {
	// 	s = sv.Addr().Interface()
	// }
	val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
	if found {
		var str string
		t := val.Interface().(time.Time)
		tooBig := false
		if !t.IsZero() {
			// zlog.Info("UpdateDur:", f.Name, f.Flags&FlagHasSeconds != 0)
			since := time.Since(t)
			str, tooBig = ztime.GetDurationString(since, f.Flags&FlagHasSeconds != 0, f.Flags&FlagHasMinutes != 0, f.Flags&FlagHasHours != 0, f.FractionDecimals)
		}
		inter := val.Interface()
		if val.CanAddr() {
			inter = val.Addr()
		}
		if tooBig {
			label.SetText("●")
			label.SetColor(zgeo.ColorRed)
		} else {
			label.SetText(str)
			setColorFromField(label, f)
		}
		callActionHandlerFunc(v, f, DataChangedAction, inter, &label.View)
	}
}

func makeFlagStack(v *FieldView, flags reflect.Value, f *Field) zview.View {
	stack := zcontainer.StackViewHor("flags")
	stack.SetMinSize(zgeo.Size{20, 20})
	spacing := f.Styling.MergeWith(v.params.Styling).Spacing
	stack.SetSpacing(spacing)
	return stack
}

func getColumnsForTime(f *Field) int {
	var c int
	for _, flag := range []FlagType{FlagHasSeconds, FlagHasMinutes, FlagHasHours, FlagHasDays, FlagHasMonths, FlagHasYears} {
		if f.Flags&flag != 0 {
			c += 3
		}
	}
	return c - 1
}

func updateFlagStack(flags reflect.Value, f *Field, view zview.View) {
	stack := view.(*zcontainer.StackView)
	// zlog.Info("zfields.updateFlagStack", Name(f))
	bso := flags.Interface().(zbits.BitsetItemsOwner)
	bitset := bso.GetBitsetItems()
	n := flags.Int()
	for _, bs := range bitset {
		name := bs.Name
		vf, _ := stack.FindViewWithName(name, false)
		if n&bs.Mask != 0 {
			if vf == nil {
				path := "images/" + f.ID + "/" + name + ".png"
				iv := zimageview.New(nil, path, zgeo.Size{16, 16})
				iv.DownsampleImages = true
				// zlog.Info("flag image:", name, iv.DownsampleImages)
				iv.SetObjectName(name) // very important as we above find it in stack
				iv.SetMinSize(zgeo.Size{16, 16})
				stack.Add(iv, zgeo.Center)
				if stack.Presented {
					stack.ArrangeChildren()
				}
				title := bs.Title
				iv.SetToolTip(title)
			}
		} else {
			if vf != nil {
				stack.RemoveNamedChild(name, false)
				stack.ArrangeChildren()
			}
		}
	}
}

func (v *FieldView) createSpecialView(rval reflect.Value, f *Field) (view zview.View, skip bool) {
	// zlog.Info("createSpecialView:", f.FieldName)
	if f.Flags&FlagIsButton != 0 {
		if v.params.HideStatic {
			return nil, true
		}
		return v.makeButton(rval, f), false
	}
	if f.WidgetName != "" && f.Kind != zreflect.KindSlice {
		w := widgeters[f.WidgetName]
		if w != nil {
			return w.Create(f), false
		}
	}
	if v.callTriggerHandler(f, CreateFieldViewAction, nil, &view) {
		return
	}
	callActionHandlerFunc(v, f, CreateFieldViewAction, rval.Addr().Interface(), &view) // this sees if actual ITEM is a field handler
	if view != nil {
		return
	}
	if f.LocalEnum != "" {
		ei, findex := FindLocalFieldWithID(v.data, f.LocalEnum)
		if zlog.ErrorIf(findex == -1, v.Hierarchy(), f.Name, f.LocalEnum) {
			return nil, true
		}
		getter, _ := ei.Interface().(zdict.ItemsGetter)
		if zlog.ErrorIf(getter == nil, "field isn't enum, not ItemGetter type", f.Name, f.LocalEnum) {
			return nil, true
		}
		enum := getter.GetItems()
		// zlog.Info("make local enum:", f.Name, f.LocalEnum, enum, ei)
		// 	continue
		// }
		//					zlog.Info("make local enum:", f.Name, f.LocalEnum)
		for i := range enum {
			if f.Flags&FlagZeroIsEmpty != 0 {
				if enum[i].Value != nil && reflect.ValueOf(enum[i].Value).IsZero() {
					// zlog.Info("Clear zero name")
					enum[i].Name = ""
				}
			}
		}

		menu := v.makeMenu(rval, f, enum)
		if menu == nil {
			zlog.Error(nil, "no local enum for", f.LocalEnum)
			return nil, true
		}
		return menu, false
		// mt := view.(zmenu.MenuType)
		//!!					mt.SelectWithValue(item.Interface)
	}
	if f.Enum != "" {
		// fmt.Println("make enum:", f.Name, item.Interface)
		enum, _ := fieldEnums[f.Enum]
		zlog.Assert(enum != nil, f.Enum, f.FieldName)
		if rval.IsZero() {
			if !v.params.MultiSliceEditInProgress {
				if len(enum) > 0 {
					rval.Set(reflect.ValueOf(enum[0].Value))
				}
			} else {
				var zero bool
				for i := range enum {
					if enum[i].Value != nil && reflect.ValueOf(enum[i].Value).IsZero() {
						zero = true
						break
					}
				}
				if !zero {
					fmt.Println("Add zero enum:", f.Name)
					enum = append(enum, zdict.Item{"", rval.Interface()})
				}
			}
		}
		view = v.makeMenu(rval, f, enum)
		// exp = zgeo.AlignmentNone
		return view, false
	}
	kind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	if kind == zreflect.KindInt && rval.Type().Name() != "BoolInd" {
		_, got := rval.Interface().(zbits.BitsetItemsOwner)
		if got {
			return makeFlagStack(v, rval, f), false
		}
	}
	_, got := rval.Interface().(UIStringer)
	if got && (f.IsStatic() || v.params.AllStatic) {
		return v.makeText(rval, f, false), false
	}
	return nil, false
}

func (v *FieldView) BuildStack(name string, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) {
	zlog.Assert(reflect.ValueOf(v.data).Kind() == reflect.Ptr, name, v.data)
	// fmt.Println("buildStack1", name, defaultAlign, v.params.SkipFieldNames)
	ForEachField(v.data, v.params.FieldParameters, v.Fields, func(index int, f *Field, val reflect.Value, sf reflect.StructField) {
		v.buildItem(f, val, sf, index, defaultAlign, cellMargin, useMinWidth)
	})
}

func (v *FieldView) buildItem(f *Field, rval reflect.Value, sf reflect.StructField, index int, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) {
	labelizeWidth := v.params.LabelizeWidth
	parentFV := ParentFieldView(v)
	if parentFV != nil && v.params.LabelizeWidth == 0 {
		labelizeWidth = parentFV.params.LabelizeWidth
	}
	exp := zgeo.AlignmentNone
	if v.params.HideStatic && f.IsStatic() {
		return
	}
	view, skip := v.createSpecialView(rval, f)
	if skip {
		return
	}
	if view == nil {
		switch f.Kind {
		case zreflect.KindStruct:
			// col, got := rval.Interface().(zgeo.Color)
			// // zlog.Info("makeStruct", item.FieldName, got)
			// // zlog.Info("make stringer?:", f.Name, got)
			// if got {
			// 	view = zcolor.NewView(col)
			// } else {
			//!!! exp = zgeo.HorExpand
			// zlog.Info("struct make field view:", f.Name, f.Kind, exp)
			vert := true
			if !f.Vertical.IsUnknown() {
				vert = f.Vertical.Bool()
			}
			// zlog.Info("struct fieldViewNew", f.ID, vert, f.Vertical)
			params := v.params
			params.Field.MergeInField(f)
			// params.Field.Flags = 0
			fieldView := fieldViewNew(f.FieldName, vert, rval.Addr().Interface(), params, zgeo.Size{}, v)
			// fmt.Printf("makeStruct2 %s %p\n", item.FieldName, fieldView.View)
			view = makeFrameIfFlag(f, fieldView)
			if view == nil {
				view = fieldView
			}
			// fmt.Printf("makeStruct3 %s %p %v %p\n", item.FieldName, view, view == nil, fieldView)
			// fieldView.parentField = f
			fieldView.BuildStack(f.FieldName, zgeo.TopLeft, zgeo.Size{}, true)

		case zreflect.KindBool:
			if f.Flags&FlagIsImage != 0 && f.IsImageToggle() && rval.Kind() == reflect.Bool {
				view = v.makeImage(rval, f)
				break
			}
			b := zbool.ToBoolInd(rval.Interface().(bool))
			exp = zgeo.AlignmentNone
			view = v.makeCheckbox(f, b)
		case zreflect.KindInt:
			if rval.Type().Name() == "BoolInd" {
				exp = zgeo.HorShrink
				view = v.makeCheckbox(f, zbool.BoolInd(rval.Int()))
			} else {
				_, got := rval.Interface().(zbits.BitsetItemsOwner)
				if got {
					view = makeFlagStack(v, rval, f)
					break
				}
				noUpdate := true
				view = v.makeText(rval, f, noUpdate)
			}

		case zreflect.KindFloat:
			noUpdate := true
			view = v.makeText(rval, f, noUpdate)

		case zreflect.KindString:
			if f.Flags&FlagIsImage != 0 {
				view = v.makeImage(rval, f)
			} else {
				if (f.MaxWidth != f.MinWidth || f.MaxWidth != 0) && f.Flags&FlagIsButton == 0 {
					exp = zgeo.HorExpand
				}
				view = v.makeText(rval, f, false)
			}

		case zreflect.KindSlice:
			getter, _ := rval.Interface().(zdict.ItemsGetter)
			if getter != nil {
				menu := v.makeMenu(rval, f, getter.GetItems())
				view = menu
				break
			}
			if f.Alignment != zgeo.AlignmentNone {
				exp = zgeo.Expand
			} else {
				exp = zgeo.AlignmentNone
			}
			vert := v.Vertical
			if labelizeWidth != 0 {
				vert = false
			}
			params := v.params
			params.Field.MergeInField(f)
			// params.Field.Flags = 0
			fv := fieldViewNew(f.FieldName, vert, rval.Addr().Interface(), params, zgeo.Size{}, v)
			view = fv
			var add zview.View
			if f.Flags&FlagIsGroup == 0 {
				// zlog.Info("BUILD static slice:", f.Name, v.Hierarchy(), v.Spacing(), fv.Spacing(), fv.params.Styling.Spacing)
				add = fv.buildRepeatedStackFromSlice(rval.Addr().Interface(), vert, f)
			} else {
				// zlog.Info("NewMenuGroup:", f.FieldName, f.TitleOrName(), f.Styling.StrokeWidth, params.Styling.StrokeWidth)
				mg := buildMenuGroup(rval.Addr().Interface(), f.ValueStoreKey, params)
				mg.SetChangedHandler(func(newID string) {
					// zlog.Info("Update group")
					callActionHandlerFunc(v, f, EditedAction, rval.Addr().Interface(), &mg.View)
				})
				fv.grouper = mg
				add = mg
			}
			fv.Add(add, zgeo.TopLeft)

		case zreflect.KindTime:
			columns := f.Columns
			if columns == 0 {
				columns = getColumnsForTime(f)
			}
			noUpdate := true
			view = v.makeText(rval, f, noUpdate)
			if f.IsStatic() || v.params.AllStatic {
				label := view.(*zlabel.Label)
				label.Columns = columns
				if f.Flags&FlagIsDuration != 0 || f.OldSecs != 0 {
					timer := ztimer.RepeatNow(1, func() bool {
						nlabel := view.(*zlabel.Label)
						if f.Flags&FlagIsDuration != 0 {
							v.updateSinceTime(nlabel, f)
						} else {
							v.updateOldTime(nlabel, f)
						}
						return true
					})
					v.AddOnRemoveFunc(timer.Stop)
					view.Native().AddOnRemoveFunc(timer.Stop)
				} else {
					if f.Format == "nice" {
						timer := ztimer.StartAt(ztime.OnTheNextHour(time.Now()), func() {
							nlabel := view.(*zlabel.Label)
							str := getTextFromNumberishItem(rval, f)
							nlabel.SetText(str)
						})
						v.AddOnRemoveFunc(timer.Stop)
					}
				}
			}

		default:
			panic(fmt.Sprintln("buildStack bad type:", f.Name, f.Kind))
		}
	}
	// zlog.Info("Hierarchy1", f.FieldName, f.Kind)
	// zlog.Info("Hierarchy:", view.Native().Hierarchy())
	zlog.Assert(view != nil)
	pt, _ := view.(zview.Pressable)
	if view != nil && pt != nil && pt.PressedHandler != nil {
		// fmt.Printf("Pressable1? %p %v\n", view, view == nil) // pt.PressedHandler,
		// zlog.Info("Pressable2:", view.Native() != nil)
		// zlog.Info("Pressable?", pt.PressedHandler(), view.Native().Hierarchy())
		ph := pt.PressedHandler()
		nowItem := rval // store item in nowItem so closures below uses right item
		pt.SetPressedHandler(func() {
			// zlog.Info("FV Pressed")
			if !callActionHandlerFunc(v, f, PressedAction, nowItem.Interface(), &view) && ph != nil {
				ph()
			}
		})
		if f.Flags&FlagLongPress != 0 {
			lph := pt.LongPressedHandler()
			pt.SetLongPressedHandler(func() {
				// zlog.Info("Field.LPH:", f.FieldName)
				if !callActionHandlerFunc(v, f, LongPressedAction, nowItem.Interface(), &view) && lph != nil {
					lph()
				}
			})
		}
	}
	// zlog.Info("BuildItem:", item.FieldName, view != nil, zlog.GetCallingStackString())
	updateItemLocalToolTip(f, v.data, view)
	if !f.Styling.DropShadow.Delta.IsNull() {
		nv := view.Native()
		nv.SetDropShadow(f.Styling.DropShadow)
	}
	view.SetObjectName(f.FieldName)
	if f.Styling.FGColor.Valid {
		view.SetColor(f.Styling.FGColor)
	}
	if f.Styling.BGColor.Valid {
		view.SetBGColor(f.Styling.BGColor)
	}
	callActionHandlerFunc(v, f, CreatedViewAction, rval.Addr().Interface(), &view)
	if f.Download != "" {
		surl := zstr.ReplaceAllCapturesWithoutMatchFunc(zstr.InDoubleSquigglyBracketsRegex, f.Download, func(fieldName string, index int) string {
			a, findex := FindLocalFieldWithID(v.data, fieldName)
			if findex == -1 {
				zlog.Error(nil, "field download", f.Download, ":", "field not found in struct:", fieldName)
				return ""
			}
			return fmt.Sprint(a.Interface())
		})
		link := zcontainer.MakeLinkedStack(surl, "", view)
		view = link
	}

	cell := &zcontainer.Cell{}
	def := defaultAlign
	all := zgeo.Left | zgeo.HorCenter | zgeo.Right
	if f.Alignment&all != 0 {
		def &= ^all
	}
	cell.Margin = cellMargin
	cell.Alignment = def | exp | f.Alignment
	if labelizeWidth != 0 || f.LabelizeWidth < 0 {
		var lstack *zcontainer.StackView
		title := f.Name
		if f.Title != "" {
			title = f.Title
		}
		if f.Flags&FlagNoTitle != 0 {
			title = ""
		}
		_, lstack, cell = zguiutil.Labelize(view, title, labelizeWidth, cell.Alignment)
		v.Add(lstack, zgeo.HorExpand|zgeo.Left|zgeo.Top)
	}
	if useMinWidth {
		cell.MinSize.W = f.MinWidth
	}
	cell.MaxSize.W = f.MaxWidth
	if labelizeWidth == 0 {
		cell.View = view
		v.AddCell(*cell, -1)
	}
}

func updateItemLocalToolTip(f *Field, structure any, view zview.View) {
	var tipField, tip string
	if zstr.HasPrefix(f.Tooltip, "./", &tipField) {
		ei, _, findex := zreflect.FieldForName(structure, true, tipField)
		if findex != -1 {
			tip = fmt.Sprint(ei.Interface())
		} else { // can't use tip == "" to check, since field might just be empty
			zlog.Error(nil, "updateItemLocalToolTip: no local field for tip", f.Name, tipField)
		}
	} else if f.Tooltip != "" {
		tip = f.Tooltip
	}
	if tip != "" {
		view.Native().SetToolTip(tip)
	}
}

func (v *FieldView) ToData(showError bool) (err error) {
	for _, f := range v.Fields {
		// fmt.Println("FV Update Item:", f.Name)
		foundView, _ := v.findNamedViewOrInLabelized(f.FieldName)
		if foundView == nil {
			// zlog.Info("FV Update no view found:", v.id, f.FieldName)
			continue
		}
		_, e := v.fieldToDataItem(&f, foundView)
		if e != nil {
			if err == nil {
				err = e
			}
		}
	}
	if showError && err != nil {
		zalert.ShowError(err)
	}
	return
}

func (v *FieldView) fieldToDataItem(f *Field, view zview.View) (value reflect.Value, err error) {
	// zlog.Info("fieldViewToDataItem1:", f.IsStatic(), f.Name, f.Index)
	if f.IsStatic() {
		return
	}
	rval, _ := zreflect.FieldForIndex(v.data, true, f.Index)

	if f.WidgetName != "" && f.Kind != zreflect.KindSlice {
		w := widgeters[f.WidgetName]
		r, _ := w.(ReadWidgeter)
		if r != nil {
			val := r.GetValue(view)
			rval.Set(reflect.ValueOf(val))
			value = rval
			return
		}
	}
	// zlog.Info("fieldViewToDataItem before:", f.IsStatic(), f.Name, f.Index, len(children), "s:")
	if f.Enum != "" || f.LocalEnum != "" {
		mv, _ := view.(*zmenu.MenuView)
		if mv != nil {
			iface := mv.CurrentValue()
			vo := reflect.ValueOf(iface)
			// zlog.Debug(iface, f.Name, iface == nil)
			if iface == nil {
				vo = reflect.Zero(rval.Type())
			}
			rval.Set(vo)
		}
		return
	}

	switch f.Kind {
	case zreflect.KindBool:
		_, got := view.(*zimageview.ImageView)
		if got {
			break
		}
		bv, _ := view.(*zcheckbox.CheckBox)
		if bv == nil {
			zcontainer.ViewRangeChildren(view, false, false, func(v zview.View) bool {
				bv, _ = v.(*zcheckbox.CheckBox)
				return bv == nil
			})
			if bv == nil {
				zlog.Fatal(nil, "Should be checkbox", view, reflect.TypeOf(view))
			}
		}
		b, _ := rval.Addr().Interface().(*bool)
		if b != nil {
			*b = bv.Value().Bool()
		}
		bi, _ := rval.Addr().Interface().(*zbool.BoolInd)
		if bi != nil {
			*bi = bv.Value()
		}

	case zreflect.KindInt:
		if rval.Type().Name() == "BoolInd" {
			bv, _ := view.(*zcheckbox.CheckBox)
			*rval.Addr().Interface().(*zbool.BoolInd) = bv.Value()
		} else {
			tv, _ := view.(*ztext.TextView)
			if f.Flags&FlagZeroIsEmpty != 0 {
				if tv.Text() == "" {
					rval.SetZero()
					v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())
					break
				}
			}
			str := tv.Text()
			if f.PackageName == "time" && rval.Type().Name() == "Duration" {
				var secs float64
				secs, err = ztime.GetSecsFromHMSString(str, f.Flags&FlagHasHours != 0, f.Flags&FlagHasMinutes != 0, f.Flags&FlagHasSeconds != 0)
				if err != nil {
					break
				}
				d := rval.Addr().Interface().(*time.Duration)
				if d != nil {
					*d = ztime.SecondsDur(secs)
				}
				value = rval
				return
			}
			var i64 int64
			if str != "" {
				i64, err = strconv.ParseInt(str, 10, 64)
				if err != nil {
					err = zlog.NewError("Error parsing", f.Name, "in", v.ObjectName())
					break
				}
			}
			zint.SetAny(rval.Addr().Interface(), i64)
			// zlog.Info("num updateShowEnableFromZeroer?", tv.Hierarchy(), i64)
			v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())
		}

	case zreflect.KindFloat:
		tv, _ := view.(*ztext.TextView)
		text := tv.Text()
		var f64 float64
		//		if text != "" {
		// zlog.Info("ParseFloat:", v.Hierarchy(), f.Name, text, f.Flags&FlagZeroIsEmpty != 0)
		if f.Flags&FlagZeroIsEmpty != 0 {
			if text == "" {
				rval.SetZero()
				v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())
				break
			}
		}
		f64, err = strconv.ParseFloat(text, 64)
		if err != nil {
			err = zlog.NewError("Error parsing", f.Name, "in", v.ObjectName())
		}
		if err != nil {
			break
		}
		//		}
		zfloat.SetAny(rval.Addr().Interface(), f64)
		v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())

	case zreflect.KindTime:
		break

	case zreflect.KindString:
		if (!f.IsStatic() && !v.params.AllStatic) && f.Flags&FlagIsImage == 0 {
			tv, _ := view.(*ztext.TextView)
			if tv == nil {
				zlog.Fatal(nil, "Copy Back string not TV:", f.Name)
			}
			text := tv.Text()
			str := rval.Addr().Interface().(*string)
			*str = text
		}

	case zreflect.KindFunc:
		break

	case zreflect.KindStruct:
		zcontainer.ViewRangeChildren(v, true, true, func(view zview.View) bool {
			if view.ObjectName() == f.FieldName {
				fv, _ := view.(*FieldView)
				if fv != nil {
					cerr := fv.ToData(false)
					if cerr != nil {
						err = cerr
					}
					return false
				}
			}
			return true
		})
		break

	default:
		panic(fmt.Sprint("bad type: ", f.Kind))
	}
	value = reflect.ValueOf(rval.Addr().Interface()).Elem() //.Interface()
	return
}

func ParentFieldView(view zview.View) *FieldView {
	for _, nv := range view.Native().AllParents() {
		fv, _ := nv.View.(*FieldView)
		if fv != nil {
			return fv
		}
	}
	return nil
}

func PresentOKCancelStruct[S any](structPtr *S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) bool) {
	slice := []S{*structPtr}
	//	zlog.Info("PresentOKCancelStruct", structType)
	PresentOKCancelStructSlice(&slice, params, title, att, func(ok bool) bool {
		if ok {
			*structPtr = slice[0]
		}
		return done(ok)
	})
}
