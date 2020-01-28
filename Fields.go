package zui

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

	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zstr"
)

type fieldType int

type FieldActionType int

const (
	FieldUpdateAction FieldActionType = iota
	FieldPressedAction
)

const (
	fieldIsStatic = 1 << iota
	fieldHasSeconds
	fieldHasMinutes
	fieldHasHours
	fieldHasDays
	fieldHasMonths
	fieldHasYears
	fieldIsImage
	fieldHasHeaderImage
	fieldsNoHeader
	fieldsLive
)
const (
	fieldTimeFlags = fieldHasSeconds | fieldHasMinutes | fieldHasHours
	fieldDateFlags = fieldHasDays | fieldHasMonths | fieldHasYears
)

type FieldsActionFunc func(i int, rowPtr interface{})

type FieldsActionType interface {
	Action(i int, rowPtr interface{})
}

type Field struct {
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
	Enum          MenuItems
	LocalEnum     string
	Size          zgeo.Size
	Flags         int
	DefaultWeight float64
	Tooltip       string
	// Type          fieldType
}

type FieldView struct {
	StackView
	parentField *Field
	fields      []Field
	structure   interface{}
}

func callFieldFunc(i int, structData interface{}, f *Field) {
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

func (f Field) IsStatic() bool {
	return (f.Flags&fieldIsStatic != 0)
}

func fieldsMakeButton(i int, structData interface{}, height float64, f *Field, item zreflect.Item) *Button {
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
	button.SetPressedHandler(func() {
		callFieldFunc(i, structData, f)
	})
	return button
}

func fieldsMakeMenu(item zreflect.Item, i int, items MenuItems, f *Field, handleAction func(i int, a FieldActionType, id string)) *MenuView {
	sn := ""
	if f.IsStatic() {
		sn = f.Name
	}
	menu := MenuViewNew(f.Name+"Menu", items, item.Value, sn)
	menu.SetMaxWidth(f.MaxWidth)
	// fmt.Println("fieldsMakeMenu:", f.IsStatic(), i, f.Name, menu.StaticName, MenuItemsLength(items))
	menu.ChangedHandler(func(id, name string, value interface{}) {
		if handleAction != nil {
			handleAction(i, FieldUpdateAction, f.ID)
		}
	})
	return menu
}

func fieldsMakeTimeView(f *Field, item zreflect.Item, i int, handleAction func(i int, a FieldActionType, id string)) View {
	t := item.Interface.(time.Time)
	format := f.Format
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	var style TextViewStyle
	str := t.Format(format)
	tv := TextViewNew(str, style)
	return tv
}

func fieldsMakeText(f *Field, item zreflect.Item, i int, handleAction func(i int, a FieldActionType, id string)) View {
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
			j = f.Alignment & (zgeo.Left | zgeo.HorCenter | zgeo.Right)
			if j == zgeo.AlignmentNone {
				j = zgeo.Left
			}
		}
		label.SetTextAlignment(j)
		label.SetPressedHandler(func() {
			handleAction(i, FieldPressedAction, f.ID)
		})
		return label
	}
	var style TextViewStyle
	tv := TextViewNew(str, style)
	if f.Flags&fieldsLive != 0 {
		tv.ContinuousUpdateCalls = true
	}
	if handleAction != nil {
		tv.ChangedHandler(func(view View) {
			handleAction(i, FieldUpdateAction, f.ID)
		})
	}
	tv.KeyHandler(func(view View, key KeyboardKey, mods KeyboardModifier) {
		fmt.Println("keyup!")
	})
	return tv
}

func fieldsMakeCheckbox(b BoolInd, i int, handleAction func(i int, a FieldActionType, id string)) View {
	c := CheckBoxNew(b)
	if handleAction != nil {
		c.ValueHandler(func(v View) {
			handleAction(i, FieldUpdateAction, v.GetObjectName())
		})
	}
	return c
}

