//go:build zui
// +build zui

package zfields

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
)

// type fieldType int

type ActionType string

// UIStringer defines a special string return function used to show a complex type as a string in fields/tables etc, instead of the complex type. String() would kick in too often
type UIStringer interface {
	ZUIString() string
}

type Widgeter interface {
	Create(f *Field) zui.View
	SetValue(view zui.View, val interface{})
}

type ReadWidgeter interface {
	GetValue(view zui.View) interface{}
}

type SetupWidgeter interface {
	SetupField(f *Field)
}

const (
	DataChangedActionPre  ActionType = "changed-pre" // called on struct before DataChangedAction on fields
	DataChangedAction     ActionType = "changed"     // called when value changed, typically programatically or edited. Called on fields with id, then on struct
	EditedAction          ActionType = "edited"      // called when value edited by user, DataChangedAction will also be called
	SetupFieldAction      ActionType = "setup"       // called when a field is being set up from a struct, view will be nil
	PressedAction         ActionType = "pressed"     // called when view is pressed, view is valid
	LongPressedAction     ActionType = "longpressed" // called when view is long-pressed, view is valid
	NewStructAction       ActionType = "newstruct"   // called when new stucture is created, for initializing. View may  be nil
	CreateFieldViewAction ActionType = "createview"  // called to create view, view is pointer to view and is returned in it
	CreatedViewAction     ActionType = "createdview" // called after view created, view is pointer to newly created view.
)

const (
	flagIsStatic = 1 << iota
	flagHasSeconds
	flagHasMinutes
	flagHasHours
	flagHasDays
	flagHasMonths
	flagHasYears
	flagIsImage
	flagIsFixed
	flagIsButton
	flagHasHeaderImage
	flagNoTitle
	flagToClipboard
	flagIsNamedSelection
	flagIsStringer
	flagIsPassword
	flagExpandFromMinSize
	flagIsDuration
	flagIsOpaque
	flagIsActions
)

const (
	flagTimeFlags = flagHasSeconds | flagHasMinutes | flagHasHours
	flagDateFlags = flagHasDays | flagHasMonths | flagHasYears
)

type Field struct {
	Index                int
	ID                   string
	ActionValue          interface{} // ActionValue is used to send other information with an action into ActionHandler / ActionFieldHandler
	Name                 string
	FieldName            string
	Title                string // name of item in row, and header if no title
	MaxWidth             float64
	MinWidth             float64
	Kind                 zreflect.TypeKind
	Vertical             zbool.BoolInd
	Alignment            zgeo.Alignment
	Justify              zgeo.Alignment
	Format               string
	Colors               []string
	ImageFixedPath       string
	HeaderImageFixedPath string
	Height               float64
	Enum                 string
	LocalEnum            string
	Size                 zgeo.Size
	HeaderSize           zgeo.Size
	Flags                int
	Tooltip              string
	UpdateSecs           float64
	LabelizeWidth        float64
	LocalEnable          string
	LocalDisable         string
	LocalShow            string
	LocalHide            string
	FontSize             float64
	FontName             string
	FontStyle            zgeo.FontStyle
	Spacing              float64
	Placeholder          string
	Columns              int
	Rows                 int
	Shadow               zgeo.DropShadow
	SortSmallFirst       zbool.BoolInd
	SortPriority         int
	IsGroup              bool
	FractionDecimals     int
	OldSecs              int
	ValueStoreKey        string
	Visible              bool
	Disabled             bool
	SetEdited            bool
	WidgetName           string
}

type ActionHandler interface {
	HandleAction(f *Field, action ActionType, view *zui.View) bool
}

type ActionFieldHandler interface {
	HandleFieldAction(f *Field, action ActionType, view *zui.View) bool
}

var widgeters = map[string]Widgeter{}

func RegisterWigeter(name string, w Widgeter) {
	widgeters[name] = w
}

func (f Field) IsStatic() bool {
	return f.Flags&flagIsStatic != 0
}

func (f *Field) SetFont(view zui.View, from *zgeo.Font) {
	to := view.(zui.TextLayoutOwner)
	size := f.FontSize
	if size == 0 {
		if from != nil {
			size = from.Size
		} else {
			size = zgeo.FontDefaultSize
		}
	}
	style := f.FontStyle
	if from != nil {
		style = from.Style
	}
	var font *zgeo.Font
	if f.FontName != "" {
		font = zgeo.FontNew(f.FontName, size, style)
	} else if from != nil {
		font = new(zgeo.Font)
		*font = *from
	} else {
		font = zgeo.FontNice(size, style)
	}
	to.SetFont(font)
}

