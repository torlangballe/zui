// The zfields package is functionality to create UI from data structures.
// With reflection, the fields of structures are used to create stacks of GUI elements.
// The 'zui' tag on struct fields is used to stylize how the gui elements are created.
// This file is mostly about how these tags are parsed into a Field, and FieldView.go
// is where the building and updating of gui from the structure and Field info is done.

package zfields

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zbits"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
)

// UIStringer defines a ZUIString() method, which if present, shows a complex type as a string in FieldViews.
// We can't just use the fmt.Stringer interface as that would be too common.
type UIStringer interface {
	ZUIString() string
}

type UISetStringer interface {
	ZUISetFromString(str string)
}

// ActionType are the types of actions any type can handle an HandleAction method with type of.
// This allows a type to handle it's  Field setup, it's creation, editing, data changed and more.
type ActionType string

// SortInfo is information about how to sort fields/columns
type SortInfo struct {
	FieldName  string
	SmallFirst bool
}

type FieldParameters struct {
	HideStatic              bool
	ForceZeroOption         bool     // ForceZeroOption makes menus (and theoretically more) have a zero, or undefined option. This is set when creating a single dialog box for a whole slice of structures.
	AllStatic               bool     // AllStatic makes even not "static" tagged fields static. Good for showing in tables etc.
	TipsAsDescriptionLabels bool     // Used for labelizing fields, todo.
	UseInValues             []string // IDs that reflect a state. Fields with UseIn set will only show if it intersecs UseInValues. Exampe: TableView sets UseInValues=[$row], field with usein:$row shows in table but not dialog.
	SkipFieldNames          []string
}

const (
	NoAction              ActionType = ""            // Add a trigger for this to get ALL actions
	DataChangedActionPre  ActionType = "changed-pre" // called on struct before DataChangedAction on fields
	DataChangedAction     ActionType = "changed"     // called when value changed, typically programatically or edited. Called on fields with id, then on struct
	EditedAction          ActionType = "edited"      // called when value edited by user, DataChangedAction will also be called
	SetupFieldAction      ActionType = "setup"       // called when a field is being set up from a struct, view will be nil
	PressedAction         ActionType = "pressed"     // called when view is pressed, view is valid
	LongPressedAction     ActionType = "longpressed" // called when view is long-pressed, view is valid
	CreateFieldViewAction ActionType = "createview"  // called to create view, view is pointer to view and is returned in it
	CreatedViewAction     ActionType = "createdview" // called after view created, view is pointer to newly created view.

	RowUseInSpecialName = "$row"
)

// The FlagType are a number of flags a field can have set, based on the struct field/tag it is created from.
type FlagType zbits.NamedBit

