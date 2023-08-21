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

func reduceLocalEnumField[S any](editStruct *S, enumField reflect.Value, index int, fromStruct reflect.Value, f *Field) {
	ei, findex := FindLocalFieldWithFieldName(editStruct, f.LocalEnum)
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

func PresentOKCancelStructSlice[S any](structSlicePtr *[]S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
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
			setupWidgeter(f)
			if !reflect.DeepEqual(sliceField.Interface(), val.Interface()) {
				if f.IsStatic() {
					if val.Kind() == reflect.Slice {
						accumilateSlice(val, sliceField)
					} else {
						val.Set(reflect.Zero(val.Type()))
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
		view, _ := fview.FindNamedViewOrInLabelized(bid)
		check, _ := view.(*zcheckbox.CheckBox)
		if check != nil {
			check.SetValue(zbool.Unknown)
		}
	}

	var wildCards zview.View
	if len(*structSlicePtr) > 1 {
		wild := zlabel.New("Use *x*->*y* to replace x with y,\nreplacing wildcards with their matches")
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
			// zlog.Info("Presented!")
			fview.HandleFocusInChildren(true, false, func(view zview.View, focused bool) {
				tv, _ := view.(*ztext.TextView)
				// zlog.Info("FOCUSED:", view.Native().Hierarchy(), tv != nil)
				wild.Show(tv != nil && tv.Text() == "")
			})
			// tv := getFocusedEmptyTextView(&fview.NativeView)
			// wild.Show(tv != nil)
		}
	}
	zalert.PresentOKCanceledView(fview, title, att, wildCards, func(ok bool) (close bool) {
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
				view, _ := fview.FindNamedViewOrInLabelized(bid)
				check, _ := view.(*zcheckbox.CheckBox)
				isCheck := (check != nil)
				if isCheck && check.Value().IsUnknown() {
					return true // skip to next
				}
				f := findFieldWithIndex(&fview.Fields, index)
				if f.IsStatic() {
					return true // continue
				}
				if params.MultiSliceEditInProgress && val.Kind() == reflect.Slice {
					if f.Enum == "" {
						zlog.Info("Skip non-enum slice in multi-edit:", sf.Name)
						return true // skip to next
					}
				}
				var wildTransformer *zstr.WildCardTransformer
				if !val.IsZero() && length > 1 && val.Kind() == reflect.String {
					var wildFrom, wildTo string
					if zstr.SplitN(val.String(), "->", &wildFrom, &wildTo) && strings.Contains(wildFrom, "*") && strings.Contains(wildTo, "*") {
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
					for i := 0; i < length; i++ {
						sliceField := sliceVal.Index(i).Field(index)
						replaced, err := wildTransformer.Transform(sliceField.String())
						if err != nil {
							zlog.Error(err)
						} else {
							sliceField.SetString(replaced)
						}
					}
					return true
				}
				if !val.IsZero() || length == 1 || isCheck {
					for i := 0; i < length; i++ {
						sliceField := sliceVal.Index(i).Field(index)
						sliceField.Set(val)
					}
				}
				return true
			})
		}
		return done(ok)
	})
}

func getFocusedEmptyTextView(parent *zview.NativeView) *ztext.TextView {
	tv, _ := parent.GetFocusedChildView(false).(*ztext.TextView)
	if tv != nil && tv.Text() == "" {
		return tv
	}
	return nil
}