func fieldsMakeImage(structData interface{}, f *Field, i int) View {
	iv := ImageViewNew("", f.Size)
	iv.SetObjectName(f.ID)
	iv.SetPressedHandler(func() {
		callFieldFunc(i, structData, f)
	})
	return iv
}

func findFieldWithIndex(fields *[]Field, index int) *Field {
	for i, f := range *fields {
		if f.Index == index {
			return &(*fields)[i]
		}
	}
	return nil
}

func fieldsUpdateStack(stack *StackView, structData interface{}, fields *[]Field) {
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
		// fmt.Println("updateStack:", f.Name, f.LocalEnum, f.Enum, f.Kind)
		view := *fview
		if f.Enum != nil && f.Kind != zreflect.KindSlice || f.LocalEnum != "" {
			menu := view.(*MenuView)
			// fmt.Println("FUS:", f.Name, item.Interface, item.Kind, reflect.ValueOf(item.Interface).Type())
			menu.SetWithIdOrValue(item.Interface)
			continue
		}
		switch f.Kind {
		case zreflect.KindSlice:
			items, got := item.Interface.(MenuItems)
			if got {
				menu := view.(*MenuView)
				menu.UpdateValues(items)
			} else if f.IsStatic() {
				mItems := item.Interface.(MenuItems)
				menu := view.(*MenuView)
				menu.UpdateValues(mItems)
			}

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
			str := item.Value.String()
			if f.Flags&fieldIsImage != 0 {
				path := ""
				if f.Kind == zreflect.KindString {
					path = str
				}
				if path == "" {
					path = f.FixedPath
				}
				iv := view.(*ImageView)
				iv.SetImage(nil, path, nil)
			} else {
				if f.IsStatic() {
					label, _ := view.(*Label)
					zlog.Assert(label != nil)
					label.SetText(str)
				} else {
					tv, _ := view.(*TextView)
					// fmt.Println("fields set text:", f.Name, str)
					if tv != nil {
						tv.SetText(str)
					}
				}
			}
		}
	}
}

func findLocalEnum(children *[]zreflect.Item, name string) *zreflect.Item {
	name = zstr.HeadUntil(name, ".")
	for i, c := range *children {
		if c.FieldName == name {
			return &(*children)[i]
		}
	}
	return nil
}