func findFieldWithIndex(fields *[]Field, index int) *Field {
	for i, f := range *fields {
		if f.Index == index {
			return &(*fields)[i]
		}
	}
	return nil
}

func findLocalFieldWithID(children *[]zreflect.Item, name string) *zreflect.Item {
	name = zstr.HeadUntil(name, ".")
	for i, c := range *children {
		if fieldNameToID(c.FieldName) == name {
			return &(*children)[i]
		}
	}
	return nil
}

func fieldNameToID(name string) string {
	return zstr.FirstToLowerWithAcronyms(name)
}

func (f *Field) makeFromReflectItem(structure interface{}, item zreflect.Item, index int, immediateEdit bool) bool {
	f.Index = index
	f.ID = fieldNameToID(item.FieldName)
	// zlog.Info("FIELD:", f.ID, item.FieldName)
	f.Kind = item.Kind
	f.FieldName = item.FieldName
	f.Alignment = zgeo.AlignmentNone
	f.UpdateSecs = -1
	f.Rows = 1
	f.SortSmallFirst = zbool.Unknown
	f.SetEdited = true
	f.Vertical = zbool.Unknown

	// zlog.Info("Field:", f.ID)
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
		n, floatErr := strconv.ParseFloat(val, 32)
		flag := zbool.FromString(val, false)
		switch key {
		case "password":
			f.Flags |= flagIsPassword
		case "setedited":
			f.SetEdited = flag
		case "format":
			f.Format = val
		case "vertical":
			f.Vertical = zbool.True
		case "horizontal":
			f.Vertical = zbool.False
		case "align":
			f.Alignment = zgeo.AlignmentFromString(val)
			// zlog.Info("ALIGN:", f.Name, val, a)
		case "nosize":
			f.Flags |= flagExpandFromMinSize
		// case "cannil"
		// f.Flags |= flagAllowNil
		case "justify":
			if val == "" {
				f.Justify = f.Alignment
			} else {
				f.Justify = zgeo.AlignmentFromString(val)
			}
		case "name":
			f.Name = origVal
		case "title":
			f.Title = origVal
		case "color":
			f.Colors = strings.Split(val, "|")
		case "height":
			if floatErr == nil {
				f.Height = n
			}
		case "width":
			if floatErr == nil {
				f.MinWidth = n
				f.MaxWidth = n
			}
		case "cols":
			if floatErr == nil {
				f.Columns = int(n)
			}
		case "rows":
			if floatErr == nil {
				f.Rows = int(n)
			}
		case "widget":
			f.WidgetName = val
		case "ascending":
			f.SortSmallFirst = zbool.True
			f.SortPriority = int(n)
		case "descending":
			f.SortSmallFirst = zbool.False
			f.SortPriority = int(n)
		case "actions":
			f.Flags |= flagIsActions
		case "size":
			f.Size.FromString(val)
			if f.Size.IsNull() {
				f.Size = zgeo.SizeBoth(n)
			}
		case "minwidth":
			if floatErr == nil {
				f.MinWidth = n
			}
		case "spacing":
			if floatErr == nil {
				f.Spacing = n
			}
		case "storekey":
			f.ValueStoreKey = val
		case "static":
			if flag || val == "" {
				f.Flags |= flagIsStatic
			}
		case "fracts":
			f.FractionDecimals = int(n)
		case "secs":
			f.Flags |= flagHasSeconds
		case "oldsecs":
			f.OldSecs = int(n)
		case "mins":
			f.Flags |= flagHasMinutes
		case "hours":
			f.Flags |= flagHasHours
		case "maxwidth":
			if floatErr == nil {
				f.MaxWidth = n
			}
		case "group":
			f.IsGroup = true
		case "fixed":
			f.Flags |= flagIsFixed
		case "opaque":
			f.Flags |= flagIsOpaque
		case "shadow":
			for _, part := range strings.Split(val, "|") {
				got := false
				if f.Shadow.Delta.IsNull() {
					err := f.Shadow.Delta.FromString(part)
					if err == nil {
						got = true
					} else {
						num, err := strconv.ParseFloat(part, 32)
						if err != nil {
							f.Shadow.Delta = zgeo.SizeBoth(num)
							got = true
						}
					}
				}
				if !got && f.Shadow.Blur == 0 {
					num, err := strconv.ParseFloat(part, 32)
					if err != nil {
						f.Shadow.Blur = float32(num)
						got = true
					}
				}
				if !got && !f.Shadow.Color.Valid {
					f.Shadow.Color = zgeo.ColorFromString(part)
				}
			}
			if f.Shadow.Delta.IsNull() {
				f.Shadow.Delta = zgeo.SizeBoth(3)
			}
			if f.Shadow.Blur == 0 {
				f.Shadow.Blur = float32(f.Shadow.Delta.Min())
			}
			if !f.Shadow.Color.Valid {
				f.Shadow.Color = zgeo.ColorBlack
			}

		case "font":
			var sign int
			for _, part := range strings.Split(val, "|") {
				if zstr.HasPrefix(part, "+", &part) {
					sign = 1
				}
				if zstr.HasPrefix(part, "-", &part) {
					sign = -1
				}
				n, _ := strconv.Atoi(part)
				if n != 0 {
					if sign != 0 {
						f.FontSize = float64(n*sign) + zgeo.FontDefaultSize
					} else {
						f.FontSize = float64(n)
					}
				} else {
					if f.FontName == "" && f.FontSize == 0 {
						f.FontName = part
					} else {
						f.FontStyle = zgeo.FontStyleFromStr(part)
					}
				}
			}

		case "image", "himage":
			var ssize, path string
			if !zstr.SplitN(val, "|", &ssize, &path) {
				ssize = val
			} else {
				path = "images/" + path
			}
			if key == "image" {
				f.Flags |= flagIsImage
				f.Size.FromString(ssize)
				f.ImageFixedPath = path
			} else {
				f.Flags |= flagHasHeaderImage
				f.HeaderSize.FromString(ssize)
				f.HeaderImageFixedPath = path
			}
		case "enum":
			if zstr.HasPrefix(val, ".", &f.LocalEnum) {
			} else {
				if fieldEnums[val] == nil {
					zlog.Error(nil, "no such enum:", val, fieldEnums)
				}
				f.Enum = val
			}
		case "notitle":
			f.Flags |= flagNoTitle
		case "tip":
			f.Tooltip = val
		case "immediate":
			f.UpdateSecs = 0
		case "upsecs":
			if floatErr == nil && n > 0 {
				f.UpdateSecs = n
			}
		case "2clip":
			f.Flags |= flagToClipboard
		case "named-selection":
			f.Flags |= flagIsNamedSelection
		case "labelize":
			f.LabelizeWidth = n
			if n == 0 {
				f.LabelizeWidth = 200
			}
		case "button":
			f.Flags |= flagIsButton
		case "enable":
			f.LocalEnable = val
		case "disable":
			if val != "" {
				f.LocalDisable = val
			} else {
				f.Disabled = true // not used yet
			}
		case "show":
			if val == "" {
				f.Visible = true
			} else {
				f.LocalShow = val
			}
		case "hide":
			if val == "" {
				f.Visible = false
			} else {
				f.LocalHide = val
			}
		case "placeholder":
			if val != "" {
				f.Placeholder = val
			} else {
				f.Placeholder = "$HAS$"
			}
		case "since":
			f.Flags |= flagIsStatic | flagIsDuration
		}
	}
	if immediateEdit {
		f.UpdateSecs = -1
	}
	if f.HeaderSize.IsNull() {
		f.HeaderSize = f.Size
	}
	if f.HeaderImageFixedPath == "" && f.Flags&flagHasHeaderImage != 0 {
		f.HeaderImageFixedPath = f.ImageFixedPath
	}
	if f.Flags&flagToClipboard != 0 && f.Tooltip == "" {
		f.Tooltip = "press to copy to Clipboard"
	}
	// zfloat.Maximize(&f.MaxWidth, f.MinWidth)
	if f.MaxWidth != 0 {
		zfloat.Minimize(&f.MinWidth, f.MaxWidth)
	}
	if f.Name == "" {
		str := zstr.PadCamelCase(item.FieldName, " ")
		str = zstr.FirstToTitleCase(str)
		f.Name = str
	}
	if f.Placeholder == "$HAS$" {
		f.Placeholder = f.Name
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
				if f.Flags&flagTimeFlags == 0 { // if no flags set, set default h,m,s
					f.Flags |= flagTimeFlags
				}
				setDurationColumns(f)
			}
			if f.Enum == "" && f.LocalEnum == "" {
				if f.MinWidth == 0 {
					f.MinWidth = 40
				}
				if f.MaxWidth == 0 {
					f.MaxWidth = 80
				}
			}
			break
		}
		fallthrough

	case zreflect.KindBool:
		if f.MinWidth == 0 {
			f.MinWidth = 20
		}
	case zreflect.KindString:
		if f.Flags&(flagHasHeaderImage|flagIsImage) != 0 {
			zfloat.Maximize(&f.MinWidth, f.HeaderSize.W)
			zfloat.Maximize(&f.MaxWidth, f.HeaderSize.W)
		}
		if f.MinWidth == 0 && f.Flags&flagIsButton == 0 && f.Enum == "" && f.LocalEnum == "" {
			f.MinWidth = 20
		}
	case zreflect.KindTime:
		if f.MaxWidth != 0 && f.MinWidth != 0 {
			break
		}
		if f.Flags&(flagTimeFlags|flagDateFlags) == 0 {
			f.Flags |= flagTimeFlags | flagDateFlags
		}
		if f.Flags&flagIsDuration != 0 {
			setDurationColumns(f)
		}
		if f.Format != "" {
			f.Columns = len(f.Format)
			if f.Format == "nice" {
				f.Columns = len(ztime.NiceFormat)
			}
			break
		}
		if f.MinWidth == 0 {
			if f.Flags&flagHasDays != 0 {
				f.Columns += 3
			}
			if f.Flags&flagHasMonths != 0 {
				f.Columns += 3
			}
			if f.Flags&flagHasYears != 0 {
				f.Columns += 3
			}
			if f.Flags&flagHasHours != 0 {
				f.Columns += 3
			}
			if f.Flags&flagHasMinutes != 0 {
				f.Columns += 3
			}
			if f.Flags&flagHasSeconds != 0 {
				f.Columns += 3
			}
			if f.MinWidth == 0 && f.Columns == 0 {
				f.MinWidth = 80
			}
		}

	case zreflect.KindFunc:
		if f.MinWidth == 0 {
			if f.Flags&flagIsImage != 0 {
				min := f.Size.W // * zscreen.GetMain().Scale
				//				min += ImageViewDefaultMargin.W * 2
				zfloat.Maximize(&f.MinWidth, min)
			}
		}
	}
	if f.WidgetName != "" {
		w := widgeters[f.WidgetName]
		if w != nil {
			sw, _ := w.(SetupWidgeter)
			if sw != nil {
				sw.SetupField(f)
			}
		}
	}

	callActionHandlerFunc(structure, f, SetupFieldAction, item.Address, nil) // need to use v.structure here, since i == -1
	// zlog.Info("Field:", f.ID, f.MinWidth, f.Size, f.MaxWidth)
	return true
}

