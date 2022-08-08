//go:build zui

package zfields

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zclipboard"
	"github.com/torlangballe/zui/zcolor"
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
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zbits"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
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
	HideStatic            bool
	ImmediateEdit         bool
	ForceZeroOption       bool // ForceZeroOption makes menus (and theoretically more) have a zero, or undefined option. This is set when creating a single dialog box for a whole slice of structures.
	AllTextStatic         bool // AllTextStatic makes even not "static" tagged fields static. Good for showing in tables etc.
	EditWithoutCallsbacks bool
}

func FieldViewParametersDefault() (f FieldViewParameters) {
	f.ImmediateEdit = true
	f.Styling = zstyle.EmptyStyling
	f.Styling.Spacing = 8
	return f
}

var fieldViewEdited = map[string]time.Time{}

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

func makeFrameIfFlag(f *Field, child zview.View) *zcontainer.StackView {
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

	// zlog.Info("fieldViewNew", id, params.Styling.StrokeWidth)
	v.data = data
	for i, item := range v.getStructItems() {
		f := EmptyField
		if f.SetFromReflectItem(data, item, i, params.ImmediateEdit) {
			// zlog.Info("fieldViewNew f:", f.Name, f.UpdateSecs)
			v.Fields = append(v.Fields, f)
		}
	}
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
	for _, c := range (v.View.(zcontainer.ContainerType)).GetChildren(false) {
		n := c.ObjectName()
		// zlog.Info("findNamedViewOrInLabelized:", v.ObjectName(), name, n)
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
		// zlog.Info("updateShowEnableFromZeroer:", f.FieldName, isZero, isShow, toID, local)
		if zstr.HasPrefix(local, "./", &id) && id == toID {
			_, fview := v.findNamedViewOrInLabelized(f.ID)
			zlog.Assert(fview != nil)
			if neg {
				isShow = !isShow
			}
			if isShow {
				fview.Show(!isZero)
			} else {
				fview.SetUsable(!isZero)
			}
			continue
		}
	}
	//TODO: handle ../ and substruct/id style
}

func doItem(item *zreflect.Item, isShow bool, view zview.View, not bool) {
	zero := item.Value.IsZero()
	if not {
		zero = !zero
	}
	if isShow {
		view.Show(!zero)
	} else {
		view.SetUsable(!zero)
	}
}

