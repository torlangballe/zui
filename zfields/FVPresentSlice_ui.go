//go:build zui

package zfields

import (
	"reflect"
	"strings"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
)

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

func reduceLocalEnumField(editStructPtr any, enumField reflect.Value, index int, fromStruct reflect.Value, f *Field) {
	ei, findex := FindLocalFieldWithFieldName(editStructPtr, f.LocalEnum)
	if zlog.ErrorIf(findex == -1, f.Name, f.LocalEnum) {
		return
	}
	zlog.Assert(ei.Kind() == reflect.Slice)
	field := zreflect.FieldForIndex(fromStruct.Interface(), FlattenIfAnonymousOrZUITag, findex)
	fromVal := field.ReflectValue
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

func reduceSliceField(reduceSlice, fromSlice reflect.Value) {
	var reduced bool
	for i := 0; i < reduceSlice.Len(); {
		rval := reduceSlice.Index(i).Interface()
		var has bool
		for j := 0; j < fromSlice.Len(); j++ {
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

func PresentOKCancelStructAnySlice(structSlicePtr any, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	sliceVal := reflect.ValueOf(structSlicePtr)
	// zlog.Info("PresentOKCancelStructAnySlice:", sliceVal.Type())
	if sliceVal.Kind() == reflect.Pointer {
		sliceVal = sliceVal.Elem()
	}
	editStructRVal := zslice.MakeAnElementOfSliceRValType(sliceVal)
	editStructRVal.Set(sliceVal.Index(0)) // we make a copy, not just first
	editStruct := editStructRVal.Addr().Interface()
	sliceLength := sliceVal.Len()
	unknownBoolViewIDs := map[string]bool{}

	// zlog.Info("PresentOKCancelStructAnySlice:", zlog.Full(editStruct))
	ForEachField(editStruct, params.FieldParameters, nil, func(each FieldInfo) bool {
		var notEqual bool
		// zlog.Info("before PresentOKCancelStructAnySlice.ForEachField:", each.Field.Name, sliceLength)
		for i := 0; i < sliceLength; i++ {
			finfo := zreflect.FieldForIndex(sliceVal.Index(i).Interface(), FlattenIfAnonymousOrZUITag, each.FieldIndex) // (fieldRefVal reflect.Value, sf reflect.StructField) {
			sliceField := finfo.ReflectValue

			if !sliceField.CanInterface() || !each.ReflectValue.CanInterface() {
				continue
			}
			setupWidgeter(each.Field)
			if !reflect.DeepEqual(sliceField.Interface(), each.ReflectValue.Interface()) {
				if each.Field.IsStatic() {
					if each.ReflectValue.Kind() == reflect.Slice {
						// zlog.Info("IsStatEq set:", each.StructField.Name, sliceField.Interface(), each.ReflectValue.Interface())
						accumilateSlice(each.ReflectValue, sliceField)
					} else {
						// zlog.Info("IsStatEq clear:", each.StructField.Name, sliceField.Interface(), each.ReflectValue.Interface())
						each.ReflectValue.Set(reflect.Zero(each.ReflectValue.Type()))
					}
				} else {
					// if each.Field.Enum != "" {
					// reduceEnumField(val, sliceField, f.Enum)
					// } else
					if each.Field.LocalEnum != "" {
						reduceLocalEnumField(editStruct, each.ReflectValue, each.FieldIndex, sliceVal.Index(i), each.Field)
					} else if each.ReflectValue.Kind() == reflect.Slice {
						// zlog.Info("IsEq Reduce:", each.StructField.Name, sliceField.Interface(), each.ReflectValue.Interface())
						reduceSliceField(each.ReflectValue, sliceField)
					} else {
						// zlog.Info("IsEq Set zero:", i, each.StructField.Name, sliceField.Interface(), each.ReflectValue.Interface())
						each.ReflectValue.Set(reflect.Zero(each.ReflectValue.Type()))
					}
				}
				notEqual = true
				break
			}
		}
		// zlog.Info("ForEach:", f.Name, val.Type(), val.Interface(), notEqual, f.Enum)
		if notEqual {
			if each.ReflectValue.Kind() == reflect.Bool {
				unknownBoolViewIDs[each.StructField.Name] = true
				// zslice.AddEmptyElementAtEnd(val.Addr().Interface())
			}
		}
		return true
	})

	params.MultiSliceEditInProgress = (sliceLength > 1)
	params.UseInValues = []string{DialogUseInSpecialName}
	fview := FieldViewNew("OkCancel", editStruct, params)
	update := true
	fview.Build(update)
	for bid := range unknownBoolViewIDs {
		view, _ := fview.FindNamedViewOrInLabelized(bid)
		check, _ := view.(*zcheckbox.CheckBox)
		if check != nil {
			check.SetValue(zbool.Unknown)
		}
	}

	var wildCards zview.View
	if sliceLength > 1 {
		wild := zlabel.New("Use *x*->*y* to replace x in each item with y,\nreplacing wildcards with their matches")
		wild.SetMaxLines(2)
		wildCards = wild // We need to
		tv := getFocusedEmptyTextView(&fview.NativeView)
		wild.Show(tv != nil)
		wild.SetCanTabFocus(false)
		wild.SetTextAlignment(zgeo.TopCenter)
		wild.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
		wild.SetColor(zgeo.ColorGray)
		wild.SetPressedDownHandler(func() {
			tv := getFocusedEmptyTextView(wild.Parent())
			if tv != nil {
				tv.SetText("*x*->*y*")
			}
		})
		att.PresentedFunc = func(win *zwindow.Window) {
			fview.HandleFocusInChildren(true, false, func(view zview.View, focused bool) {
				fvp := ParentFieldView(view)
				zlog.Info("FOC FV:", fvp.Hierarchy(), fview.Hierarchy())
				focusInTopFV := (fvp == fview)
				tv, _ := view.(*ztext.TextView)
				wild.Show(focusInTopFV && tv != nil && tv.Text() == "")
			})
		}
	}
	originalStruct := zreflect.CopyAny(editStruct)
	zalert.PresentOKCanceledView(fview, title, att, wildCards, func(ok bool) (close bool) {
		if ok {
			err := fview.ToData(true)
			if err != nil {
				return false
			}
			ForEachField(editStruct, params.FieldParameters, nil, func(each FieldInfo) bool {
				// zlog.Info("origFieldReflectValue1:", each.Field.Name, each.ReflectValue.Interface(), each.Field.Flags, each.Field.IsStatic(), FlagIsButton, FlagIsStatic)
				if each.StructField.Tag.Get("zui") == "-" {
					return true // skip to next
				}
				bid := each.StructField.Name
				view, _ := fview.FindNamedViewOrInLabelized(bid)
				check, _ := view.(*zcheckbox.CheckBox)
				isCheck := (check != nil)
				if isCheck && check.Value().IsUnknown() {
					return true // skip to next
				}
				f := findFieldWithIndex(&fview.Fields, each.FieldIndex)
				if f.IsStatic() {
					return true // means continue
				}
				finfo, _ := zreflect.FieldForName(originalStruct, FlattenIfAnonymousOrZUITag, each.StructField.Name)
				origFieldReflectValue := finfo.ReflectValue
				// zlog.Info("origFieldReflectValue:", each.Field.Name, origFieldReflectValue.Interface())
				if params.MultiSliceEditInProgress && each.ReflectValue.Kind() == reflect.Slice {
					if f.Enum == "" && f.StringSep == "" { // todo: let it through if its a UISetStringer
						// zlog.Info("Skip non-enum slice in multi-edit:", each.StructField.Name)
						return true // skip to next
					}
				}
				var wildTransformer *zstr.WildCardTransformer
				if !origFieldReflectValue.IsZero() && sliceLength > 1 && origFieldReflectValue.Kind() == reflect.String {
					var wildFrom, wildTo string
					if zstr.SplitN(each.ReflectValue.String(), "->", &wildFrom, &wildTo) && strings.Contains(wildFrom, "*") && strings.Contains(wildTo, "*") {
						tv, _ := view.(*ztext.TextView)
						if tv != nil {
							wildTransformer, err = zstr.NewWildCardTransformer(wildFrom, wildTo)
							if err != nil {
								zalert.ShowError(err)
							}
						}
					}
				}
				if wildTransformer != nil {
					for i := 0; i < sliceLength; i++ {
						finfo := zreflect.FieldForIndex(sliceVal.Index(i).Addr().Interface(), FlattenIfAnonymousOrZUITag, each.FieldIndex) // (fieldRefVal reflect.Value, sf reflect.StructField) {
						sliceField := finfo.ReflectValue
						replaced, err := wildTransformer.Transform(sliceField.String())
						if err != nil {
							zlog.Error(err)
						} else {
							sliceField.SetString(replaced)
						}
					}
					return true
				}
				if !(origFieldReflectValue.IsZero() && each.ReflectValue.IsZero()) || sliceLength == 1 || isCheck {
					for i := 0; i < sliceLength; i++ {
						itemRVal := sliceVal.Index(i)
						if itemRVal.Kind() != reflect.Pointer {
							itemRVal = itemRVal.Addr()
						}
						finfo := zreflect.FieldForIndex(itemRVal.Interface(), FlattenIfAnonymousOrZUITag, each.FieldIndex) // (fieldRefVal reflect.Value, sf reflect.StructField) {
						sliceField := finfo.ReflectValue
						sliceField.Set(each.ReflectValue)
					}
				}
				return true
			})
		}
		return done(ok)
	})
}

func PresentOKCancelStructSlice[S any](structSlicePtr *[]S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	PresentOKCancelStructAnySlice(structSlicePtr, params, title, att, done)
}

func getFocusedEmptyTextView(parent *zview.NativeView) *ztext.TextView {
	tv, _ := parent.GetFocusedChildView(false).(*ztext.TextView)
	if tv != nil && tv.Text() == "" {
		return tv
	}
	return nil
}
