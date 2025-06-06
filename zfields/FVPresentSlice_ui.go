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

func accumulateSlice(accSlice, fromSlice reflect.Value) {
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

// reduceLocalEnumField removes all enum key/vals in enumField that are not in fromVal.
// It is run on local enum fields for each slice value om enumField, except first, which is editStructPtr.
func reduceLocalEnumField(editStructPtr any, enumField reflect.Value, index int, fromStruct reflect.Value, f *Field) {
	ei, findex := FindLocalFieldWithFieldName(editStructPtr, f.LocalEnum)
	if zlog.ErrorIf(findex == -1, f.Name, f.LocalEnum) {
		return
	}
	zslice.CopyTo(ei.Addr().Interface(), ei.Interface())
	zlog.Assert(ei.Kind() == reflect.Slice)
	field := zreflect.FieldForIndex(fromStruct.Interface(), FlattenIfAnonymousOrZUITag, findex)
	fromVal := field.ReflectValue
	// zlog.Info("reduceLocalEnumField", f.Name, findex, ei.Len(), ei.Type(), fromStruct.Type(), fromStruct.Kind())
	for i := 0; i < ei.Len(); {
		eval := ei.Index(i)
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
		}
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

func EditStructAnySlice(structSlicePtr any, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	EditOrViewStructAnySlice(structSlicePtr, false, params, title, att, done)
}

func ViewStructAnySlice(structSlicePtr any, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	EditOrViewStructAnySlice(structSlicePtr, true, params, title, att, done)
}

func EditOrViewStructAnySlice(structSlicePtr any, isReadOnly bool, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	sliceVal := reflect.ValueOf(structSlicePtr)
	// zlog.Info("PresentEditOrViewStructAnySlice:", sliceVal.Type())
	if sliceVal.Kind() == reflect.Pointer {
		sliceVal = sliceVal.Elem()
	}
	editStructRVal := zslice.MakeAnElementOfSliceRValType(sliceVal)
	editStructRVal.Set(sliceVal.Index(0)) // we make a copy, not just first
	editStruct := editStructRVal.Addr().Interface()
	sliceLength := sliceVal.Len()
	unknownBoolViewIDs := map[string]bool{}
	params.FieldParameters.UseInValues = []string{DialogUseInSpecialName}
	params.MultiSliceEditInProgress = (sliceLength > 1)
	wasAllNotZero := map[string]bool{}

	ForEachField(editStruct, params.FieldParameters, nil, func(each FieldInfo) bool {
		var notEqual bool
		var notZero bool
		for i := 0; i < sliceLength; i++ {
			finfo := zreflect.FieldForIndex(sliceVal.Index(i).Interface(), FlattenIfAnonymousOrZUITag, each.FieldIndex) // (fieldRefVal reflect.Value, sf reflect.StructField) {
			sliceField := finfo.ReflectValue

			if !sliceField.CanInterface() || !each.ReflectValue.CanInterface() {
				continue
			}
			setupWidgeter(each.Field)
			if each.Field.LocalEnum != "" {
				if i != 0 {
					reduceLocalEnumField(editStruct, each.ReflectValue, each.FieldIndex, sliceVal.Index(i), each.Field)
				}
				continue
			}
			if each.ReflectValue.Kind() == reflect.Slice {
				if i != 0 {
					reduceSliceField(each.ReflectValue, sliceField)
				}
				continue
			} else if !each.ReflectValue.IsZero() {
				notZero = true
			}
			if !reflect.DeepEqual(sliceField.Interface(), each.ReflectValue.Interface()) {
				if each.Field.IsStatic() {
					if each.ReflectValue.Kind() == reflect.Slice {
						// zlog.Info("IsStatEq set:", each.StructField.Name, sliceField.Interface(), each.ReflectValue.Interface())
						accumulateSlice(each.ReflectValue, sliceField)
					} else {
						// zlog.Info("IsStatEq clear:", each.StructField.Name, sliceField.Interface(), each.ReflectValue.Interface())
						each.ReflectValue.Set(reflect.Zero(each.ReflectValue.Type()))
					}
				} else {
					each.ReflectValue.Set(reflect.Zero(each.ReflectValue.Type()))
				}
				notEqual = true
				break
			}
		}
		// zlog.Info("ForEach:", each.Field.Name, each.ReflectValue.Type(), each.ReflectValue.Interface(), notEqual, each.Field.Enum)
		if notEqual {
			if each.ReflectValue.Kind() == reflect.Bool {
				unknownBoolViewIDs[each.StructField.Name] = true
				// zslice.AddEmptyElementAtEnd(val.Addr().Interface())
			}
		} else if notZero {
			wasAllNotZero[each.StructField.Name] = true
		}
		return true
	})

	fview := FieldViewNew("OkCancel", editStruct, params)
	fview.SetCanTabFocus(true) // we need view to have focus to avoid things below modal still having focus
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
		wild.SetPressedDownHandler("", 0, func() {
			tv := getFocusedEmptyTextView(wild.Parent())
			if tv != nil {
				tv.SetText("*x*->*y*")
			}
		})
		att.PresentedFunc = func(win *zwindow.Window) {
			fview.HandleFocusInChildren(true, false, func(view zview.View, focused bool) {
				fvp := ParentFieldView(view)
				// zlog.Info("FOC FV:", fvp.Hierarchy(), fview.Hierarchy())
				focusInTopFV := (fvp == fview)
				tv, _ := view.(*ztext.TextView)
				wild.Show(focusInTopFV && tv != nil && tv.Text() == "")
			})
		}
	}
	// originalStruct := zreflect.CopyAny(editStruct)
	if isReadOnly {
		att.FocusView = fview
		zpresent.PresentTitledView(fview, title, att, nil, nil)
		return
	}
	// zlog.Info("EDIT Struct:", zlog.Full(originalStruct))
	zalert.PresentOKCanceledView(fview, title, att, wildCards, func(ok bool) (close bool) {
		var doClose = new(zbool.BoolInd)
		if ok {
			err := fview.ToData(true)
			if err != nil {
				return false
			}
			hasRequiredGroups := map[string][]string{}
			ForEachField(editStruct, params.FieldParameters, nil, func(each FieldInfo) bool {
				// zlog.Info("origFieldReflectValue1:", each.Field.Name, each.ReflectValue.Interface(), each.Field.Flags, each.Field.IsStatic(), FlagIsButton, FlagIsStatic)
				if each.StructField.Tag.Get("zui") == "-" {
					zlog.Info("SHOULD THIS HAPPEN?")
					return true // skip to next
				}
				if !params.MultiSliceEditInProgress && each.Field.Required != "" {
					zero := each.ReflectValue.IsZero()
					if each.Field.Required == RequiredSingleValue {
						if zero {
							zalert.Show("Field '" + each.Field.TitleOrName() + "' can't be empty")
							doClose.FromBool(false)
							return false
						}
					} else {
						if zero {
							g, has := hasRequiredGroups[each.Field.Required]
							// zlog.Info("Has:", has, g, each.Field.Required, each.StructField.Name)
							if !has || len(g) > 0 {
								hasRequiredGroups[each.Field.Required] = append(hasRequiredGroups[each.Field.Required], each.Field.TitleOrName())
							}
						} else {
							hasRequiredGroups[each.Field.Required] = []string{}
						}
					}

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
				//! finfo, _ := zreflect.FieldForName(originalStruct, FlattenIfAnonymousOrZUITag, each.StructField.Name)
				//! origFieldReflectValue := finfo.ReflectValue
				// zlog.Info("origFieldReflectValue:", each.Field.Name, origFieldReflectValue.Interface())
				if params.MultiSliceEditInProgress && each.ReflectValue.Kind() == reflect.Slice {
					if f.Enum == "" && f.StringSep == "" { // todo: let it through if its a UISetStringer
						// zlog.Info("Skip non-enum slice in multi-edit:", each.StructField.Name)
						return true // skip to next
					}
				}
				var wildTransformer *zstr.WildCardTransformer
				var wildFrom, wildTo string
				if sliceLength > 1 && each.ReflectValue.Kind() == reflect.String { // !origFieldReflectValue.IsZero() &&
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
							zlog.Error("transform", err)
						} else {
							sliceField.SetString(replaced)
						}
					}
					return true
				}
				// zlog.Info("each.ReflectValue:", each.Field.Name, each.ReflectValue.Interface())
				//				if !(origFieldReflectValue.IsZero() && each.ReflectValue.IsZero()) || sliceLength == 1 || isCheck {
				if (!each.ReflectValue.IsZero() || wasAllNotZero[each.StructField.Name]) || sliceLength == 1 || isCheck {
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
			for _, fields := range hasRequiredGroups {
				if len(fields) > 0 {
					doClose.FromBool(false)
					zalert.Show("All of fields:", strings.Join(fields, "/"), "can't be empty")
					break
				}
			}
		}
		if ok && !doClose.IsUnknown() {
			return doClose.IsTrue()
		}
		return done(ok)
	})
}

func EditStructSlice[S any](structSlicePtr *[]S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	EditOrViewStructSlice(structSlicePtr, false, params, title, att, done)
}

func ViewStructSlice[S any](structSlicePtr *[]S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	EditOrViewStructSlice(structSlicePtr, true, params, title, att, done)
}

func EditOrViewStructSlice[S any](structSlicePtr *[]S, isReadOnly bool, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	EditOrViewStructAnySlice(structSlicePtr, isReadOnly, params, title, att, done)
}

func getFocusedEmptyTextView(parent *zview.NativeView) *ztext.TextView {
	tv, _ := parent.GetFocusedChildView(false).(*ztext.TextView)
	if tv != nil && tv.Text() == "" {
		return tv
	}
	return nil
}
