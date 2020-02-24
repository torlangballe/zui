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
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

type fieldType int

type FieldActionType string

const (
	FieldDataChangedAction FieldActionType = "changed"
	FieldEditedAction      FieldActionType = "edited"
	FieldSetupAction       FieldActionType = "setup"
	FieldPressedAction     FieldActionType = "pressed"
	FieldCreateAction      FieldActionType = "create"
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
	fieldToClipboard
	fieldIsMenuedGroup
	fieldIsTabGroup
)
const (
	fieldTimeFlags = fieldHasSeconds | fieldHasMinutes | fieldHasHours
	fieldDateFlags = fieldHasDays | fieldHasMonths | fieldHasYears
)

//type FieldsActionFunc func(i int, rowPtr interface{})

// type FieldsActionType interface {
// 	Action(i int, rowPtr interface{})
// }

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
	UpdateSecs    float64
	LabelizeWidth float64
	// Type          fieldType
}

type FieldActionHandler interface {
	HandleAction(id string, i int, f *Field, action FieldActionType, view *View) bool
}

type FieldOwner struct {
	fields        []Field
	structure     interface{} // structure of ALL, not just a row
	handleUpdate  func(edited bool, i int)
	labelizeWidth float64
}

func callFieldHandlerFunc(structure interface{}, fid string, i int, f *Field, action FieldActionType, view *View) bool {
	fh, _ := structure.(FieldActionHandler)
	fmt.Println("callFieldHandler1", fh)
	fmt.Printf("callFieldHandler: %+v\n", structure)
	if fh != nil {
		return fh.HandleAction(fid, i, f, action, view)
	}
	return false
}

func (fo FieldOwner) FieldFindWithID(id string) *Field {
	for i, f := range fo.fields {
		if f.ID == id {
			return &fo.fields[i]
		}
	}
	return nil
}

type FieldView struct {
	StackView
	FieldOwner
	parentField *Field
}

/*
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
	// action, _ := item.Interface.(FieldsActionType)
	// if action != nil { //!item.Value.IsNil() {
	// 	go action.Action(i, structData)
	// }
}
*/

func (f Field) IsStatic() bool {
	return (f.Flags&fieldIsStatic != 0)
}

func fieldsMakeButton(structure interface{}, item zreflect.Item, f *Field, i int, height float64) *Button {
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
		view := View(button)
		callFieldHandlerFunc(structure, f.ID, i, f, FieldPressedAction, &view)
	})
	return button
}

func fieldsMakeMenu(structure interface{}, item zreflect.Item, f *Field, i int, items MenuItems) *MenuView {
	menu := MenuViewNew(f.Name+"Menu", items, item.Value, f.IsStatic())
	menu.SetMaxWidth(f.MaxWidth)
	// fmt.Println("fieldsMakeMenu:", f.IsStatic(), i, f.Name, menu.StaticName, MenuItemsLength(items))
	menu.ChangedHandler(func(id, name string, value interface{}) {
		iface := menu.GetCurrentIdOrValue()
		//		zlog.Debug(iface, f.Name)
		item.Value.Set(reflect.ValueOf(iface))
		callFieldHandlerFunc(structure, f.ID, i, f, FieldEditedAction, nil)
	})
	return menu
}

func fieldsMakeTimeView(structure interface{}, item zreflect.Item, f *Field, i int) View {
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

func fieldsMakeText(structure interface{}, item zreflect.Item, f *Field, i int) View {
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
			fmt.Println("label ressed:", f.Name)
			view := View(label)
			callFieldHandlerFunc(structure, f.ID, i, f, FieldPressedAction, &view)
		})
		if f.Flags&fieldToClipboard != 0 {
			label.SetPressedHandler(func() {
				text := label.Text()
				PasteboardSetString(text)
				label.SetText("📋 " + text)
				ztimer.StartIn(0.6, true, func() {
					label.SetText(text)
				})
			})
		}
		return label
	}
	var style TextViewStyle
	tv := TextViewNew(str, style)
	tv.UpdateSecs = f.UpdateSecs
	tv.ChangedHandler(func(view View) {
		// fmt.Println("Changed txt", structure)
		fieldViewToDataItem(item, f, tv, true)
		view = View(tv)
		callFieldHandlerFunc(structure, f.ID, i, f, FieldEditedAction, &view)
	})
	tv.KeyHandler(func(view View, key KeyboardKey, mods KeyboardModifier) {
		fmt.Println("keyup!")
	})
	return tv
}