const (
	FlagIsStatic             FlagType = 1 << iota // FlagIsStatic means this this field should not be editable
	FlagHasSeconds                                // FlagHasSeconds means it's a time/duration where seconds should be shown/used
	FlagHasMinutes                                // FlagHasMinutes is the same but for minutes
	FlagHasHours                                  // FlagHasMinutes is the same but for hours
	FlagHasDays                                   // FlagHasMinutes is the same but for days of the month
	FlagHasMonths                                 // FlagHasMinutes is the same but for months
	FlagHasYears                                  // FlagHasMinutes is the same but for years
	FlagIsImage                                   // FlagIsImage means the field is an image. It is typically a string with a local served image file, or an external URL.
	FlagIsFixed                                   // FlagIsFixed means that an image's path/url has a fixed url in tag, not in field's string value, or an editable slice can't be added to/removed from.
	FlagIsButton                                  // FlagIsButton means the field is actually a button. It's type is irrelevant. Will call the PressedAction
	FlagHasHeaderImage                            // FlagHasHeaderImage is true true if it has a an image for showing in header
	FlagNoTitle                                   // FlagNoTitle i set when we don't use FieldName as a title, show nothing
	FlagToClipboard                               // FlagToClipboard: If gui item is pressed, contents pasted to clipboard, with copy icon shown briefly
	FlagIsPassword                                // Set if a text field is a password, shown as •••• and with special keyboard and auto password fill etc
	FlagIsDuration                                // Means a time should be shown as a duration. If it is static or OldSecs is set, it will repeatedly show the duration since it
	FlagIsOpaque                                  // FlagIsOpaque means entire view will be covered when drawn
	FlagIsActions                                 // FlagIsActions means a menu created from an enum is actions and not a value to set
	FlagHasFrame                                  // FlagHasFrame is set if for the "frame" tag on a struct. A border is drawn around it.
	FlagIsGroup                                   // The "group" tag on a slice sets FlagIsGroup, and one of slice items is shown with a menu to choose between them. FlagHasFrame is set.
	FlagGroupSingle                               // if The "group" tag has "single" option, a group of slices is shown one at a time with a menu to choose which one to view.
	FlagFrameIsTitled                             // If FlagFrameIsTitled is set the frame has a title shown, set if "titled specified for group or frame tag"
	FlagFrameTitledOnFrame                        // FlagFrameTitledOnFrame is set if the group or frame zui tag have the "ontag" value. The title is drawn inset into frame border then.
	FlagSkipIndicator                             // If FlagSkipIndicator is set as value on a group tag, the indicator field is not shown within, as it is shown in the menu.
	FlagLongPress                                 // If FlagLongPress is set this button/image etc handles long-press
	FlagDisableAutofill                           // FlagDisableAutofill if set makes a text field not autofill
	FlagIsSearchable                              // This field can be used to search in tables etc
	FlagIsUseInValue                              // This value is set as a string to InNames before entire struct is created
	FlagAllowEmptyAsZero                          // This shows the empty value as nothing. So int 0 would be shown as "" in text
	FlagIsForZDebugOnly                           // Set if "zdebug" tag. Only used if zui.DebugOwnerMode true
	FlagIsRebuildAllOnChange                      // If set, and this item is edited, rebuild the FieldView
	FlagIsURL                                     // Field is string, and it's a url
)

const (
	flagTimeFlags = FlagHasSeconds | FlagHasMinutes | FlagHasHours
	flagDateFlags = FlagHasDays | FlagHasMonths | FlagHasYears
)

type Field struct {
	Index int // Index is the position in the total amount of fields (inc anonymous) in struct
	// ID                   string // ID is string from field's name using fieldNameToID(). TODO: Use this less, use FieldName more, as we are 100% sure what that is
	ActionValue          any // ActionValue is used to send other information with an action into ActionHandler / ActionFieldHandler
	Name                 string
	FieldName            string //
	PackageName          string // the name of the package struct the field is in is from
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
	OffImagePath         string
	HeaderImageFixedPath string
	Height               float64
	Enum                 string
	LocalEnum            string
	Size                 zgeo.Size
	HeaderSize           zgeo.Size
	Flags                FlagType
	Tooltip              string
	Description          string
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
	UseIn                []string // If UseIn set, field will only be made if FieldView has paramater UseInValues with corresponding entry
	Styling              zstyle.Styling
	CustomFields         map[string]string // CustomFields are anything not parsed by SetFromReflectItem TODO: Rename to custom options or something
	Download             string
	StringSep            string // "sep": if set value is actually a slice, set/got from string separated by StringSep, no value given is space as separator.
}

var EmptyField = Field{
	Styling:      zstyle.EmptyStyling,
	CustomFields: map[string]string{},
}

var flagsNameMap = zbits.NamedBitMap{
	"HasSeconds":      uint64(FlagHasSeconds),
	"HasMinutes":      uint64(FlagHasMinutes),
	"HasHours":        uint64(FlagHasHours),
	"HasDays":         uint64(FlagHasDays),
	"HasMonths":       uint64(FlagHasMonths),
	"HasYears":        uint64(FlagHasYears),
	"IsImage":         uint64(FlagIsImage),
	"IsFixed":         uint64(FlagIsFixed),
	"IsButton":        uint64(FlagIsButton),
	"HasHeaderImage":  uint64(FlagHasHeaderImage),
	"NoTitle":         uint64(FlagNoTitle),
	"ToClipboard":     uint64(FlagToClipboard),
	"IsPassword":      uint64(FlagIsPassword),
	"IsDuration":      uint64(FlagIsDuration),
	"IsOpaque":        uint64(FlagIsOpaque),
	"IsActions":       uint64(FlagIsActions),
	"FrameIsTitled":   uint64(FlagFrameIsTitled),
	"IsGroup":         uint64(FlagIsGroup),
	"HasFrame":        uint64(FlagHasFrame),
	"SkipIndicator":   uint64(FlagSkipIndicator),
	"LongPress":       uint64(FlagLongPress),
	"DisableAutofill": uint64(FlagDisableAutofill),
}

