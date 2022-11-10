// The zfields package is functionality to create GUI from data structures.
// With reflection, the fields of structures are used to create stacks of GUI elements.
// The 'zui' tag on struct fields is used to stylize how the gui elements are created.
// This file is mostly about how these tags are parsed into a Field, and FieldView.go
// is where the building and updating of gui from the structure and Field info is done.

//go:build zui

package zfields

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui/zheader"
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
)

// type fieldType int

// UIStringer defines a ZUIString() method, which if present, shows a complex type as a string in FieldViews.
// We can't just use the fmt.Stringer interface as that would be too common.
type UIStringer interface {
	ZUIString() string
}

// Widgeter is an interface to make a type create it's own view when build with zfields package.
// It is registered with the RegisterWigeter function, and specified with the zui:"widget:xxx" tag.
type Widgeter interface {
	IsStatic() bool
	Create(f *Field) zview.View
	SetValue(view zview.View, val any)
}

// ReadWidgeter is a Widgeter that also can return it's value
type ReadWidgeter interface {
	GetValue(view zview.View) any
}

// SetupWidgeter is a Widgeter that also can setup it's field before creation
type SetupWidgeter interface {
	SetupField(f *Field)
}

// ActionType are the types of actions any type can handle an HandleAction method with type of.
// This allows a type to handle it's  Field setup, it's creation, editing, data changed and more.
type ActionType string

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

// The FlagType are a number of flags a field can have set, based on the struct field/tag it is created from.
type FlagType int64

const (
	FlagIsStatic           FlagType = 1 << iota // FlagIsStatic means this this field should not be editable
	FlagHasSeconds                              // FlagHasSeconds means it's a time/duration where seconds should be shown/used
	FlagHasMinutes                              // FlagHasMinutes is the same but for minutes
	FlagHasHours                                // FlagHasMinutes is the same but for hours
	FlagHasDays                                 // FlagHasMinutes is the same but for days of the month
	FlagHasMonths                               // FlagHasMinutes is the same but for months
	FlagHasYears                                // FlagHasMinutes is the same but for years
	FlagIsImage                                 // FlagIsImage means the field is an image. It is typically a string with a local served image file, or an external URL.
	FlagIsFixed                                 // FlagIsFixed means that an image's path/url has a fixed url in tag, not in field's string value
	FlagIsButton                                // FlagIsButton means the field is actually a button. It's type is irrelevant. Will call the PressedAction
	FlagHasHeaderImage                          // FlagHasHeaderImage is true true if it has a an image for showing in header
	FlagNoTitle                                 // FlagNoTitle i set when we don't use FieldName as a title, show nothing
	FlagToClipboard                             // FlagToClipboard: If gui item is pressed, contents pasted to clipboard, with copy icon shown briefly
	FlagIsPassword                              // Set if a text field is a password, shown as •••• and with special keyboard and auto password fill etc
	FlagIsDuration                              // Means a time should be shown as a duration. If it is static or OldSecs is set, it will repeatedly show the duration since it
	FlagIsOpaque                                // FlagIsOpaque means entire view will be covered when drawn
	FlagIsActions                               // FlagIsActions means a menu created from an enum is actions and not a value to set
	FlagHasFrame                                // FlagHasFrame is set if for the "frame" tag on a struct. A border is drawn around it.
	FlagIsGroup                                 // The "group" tag on a slice sets FlagIsGroup, and one of slice items is shown with a menu to choose between them. FlagHasFrame is set.
	FlagFrameIsTitled                           // If FlagFrameIsTitled is set the frame has a title shown, set if "titled specified for group or frame tag"
	FlagFrameTitledOnFrame                      // FlagFrameTitledOnFrame is set if the group or frame zui tag have the "ontag" value. The title is drawn inset into frame border then.
	FlagSkipIndicator                           // If FlagSkipIndicator is set as value on a group tag, the indicator field is not shown within, as it is shown in the menu.
	FlagLongPress                               // If FlagLongPress is set this button/image etc handles long-press
	FlagDisableAutofill                         // FlagDisableAutofill if set makes a text field not autofill
)

const (
	flagTimeFlags = FlagHasSeconds | FlagHasMinutes | FlagHasHours
	flagDateFlags = FlagHasDays | FlagHasMonths | FlagHasYears
)