func fieldsBuildStack(fv *FieldView, stack *StackView, structData interface{}, parentField *Field, fields *[]Field, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool, inset float64, i int, handleAction func(i int, a FieldActionType, id string)) {
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
			ei := findLocalEnum(&rootItems.Children, f.LocalEnum)
			if !zlog.ErrorIf(ei == nil, f.Name, f.LocalEnum) {
				enum, _ := ei.Interface.(MenuItems)
				if zlog.ErrorIf(enum == nil) {
					continue
				}
				// fmt.Println("make local enum:", f.Name, f.LocalEnum, i, MenuItemsLength(enum))
				menu := fieldsMakeMenu(item, i, enum, f, handleAction)
				if menu == nil {
					zlog.Error(nil, "no local enum for", f.LocalEnum)
					continue
				}
				view = menu
				menu.SetWithIdOrValue(item.Interface)
			}
		} else if f.Enum != nil {
			// fmt.Printf("make enum: %s %v %v\n", f.Name, f.Enum, item)
			view = fieldsMakeMenu(item, i, f.Enum, f, handleAction)
			exp = zgeo.AlignmentNone
		} else {
			switch f.Kind {
			case zreflect.KindStruct:
				// fmt.Println("struct:", f.Kind, j, item.Value)
				childStruct := item.Address
				fieldView := fieldViewNew(f.Name, false, childStruct, 10, zgeo.Size{})
				fieldView.parentField = f
				view = fieldView
				fieldsBuildStack(fieldView, &fieldView.StackView, fieldView.structure, fieldView.parentField, &fieldView.fields, zgeo.Left|zgeo.Top, zgeo.Size{}, true, 5, 0, handleAction)

			case zreflect.KindBool:
				b := BoolIndFromBool(item.Value.Interface().(bool))
				view = fieldsMakeCheckbox(b, i, handleAction)

			case zreflect.KindInt:
				if item.TypeName == "BoolInd" {
					exp = zgeo.HorShrink
					view = fieldsMakeCheckbox(BoolInd(item.Value.Int()), i, handleAction)
				} else {
					view = fieldsMakeText(f, item, i, handleAction)
				}

			case zreflect.KindFloat:
				view = fieldsMakeText(f, item, i, handleAction)

			case zreflect.KindString:
				if f.Flags&fieldIsImage != 0 {
					view = fieldsMakeImage(structData, f, i)
				} else {
					exp = zgeo.HorExpand
					view = fieldsMakeText(f, item, i, handleAction)
				}

			case zreflect.KindFunc:
				if f.Flags&fieldIsImage != 0 {
					view = fieldsMakeImage(structData, f, i)
				} else {
					view = fieldsMakeButton(i, structData, f.Height, f, item)
				}

			case zreflect.KindSlice:
				items, got := item.Interface.(MenuItems)
				if got {
					menu := fieldsMakeMenu(item, i, items, f, handleAction)
					view = menu
					break
				}
				view = fieldsMakeText(f, item, i, handleAction)
				break

			case zreflect.KindTime:
				view = fieldsMakeTimeView(f, item, i, handleAction)

			default:
				panic(fmt.Sprintln("fieldsBuildStack bad type:", f.Name, f.Kind))
			}
		}
		var tipField, tip string
		if zstr.HasPrefix(f.Tooltip, ".", &tipField) {
			for _, ei := range rootItems.Children {
				if ei.FieldName == tipField {
					tip = fmt.Sprint(ei.Interface)
					break
				}
			}
		} else if f.Tooltip != "" {
			tip = f.Tooltip
		}
		if tip != "" {
			ViewGetNative(view).SetToolTip(tip)
		}
		view.SetObjectName(f.ID)
		cell := ContainerViewCell{}
		cell.Margin = cellMargin
		def := defaultAlign
		all := zgeo.Left | zgeo.HorCenter | zgeo.Right
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
			// fmt.Println("Cell Width:", f.Name, cell.MinSize.W)
		}
		cell.View = view
		cell.MaxSize.W = f.MaxWidth
		if cell.MinSize.W != 0 && (j == 0 || j == len(rootItems.Children)-1) {
			cell.MinSize.W += inset
		}
		//		fmt.Println("Add Field Item:", cell.View.GetObjectName(), cell.Alignment, f.MinWidth, cell.MinSize.W)
		stack.AddCell(cell, -1)
	}
}

func (f *Field) makeFromReflectItem(item zreflect.Item, index int) bool {
	f.Index = index
	f.ID = zstr.FirstToLower(item.FieldName)
	f.Name = item.FieldName
	f.Kind = item.Kind
	f.Alignment = zgeo.AlignmentNone

	for _, part := range zreflect.GetTagAsMap(item.Tag)["zui"] {
		if part == "-" {
			return false
		}
		var key, val string
		if !zstr.SplitN(part, ":", &key, &val) {
			key = part
		}
		key = strings.TrimSpace(key)
		origVal := val
		val = strings.TrimSpace(val)
		align := zgeo.AlignmentFromString(val)
		n, floatErr := strconv.ParseFloat(val, 32)
		flag := zstr.StrToBool(val, false)
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
			f.Title = origVal
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
		case "width":
			if floatErr == nil {
				f.MinWidth = n
				f.MaxWidth = n
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
		case "image", "himage":
			var ssize string
			if !zstr.SplitN(val, "|", &ssize, &f.FixedPath) {
				ssize = val
			}
			if key == "image" {
				f.Flags |= fieldIsImage
			} else {
				f.Flags |= fieldHasHeaderImage
			}
			f.Size.FromString(ssize)
		case "enum":
			if zstr.HasPrefix(val, ".", &f.LocalEnum) {
			} else {
				if fieldEnums[val] == nil {
					zlog.Error(nil, "no such enum:", val)
				}
				f.Enum, _ = fieldEnums[val]
			}
		case "noheader":
			f.Flags |= fieldsNoHeader
		case "tip":
			f.Tooltip = val
		case "live":
			f.Flags |= fieldsLive
		}
	}
	zfloat.Maximize(&f.MaxWidth, f.MinWidth)
	zfloat.Minimize(&f.MinWidth, f.MaxWidth)
	if f.Title == "" {
		str := zstr.PadCamelCase(f.Name, " ")
		str = zstr.FirstToTitleCase(str)
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
			if f.MinWidth == 0 {
				f.MinWidth = 60
			}
			if f.MaxWidth == 0 {
				f.MaxWidth = 80
			}
			break
		}
		fallthrough

	case zreflect.KindBool:
		if f.MinWidth == 0 {
			f.MinWidth = 20
		}
	case zreflect.KindString:
		if f.Flags&(fieldHasHeaderImage|fieldIsImage) != 0 {
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
				min := f.Size.W // * ScreenMain().Scale
				//				min += ImageViewDefaultMargin.W * 2
				zfloat.Maximize(&f.MinWidth, min)
			}
		}
	}
	return true
}

