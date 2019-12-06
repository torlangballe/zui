package zgo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztime"

	"github.com/torlangballe/zutil/ustr"
	"github.com/torlangballe/zutil/zreflect"
)

type fieldType int

const (
	fieldIsStatic = 1 << iota
	fieldHasSeconds
	fieldHasMinutes
	fieldHasHours
	fieldHasDays
	fieldHasMonths
	fieldHasYears
	fieldIsImage
	fieldsNoHeader
)
const (
	fieldTimeFlags = fieldHasSeconds | fieldHasMinutes | fieldHasHours
	fieldDateFlags = fieldHasDays | fieldHasMonths | fieldHasYears
)

type FieldsActionFunc func(i int, rowPtr interface{})

type FieldsActionType interface {
	Action(i int, rowPtr interface{})
}

type field struct {
	Index         int
	ID            string
	Name          string
	Title         string // name of item in row, and header if no title
	Width         float64
	MaxWidth      float64
	MinWidth      float64
	Kind          zreflect.TypeKind
	Alignment     zgeo.Alignment
	Justify       zgeo.Alignment
	Format        string
	Color         string
	FixedPath     string
	Password      bool
	Height        float64
	Weight        float64
	Enum          zdict.Dict
	LocalEnum     string
	Size          zgeo.Size
	Flags         int
	DefaultWeight float64
	// Type          fieldType
}

type FieldView struct {
	StackView
	parentField *field
	fields      []field
	structure   interface{}
}

func callFieldFunc(i int, structData interface{}, f *field) {
	unnestAnon := true
	recursive := false
	rootItems, _ := zreflect.ItterateStruct(structData, unnestAnon, recursive)
	item := rootItems.Children[f.Index]
	if item.Kind == zreflect.KindFunc && !item.Value.IsNil() {
		var args []reflect.Value
		if reflect.TypeOf(item.Interface).NumIn() == 2 {
			args = []reflect.Value{
				reflect.ValueOf(i),
				reflect.ValueOf(structData),
			}
		}
		go item.Value.Call(args)
		return
	}
	action, _ := item.Interface.(FieldsActionType)
	if action != nil { //!item.Value.IsNil() {
		go action.Action(i, structData)
	}
}