func findIDInStructItems(items []zreflect.Item, id string) *zreflect.Item {
	for i, item := range items {
		// zlog.Info("findIDInStructItems:", fieldNameToID(item.FieldName), id)
		if fieldNameToID(item.FieldName) == id {
			return &items[i]
		}
	}
	return nil
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

func (v *FieldView) updateShowEnableOnView(view zview.View, isShow bool, toID string) {
	// zlog.Info("updateShowOrEnable:", isShow, toID, len(v.Fields))
	for _, f := range v.Fields {
		if f.ID != toID {
			continue
		}
		var prefix, id string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		if zstr.HasPrefix(local, "./", &id) {
			// zlog.Info("local:", toID, f.FieldName, id)
			fitem := findIDInStructItems(v.getStructItems(), id)
			if fitem != nil {
				doItem(fitem, isShow, view, neg)
			}
			continue
		}
		if zstr.SplitN(local, "/", &prefix, &id) && prefix == f.ID {
			fv := view.(*FieldView)
			if fv == nil {
				zlog.Error(nil, "updateShowOrEnable: not field view:", f.FieldName, local, v.ObjectName)
				return
			}
			// zlog.Info("sub in:", toID, prefix, id, f.ID, prefix == f.ID)
			fitem := findIDInStructItems(fv.getStructItems(), id)
			if fitem != nil {
				// zlog.Info("subin:", id, fv.ObjectName(), fitem != nil)
				doItem(fitem, isShow, view, neg)
			}
			continue
		}
		if zstr.HasPrefix(local, "../", &id) && v.parent != nil {
			fitem := findIDInStructItems(v.parent.getStructItems(), id)
			if fitem != nil {
				doItem(fitem, isShow, view, neg)
				// zlog.Info("sub back:", toID, id)
			}
			continue
		}
	}
}

func (v *FieldView) Update(data any, dontOverwriteEdited bool) {
	if data != nil {
		v.data = data
	}
	// zlog.Info("fv.Update:", v.ObjectName(), dontOverwriteEdited, IsFieldViewEditedRecently(v))
	if dontOverwriteEdited && IsFieldViewEditedRecently(v) {
		zlog.Info("FV No Update, edited", v.Hierarchy())
		return
	}
	children := v.getStructItems()
	fh, _ := v.data.(ActionHandler)
	sview := v.View
	if fh != nil {
		fh.HandleAction(nil, DataChangedActionPre, &sview)
	}
	// fmt.Println("FV Update", v.id, len(children))
	// fmt.Printf("FV Update: %s %d %+v\n", v.id, len(children), v.data)
	for i, item := range children {
		var valStr string
		f := findFieldWithIndex(&v.Fields, i)
		if f == nil {
			// zlog.Info("FV Update no index found:", i, v.id)
			continue
		}
		// fmt.Println("FV Update Item:", v.Hierarchy(), f.Name, item.Kind)
		fview, flabelized := v.findNamedViewOrInLabelized(f.ID)
		// zlog.Info("fv.UpdateF:", v.Hierarchy(), f.ID, f.FieldName, fview != nil)
		if fview == nil {
			// zlog.Info("FV Update no view found:", i, v.id, f.ID)
			continue
		}
		v.updateShowEnableOnView(flabelized, true, fview.ObjectName())
		v.updateShowEnableOnView(flabelized, false, fview.ObjectName())
		called := callActionHandlerFunc(v, f, DataChangedAction, item.Address, &fview)
		// zlog.Info("fv.Update:", v.ObjectName(), f.ID, called)
		if called {
			// fmt.Println("FV Update called", v.id, f.Kind, f.ID)
			continue
		}
		if f.Kind != zreflect.KindSlice {
			// zlog.Info("fv update !slice:", f.Name, reflect.ValueOf(fview).Type())
			w := widgeters[f.WidgetName]
			if w != nil {
				w.SetValue(fview, item.Interface)
				continue
			}
		}
		menuType, _ := fview.(zmenu.MenuType)
		if menuType != nil && ((f.Enum != "" && f.Kind != zreflect.KindSlice) || f.LocalEnum != "") {
			var enum zdict.Items
			// zlog.Info("Update FV: Menu:", f.Name, f.Enum, f.LocalEnum)
			if f.Enum != "" {
				enum, _ = fieldEnums[f.Enum]
				// zlog.Info("UpdateStack Enum:", f.Name)
				// zdict.DumpNamedValues(enum)
			} else {
				ei := findLocalFieldWithID(&children, f.LocalEnum)
				zlog.Assert(ei != nil, f.Name, f.LocalEnum)
				enum = ei.Interface.(zdict.ItemsGetter).GetItems()
			}
			if v.params.ForceZeroOption {
				var rtype reflect.Type
				var hasZero bool
				for _, item := range enum {
					rval := reflect.ValueOf(item.Value)
					if item.Value != nil {
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

			// zlog.Assert(enum != nil, f.Name, f.LocalEnum, f.Enum)
			// zlog.Info("Update FV: Menu2:", f.Name, enum, item.Interface)
			menuType.UpdateItems(enum, []interface{}{item.Interface})
			continue
		}
		updateItemLocalToolTip(f, children, fview)
		if f.IsStatic() || v.params.AllTextStatic {
			zuistringer, _ := item.Interface.(UIStringer)
			if zuistringer != nil {
				label, _ := fview.(*zlabel.Label)
				if label != nil {
					label.SetText(zuistringer.ZUIString())
					continue
				}
			}
		}
		switch f.Kind {
		case zreflect.KindSlice:
			// zlog.Info("updateSliceFieldView:", v.Hierarchy())
			// val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
			// fmt.Printf("updateSliceFieldView: %s %p %p %v %p\n", v.id, item.Interface, val.Interface(), found, fview)
			fv := fview.(*FieldView)
			fv.data = item.Address
			hash := zstr.HashAnyToInt64(reflect.ValueOf(fv.data).Elem())
			// zlog.Info("update any SliceValue:", f.Name, hash, fv.dataHash, reflect.ValueOf(fv.data).Elem())
			sameHash := (fv.dataHash == hash)
			fv.dataHash = hash
			getter, _ := item.Interface.(zdict.ItemsGetter)
			// zlog.Info("fv update slice:", f.Name, reflect.ValueOf(fview).Type(), getter != nil)
			if !sameHash && getter != nil {
				items := getter.GetItems()
				mt := fview.(zmenu.MenuType)
				// zlog.Info("fv update slice:", f.Name, len(items), mt != nil, reflect.ValueOf(fview).Type())
				if mt != nil {
					// assert menu is static...
					mt.UpdateItems(items, nil)
				}
			}
			if f.Flags&FlagIsGroup == 0 {
				stack := fv.GetChildren(true)[0].(*zcontainer.StackView)
				if sameHash {
					fv.updateSliceElementData(item.Address, stack)
					break
				}
				if menuType != nil {
					break
				}
				vert := v.Vertical
				if v.params.LabelizeWidth != 0 {
					vert = false
				}
				fv.updateSliceValue(item.Address, stack, vert, f, false)
			} else {
				// zlog.Info("UpdateGSlice:", fv.Hierarchy(), item.Interface)
				mg := fv.grouper.(*zgroup.MenuGroupView) // we have to convert it to a MenuGroupView for replace to work, for comparison
				if !sameHash {
					changedIDs := mg.HasDataChangedInIDsFunc()
					if len(changedIDs) != 0 && !(len(changedIDs) == 1 && changedIDs[0] == mg.GetCurrentID()) { // if it's not the id of current slice element, we di a full rebuild
						// zlog.Info("UpdateSliceChanged in other item:", changedIDs)
						repopulateMenuGroup(mg, item.Address, fv.params)
						break
					}
				}
				updateMenuGroupSlice(mg, item.Address, dontOverwriteEdited)
			}
		case zreflect.KindTime:
			tv, _ := fview.(*ztext.TextView)
			if tv != nil && tv.IsEditing() {
				break
			}
			if f.Flags&FlagIsDuration != 0 {
				// val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
				// if found {
				// t := val.Interface().(time.Time)
				// fmt.Println("FV Update Time Dur", v.id, time.Since(t))
				// }
				v.updateSinceTime(fview.(*zlabel.Label), f)
				break
			}
			valStr = getTimeString(item, f)
			to := fview.(ztext.LayoutOwner)
			to.SetText(valStr)

		case zreflect.KindStruct:
			// zlog.Info("Update Struct:", item.FieldName)
			fv, _ := fview.(*FieldView)
			if fv == nil {
				fv = findSubFieldView(fview, "")
			}
			if fv == nil {
				break
			}
			// zlog.Info("Update Struct2:", item.FieldName, fv.Hierarchy())
			fv.Update(item.Address, dontOverwriteEdited)

		case zreflect.KindBool:
			cv, _ := fview.(*zcheckbox.CheckBox) // it might be a button or something instead
			if cv != nil {
				b := zbool.ToBoolInd(item.Value.Interface().(bool))
				v := cv.Value()
				if v != b {
					cv.SetValue(b)
				}
			}

		case zreflect.KindInt, zreflect.KindFloat:
			_, got := item.Interface.(zbits.BitsetItemsOwner)
			if got {
				updateFlagStack(item, f, fview)
			}

			valStr = getTextFromNumberishItem(item, f)
			if f.IsStatic() || v.params.AllTextStatic {
				label, _ := fview.(*zlabel.Label)
				if label != nil {
					label.SetText(valStr)
				}
				break
			}
			tv, _ := fview.(*ztext.TextView)
			if tv != nil {
				if tv.IsEditing() {
					break
				}
				tv.SetText(valStr)
			}

		case zreflect.KindString, zreflect.KindFunc:
			valStr = item.Value.String()
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
				io := fview.(zimage.Owner)
				io.SetImage(nil, path, nil)
			} else {
				if f.IsStatic() || v.params.AllTextStatic {
					label, _ := fview.(*zlabel.Label)
					if label != nil {
						if f.Flags&FlagIsFixed != 0 {
							valStr = f.Name
						}
						label.SetText(valStr)
					}
				} else {
					tv, _ := fview.(*ztext.TextView)
					if tv != nil {
						if tv.IsEditing() {
							break
						}
						tv.SetText(valStr)
					}
				}
			}
		}
		gb := zgroup.GetAncestorGroupBase(fview)
		if gb != nil {
			// zlog.Info("UpdateIndicator:", fview.Native().Hierarchy(), gb.Hierarchy())
			data := gb.Data.(*zgroup.SliceGroupData)
			if data.IndicatorID == f.FieldName {
				if gb.UpdateCurrentIndicatorFunc != nil {
					gb.UpdateCurrentIndicatorFunc(fmt.Sprint(valStr))
				}
			}
		}
	}
	// call general one with no id. Needs to be after above loop, so values set
	if fh != nil {
		fh.HandleAction(nil, DataChangedAction, &sview)
	}
}

func updateMenuGroupSlice(mg *zgroup.MenuGroupView, slicePtr any, dontOverwriteEdited bool) {
	data := mg.Data.(*zgroup.SliceGroupData)
	data.SlicePtr = slicePtr
	id := mg.GetCurrentID()
	if id != "" {
		fv := mg.ChildView.(*FieldView)
		i := zgroup.IndexForIDFromSlice(slicePtr, id)
		zlog.Assert(i != -1)
		a := reflect.ValueOf(slicePtr).Elem().Index(i).Addr().Interface()
		// fmt.Printf("updateMenuGroupSlice: %s %s %p\n", mg.Hierarchy(), fv.Hierarchy(), a)
		fv.Update(a, dontOverwriteEdited)
	}
}

func findSubFieldView(view zview.View, optionalID string) (fv *FieldView) {
	zcontainer.ViewRangeChildren(view, true, true, func(view zview.View) bool {
		f, _ := view.(*FieldView)
		if f != nil {
			if optionalID == "" || f.id == optionalID {
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
		if v.params.EditWithoutCallsbacks {
			return true
		}
		// fmt.Printf("call edit ActionHandlerFunc2: %s %p\n", f.Name, v.data)
	}
	// zlog.Info("callActionHandlerFunc:", f.ID, f.Name, action)
	direct := (action == CreateFieldViewAction || action == SetupFieldAction)
	// zlog.Info("callActionHandlerFunc  get sub:", f.ID, f.Name, action)
	// data := v.getSubStruct(structID, direct)
	// zlog.Info("callFieldHandler1", action, f.Name, data != nil, reflect.ValueOf(data))
	fh, _ := v.data.(ActionHandler)
	var result bool
	if fh != nil {
		result = fh.HandleAction(f, action, view)
	}
	// zlog.Info("callActionHandlerFunc2:", f.ID, f.Name, action)

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
						fv.dataHash = zstr.HashAnyToInt64(reflect.ValueOf(fv.data).Elem())
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
		aih, _ := fieldValue.(ActionFieldHandler)
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
			aih, _ = fieldAddress.(ActionFieldHandler)
		}
		// zlog.Info("callActionHandlerFunc bottom:", f.Name, action, aih != nil, reflect.ValueOf(fieldValue).Type(), reflect.ValueOf(fieldValue).Kind())
		if aih != nil {
			result = aih.HandleFieldAction(f, action, view)
			// zlog.Info("callActionHandlerFunc bottom:", f.Name, action, result, view, aih)
		}
	}
	// zlog.Info("callActionHandlerFunc top done:", f.ID, f.Name, action)
	return result
}

func (v *FieldView) findFieldWithID(id string) *Field {
	for i, f := range v.Fields {
		zlog.Info("FFWID:", v.ObjectName(), f.ID, id, f.ID == id)
		if f.ID == id {
			return &v.Fields[i]
		}
	}
	return nil
}

func (fv *FieldView) makeButton(item zreflect.Item, f *Field) *zshape.ImageButtonView {
	// zlog.Info("makeButton:", f.Name, f.Height)
	format := f.Format
	if format == "" {
		format = "%v"
	}
	color := "gray"
	if len(f.Colors) != 0 {
		color = f.Colors[0]
	}
	name := f.Name
	if f.Title != "" {
		name = f.Title
	}
	s := zgeo.Size{20, 22}
	if f.Height != 0 {
		s.H = f.Height
	}
	button := zshape.ImageButtonViewNew(name, color, s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
	button.SetTextColor(zgeo.ColorBlack)
	button.TextXMargin = 0
	return button
}

func (v *FieldView) makeMenu(item zreflect.Item, f *Field, items zdict.Items) zview.View {
	var view zview.View
	if f.IsStatic() || item.IsSlice {
		multi := item.IsSlice
		// zlog.Info("FV Menu Make static:", f.ID, f.Format, f.Name)
		vals := []interface{}{item.Interface}
		isImage := (f.ImageFixedPath != "")
		shape := zshape.TypeRoundRect
		if isImage {
			shape = zshape.TypeNone
		}
		var mItems []zmenu.MenuedOItem
		for i := range items {
			var m zmenu.MenuedOItem
			for j := range vals {
				if reflect.DeepEqual(items[i], vals[j]) {
					m.Selected = true
					break
				}
			}
			if f.Flags&FlagIsActions != 0 {
				m.IsAction = true
			}
			m.Name = items[i].Name
			m.Value = items[i].Value
			mItems = append(mItems, m)
		}
		menuOwner := zmenu.NewMenuedOwner()
		menuOwner.IsStatic = f.IsStatic()
		menuOwner.IsMultiple = multi
		menuOwner.StoreKey = f.ValueStoreKey
		menu := zmenu.MenuOwningButtonCreate(menuOwner, mItems, shape)
		if isImage {
			menu.SetImage(nil, f.ImageFixedPath, nil)
			menu.ImageAlign = zgeo.Center | zgeo.Proportional
			// zlog.Info("FV Menued:", f.ID, f.Size)
			menu.ImageMaxSize = f.Size
		} else {
			// menu.SetPillStyle()
			if len(f.Colors) != 0 {
				menu.SetColor(zgeo.ColorFromString(f.Colors[0]))
			}
		}
		view = menu
		// zlog.Info("Make Menu Format", f.Name, f.Format)
		if f.Format != "" {
			if f.Format == "-" {
				menuOwner.GetTitle = func(icount int) string {
					return ""
				}
			} else if f.Format == "%d" {
				menuOwner.GetTitle = func(icount int) string {
					// zlog.Info("fv menu gettitle2:", f.FieldName, f.Format, icount)
					return strconv.Itoa(icount)
				}
			} else {
				menuOwner.GetTitle = func(icount int) string {
					// zlog.Info("fv menu gettitle:", f.FieldName, f.Format, icount)
					return zwords.PluralWordWithCount(f.Format, float64(icount), "", "", 0)
				}
			}
		}
		menuOwner.SelectedHandler = func() {
			v.fieldToDataItem(f, menu, false)
			if menuOwner.IsStatic {
				sel := menuOwner.SelectedItem()
				if sel != nil {
					kind := reflect.ValueOf(sel.Value).Kind()
					// zlog.Info("action pressed", kind, sel.Name, "val:", sel.Value)
					if kind != reflect.Ptr && kind != reflect.Struct {
						nf := *f
						nf.ActionValue = sel.Value
						callActionHandlerFunc(v, &nf, PressedAction, item.Interface, &view)
					}
				}
			} else {
				callActionHandlerFunc(v, f, EditedAction, item.Interface, &view)
			}
		}
	} else {
		menu := zmenu.NewView(f.Name+"Menu", items, item.Interface)
		menu.SetMaxWidth(f.MaxWidth)
		view = menu
		menu.SetSelectedHandler(func() {
			// valInterface, _ := v.fieldToDataItem(f, menu, false)
			v.fieldToDataItem(f, menu, false)
			callActionHandlerFunc(v, f, EditedAction, item.Interface, &view)
		})
	}
	return view
}

func getTimeString(item zreflect.Item, f *Field) string {
	var str string
	t := item.Interface.(time.Time)
	if t.IsZero() {
		return ""
	}
	format := f.Format
	secs := (f.Flags&FlagHasSeconds != 0)
	if format == "" {
		format = "15:04 02-Jan-06"
		if secs {
			format = "15:04:03 02-Jan-06"
		}
	}
	if format == "nice" {
		str = ztime.GetNice(t, f.Flags&FlagHasSeconds != 0)
	} else {
		str = t.Format(format)
	}
	return str
}

func getTextFromNumberishItem(item zreflect.Item, f *Field) string {
	isDurTime := item.Kind == zreflect.KindTime && f.Flags&FlagIsDuration != 0
	// zlog.Info("makeTextTime:", f.Name, isDurTime, f.Format)
	if item.Kind == zreflect.KindTime && !isDurTime {
		return getTimeString(item, f)
	}
	if isDurTime || item.Package == "time" && item.TypeName == "Duration" {
		var t float64
		if isDurTime {
			t = ztime.Since(item.Interface.(time.Time))
		} else {
			t = ztime.DurSeconds(time.Duration(item.Value.Int()))
		}
		return ztime.GetSecsAsHMSString(t, f.Flags&FlagHasSeconds != 0, 0)
		// zlog.Info("makeTextTime:", str, f.Name)
	}
	format := f.Format
	switch format {
	case "memory":
		b, err := zint.GetAny(item.Value.Interface())
		if err == nil {
			return zwords.GetMemoryString(b, "", 1)
		}
	case "storage":
		b, err := zint.GetAny(item.Value.Interface())
		if err == nil {
			return zwords.GetStorageSizeString(b, "", 1)
		}
	case "bps":
		b, err := zint.GetAny(item.Value.Interface())
		if err == nil {
			return zwords.GetBandwidthString(b, "", 1)
		}
	case "":
		format = "%v"
	}
	return fmt.Sprintf(format, item.Value.Interface())
}

func (v *FieldView) makeText(item zreflect.Item, f *Field, noUpdate bool) zview.View {
	// fmt.Printf("make Text %s %s %s %p\n:", item.FieldName, f.Title, f.Name, v.data)
	str := getTextFromNumberishItem(item, f)
	if f.IsStatic() || v.params.AllTextStatic {
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
	if f.Flags&FlagIsPassword != 0 {
		style.KeyboardType = zkeyboard.TypePassword
	}
	tv := ztext.NewView(str, style, cols, f.Rows)
	tv.SetObjectName(f.ID)
	f.SetFont(tv, nil)
	tv.UpdateSecs = f.UpdateSecs
	if !noUpdate && tv.UpdateSecs == -1 {
		tv.UpdateSecs = 4
	}
	tv.SetPlaceholder(f.Placeholder)
	tv.SetChangedHandler(func() {
		v.fieldToDataItem(f, tv, true)
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
	cv.SetObjectName(f.ID)
	cv.SetValueHandler(func() {
		val, _ := v.fieldToDataItem(f, cv, true)
		v.updateShowEnableFromZeroer(val.IsZero(), true, cv.ObjectName())
		v.updateShowEnableFromZeroer(val.IsZero(), false, cv.ObjectName())
		view := zview.View(cv)
		callActionHandlerFunc(v, f, EditedAction, val.Interface(), &view)
	})
	return cv
}

func (v *FieldView) makeImage(item zreflect.Item, f *Field) zview.View {
	iv := zimageview.New(nil, "", f.Size)
	iv.DownsampleImages = true
	iv.SetMinSize(f.Size)
	iv.SetObjectName(f.ID)
	iv.OpaqueDraw = (f.Flags&FlagIsOpaque != 0)
	if len(f.Colors) > 0 {
		iv.EmptyColor = zgeo.ColorFromString(f.Colors[0])
	}
	return iv
}

func setColorFromField(view zview.View, f *Field) {
	col := zstyle.DefaultFGColor()
	if len(f.Colors) != 0 {
		col = zgeo.ColorFromString(f.Colors[0])
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
		tooBig := true
		if !t.IsZero() {
			// zlog.Info("DUR-FROM:", t)
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

func makeFlagStack(flags zreflect.Item, f *Field) zview.View {
	stack := zcontainer.StackViewHor("flags")
	stack.SetMinSize(zgeo.Size{20, 20})
	stack.SetSpacing(2)
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

func updateFlagStack(flags zreflect.Item, f *Field, view zview.View) {
	stack := view.(*zcontainer.StackView)
	// zlog.Info("zfields.updateFlagStack", Name(f))
	bso := flags.Interface.(zbits.BitsetItemsOwner)
	bitset := bso.GetBitsetItems()
	n := flags.Value.Int()
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

func (v *FieldView) createSpecialView(item zreflect.Item, f *Field, children []zreflect.Item) (view zview.View, skip bool) {
	if f.Flags&FlagIsButton != 0 {
		if v.params.HideStatic {
			return nil, true
		}
		return v.makeButton(item, f), false
	}
	// if f.WidgetName != "" {
	// 	zlog.Info("createSpecialView?:", f.WidgetName)
	// }
	if f.Kind != zreflect.KindSlice && f.WidgetName != "" {
		// zlog.Info("createSpecialView:", f.WidgetName)
		w := widgeters[f.WidgetName]
		if w != nil {
			if w.IsStatic() && v.params.HideStatic {
				return nil, true
			}
			return w.Create(f), false
		}
	}
	callActionHandlerFunc(v, f, CreateFieldViewAction, item.Address, &view) // this sees if actual ITEM is a field handler
	if view != nil {
		return
	}
	if f.LocalEnum != "" {
		ei := findLocalFieldWithID(&children, f.LocalEnum)
		if zlog.ErrorIf(ei == nil, f.Name, f.LocalEnum) {
			return nil, true
		}
		getter, _ := ei.Interface.(zdict.ItemsGetter)
		if zlog.ErrorIf(getter == nil, "field isn't enum, not ItemGetter type", f.Name, f.LocalEnum) {
			return nil, true
		}
		enum := getter.GetItems()
		// zlog.Info("make local enum:", f.Name, f.LocalEnum, enum, ei)
		// 	continue
		// }
		//					zlog.Info("make local enum:", f.Name, f.LocalEnum)
		menu := v.makeMenu(item, f, enum)
		if menu == nil {
			zlog.Error(nil, "no local enum for", f.LocalEnum)
			return nil, true
		}
		return menu, false
		// mt := view.(zmenu.MenuType)
		//!!					mt.SelectWithValue(item.Interface)
	}
	if f.Enum != "" {
		// fmt.Println("make enum:", f.Name)
		enum, _ := fieldEnums[f.Enum]
		zlog.Assert(enum != nil, f.Enum, f.FieldName)
		view = v.makeMenu(item, f, enum)
		// exp = zgeo.AlignmentNone
		return view, false
	}
	_, got := item.Interface.(UIStringer)
	if got && (f.IsStatic() || v.params.AllTextStatic) {
		return v.makeText(item, f, false), false
	}
	return nil, false
}

func (v *FieldView) BuildStack(name string, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) {
	zlog.Assert(reflect.ValueOf(v.data).Kind() == reflect.Ptr, name, v.data)
	// fmt.Println("buildStack1", name, defaultAlign, useMinWidth)
	children := v.getStructItems()
	for j, item := range children {
		if zstr.IndexOf(item.FieldName, v.params.SkipFieldNames) != -1 {
			continue
		}
		f := findFieldWithIndex(&v.Fields, j)
		if f == nil {
			//			zlog.Error(nil, "no field for index", j)
			continue
		}
		v.buildItem(f, item, j, children, defaultAlign, cellMargin, useMinWidth)
	}
}

func (v *FieldView) buildItem(f *Field, item zreflect.Item, index int, children []zreflect.Item, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) {
	labelizeWidth := v.params.LabelizeWidth
	parentFV := ParentFieldView(v)
	if parentFV != nil && v.params.LabelizeWidth == 0 {
		labelizeWidth = parentFV.params.LabelizeWidth
	}
	exp := zgeo.AlignmentNone
	if v.params.HideStatic && f.IsStatic() {
		return
	}
	// zlog.Info("   buildStack2", j, f.Name, item)
	// if f.FieldName == "CPU" {
	// 	zlog.Info("   buildStack1.2", j, item.Value.Len())
	// }

	view, skip := v.createSpecialView(item, f, children)
	if skip {
		return
	}
	if view == nil {
		switch f.Kind {
		case zreflect.KindStruct:
			// zlog.Info("make stringer?:", f.Name, got)
			col, got := item.Interface.(zgeo.Color)
			if got {
				view = zcolor.NewView(col)
			} else {
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
				fieldView := fieldViewNew(f.ID, vert, item.Address, params, zgeo.Size{}, v)
				view = makeFrameIfFlag(f, fieldView)
				if view == nil {
					view = fieldView
				}
				// fieldView.parentField = f
				fieldView.BuildStack(f.ID, zgeo.TopLeft, zgeo.Size{}, true)
			}

		case zreflect.KindBool:
			b := zbool.ToBoolInd(item.Value.Interface().(bool))
			exp = zgeo.AlignmentNone
			view = v.makeCheckbox(f, b)

		case zreflect.KindInt:
			if item.TypeName == "BoolInd" {
				exp = zgeo.HorShrink
				view = v.makeCheckbox(f, zbool.BoolInd(item.Value.Int()))
			} else {
				_, got := item.Interface.(zbits.BitsetItemsOwner)
				if got {
					view = makeFlagStack(item, f)
					break
				}
				noUpdate := true
				view = v.makeText(item, f, noUpdate)
			}

		case zreflect.KindFloat:
			noUpdate := true
			view = v.makeText(item, f, noUpdate)

		case zreflect.KindString:
			if f.Flags&FlagIsImage != 0 {
				view = v.makeImage(item, f)
			} else {
				if (f.MaxWidth != f.MinWidth || f.MaxWidth != 0) && f.Flags&FlagIsButton == 0 {
					exp = zgeo.HorExpand
				}
				view = v.makeText(item, f, false)
			}

		case zreflect.KindSlice:
			getter, _ := item.Interface.(zdict.ItemsGetter)
			if getter != nil {
				menu := v.makeMenu(item, f, getter.GetItems())
				view = menu
				break
			}
			//				zlog.Info("Make slice:", v.ObjectName(), f.FieldName, , labelizeWidth)
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
			fv := fieldViewNew(f.ID, vert, item.Address, params, zgeo.Size{}, v)
			view = fv
			var add zview.View
			if f.Flags&FlagIsGroup == 0 {
				add = fv.buildRepeatedStackFromSlice(item.Address, vert, f)
				// zlog.Info("BUILD static slice:", f.Name, v.Hierarchy(), v.data != nil)
			} else {
				// zlog.Info("NewMenuGroup:", f.FieldName, f.TitleOrName(), f.Styling.StrokeWidth, params.Styling.StrokeWidth)
				mg := buildMenuGroup(item.Address, f.TitleOrName(), params)
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
			view = v.makeText(item, f, noUpdate)
			if f.IsStatic() || v.params.AllTextStatic {
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
				}
			}

		default:
			panic(fmt.Sprintln("buildStack bad type:", f.Name, f.Kind))
		}
	}
	pt, _ := view.(zview.Pressable)
	if pt != nil {
		ph := pt.PressedHandler()
		nowItem := item // store item in nowItem so closures below uses right item
		pt.SetPressedHandler(func() {
			// zlog.Info("FV Pressed")
			if !callActionHandlerFunc(v, f, PressedAction, nowItem.Interface, &view) && ph != nil {
				ph()
			}
		})
		// // lph := pt.LongPressedHandler()
		// pt.SetLongPressedHandler(func() {
		// 	// zlog.Info("Field.LPH:", f.ID)
		// 	if !callActionHandlerFunc(v, f, LongPressedAction, nowItem.Interface, &view) && lph != nil {
		// 		lph()
		// 	}
		// })
	}
	// zlog.Info("BuildItem:", item.FieldName, view != nil, zlog.GetCallingStackString())
	updateItemLocalToolTip(f, children, view)
	if !f.Styling.DropShadow.Delta.IsNull() {
		nv := view.Native()
		nv.SetDropShadow(f.Styling.DropShadow)
	}
	view.SetObjectName(f.ID)
	if len(f.Colors) != 0 {
		view.SetColor(zgeo.ColorFromString(f.Colors[0]))
	}
	callActionHandlerFunc(v, f, CreatedViewAction, item.Address, &view)
	cell := &zcontainer.Cell{}
	def := defaultAlign
	all := zgeo.Left | zgeo.HorCenter | zgeo.Right
	if f.Alignment&all != 0 {
		def &= ^all
	}
	cell.Margin = cellMargin
	cell.Alignment = def | exp | f.Alignment
	if labelizeWidth != 0 {
		var lstack *zcontainer.StackView
		title := f.Name
		if f.Title != "" {
			title = f.Title
		}
		if f.Flags&FlagNoTitle != 0 {
			title = ""
		}
		_, lstack, cell = zlabel.Labelize(view, title, labelizeWidth, cell.Alignment)
		v.AddView(lstack, zgeo.HorExpand|zgeo.Left|zgeo.Top)
	}
	if useMinWidth {
		cell.MinSize.W = f.MinWidth
	}
	cell.MaxSize.W = f.MaxWidth
	if f.Flags&FlagExpandFromMinSize != 0 {
		cell.ExpandFromMinSize = true
	}
	if labelizeWidth == 0 {
		cell.View = view
		v.AddCell(*cell, -1)
	}
}

func updateItemLocalToolTip(f *Field, children []zreflect.Item, view zview.View) {
	var tipField, tip string
	found := false
	if zstr.HasPrefix(f.Tooltip, ".", &tipField) {
		for _, ei := range children {
			// zlog.Info("updateItemLocalToolTip:", fieldNameToID(ei.FieldName), tipField)
			if fieldNameToID(ei.FieldName) == tipField {
				tip = fmt.Sprint(ei.Interface)
				found = true
				break
			}
		}
		if !found { // can't use tip == "" to check, since field might just be empty
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
		fview, _ := v.findNamedViewOrInLabelized(f.ID)
		if fview == nil {
			// zlog.Info("FV Update no view found:", v.id, f.ID)
			continue
		}
		_, e := v.fieldToDataItem(&f, fview, showError)
		if e != nil {
			if err == nil {
				err = e
			}
			showError = false
		}
	}
	return
}

func (v *FieldView) fieldToDataItem(f *Field, view zview.View, showError bool) (value reflect.Value, err error) {
	if f.IsStatic() {
		return
	}
	children := v.getStructItems()
	// zlog.Info("fieldViewToDataItem before:", f.Name, f.Index, len(children), "s:", data)
	item := children[f.Index]
	if (f.Enum != "" || f.LocalEnum != "") && !f.IsStatic() {
		mv, _ := view.(*zmenu.MenuView)
		if mv != nil {
			iface := mv.CurrentValue()
			vo := reflect.ValueOf(iface)
			// zlog.Debug(iface, f.Name, iface == nil)
			if iface == nil {
				vo = reflect.Zero(item.Value.Type())
			}
			item.Value.Set(vo)
		}
		return
	}

	switch f.Kind {
	case zreflect.KindBool:
		bv, _ := view.(*zcheckbox.CheckBox)
		if bv == nil {
			panic("Should be checkbox")
		}
		b, _ := item.Address.(*bool)
		if b != nil {
			*b = bv.Value().Bool()
			// zlog.Info("SetCheck:", bv.Value(), *b, value)
		}
		bi, _ := item.Address.(*zbool.BoolInd)
		if bi != nil {
			*bi = bv.Value()
		}

	case zreflect.KindInt:
		if item.TypeName == "BoolInd" {
			bv, _ := view.(*zcheckbox.CheckBox)
			*item.Address.(*zbool.BoolInd) = bv.Value()
		} else {
			tv, _ := view.(*ztext.TextView)
			str := tv.Text()
			if item.Package == "time" && item.TypeName == "Duration" {
				var secs float64
				secs, err = ztime.GetSecsFromHMSString(str, f.Flags&FlagHasHours != 0, f.Flags&FlagHasMinutes != 0, f.Flags&FlagHasSeconds != 0)
				if err != nil {
					break
				}
				d := item.Address.(*time.Duration)
				if d != nil {
					*d = ztime.SecondsDur(secs)
				}
				return
			}
			var i64 int64
			if str != "" {
				i64, err = strconv.ParseInt(str, 10, 64)
				if err != nil {
					break
				}
			}
			zint.SetAny(item.Address, i64)
		}

	case zreflect.KindFloat:
		tv, _ := view.(*ztext.TextView)
		var f64 float64
		f64, err = strconv.ParseFloat(tv.Text(), 64)
		if err != nil {
			break
		}
		zfloat.SetAny(item.Address, f64)
		// zlog.Info("fieldToDataItem float", f.FieldName, view.ObjectName(), tv.Text(), f64, err, item)
		// fmt.Printf("fieldToDataItem struct: %+v\n", item.Value.Interface())

	case zreflect.KindTime:
		break

	case zreflect.KindString:
		if (!f.IsStatic() && !v.params.AllTextStatic) && f.Flags&FlagIsImage == 0 {
			tv, _ := view.(*ztext.TextView)
			if tv == nil {
				zlog.Fatal(nil, "Copy Back string not TV:", f.Name)
			}
			text := tv.Text()
			str := item.Address.(*string)
			*str = text
		}

	case zreflect.KindFunc:
		break

	case zreflect.KindStruct:
		fv, _ := view.(*FieldView)
		zlog.Info("ToData struct:", f.Name, fv != nil)

	default:
		panic(fmt.Sprint("bad type: ", f.Kind))
	}

	if showError && err != nil {
		zalert.ShowError(err)
	}
	value = reflect.ValueOf(item.Address).Elem() //.Interface()
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

func (fv *FieldView) getStructItems() []zreflect.Item {
	k := reflect.ValueOf(fv.data).Kind()
	// zlog.Info("getStructItems", direct, k, sub)
	zlog.Assert(k == reflect.Ptr, "not pointer", k)
	options := zreflect.Options{UnnestAnonymous: true, Recursive: false}
	rootItems, err := zreflect.ItterateStruct(fv.data, options)
	if err != nil {
		panic(err)
	}
	// zlog.Info("Get Struct Items sub:", len(rootItems.Children))
	return rootItems.Children
}

func PresentOKCancelStruct[S any](structPtr *S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) bool) {
	slice := []S{*structPtr}
	//	zlog.Info("PresentOKCancelStruct", structType)
	PresentOKCancelStructSlice(&slice, params, title, att, func(ok bool) bool {
		if ok {
			zlog.Info("OK:", slice[0])
			*structPtr = slice[0]
		}
		return done(ok)
	})
}

/*
func PresentOKCancelStruct_Old(structPtr interface{}, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) bool) {
	// var structCopy any
	// err := zreflect.DeepCopy(&structCopy, reflect.ValueOf(structPtr).Elem())
	// zlog.AssertNotError(err, "copy")
	// zlog.Info("Copy:", structCopy)
	fview := FieldViewNew("OkCancel", structPtr, params)
	update := true
	fview.Build(update)
	params.EditWithoutCallsbacks = true
	zalert.PresentOKCanceledView(fview, title, att, func(ok bool) bool {
		if ok {
			err := fview.ToData(true)
			if err != nil {
				return false
			}
		}
		return done(ok)
	})
}
*/