type Field struct {
	Index                int    // Index is the position in the total amount of fields (inc anonymous) in struct
	ID                   string // ID is string from field's name using fieldNameToID(). TODO: Use this less, use FieldName more, as we are 100% sure what that is
	ActionValue          any    // ActionValue is used to send other information with an action into ActionHandler / ActionFieldHandler
	Name                 string // Name is 
	FieldName            string //
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
	Flags                FlagType
	Tooltip              string
	UpdateSecs           float64
	LabelizeWidth        float64
	LocalEnable          string
	LocalDisable         string
	LocalShow            string
	LocalHide            string
	Placeholder          string
	Columns              int
	Rows                 int
	SortSmallFirst       zbool.BoolInd
	SortPriority         int
	FractionDecimals     int
	OldSecs              int
	ValueStoreKey        string
	Visible              bool
	Disabled             bool
	SetEdited            bool
	WidgetName           string
	BrancherType         string
	SkipFieldNames       []string
	Styling              zstyle.Styling
}

var EmptyField = Field{
	Styling: zstyle.EmptyStyling,
}

type ActionHandler interface {
	HandleAction(f *Field, action ActionType, view *zview.View) bool
}

// type ActionFieldHandler interface {
// 	HandleFieldAction(f *Field, action ActionType, view *zview.View) bool
// }

var widgeters = map[string]Widgeter{}

var flagsList = []zbits.BitsetItem{
	zbits.BSItem("HasSeconds", int64(FlagHasSeconds)),
	zbits.BSItem("HasMinutes", int64(FlagHasMinutes)),
	zbits.BSItem("HasHours", int64(FlagHasHours)),
	zbits.BSItem("HasDays", int64(FlagHasDays)),
	zbits.BSItem("HasMonths", int64(FlagHasMonths)),
	zbits.BSItem("HasYears", int64(FlagHasYears)),
	zbits.BSItem("IsImage", int64(FlagIsImage)),
	zbits.BSItem("IsFixed", int64(FlagIsFixed)),
	zbits.BSItem("IsButton", int64(FlagIsButton)),
	zbits.BSItem("HasHeaderImage", int64(FlagHasHeaderImage)),
	zbits.BSItem("NoTitle", int64(FlagNoTitle)),
	zbits.BSItem("ToClipboard", int64(FlagToClipboard)),
	// zbits.BSItem("IsNamedSelection", int64(FlagIsNamedSelection)),
	zbits.BSItem("IsPassword", int64(FlagIsPassword)),
	zbits.BSItem("IsDuration", int64(FlagIsDuration)),
	zbits.BSItem("IsOpaque", int64(FlagIsOpaque)),
	zbits.BSItem("IsActions", int64(FlagIsActions)),
	zbits.BSItem("FrameIsTitled", int64(FlagFrameIsTitled)),
	zbits.BSItem("IsGroup", int64(FlagIsGroup)),
	zbits.BSItem("HasFrame", int64(FlagHasFrame)),
	zbits.BSItem("SkipIndicator", int64(FlagSkipIndicator)),
	zbits.BSItem("LongPress", int64(FlagLongPress)),
	zbits.BSItem("DisableAutofill", int64(FlagDisableAutofill)),
}

func (f FlagType) String() string {
	return zbits.Int64ToStringFromList(int64(f), flagsList)
}

func RegisterWigeter(name string, w Widgeter) {
	widgeters[name] = w
}

func (f Field) IsStatic() bool {
	return f.Flags&FlagIsStatic != 0
}