func fieldsMakeCheckbox(structure interface{}, item zreflect.Item, f *Field, i int, b BoolInd) View {
	cv := CheckBoxNew(b)
	cv.ValueHandler(func(v View) {
		fieldViewToDataItem(item, f, cv, true)
		view := View(cv)
		callFieldHandlerFunc(structure, f.ID, i, f, FieldEditedAction, &view)
	})
	return cv
}

func fieldsMakeImage(structure interface{}, item zreflect.Item, f *Field, i int) View {
	iv := ImageViewNew("", f.Size)
	iv.SetObjectName(f.ID)
	iv.SetPressedHandler(func() {
		view := View(iv)
		callFieldHandlerFunc(structure, f.ID, i, f, FieldPressedAction, &view)
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

func fieldsUpdateStack(fo FieldOwner, stack *StackView, structData interface{}) {
	unnestAnon := true
	recursive := false
	rootItems, err := zreflect.ItterateStruct(structData, unnestAnon, recursive)

	if err != nil {
		panic(err)
	}
	for i, item := range rootItems.Children {
		f := findFieldWithIndex(&fo.fields, i)
		if f == nil {
			continue
		}
		fview := stack.FindViewWithName(f.ID, true)
		if fview == nil {
			continue
		}
		view := *fview
		called := callFieldHandlerFunc(structData, f.ID, 0, f, FieldDataChangedAction, &view)
		// fmt.Println("updateStack:", f.Name, f.Kind, called)
		if called {
			continue
		}
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
			v := cv.Value()
			if v != b {
				cv.SetValue(b)
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
					//  fmt.Println("fields set text:", f.Name, str)
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

func refreshFieldMenuedGroup(name string, menu *MenuView, fieldView *FieldView, item zreflect.Item) {
	menu.Empty()

	mItems, _ := item.Interface.(MenuItems)		

	menu.items = mItems
	menu.AddAction("$title",  name + ":")
	menu.AddAction("add", "Add")

	c := mItems.Count()
	if c > 0 {
		menu.AddAction("remove", "Remove Current")
		menu.SetWithID("$title")
		menu.AddSeparator()
	}
	fieldView.Show(c != 0)

	menu.updateVals(mItems, nil)
}

func makeFieldMenuedGroup(fo FieldOwner, stack *StackView, item zreflect.Item, f *Field, i int, defaultAlign zgeo.Alignment, cellMargin zgeo.Size) View {
	fmt.Println("makeFieldMenuedGroup", f.Name, f.LabelizeWidth)
	vert := StackViewVert("menued-group-vert")
	bar := StackViewHor("bar")
	vert.Add(zgeo.Left | zgeo.Top | zgeo.HorExpand, bar)


	menu := MenuViewNew("menu", nil, nil, false)
	var s interface{}
	if item.Value.Len() == 0 {
		s = reflect.New(item.Value.Type().Elem()).Interface()
		// zlog.Info("new item", item.Value.Kind(), item.Value.Type().Elem(), reflect.ValueOf(s).Type(), reflect.ValueOf(s).Kind())
	} else {
		s = item.Value.Index(0).Addr().Interface()
		// zlog.Info("slice item", s)
	}
	fv := FieldViewNew(item.FieldName, s, f.LabelizeWidth)
	fv.Build()

	menu.ChangedHandler(func(id string, name string, value interface{}) {
		fmt.Println("changed:", id, name, value)
		switch id {
		case "add":
			a := reflect.New(item.Value.Type().Elem()).Elem()
			nv := reflect.Append(item.Value, a) 
			item.Value.Addr().Elem().Set(nv)
			item.Interface = item.Value.Interface()		
			callFieldHandlerFunc(a.Interface(), f.ID, i, f, FieldEditedAction, nil)
		}
		refreshFieldMenuedGroup(item.FieldName, menu, fv, item)
	})

	bar.Add(zgeo.Left | zgeo.VertCenter, menu) //  | zgeo.HorExpand

	vert.Add(zgeo.Left | zgeo.Top | zgeo.Expand, fv)
	refreshFieldMenuedGroup(item.FieldName, menu, fv, item)

	return vert
}

func fieldsBuildStack(fo FieldOwner, stack *StackView, structData interface{}, parentField *Field, fields *[]Field, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool, inset float64, i int) {
	unnestAnon := true
	recursive := false
	rootItems, err := zreflect.ItterateStruct(structData, unnestAnon, recursive)
	//  fmt.Println("fieldsBuildStack", len(rootItems.Children), err, len(*fields), fo.labelizeWidth)
	if err != nil {
		panic(err)
	}
	for j, item := range rootItems.Children {
		exp := zgeo.AlignmentNone
		f := findFieldWithIndex(fields, j)
		if f == nil {
			zlog.Error(nil, "no field for index", j)
			continue
		}
		// fmt.Println("   fieldsBuildStack2", j, f.Name)

		var view View
		if f.Flags & fieldIsMenuedGroup != 0 {
			view = makeFieldMenuedGroup(fo, stack, item, f, i, defaultAlign, cellMargin)
		} else {
			callFieldHandlerFunc(item, f.ID, i, f, FieldCreateAction, &view) // this sees if actual ITEM is a field handler
		}
		if view == nil {
			callFieldHandlerFunc(structData, f.ID, i, f, FieldCreateAction, &view)
		}
		if view != nil {
		} else if f.LocalEnum != "" {
			ei := findLocalEnum(&rootItems.Children, f.LocalEnum)
			if !zlog.ErrorIf(ei == nil, f.Name, f.LocalEnum) {
				enum, _ := ei.Interface.(MenuItems)
				if zlog.ErrorIf(enum == nil) {
					continue
				}
				// fmt.Println("make local enum:", f.Name, f.LocalEnum, i, MenuItemsLength(enum))
				menu := fieldsMakeMenu(structData, item, f, i, enum)
				if menu == nil {
					zlog.Error(nil, "no local enum for", f.LocalEnum)
					continue
				}
				view = menu
				menu.SetWithIdOrValue(item.Interface)
			}
		} else if f.Enum != nil {
			// fmt.Printf("make enum: %s %v %v\n", f.Name, f.Enum, item)
			view = fieldsMakeMenu(structData, item, f, i, f.Enum)
			exp = zgeo.AlignmentNone
		} else {
			switch f.Kind {
			case zreflect.KindStruct:
				// fmt.Println("struct:", f.Kind, j, item.Value)
				childStruct := item.Address
				fieldView := fieldViewNew(f.Name, false, childStruct, 10, zgeo.Size{}, f.LabelizeWidth)
				fieldView.parentField = f
				view = fieldView
				fieldsBuildStack(fo, &fieldView.StackView, fieldView.structure, fieldView.parentField, &fieldView.fields, zgeo.Left|zgeo.Top, zgeo.Size{}, true, 5, 0)

			case zreflect.KindBool:
				b := BoolIndFromBool(item.Value.Interface().(bool))
				view = fieldsMakeCheckbox(structData, item, f, i, b)

			case zreflect.KindInt:
				if item.TypeName == "BoolInd" {
					exp = zgeo.HorShrink
					view = fieldsMakeCheckbox(structData, item, f, i, BoolInd(item.Value.Int()))
				} else {
					view = fieldsMakeText(structData, item, f, i)
				}

			case zreflect.KindFloat:
				view = fieldsMakeText(structData, item, f, i)

			case zreflect.KindString:
				if f.Flags&fieldIsImage != 0 {
					view = fieldsMakeImage(structData, item, f, i)
				} else {
					exp = zgeo.HorExpand
					view = fieldsMakeText(structData, item, f, i)
				}

			case zreflect.KindFunc:
				if f.Flags&fieldIsImage != 0 {
					view = fieldsMakeImage(structData, item, f, i)
				} else {
					view = fieldsMakeButton(structData, item, f, i, f.Height)
				}

			case zreflect.KindSlice:
				items, got := item.Interface.(MenuItems)
				if got {
					menu := fieldsMakeMenu(structData, item, f, i, items)
					view = menu
					break
				}
				view = fieldsMakeText(structData, item, f, i)
				break

			case zreflect.KindTime:
				view = fieldsMakeTimeView(structData, item, f, i)

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

		cell := &ContainerViewCell{}
		if fo.labelizeWidth != 0 {
			_, cell = Labelize(stack, view, f.Title, fo.labelizeWidth)			
		}
		cell.Margin = cellMargin
		def := defaultAlign
		all := zgeo.Left | zgeo.HorCenter | zgeo.Right
		if f.Alignment&all != 0 {
			def &= ^all
		}
		cell.Alignment = def | exp | f.Alignment
		//  fmt.Println("field align:", f.Alignment, f.Name, def|exp, cell.Alignment, int(cell.Alignment))
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
		// fmt.Println("Add Field Item:", cell.View.ObjectName(), cell.Alignment, f.MinWidth, cell.MinSize.W, cell.MaxSize)
		if fo.labelizeWidth == 0 {
			stack.AddCell(*cell, -1)
		}
	}
}

func (f *Field) makeFromReflectItem(fo FieldOwner, structure interface{}, item zreflect.Item, index int) bool {
	f.Index = index
	f.ID = zstr.FirstToLowerWithAcronyms(item.FieldName)
	f.Name = item.FieldName
	f.Kind = item.Kind
	f.Alignment = zgeo.AlignmentNone
	f.UpdateSecs = 4

	// fmt.Println("Field:", f.ID)
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
		case "immediate":
			f.UpdateSecs = 0
		case "upsecs":
			if floatErr == nil && n > 0 {
				f.UpdateSecs = n
			}
		case "2clip":
			f.Flags |= fieldToClipboard
		case "menued-group":
			mi, _ := item.Interface.(MenuItems)		
			zlog.Assert(mi != nil)
			f.Flags |= fieldIsMenuedGroup
		case "labelize":
			f.LabelizeWidth = n
			if n == 0 {
				f.LabelizeWidth = 200
			}
		}
	}
	if f.Flags&fieldToClipboard != 0 && f.Tooltip == "" {
		f.Tooltip = "press to copy to pasteboard"
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
	callFieldHandlerFunc(structure, f.ID, -1, f, FieldSetupAction, nil)
	return true
}

func fieldViewNew(name string, vertical bool, structure interface{}, spacing float64, marg zgeo.Size, labelizeWidth float64) *FieldView {
	v := &FieldView{}
	v.StackView.init(v, name)
	v.SetSpacing(12)
	v.SetMargin(zgeo.RectFromMinMax(marg.Pos(), marg.Pos().Negative()))
	v.Vertical = vertical
	v.structure = structure
	v.FieldOwner.labelizeWidth = labelizeWidth
	unnestAnon := true
	recursive := false
	froot, err := zreflect.ItterateStruct(structure, unnestAnon, recursive)
	if err != nil {
		panic(err)
	}
	fmt.Println("FieldViewNew", name, len(froot.Children), labelizeWidth)
	for i, item := range froot.Children {
		var f Field
		if f.makeFromReflectItem(v.FieldOwner, structure, item, i) {
			v.fields = append(v.fields, f)
		}
	}
	return v
}

func (v *FieldView) Build() {
	fieldsBuildStack(v.FieldOwner, &v.StackView, v.structure, v.parentField, &v.fields, zgeo.Left|zgeo.Top, zgeo.Size{}, true, 5, 0) // Size{6, 4}
}

func FieldViewNew(name string, structure interface{}, labelizeWidth float64) *FieldView {
	v := fieldViewNew(name, true, structure, 12, zgeo.Size{10, 10}, labelizeWidth)
	return v
}

func fieldViewToDataItem(item zreflect.Item, f *Field, view View, showError bool) error {
	var err error

	if f.Flags&fieldIsStatic != 0 {
		return nil
	}
	if (f.Enum != nil || f.LocalEnum != "") && !f.IsStatic() {
		mv, _ := view.(*MenuView)
		if mv != nil {
			iface := mv.GetCurrentIdOrValue()
			zlog.Debug(iface, f.Name)
			item.Value.Set(reflect.ValueOf(iface))
		}
		return nil
	}

	switch f.Kind {
	case zreflect.KindBool:
		bv, _ := view.(*CheckBox)
		if bv == nil {
			panic("Should be switch")
		}
		b, _ := item.Address.(*bool)
		if b != nil {
			*b = bv.Value().Value()
		}
		bi, _ := item.Address.(*BoolInd)
		if bi != nil {
			*bi = bv.Value()
		}

	case zreflect.KindInt:
		if f.Flags&fieldIsStatic == 0 {
			if item.TypeName == "BoolInd" {
				bv, _ := view.(*CheckBox)
				*item.Address.(*bool) = bv.Value().Value()
			} else {
				tv, _ := view.(*TextView)
				str := tv.Text()
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
			f64, err = strconv.ParseFloat(tv.Text(), 64)
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
			text := tv.Text()
			str := item.Address.(*string)
			*str = text
		}

	case zreflect.KindFunc:
		break

	default:
		panic(fmt.Sprint("bad type: ", f.Kind))
	}

	if showError && err != nil {
		AlertShowError("", err)
	}

	return err
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