// callSetupWidgeter is called to set gui widgets registered for use in zui tags.
// It is dependent on a gui, so injected with this func variable.
var (
	callSetupWidgeter func(f *Field)
	fieldEnums        = map[string]zdict.Items{}
)

func (f FlagType) String() string {
	return zbits.NamedBit(f).ToString(flagsNameMap)
}

func (f *Field) DebugName() string {
	if f == nil {
		return "nil"
	}
	return f.FieldName
}

func (f Field) IsStatic() bool {
	return f.HasFlag(FlagIsStatic)
}

func (f Field) HasFlag(flag FlagType) bool {
	return f.Flags&flag != 0
}

func findFieldWithIndex(fields *[]Field, index int) *Field {
	for i, f := range *fields {
		if f.Index == index {
			return &(*fields)[i]
		}
	}
	return nil
}

func FindLocalFieldWithFieldName(structure any, name string) (reflect.Value, int) {
	name = zstr.HeadUntil(name, ".")
	fval, _, i := zreflect.FieldForName(structure, true, name)
	return fval, i
}

// func fieldNameToID(name string) string {
// 	return zstr.FirstToLowerWithAcronyms(name)
// }

func (f *Field) SetFromReflectValue(rval reflect.Value, sf reflect.StructField, index int, params FieldParameters) bool {
	f.Index = index
	//	f.ID = fieldNameToID(sf.Name)
	fTypeName := rval.Type().Name()
	f.Kind = zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	f.FieldName = sf.Name
	// zlog.Info("FIELD:", f.FieldName)
	f.Alignment = zgeo.AlignmentNone
	f.UpdateSecs = -1
	f.Rows = 1
	f.SortSmallFirst = zbool.Unknown
	f.SetEdited = true
	f.Vertical = zbool.Unknown
	f.PackageName = rval.Type().PkgPath()
	var skipping bool
	// zlog.Info("Packagename:", f.PackageName, f.FieldName)
	// zlog.Info("Field:", f.ID)
	for _, part := range zreflect.GetTagAsMap(string(sf.Tag))["zui"] {
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
		barParts := strings.Split(val, "|")
		if key == "IN" {
			skipping = !zstr.SlicesIntersect(params.UseInValues, barParts)
		}
		if skipping {
			continue
		}
		n, floatErr := strconv.ParseFloat(val, 32)
		flag := zbool.FromString(val, false)
		switch key {
		case "search":
			f.Flags |= FlagIsSearchable
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
		case "url":
			f.Flags |= FlagIsURL
		case "usein":
			f.UseIn = barParts
		case "rebuild":
			f.Flags |= FlagIsRebuildAllOnChange
		case "sep":
			f.StringSep = val
			if val == "" {
				f.StringSep = " "
			}
		case "isuseinval":
			f.Flags |= FlagIsUseInValue
		case "color":
			f.Colors = barParts
			if len(f.Colors) == 1 {
				f.Styling.FGColor.SetFromString(f.Colors[0])
			}
		case "bgcolor":
			f.Styling.BGColor.SetFromString(val)
		case "download":
			f.Download = val
		case "zdebug":
			f.Flags |= FlagIsForZDebugOnly
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
			f.Size, _ = zgeo.SizeFromString(val)
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
		case "allowempty":
			f.Flags |= FlagAllowEmptyAsZero
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
			for _, part := range barParts {
				switch part {
				case "titled":
					f.Flags |= FlagFrameIsTitled
				case "skipindicator":
					f.Flags |= FlagSkipIndicator
				case "onframe":
					f.Flags |= FlagFrameTitledOnFrame
				case "single":
					f.Flags |= FlagGroupSingle
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
					var err error
					f.Styling.DropShadow.Delta, err = zgeo.SizeFromString(part)
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

		case "off":
			f.OffImagePath = "images/" + val
		case "image", "himage":
			var size zgeo.Size
			var path string
			for _, part := range barParts {
				var s zgeo.Size
				s, err := zgeo.SizeFromString(part)
				if err == nil {
					size = s
					// zlog.Info("Got size:", part, size)
				} else {
					path = "images/" + part
				}
			}
			if key == "image" {
				f.Size = size
				f.Flags |= FlagIsImage
				f.ImageFixedPath = path
			} else {
				f.Flags |= FlagHasHeaderImage
				f.HeaderSize = size
				f.HeaderImageFixedPath = path
			}
		case "enum":
			if zstr.HasPrefix(val, "./", &f.LocalEnum) {
			} else {
				_, got := fieldEnums[val]
				if !got {
					zlog.Error(nil, "no such enum:", val, fieldEnums)
				}
				f.Enum = val
			}
		case "notitle":
			f.Flags |= FlagNoTitle
		case "tip":
			f.Tooltip = val
		case "desc":
			f.Description = val
		case "immediate":
			f.UpdateSecs = 0
		case "upsecs":
			if floatErr == nil && n > 0 {
				f.UpdateSecs = n
			}
		case "2clip":
			f.Flags |= FlagToClipboard
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
				f.Placeholder = "$HAS$" // set to this special value to set to name once set
			}
		case "since":
			f.Flags |= FlagIsStatic | FlagIsDuration
		default:
			f.CustomFields[key] = val
		}
	}
	if rval.Type() == reflect.TypeOf(zgeo.Color{}) {
		f.WidgetName = "zcolor"
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
		str := zstr.PadCamelCase(sf.Name, " ")
		str = zlocale.FirstToTitleCaseExcept(str, "")
		f.Name = str
	}
	if f.Placeholder == "$HAS$" {
		f.Placeholder = f.Name
	}

	switch f.Kind {
	case zreflect.KindFloat:
		if f.MinWidth == 0 {
			f.MinWidth = 64
		}
		if f.MaxWidth == 0 {
			f.MaxWidth = 64
		}
	case zreflect.KindInt:
		if fTypeName != "BoolInd" {
			if sf.PkgPath == "time" && fTypeName == "Duration" {
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
			f.Columns += 3 // in case we show offset
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
	if f.WidgetName != "" && callSetupWidgeter != nil {
		callSetupWidgeter(f)
	}
	// zlog.Info("Field:", f.Name, f.Flags&FlagHasSeconds != 0)
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

func SetEnum(name string, enum zdict.Items) {
	fieldEnums[name] = enum
}

func GetEnum(name string) zdict.Items {
	return fieldEnums[name]
}

func SetEnumItems(name string, nameValPairs ...any) {
	var dis zdict.Items

	for i := 0; i < len(nameValPairs); i += 2 {
		var di zdict.Item
		di.Name = fmt.Sprint(nameValPairs[i])
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

func AddAnyToEnum[S any](name string, vals ...S) {
	var items zdict.Items
	for _, v := range vals {
		s := fmt.Sprint(v)
		item := zdict.Item{s, v}
		items = append(items, item)
	}
	fieldEnums[name] = items
}

func addNamesOfEnumValue(enumTitles map[string]mapValueToName, slice any, f Field) {
	enum := fieldEnums[f.Enum]
	val := reflect.ValueOf(slice)
	slen := val.Len()
	m := enumTitles[f.FieldName]
	if m == nil {
		m = mapValueToName{}
		enumTitles[f.FieldName] = m
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

func getSortCache(slice any, fields []Field, sortOrder []SortInfo) (fieldMap map[string]*Field, enumTitles map[string]mapValueToName) {
	fieldMap = map[string]*Field{}
	enumTitles = map[string]mapValueToName{}

	for _, s := range sortOrder {
		for i, f := range fields {
			if f.FieldName == s.FieldName {
				fieldMap[f.FieldName] = &fields[i]
				// zlog.Info("ADD2cache:", f.FieldName)
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

func SortSliceWithFields(slice any, fields []Field, sortOrder []SortInfo) {
	// start := time.Now()
	fieldMap, enumTitles := getSortCache(slice, fields, sortOrder)
	// fmt.Printf("FieldMap: %+v %+v\n", fieldMap, sortOrder)
	val := reflect.ValueOf(slice)
	// zlog.Info("SORT:", sortOrder, enumTitles, val.Len())
	sort.SliceStable(slice, func(i, j int) bool {
		ei := val.Index(i).Addr().Interface()
		ej := val.Index(j).Addr().Interface()
		ic, ierr := zreflect.ItterateStruct(ei, zreflect.Options{UnnestAnonymous: true})
		jc, jerr := zreflect.ItterateStruct(ej, zreflect.Options{UnnestAnonymous: true})
		zlog.Assert(ierr == nil && jerr == nil, ierr, jerr)
		for _, s := range sortOrder {
			f := fieldMap[s.FieldName]
			// zlog.Info("SORTING:", i, j, s.FieldName, f != nil)
			iitem := ic.Children[f.Index]
			jitem := jc.Children[f.Index]
			sliceEnumNames := enumTitles[f.FieldName]
			if sliceEnumNames != nil {
				ni := sliceEnumNames[iitem.Interface]
				nj := sliceEnumNames[jitem.Interface]
				r := (zstr.CaselessCompare(ni, nj) < 0) == s.SmallFirst
				if ni == nj {
					// zlog.Info("sliceEnumNames same:", r, s.FieldName, i, ni, j, nj)
					continue
				}
				// zlog.Info("sliceEnumNames:", r, s.FieldName, i, ni, j, nj)
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

// FN is convenience method to get FieldName from a field if any (used often in HandleAction methods).
func (f *Field) FN() string {
	if f != nil {
		return f.FieldName
	}
	return "nil"
}

// Name is convenience method to get name from a field if any (used often in HandleAction methods for debugging).
func Name(f *Field) string {
	if f != nil {
		return f.Name
	}
	return ""
}

func ForEachField(structure any, params FieldParameters, fields []Field, got func(index int, f *Field, val reflect.Value, sf reflect.StructField)) {
	if len(fields) == 0 {
		zreflect.ForEachField(structure, true, func(index int, val reflect.Value, sf reflect.StructField) bool {
			f := EmptyField
			if !f.SetFromReflectValue(val, sf, index, params) {
				return true
			}
			fields = append(fields, f)
			return true
		})
	}
	zreflect.ForEachField(structure, true, func(index int, val reflect.Value, sf reflect.StructField) bool {
		f := findFieldWithIndex(&fields, index)
		if f == nil {
			return true
		}
		if f.Flags&FlagIsUseInValue != 0 {
			zstr.AddToSet(&params.UseInValues, fmt.Sprint(val.Interface()))
		}
		return true
	})
	zreflect.ForEachField(structure, true, func(index int, val reflect.Value, sf reflect.StructField) bool {
		if zstr.IndexOf(sf.Name, params.SkipFieldNames) != -1 {
			return true
		}
		f := findFieldWithIndex(&fields, index)
		if f == nil {
			return true
		}
		if f.Flags&FlagIsForZDebugOnly != 0 && !zui.DebugOwnerMode {
			return true
		}
		isInRow := zstr.StringsContain(params.UseInValues, RowUseInSpecialName)
		wantsDialog := zstr.StringsContain(f.UseIn, "$dialog")
		// zlog.Info("IFIn:", f.Name, wantsDialog, len(f.UseIn), isInRow, zstr.SlicesIntersect(f.UseIn, params.UseInValues))
		if !(len(f.UseIn) == 0 || (zstr.SlicesIntersect(f.UseIn, params.UseInValues) || (isInRow && !wantsDialog))) {
			return true
		}
		got(index, f, val, sf)
		return true
	})
}

func FindIndicatorOfSlice(slicePtr any) string {
	s := zslice.MakeAnElementOfSliceType(slicePtr)
	_, f, got := FindIndicatorRValOfStruct(s)
	if got {
		return f.Name
	}
	return ""
}

func FindIndicatorRValOfStruct(structPtr any) (rval reflect.Value, field *Field, got bool) {
	// fmt.Printf("CreateSliceGroupOwner %s %+v\n", grouper.GetGroupBase().Hierarchy(), s)
	ForEachField(structPtr, FieldParameters{}, nil, func(index int, f *Field, val reflect.Value, sf reflect.StructField) {
		for _, part := range zreflect.GetTagAsMap(string(sf.Tag))["zui"] {
			if part == "indicator" {
				rval = val
				field = f
				got = true
			}
		}
	})
	return
}

func (f *Field) IsImageToggle() bool {
	return f.Flags&FlagIsImage != 0 && f.OffImagePath != ""
}