func fieldViewNew(name string, vertical bool, structure interface{}, spacing float64, marg zgeo.Size) *FieldView {
	// fmt.Println("FieldViewNew", name)
	v := &FieldView{}
	v.StackView.init(v, name)
	v.SetSpacing(12)
	v.SetMargin(zgeo.RectFromMinMax(marg.Pos(), marg.Pos().Negative()))
	v.Vertical = vertical
	v.structure = structure
	unnestAnon := false
	recursive := false
	froot, err := zreflect.ItterateStruct(structure, unnestAnon, recursive)
	if err != nil {
		panic(err)
	}
	for i, item := range froot.Children {
		var f Field
		if f.makeFromReflectItem(item, i) {
			v.fields = append(v.fields, f)
		}
	}
	return v
}

func (v *FieldView) Build(handleAction func(i int, a FieldActionType, id string)) {
	fieldsBuildStack(v, &v.StackView, v.structure, v.parentField, &v.fields, zgeo.Left|zgeo.Top, zgeo.Size{}, true, 5, 0, handleAction) // Size{6, 4}
}

func FieldViewNew(name string, structure interface{}) *FieldView {
	v := fieldViewNew(name, true, structure, 12, zgeo.Size{10, 10})
	return v
}

func (v *FieldView) CopyBack(showError bool) error {
	return FieldsCopyBack(v.structure, v.fields, v, showError)
}

func FieldsCopyBack(structure interface{}, fields []Field, ct ContainerType, showError bool) error {
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
		if (f.Enum != nil || f.LocalEnum != "") && !f.IsStatic() {
			mv, _ := view.(*MenuView)
			if mv != nil {
				iface := mv.GetCurrentIdOrValue()
				zlog.Debug(iface, f.Name)
				item.Value.Set(reflect.ValueOf(iface))
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
				if tv == nil {
					zlog.Fatal(nil, "Copy Back string not TV:", f.Name)
				}
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

var fieldEnums = map[string]MenuItems{}

func FieldsSetEnum(name string, enum MenuItems) {
	fieldEnums[name] = enum
}

func FieldsSetEnumItems(name string, nameValPairs ...interface{}) {
	var dis zdict.Items

	for i := 0; i < len(nameValPairs); i += 2 {
		var di zdict.Item
		di.Name = nameValPairs[i].(string)
		di.Value = nameValPairs[i+1]
		dis = append(dis, di)
	}
	fieldEnums[name] = dis
}

func FieldsAddStringBasedEnum(name string, vals ...interface{}) {
	var items zdict.Items
	for _, v := range vals {
		n := fmt.Sprintf("%v", v)
		i := zdict.Item{n, v}
		items = append(items, i)
	}
	fieldEnums[name] = items
}