func fieldsMakeButton(i int, structData interface{}, height float64, f *field, item zreflect.Item) *Button {
	format := f.Format
	if format == "" {
		format = "%v"
	}
	color := f.Color
	if color == "" {
		color = "gray"
	}
	name := f.Name
	if f.Title != "" {
		name = f.Title
	}
	button := ButtonNew(name, color, zgeo.Size{40, height}, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
	button.TextInfo.Color = zgeo.ColorRed
	button.PressedHandler(func() {
		callFieldFunc(i, structData, f)
	})
	return button
}

func fieldsMakeEnumMenu(f *field, item zreflect.Item, i int, handleUpdate func(i int)) View {
	menu := MenuViewNew(f.Enum, item.Value)
	menu.ChangedHandler(func(key string, val interface{}) {
		if handleUpdate != nil {
			handleUpdate(i)
		}
	})
	return menu
}

func fieldsMakeLocalEnumMenu(item zreflect.Item, enumIndex int, enum zdict.Dict, handleUpdate func(i int)) View {
	menu := MenuViewNew(enum, item.Value)
	menu.ChangedHandler(func(key string, val interface{}) {
		if handleUpdate != nil {
			handleUpdate(enumIndex)
		}
	})
	return menu
}

func fieldsMakeTimeView(f *field, item zreflect.Item, i int, handleUpdate func(i int)) View {
	t := item.Interface.(time.Time)
	format := f.Format
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	str := t.Format(format)
	tv := TextViewNew(str)
	return tv
}

func fieldsMakeText(f *field, item zreflect.Item, i int, handleUpdate func(i int)) View {
	str := ""
	if item.Package == "time" && item.TypeName == "Duration" {
		t := ztime.DurSeconds(time.Duration(item.Value.Int()))
		str = ztime.GetSecsAsHMSString(t, f.Flags&fieldHasHours != 0, f.Flags&fieldHasMinutes != 0, f.Flags&fieldHasSeconds != 0, 0)
	} else {
		format := f.Format
		if format == "" {
			format = "%v"
		}
		str = fmt.Sprintf(format, item.Value.Interface())
	}
	if f.Flags&fieldIsStatic != 0 {
		label := LabelNew(str)
		j := f.Justify
		if j == zgeo.AlignmentNone {
			j = f.Alignment & (zgeo.AlignmentLeft | zgeo.AlignmentHorCenter | zgeo.AlignmentRight)
			if j == zgeo.AlignmentNone {
				j = zgeo.AlignmentLeft
			}
		}
		label.TextAlignment(j)
		return label
	}
	tv := TextViewNew(str)
	if handleUpdate != nil {
		tv.ChangedHandler(func(view View) {
			handleUpdate(i)
		})
	}
	tv.KeyHandler(func(view View, key int) {
		fmt.Println("keyup!")
	})
	return tv
}

func fieldsMakeCheckbox(b BoolInd, i int, handleUpdate func(i int)) View {
	c := CheckBoxNew(b)
	if handleUpdate != nil {
		c.ValueHandler(func(v View) {
			handleUpdate(i)
		})
	}
	return c
}

func fieldsMakeImage(structData interface{}, f *field, i int) View {
	iv := ImageViewNew("", f.Size)
	iv.ObjectName(f.ID)
	iv.PressedHandler(func() {
		callFieldFunc(i, structData, f)
	})
	return iv
}

func findFieldWithIndex(fields *[]field, index int) *field {
	for i, f := range *fields {
		if f.Index == index {
			return &(*fields)[i]
		}
	}
	return nil
}

func fieldsUpdateStack(stack *StackView, structData interface{}, fields *[]field) {
	unnestAnon := true
	recursive := true
	rootItems, err := zreflect.ItterateStruct(structData, unnestAnon, recursive)

	if err != nil {
		panic(err)
	}
	for i, item := range rootItems.Children {
		f := findFieldWithIndex(fields, i)
		if f == nil {
			continue
		}
		fview := stack.FindViewWithName(f.ID, false)
		if fview == nil {
			continue
		}
		view := *fview
		switch f.Kind {
		case zreflect.KindStruct:
			break
		case zreflect.KindBool:
			b := BoolIndFromBool(item.Value.Interface().(bool))
			cv := view.(*CheckBox)
			v := cv.GetValue()
			if v != b {
				cv.Value(b)
			}

		case zreflect.KindString, zreflect.KindFunc:
			if f.Flags&fieldIsImage != 0 {
				path := ""
				if f.Kind == zreflect.KindString {
					path = item.Value.String()
				}
				if path == "" {
					path = f.FixedPath
				}
				iv := view.(*ImageView)
				iv.SetImage(nil, path, nil)
			} else {

			}
		}

	}
}

func fieldsBuildStack(fv *FieldView, stack *StackView, structData interface{}, parentField *field, fields *[]field, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool, inset float64, i int, handleUpdate func(i int)) {
	// if fv != nil && fv.parentField != nil {
	// 	fmt.Println("fieldsBuildStack2:", fv.GetObjectName())
	// }
	unnestAnon := true
	recursive := true
	rootItems, err := zreflect.ItterateStruct(structData, unnestAnon, recursive)
	if err != nil {
		panic(err)
	}
	for j, item := range rootItems.Children {
		exp := zgeo.AlignmentNone
		var view View
		f := findFieldWithIndex(fields, j)
		if f == nil {
			continue
		}
		if f.LocalEnum != "" {
			for eIndex, ei := range rootItems.Children {
				if ei.FieldName == f.LocalEnum {
					enum := ei.Interface.(zdict.Dict)
					zlog.Assert(enum != nil)
					view = fieldsMakeLocalEnumMenu(item, eIndex, enum, handleUpdate)
				}
			}
			if view == nil {
				zlog.Error(nil, "no local enum for", f.LocalEnum)
				continue
			}
		} else if f.Enum != nil {
			view = fieldsMakeEnumMenu(f, item, i, handleUpdate)
			exp = zgeo.AlignmentHorShrink
		} else {
			switch f.Kind {
			case zreflect.KindStruct:
				// fmt.Println("struct:", f.Kind, j, item.Value)
				childStruct := item.Address
				fieldView := fieldViewNew(f.Name, false, childStruct, 10, zgeo.Size{})
				fieldView.parentField = f
				view = fieldView
				fieldsBuildStack(fieldView, &fieldView.StackView, fieldView.structure, fieldView.parentField, &fieldView.fields, zgeo.AlignmentLeft|zgeo.AlignmentTop, zgeo.Size{}, true, 5, 0, handleUpdate)

			case zreflect.KindBool:
				b := BoolIndFromBool(item.Value.Interface().(bool))
				view = fieldsMakeCheckbox(b, i, handleUpdate)

			case zreflect.KindInt:
				if item.TypeName == "BoolInd" {
					exp = zgeo.AlignmentHorShrink
					view = fieldsMakeCheckbox(BoolInd(item.Value.Int()), i, handleUpdate)
				} else {
					view = fieldsMakeText(f, item, i, handleUpdate)
				}

			case zreflect.KindFloat:
				view = fieldsMakeText(f, item, i, handleUpdate)

			case zreflect.KindString:
				if f.Flags&fieldIsImage != 0 {
					view = fieldsMakeImage(structData, f, i)
				} else {
					exp = zgeo.AlignmentHorExpand
					view = fieldsMakeText(f, item, i, handleUpdate)
				}

			case zreflect.KindFunc:
				if f.Flags&fieldIsImage != 0 {
					view = fieldsMakeImage(structData, f, i)
				} else {
					view = fieldsMakeButton(i, structData, f.Height, f, item)
				}

			case zreflect.KindSlice:
				view = fieldsMakeText(f, item, i, handleUpdate)
				break

			case zreflect.KindTime:
				view = fieldsMakeTimeView(f, item, i, handleUpdate)

			default:
				panic(fmt.Sprint("fieldsBuildStack bad type: ", f.Kind))
			}
		}
		view.ObjectName(f.ID)
		cell := ContainerViewCell{}
		cell.Margin = cellMargin
		def := defaultAlign
		all := zgeo.AlignmentLeft | zgeo.AlignmentHorCenter | zgeo.AlignmentRight
		if f.Alignment&all != 0 {
			def &= ^all
		}
		cell.Alignment = def | exp | f.Alignment
		// fmt.Println("field align:", f.Alignment, f.Name, def|exp, cell.Alignment, int(cell.Alignment))
		cell.Weight = f.Weight
		if parentField != nil && cell.Weight == 0 {
			cell.Weight = parentField.DefaultWeight
		}
		if useMinWidth {
			cell.MinSize.W = f.MinWidth
		}
		cell.View = view
		cell.MinSize.W = f.MinWidth //- v.ColumnMargin*2
		cell.MaxSize.W = f.MaxWidth
		if exp != zgeo.AlignmentHorExpand && (j == 0 || j == len(rootItems.Children)-1) {
			cell.MinSize.W -= inset
		}
		//		fmt.Println("Add Field Item:", cell.View.GetObjectName(), cell.Alignment, f.MinWidth, cell.MinSize.W)
		stack.AddCell(cell, -1)
	}
}

func (f *field) makeFromReflectItem(item zreflect.Item, index int) bool {
	f.Index = index
	f.ID = ustr.FirstToLower(item.FieldName)
	f.Name = item.FieldName
	f.Kind = item.Kind
	f.Alignment = zgeo.AlignmentNone

	for _, part := range zreflect.GetTagAsMap(item.Tag)["zui"] {
		if part == "-" {
			return false
		}
		var key, val string
		if !ustr.SplitN(part, ":", &key, &val) {
			key = part
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		align := zgeo.AlignmentFromString(val)
		n, floatErr := strconv.ParseFloat(val, 32)
		flag := ustr.StrToBool(val, false)
		switch key {
		case "align":
			if align != zgeo.AlignmentNone {
				f.Alignment = align
			}
			// fmt.Println("ALIGN:", f.Name, val, a)
		case "justify":
			if align != zgeo.AlignmentNone {
				f.Justify = align
			}
		case "title":
			f.Title = val
		case "color":
			f.Color = val
		case "height":
			if floatErr == nil {
				f.Height = n
			}
		case "weight":
			if floatErr == nil {
				f.Weight = n
			}
		case "weights":
			if floatErr == nil {
				f.DefaultWeight = n
			}
		case "minwidth":
			if floatErr == nil {
				f.MinWidth = n
			}
		case "static":
			if flag || val == "" {
				f.Flags |= fieldIsStatic
			}
		case "secs":
			f.Flags |= fieldHasSeconds
		case "mins":
			f.Flags |= fieldHasMinutes
		case "hours":
			f.Flags |= fieldHasHours
		case "maxwidth":
			if floatErr == nil {
				f.MaxWidth = n
			}
		case "image":
			var ssize string
			if !ustr.SplitN(val, "/", &ssize, &f.FixedPath) {
				ssize = val
			}
			f.Flags |= fieldIsImage
			f.Size.FromString(ssize)
		case "enum":
			if ustr.HasPrefix(val, ".", &f.LocalEnum) {
			} else {
				if fieldEnums[val] == nil {
					zlog.Error(nil, "no such enum:", val)
				}
				f.Enum, _ = fieldEnums[val]
			}
		case "noheader":
			f.Flags |= fieldsNoHeader
		}
	}
	if f.Title == "" {
		str := ustr.PadCamelCase(f.Name, " ")
		str = ustr.FirstToTitleCase(str)
		f.Title = str

	}

	switch item.Kind {
	case zreflect.KindFloat:
		if f.MinWidth == 0 {
			f.MinWidth = 64
		}
		if f.MaxWidth == 0 {
			f.MaxWidth = 64
		}
	case zreflect.KindInt:
		if item.TypeName != "BoolInd" {
			if item.Package == "time" && item.TypeName == "Duration" {
				if f.Flags&fieldTimeFlags == 0 {
					f.Flags |= fieldTimeFlags
				}
			}
			if f.MaxWidth == 0 {
				f.MinWidth = 64
			}
			if f.MinWidth == 0 {
				f.MaxWidth = 64
			}
			break
		}
		fallthrough

	case zreflect.KindBool:
		if f.MinWidth == 0 {
			f.MinWidth = 20
		}
	case zreflect.KindString:
		if f.Flags&fieldIsImage != 0 {
			f.MinWidth = f.Size.W
			f.MaxWidth = f.Size.W
		}
		if f.MinWidth == 0 {
			f.MinWidth = 100
		}
	case zreflect.KindTime:
		if f.MinWidth == 0 {
			f.MinWidth = 80
		}
		if f.Flags&(fieldTimeFlags|fieldDateFlags) == 0 {
			f.Flags |= fieldTimeFlags | fieldDateFlags
		}
		dig2 := 20.0
		if f.MinWidth == 0 {
			if f.Flags&fieldHasSeconds != 0 {
				f.MaxWidth += dig2

			}
			if f.Flags&fieldHasMinutes != 0 {
				f.MaxWidth += dig2

			}
			if f.Flags&fieldHasHours != 0 {
				f.MaxWidth += dig2

			}
			if f.Flags&fieldHasDays != 0 {
				f.MaxWidth += dig2

			}
			if f.Flags&fieldHasMonths != 0 {
				f.MaxWidth += dig2

			}
			if f.Flags&fieldHasYears != 0 {
				f.MaxWidth += dig2 * 2
			}
		}
		//		fmt.Println("Time max:", f.MaxWidth)

	case zreflect.KindFunc:
		if f.MinWidth == 0 {
			if f.Flags&fieldIsImage != 0 {
				f.MinWidth = f.Size.W * ScreenMain().Scale
			}
		}
	}
	return true
}

func fieldViewNew(name string, vertical bool, structure interface{}, spacing float64, marg zgeo.Size) *FieldView {
	// fmt.Println("FieldViewNew", name)
	v := &FieldView{}
	v.StackView.init(v, name)
	v.Spacing(12)
	v.Margin(zgeo.RectFromMinMax(marg.Pos(), marg.Pos().Negative()))
	v.Vertical = vertical
	v.structure = structure
	unnestAnon := false
	recursive := false
	froot, err := zreflect.ItterateStruct(structure, unnestAnon, recursive)
	if err != nil {
		panic(err)
	}
	for i, item := range froot.Children {
		var f field
		if f.makeFromReflectItem(item, i) {
			v.fields = append(v.fields, f)
		}
	}
	return v
}

func (v *FieldView) Build(handleUpdate func(i int)) {
	fieldsBuildStack(v, &v.StackView, v.structure, v.parentField, &v.fields, zgeo.AlignmentLeft|zgeo.AlignmentTop, zgeo.Size{}, true, 5, 0, handleUpdate) // Size{6, 4}
}

func FieldViewNew(name string, structure interface{}) *FieldView {
	v := fieldViewNew(name, true, structure, 12, zgeo.Size{10, 10})
	return v
}

func (v *FieldView) CopyBack(showError bool) error {
	return FieldsCopyBack(v.structure, v.fields, v, showError)
}

func FieldsCopyBack(structure interface{}, fields []field, ct ContainerType, showError bool) error {
	var err error
	unnestAnon := true
	recursive := true
	rootItems, err := zreflect.ItterateStruct(structure, unnestAnon, recursive)
	if err != nil {
		panic(err)
	}
	children := ct.GetChildren()
	j := 0
	for i, item := range rootItems.Children {
		f := findFieldWithIndex(&fields, i)
		if f == nil {
			continue
		}

		view := children[j]
		j++
		if f.Flags&fieldIsStatic != 0 {
			continue
		}
		if f.Enum != nil {
			mv, _ := view.(*MenuView)
			if mv != nil {
				_, val := mv.NameAndValue()
				a := reflect.ValueOf(item.Address).Elem()
				rval := reflect.ValueOf(val)
				a.Set(rval)
			}
			continue
		}

		switch f.Kind {
		case zreflect.KindStruct:
			fv, _ := view.(*FieldView)
			if fv != nil {
				fv.CopyBack(showError)
			}

		case zreflect.KindBool:
			bv, _ := view.(*CheckBox)
			if bv == nil {
				panic("Should be switch")
			}
			b, _ := item.Address.(*bool)
			if b != nil {
				*b = bv.GetValue().Value()
			}
			bi, _ := item.Address.(*BoolInd)
			if bi != nil {
				*bi = bv.GetValue()
			}

		case zreflect.KindInt:
			if f.Flags&fieldIsStatic == 0 {
				if item.TypeName == "BoolInd" {
					bv, _ := view.(*CheckBox)
					*item.Address.(*bool) = bv.GetValue().Value()
				} else {
					tv, _ := view.(*TextView)
					str := tv.GetText()
					if item.Package == "time" && item.TypeName == "Duration" {
						var secs float64
						secs, err = ztime.GetSecsFromHMSString(str, f.Flags&fieldHasHours != 0, f.Flags&fieldHasMinutes != 0, f.Flags&fieldHasSeconds != 0)
						if err != nil {
							break
						}
						d := item.Address.(*time.Duration)
						if d != nil {
							*d = ztime.SecondsDur(secs)
						}
						return nil
					}
					var i64 int64
					i64, err = strconv.ParseInt(str, 10, 64)
					if err != nil {
						break
					}
					zint.SetAny(item.Address, i64)
				}
			}

		case zreflect.KindFloat:
			if f.Flags&fieldIsStatic == 0 {
				tv, _ := view.(*TextView)
				var f64 float64
				f64, err = strconv.ParseFloat(tv.GetText(), 64)
				if err != nil {
					break
				}
				zfloat.SetAny(item.Address, f64)
			}

		case zreflect.KindTime:
			break

		case zreflect.KindString:
			if f.Flags&fieldIsStatic == 0 && f.Flags&fieldIsImage == 0 {
				tv, _ := view.(*TextView)
				text := tv.GetText()
				str := item.Address.(*string)
				*str = text
			}

		case zreflect.KindFunc:
			break

		default:
			panic(fmt.Sprint("bad type: ", f.Kind))
		}
	}
	if showError && err != nil {
		AlertShowError("", err)
	}

	return nil
}

var fieldEnums = map[string]zdict.Dict{}

func FieldsAddEnum(name string, nameVals zdict.Dict) {
	fieldEnums[name] = nameVals
}

func FieldsAddStringBasedEnum(name string, vals ...interface{}) {
	m := zdict.Dict{}
	for _, v := range vals {
		n := fmt.Sprintf("%v", v)
		m[n] = v
	}
	fieldEnums[name] = m
}