func setDurationColumns(f *Field) {
	if f.Flags&flagHasMinutes != 0 {
		f.Columns += 3
	}
	if f.Flags&flagHasSeconds != 0 {
		f.Columns += 3
	}
	if f.Flags&flagHasHours != 0 {
		f.Columns += 3
	}
	if f.Flags&flagHasDays != 0 {
		f.Columns += 3
	}
}

var fieldEnums = map[string]zdict.Items{}

func SetEnum(name string, enum zdict.Items) {
	fieldEnums[name] = enum
}

func SetEnumItems(name string, nameValPairs ...interface{}) {
	var dis zdict.Items

	for i := 0; i < len(nameValPairs); i += 2 {
		var di zdict.Item
		di.Name = nameValPairs[i].(string)
		di.Value = nameValPairs[i+1]
		dis = append(dis, di)
	}
	fieldEnums[name] = dis
}

func AddStringBasedEnum(name string, vals ...interface{}) {
	var items zdict.Items
	for _, v := range vals {
		n := fmt.Sprintf("%v", v)
		i := zdict.Item{n, v}
		items = append(items, i)
	}
	fieldEnums[name] = items
}

func addNamesOfEnumValue(enumTitles map[string]mapValueToName, slice interface{}, f Field) {
	enum := fieldEnums[f.Enum]
	val := reflect.ValueOf(slice)
	slen := val.Len()
	m := enumTitles[f.ID]
	if m == nil {
		m = mapValueToName{}
		enumTitles[f.ID] = m
	}
	for i := 0; i < slen; i++ {
		fi := val.Index(i).Addr().Interface()
		ic, ierr := zreflect.ItterateStruct(fi, zreflect.Options{UnnestAnonymous: true})
		zlog.Assert(ierr == nil)
		item := ic.Children[f.Index]
		di := enum.FindValue(item.Interface)
		if di != nil {
			m[item.Interface] = di.Name
		}
	}
}