func (f *Field) SetFont(view zview.View, from *zgeo.Font) {
	to := view.(ztext.LayoutOwner)
	size := f.Styling.Font.Size
	if size <= 0 {
		if from != nil {
			size = from.Size
		} else {
			size = zgeo.FontDefaultSize
		}
	}
	style := f.Styling.Font.Style
	if from != nil {
		style = from.Style
	}
	if style == zgeo.FontStyleUndef {
		style = zgeo.FontStyleNormal
	}
	var font *zgeo.Font
	if f.Styling.Font.Name != "" {
		font = zgeo.FontNew(f.Styling.Font.Name, size, style)
	} else if from != nil {
		font = new(zgeo.Font)
		*font = *from
	} else {
		font = zgeo.FontNice(size, style)
	}
	// zlog.Info("Field SetFont:", view.Native().Hierarchy(), *font)
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

func (f *Field) SetFromReflectItem(structure any, item zreflect.Item, index int, immediateEdit bool) bool {
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
			f.Flags |= FlagIsPassword
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
		// case "cannil"
		// f.Flags |= flagAllowNil
		case "brancher":
			f.BrancherType = val
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
		case "skip":
			f.SkipFieldNames = strings.Split(val, "|")
		case "color":
			f.Colors = strings.Split(val, "|")
			if len(f.Colors) == 1 {
				f.Styling.FGColor.SetFromString(f.Colors[0])
			}
		case "bgcolor":
			f.Styling.BGColor.SetFromString(val)
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
			f.Flags |= FlagIsActions
		case "noautofill":
			f.Flags |= FlagDisableAutofill
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
				f.Styling.Spacing = n
			}
		case "storekey":
			f.ValueStoreKey = val
		case "static":
			if flag || val == "" {
				f.Flags |= FlagIsStatic
			}
		case "fracts":
			f.FractionDecimals = int(n)
		case "secs":
			f.Flags |= FlagHasSeconds
		case "oldsecs":
			f.OldSecs = int(n)
		case "mins":
			f.Flags |= FlagHasMinutes
		case "hours":
			f.Flags |= FlagHasHours
		case "maxwidth":
			if floatErr == nil {
				f.MaxWidth = n
			}
		case "longpress":
			f.Flags |= FlagLongPress
		case "group":
			for _, part := range strings.Split(val, "|") {
				switch part {
				case "titled":
					f.Flags |= FlagFrameIsTitled
				case "skipindicator":
					f.Flags |= FlagSkipIndicator
				case "onframe":
					f.Flags |= FlagFrameTitledOnFrame
				}
			}
			f.Flags |= FlagIsGroup | FlagHasFrame
		case "frame":
			// zlog.Info("Frame:", f.Name, f.FieldName)
			f.Flags |= FlagHasFrame
			for _, part := range strings.Split(val, "|") {
				switch part {
				case "titled":
					f.Flags |= FlagFrameIsTitled
				case "onframe":
					f.Flags |= FlagFrameTitledOnFrame
				}
			}
		case "fixed":
			f.Flags |= FlagIsFixed
		case "opaque":
			f.Flags |= FlagIsOpaque
		case "shadow":
			for _, part := range strings.Split(val, "|") {
				got := false
				if f.Styling.DropShadow.Delta.IsNull() {
					err := f.Styling.DropShadow.Delta.FromString(part)
					if err == nil {
						got = true
					} else {
						num, err := strconv.ParseFloat(part, 32)
						if err != nil {
							f.Styling.DropShadow.Delta = zgeo.SizeBoth(num)
							got = true
						}
					}
				}
				if !got && f.Styling.DropShadow.Blur == 0 {
					num, err := strconv.ParseFloat(part, 32)
					if err != nil {
						f.Styling.DropShadow.Blur = float32(num)
						got = true
					}
				}
				if !got && !f.Styling.DropShadow.Color.Valid {
					f.Styling.DropShadow.Color.SetFromString(part)
				}
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
						f.Styling.Font.Size = float64(n*sign) + zgeo.FontDefaultSize
					} else {
						f.Styling.Font.Size = float64(n)
					}
				} else {
					f.Styling.Font.Style = zgeo.FontStyleFromStr(part)
					if f.Styling.Font.Style == zgeo.FontStyleUndef {
						f.Styling.Font.Name = part
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
				f.Flags |= FlagIsImage
				f.Size.FromString(ssize)
				f.ImageFixedPath = path
			} else {
				f.Flags |= FlagHasHeaderImage
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
			f.Flags |= FlagNoTitle
		case "tip":
			f.Tooltip = val
		case "immediate":
			f.UpdateSecs = 0
		case "upsecs":
			if floatErr == nil && n > 0 {
				f.UpdateSecs = n
			}
		case "2clip":
			f.Flags |= FlagToClipboard
		// case "named-selection":
		// 	f.Flags |= FlagIsNamedSelection
		case "labelize":
			f.LabelizeWidth = n
			if n == 0 {
				f.LabelizeWidth = 200
			}
		case "button":
			f.Flags |= FlagIsButton
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
			f.Flags |= FlagIsStatic | FlagIsDuration
		}
	}
	if immediateEdit {
		f.UpdateSecs = 0
	}
	if f.HeaderSize.IsNull() {
		f.HeaderSize = f.Size
	}
	if f.HeaderImageFixedPath == "" && f.Flags&FlagHasHeaderImage != 0 {
		f.HeaderImageFixedPath = f.ImageFixedPath
	}
	if f.Flags&FlagToClipboard != 0 && f.Tooltip == "" {
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
		if f.Flags&(FlagHasHeaderImage|FlagIsImage) != 0 {
			zfloat.Maximize(&f.MinWidth, f.HeaderSize.W)
			zfloat.Maximize(&f.MaxWidth, f.HeaderSize.W)
		}
		if f.MinWidth == 0 && f.Flags&FlagIsButton == 0 && f.Enum == "" && f.LocalEnum == "" {
			f.MinWidth = 20
		}
	case zreflect.KindTime:
		if f.MaxWidth != 0 && f.MinWidth != 0 {
			break
		}
		if f.Flags&(flagTimeFlags|flagDateFlags) == 0 {
			f.Flags |= flagTimeFlags | flagDateFlags
		}
		if f.Flags&FlagIsDuration != 0 {
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
			if f.Flags&FlagHasDays != 0 {
				f.Columns += 3
			}
			if f.Flags&FlagHasMonths != 0 {
				f.Columns += 3
			}
			if f.Flags&FlagHasYears != 0 {
				f.Columns += 3
			}
			if f.Flags&FlagHasHours != 0 {
				f.Columns += 3
			}
			if f.Flags&FlagHasMinutes != 0 {
				f.Columns += 3
			}
			if f.Flags&FlagHasSeconds != 0 {
				f.Columns += 3
			}
			if f.Columns == 0 {
				f.MinWidth = 90
			} else {
				f.Columns += 2 // just add a bit of padding
			}
		}

	case zreflect.KindFunc:
		if f.MinWidth == 0 {
			if f.Flags&FlagIsImage != 0 {
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

	var fv FieldView
	fv.data = structure
	callActionHandlerFunc(&fv, f, SetupFieldAction, item.Address, nil) // need to use v.structure here, since i == -1
	// zlog.Info("Field:", f.ID, f.MinWidth, f.Size, f.MaxWidth)
	return true
}

// MergeInField copies in values from the Field *n* to *f*, overwriting except where *n* has undefined value
func (f *Field) MergeInField(n *Field) {
	oldStylng := f.Styling
	*f = *n
	f.Styling = oldStylng.MergeWith(n.Styling)
}

// func (f *Field) Styling() zstyle.Styling {
// 	s := zstyle.EmptyStyling
// 	if len(f.Colors) > 0 {
// 		s.FGColor = zgeo.ColorFromString(f.Colors[0])
// 	}
// 	s.DropShadow = f.Shadow
// 	s.Font = f.Font
// 	return s
// }

func (f *Field) TitleOrName() string {
	if f.Title != "" {
		return f.Title
	}
	return f.Name
}

func setDurationColumns(f *Field) {
	if f.Flags&FlagHasMinutes != 0 {
		f.Columns += 3
	}
	if f.Flags&FlagHasSeconds != 0 {
		f.Columns += 3
	}
	if f.Flags&FlagHasHours != 0 {
		f.Columns += 3
	}
	if f.Flags&FlagHasDays != 0 {
		f.Columns += 3
	}
}

var fieldEnums = map[string]zdict.Items{}

func SetEnum(name string, enum zdict.Items) {
	fieldEnums[name] = enum
}

func SetEnumItems(name string, nameValPairs ...any) {
	var dis zdict.Items

	for i := 0; i < len(nameValPairs); i += 2 {
		var di zdict.Item
		di.Name = nameValPairs[i].(string)
		di.Value = nameValPairs[i+1]
		dis = append(dis, di)
	}
	fieldEnums[name] = dis
}

func AddStringBasedEnum(name string, vals ...string) {
	var items zdict.Items
	for _, s := range vals {
		item := zdict.Item{s, s}
		items = append(items, item)
	}
	fieldEnums[name] = items
}

func AddAny2StringBasedEnum[S any](name string, vals ...S) {
	var items zdict.Items
	for _, v := range vals {
		s := fmt.Sprint(v)
		item := zdict.Item{s, s}
		items = append(items, item)
	}
	fieldEnums[name] = items
}

func addNamesOfEnumValue(enumTitles map[string]mapValueToName, slice any, f Field) {
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

type mapValueToName map[any]string

func getSortCache(slice any, fields []Field, sortOrder []zheader.SortInfo) (fieldMap map[string]*Field, enumTitles map[string]mapValueToName) {
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

func SortSliceWithFields(slice any, fields []Field, sortOrder []zheader.SortInfo) {
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