type mapValueToName map[interface{}]string

func getSortCache(slice interface{}, fields []Field, sortOrder []zui.SortInfo) (fieldMap map[string]*Field, enumTitles map[string]mapValueToName) {
	fieldMap = map[string]*Field{}
	enumTitles = map[string]mapValueToName{}

	for _, s := range sortOrder {
		for i, f := range fields {
			if f.ID == s.ID {
				fieldMap[f.ID] = &fields[i]
				// zlog.Info("ADD2cache:", f.ID)
				if f.LocalEnum != "" {
					// name := zstr.HeadUntil(f.LocalEnum, ".")
					// for _, f2 := range fields {
					// 	if f.FieldName == name {
					// 		// enum := ei.Interface.(zdict.ItemsGetter).GetItems()
					// 		// getNameOfEnumValue(slice, enum)
					// 		break
					// 	}
					// }
				} else if f.Enum != "" {
					addNamesOfEnumValue(enumTitles, slice, f)
					// zlog.Info("addNamesOfEnumValue:", f.Name, enumTitles)
				}
				break
			}
		}
	}
	return
}

func SortSliceWithFields(slice interface{}, fields []Field, sortOrder []zui.SortInfo) {
	// start := time.Now()
	fieldMap, enumTitles := getSortCache(slice, fields, sortOrder)
	// fmt.Printf("FieldMap: %+v %+v\n", fieldMap, sortOrder)
	// zlog.Info("SORT:", sortOrder, enumTitles)
	val := reflect.ValueOf(slice)
	sort.SliceStable(slice, func(i, j int) bool {
		ei := val.Index(i).Addr().Interface()
		ej := val.Index(j).Addr().Interface()
		ic, ierr := zreflect.ItterateStruct(ei, zreflect.Options{UnnestAnonymous: true})
		jc, jerr := zreflect.ItterateStruct(ej, zreflect.Options{UnnestAnonymous: true})
		zlog.Assert(ierr == nil && jerr == nil, ierr, jerr)
		for _, s := range sortOrder {
			f := fieldMap[s.ID]
			// zlog.Info("SORTING:", i, j, s.ID, f != nil)
			iitem := ic.Children[f.Index]
			jitem := jc.Children[f.Index]
			sliceEnumNames := enumTitles[f.ID]
			if sliceEnumNames != nil {
				ni := sliceEnumNames[iitem.Interface]
				nj := sliceEnumNames[jitem.Interface]
				r := (zstr.CaselessCompare(ni, nj) < 0) == s.SmallFirst
				if ni == nj {
					// zlog.Info("sliceEnumNames same:", r, s.ID, i, ni, j, nj)
					continue
				}
				// zlog.Info("sliceEnumNames:", r, s.ID, i, ni, j, nj)
				return r
			}
			switch iitem.Kind {
			case zreflect.KindBool:
				ia := iitem.Interface.(bool)
				ja := jitem.Interface.(bool)
				if ia == ja {
					continue
				}
				return (ia == false) == s.SmallFirst

			case zreflect.KindInt:
				ia, ierr := zint.GetAny(iitem.Interface)
				ja, jerr := zint.GetAny(jitem.Interface)
				zlog.Assert(ierr == nil && jerr == nil, ierr, jerr)
				if ia == ja {
					continue
				}
				return (ia < ja) == s.SmallFirst
			case zreflect.KindFloat:
				ia, ierr := zfloat.GetAny(iitem.Interface)
				ja, jerr := zfloat.GetAny(jitem.Interface)
				zlog.Assert(ierr == nil && jerr == nil, ierr, jerr)
				if ia == ja {
					continue
				}
				return (ia < ja) == s.SmallFirst
			case zreflect.KindString:
				ia, got := iitem.Interface.(string)
				ja, _ := jitem.Interface.(string)
				if !got {
					ia = fmt.Sprint(iitem.Interface)
					ja = fmt.Sprint(iitem.Interface)
				}
				if ia == ja {
					continue
				}
				// zlog.Info("sort:", i, j, ia, "<", ja, "less:", zstr.CaselessCompare(ia, ja) < 0, s.SmallFirst)
				return (zstr.CaselessCompare(ia, ja) < 0) == s.SmallFirst
			case zreflect.KindTime:
				ia := iitem.Interface.(time.Time)
				ja := jitem.Interface.(time.Time)
				if ia == ja {
					continue
				}
				return (ia.Sub(ja) < 0) == s.SmallFirst
			default:
				continue
			}
		}
		// zlog.Fatal(nil, "No sort fields set for struct")
		return false
	})
	// zlog.Info("SORT TIME:", time.Since(start))
}

// ID is convenience method to get id from a field if any (used often in HandleAction methods).
func ID(f *Field) string {
	if f != nil {
		return f.ID
	}
	return ""
}

// Name is convenience method to get name from a field if any (used often in HandleAction methods for debugging).
func Name(f *Field) string {
	if f != nil {
		return f.Name
	}
	return ""
}
