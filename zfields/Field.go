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
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zbits"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdebug"
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
	ZUIString(allowEmpty bool) string
}

// UISetStringer is like UIStringer, but for allowing a type to be edited as a string, then converted from that string.
type UISetStringer interface {
	ZUISetFromString(str string)
}

// SortInfo is information about how to sort fields/columns. See SortSliceWithFields().
type SortInfo struct {
	FieldName  string
	SmallFirst bool
}

// FieldParameters are parameters for a FieldView's presentation.
type FieldParameters struct {
	HideStatic      bool
	ForceZeroOption bool     // ForceZeroOption makes menus (and theoretically more) have a zero, or undefined option. This is set when creating a single dialog box for a whole slice of structures.
	AllStatic       bool     // AllStatic makes even not "static" tagged fields static. Good for showing in tables etc.
	UseInValues     []string // IDs that reflect a state. Fields with UseIn set will only show if it intersects UseInValues. Example: TableView sets UseInValues=[$row], field with usein:$row shows in table but not dialog.
	SkipFieldNames  []string
}

// ActionType are the types of actions any type can handle an HandleAction method with type of.
// This allows a type to handle its  Field setup, its creation, editing, data changed and more.
type ActionType string

// DocumentationLink is used to flag a string to create doc links
type DocumentationLink string

const (
	NoAction              ActionType = ""             // Add a trigger for this to get ALL actions
	DataChangedActionPre  ActionType = "changed-pre"  // called on struct before DataChangedAction on fields
	DataChangedAction     ActionType = "changed"      // called when value changed, typically programmatically or edited. Called on fields with id, then on struct
	EditedAction          ActionType = "edited"       // called when value edited by user, DataChangedAction will also be called
	SetupFieldAction      ActionType = "setup"        // called when a field is being set up from a struct, view will be nil
	PressedAction         ActionType = "pressed"      // called when view is pressed, view is valid
	LongPressedAction     ActionType = "long-pressed" // called when view is long-pressed, view is valid
	CreateFieldViewAction ActionType = "create-view"  // called to create view, view is pointer to view and is returned in it
	CreatedViewAction     ActionType = "created-view" // called after view created, view is pointer to newly created view.

	RowUseInSpecialName    = "$row"
	DialogUseInSpecialName = "$dialog"
)

// The FlagType stores Field boolean options, based on the tag it is created from.
type FlagType zbits.NamedBit

const (
	FlagIsStatic                 FlagType = 1 << iota // FlagIsStatic means this this field should not be editable
	FlagHasSeconds                                    // FlagHasSeconds means its a time/duration where seconds should be shown/used
	FlagHasMinutes                                    // FlagHasMinutes is the same but for minutes
	FlagHasHours                                      // FlagHasMinutes is the same but for hours
	FlagHasDays                                       // FlagHasMinutes is the same but for days of the month
	FlagHasMonths                                     // FlagHasMinutes is the same but for months
	FlagHasYears                                      // FlagHasMinutes is the same but for years
	FlagIsImage                                       // FlagIsImage means the field is an image. It is typically a string with a local served image file, or an external URL.
	FlagIsFixed                                       // FlagIsFixed means that an image's path/url has a fixed url in tag, not in field's string value, or an editable slice can't be added to/removed from. Or it can mean an existing password.
	FlagIsButton                                      // FlagIsButton means the field is actually a button. its type is irrelevant. Will call the PressedAction
	FlagHasHeaderImage                                // FlagHasHeaderImage is true true if it has a an image for showing in header
	FlagNoTitle                                       // FlagNoTitle i set when we don't use FieldName as a title, show nothing
	FlagToClipboard                                   // FlagToClipboard: If gui item is pressed, contents pasted to clipboard, with copy icon shown briefly
	FlagIsPassword                                    // Set if a text field is a password, shown as •••• and with special keyboard and auto password fill etc. password:existing sets FlagIsFixed, is's an existing password.
	FlagIsDuration                                    // Means a time should be shown as a duration. If it is static or OldSecs is set, it will repeatedly show the duration since it
	FlagIsOpaque                                      // FlagIsOpaque means entire view will be covered when drawn
	FlagIsActions                                     // FlagIsActions means a menu created from an enum is actions and not a value to set
	FlagHasFrame                                      // FlagHasFrame is set if for the "frame" tag on a struct. A border is drawn around it.
	FlagIsGroup                                       // The "group" tag on a slice sets FlagIsGroup, and one of slice items is shown with a menu to choose between them. FlagHasFrame is set.
	FlagGroupSingle                                   // if The "group" tag has "single" option, a group of slices is shown one at a time with a menu to choose which one to view.
	FlagFrameIsTitled                                 // If FlagFrameIsTitled is set the frame has a title shown, set if "titled specified for group or frame tag"
	FlagFrameTitledOnFrame                            // FlagFrameTitledOnFrame is set if the group or frame zui tag have the "ontag" value. The title is drawn inset into frame border then.
	FlagSkipIndicator                                 // If FlagSkipIndicator is set as value on a group tag, the indicator field is not shown within, as it is shown in the menu.
	FlagPress                                         // If FlagLongPress is set this button/image etc handles press
	FlagLongPress                                     // If FlagLongPress is set this button/image etc handles long-press
	FlagDisableAutofill                               // FlagDisableAutofill if set makes a text field not autofill
	FlagIsSearchable                                  // This field can be used to search in tables etc
	FlagIsUseInValue                                  // ????This value is set as a string to InNames before entire struct is created
	FlagAllowEmptyAsZero                              // This shows the empty value as nothing. So int 0 would be shown as "" in text
	FlagZeroIsBig                                     // If set, a zero value is considered big, currenlty used in sorting
	FlagIsForZDebugOnly                               // Set if "zdebug" tag. Only used if zui.DebugOwnerMode true
	FlagIsRebuildAllOnChange                          // If set, and this item is edited, rebuild the FieldView
	FlagIsURL                                         // (Field is string, and it's a url) OR (it has Path set to fixed URL)
	FlagIsDocumentation                               // It is a .Path link to Documentation view.
	FlagIsDownload                                    // If set, the gui control made can be pressed to download, using "path", is audio, it might need to be long-pressed as pressing plays
	FlagIsLabelize                                    // Set to force rows of <label> view [desc] in children
	FlagLabelizeWithDescriptions                      // Set to make labelized rows add a description to far right, if FlagIsLabelize
	FlagShowSliceCount                                // Set to show a count of items in slice. Typically used on rows. Sets FlagIsStatic.
	FlagShowPopup                                     // press to show a popup of contents
	FlagIsLockable                                    // Show a lock icon to right of item when labelized. Disables/Hides.
	FlagHeaderLockable                                // Show a lock icon on header, for locking selected rows.
	FlagIsStart                                       // This field represents a start value, probably a time, and so far for if FlagHeaderLockable.
	FlagIsEnd                                         // This field represents an end value, probably a time, and so far for if FlagHeaderLockable.
	FlagDontJustifyHeader                             // If set, header is default justified, not using Field.Justify
	FlagCheckerCell                                   // Ever other column with this is darkened a bit.
	FlagFutureInvalid                                 // For time, show red if time is future.
	FlagPastInvalid                                   // For time, show red if time is future.
	FlagHasDefault                                    // If true Field.Default string is used for default value of field. Can be parsed to numbers too.
	FlagIsOpen                                        // This field can open the struct if in a table or something else that handles it.
	FlagIsOpener                                      // Flag IsOpen, and is set to a view or edit icon by table or something.
	FlagShowIfExtraSpace                              // When building a row (for now), field is added with ShowIfExtraSpace of sum of widths of self and similar onces before it
	FlagDontLabelize                                  // When we are labelizing items, use full space with no label for this one.
	FlagOmitZero                                      // Like for json, skip an item if it is zero, but only if static
)

const (
	flagTimeFlags      = FlagHasSeconds | FlagHasMinutes | FlagHasHours
	flagDateFlags      = FlagHasDays | FlagHasMonths | FlagHasYears
	FlagIsTimeBoundary = FlagIsStart | FlagIsEnd
)

// The Field struct stores information about a struct field, based on zui:"" tags, used for displaying it.
type Field struct {
	Index int // Index is the position in the total amount of fields (inc anonymous) in struct.
	// ActionValue          any    // ActionValue is used to send other information with an action into ActionHandler / ActionFieldHandler.
	FieldName            string            // The FieldName is the exact name of the struct field.
	PackageName          string            // The name of the package struct the field variable type is in.
	Name                 string            // zui:"name". Name is generated from the struct fields name, but can be overridden with the name tag.
	Title                string            // zui:"title". Uses instead of Name of item in label in list, and in header if Header not set
	MaxWidth             float64           // zui:"width/maxwidth"
	MinWidth             float64           // zui:"width/minwidth"
	Kind                 zreflect.TypeKind // Kind stores the general kind of value the field is.
	Vertical             zbool.BoolInd     // Used to override layout, typically if this field is a struct. Unknown means do default.
	Alignment            zgeo.Alignment    // zui:"align". see zgeo.Alignment for string values. Aligns item to parent
	Justify              zgeo.Alignment    // zui:"justify". Justifies text etc within field.
	Format               string            // zui:"format". Used to format numbers, time. See below.
	Colors               []string          // zui:"color". Colors for field, can be multiple.
	ImageFixedPath       string            // zui:""
	OffImagePath         string            // zui:""
	HeaderImageFixedPath string            // zui:""
	Header               string            // zui:"header". Uses as header column name, overriding of title/name.
	Path                 string            // zui:""
	Height               float64           // zui:""
	Enum                 string            // zui:""
	LocalEnum            string            // zui:""
	Size                 zgeo.Size         // zui:""
	HeaderSize           zgeo.Size         // zui:""
	Margin               zgeo.Size         // zui:""
	Flags                FlagType          // zui:""
	Tooltip              string            // zui:""
	Description          string            // zui:""
	UpdateSecs           float64           // zui:""
	LocalEnable          string            // zui:""
	LocalDisable         string            // zui:""
	LocalShow            string            // zui:""
	LocalHide            string            // zui:""
	Placeholder          string            // zui:""
	Columns              int               // zui:""
	Rows                 int               // zui:""
	SortSmallFirst       zbool.BoolInd     // zui:""
	SortPriority         int               // zui:""
	FractionDecimals     int               // zui:""
	OldSecs              int               // zui:""
	ValueStoreKey        string            // zui:""
	Visible              bool              // zui:""
	Disabled             bool              // zui:""
	SetEdited            bool              // zui:""
	WidgetName           string            // zui:""
	UseIn                []string          // If UseIn set, field will only be made if FieldView has paramater UseInValues with corresponding entry.
	Styling              zstyle.Styling    // zui:""
	CustomFields         map[string]string // CustomFields are anything not parsed by SetFromReflectItem TODO: Rename to custom options or something.
	StringSep            string            // "sep": if set value is actually a slice, set/got from string separated by StringSep, no value given is space as separator.
	RPCCall              string            // an RPC method to Call, typically on press of a button.
	Filters              []string          // Registered filters (| separated). Currently used for textview fields to filter text in/output. Built in: $lower $upper $uuid $hex $alpha $num $alphanum.
	ZeroText             string            // Text to replace a zero value with set with "allowempty" tag.
	MaxText              string            // Text to replace a "maximum" value with set with "allowempty" tag.
	Wrap                 string            // How to wrap if text. As in ztextinfo.WrapType.String().
	Default              string            // Default value for field. May be numbers or strings.
	Ask                  string            // If present, buttons etc show a Yes dialog with this before triggering
	Prefix               string            // Added to static text
	Suffix               string            // Added to static text
	Required             string            // If set, fields must be non-zero after editing. If Required is not RequiredSingleValue, it is a group id where at least one field with this Required group has to be non-zero.
	Radio                string            // If set, value is an enum name. Field must be value type of enum.
	WhenMods             zkeyboard.Modifier
	Transformer          string // A transformer that turns this field, if static, to a string, registed with RegisterFieldTransformer()
}

const (
	MemoryFormat        = "memory"
	StorageFormat       = "storage"
	BPSFormat           = "bps"   // bits or bytes pr sec
	HumanFormat         = "human" // human-readable int
	RequiredSingleValue = "$single"
	IsRadioValue        = "$radio"
)

var EmptyField = Field{
	Styling:      zstyle.EmptyStyling,
	CustomFields: map[string]string{},
}

var flagsNameMap = zbits.NamedBitMap{
	"HasSeconds":               uint64(FlagHasSeconds),
	"HasMinutes":               uint64(FlagHasMinutes),
	"HasHours":                 uint64(FlagHasHours),
	"HasDays":                  uint64(FlagHasDays),
	"HasMonths":                uint64(FlagHasMonths),
	"HasYears":                 uint64(FlagHasYears),
	"IsImage":                  uint64(FlagIsImage),
	"IsFixed":                  uint64(FlagIsFixed),
	"IsButton":                 uint64(FlagIsButton),
	"IsStatic":                 uint64(FlagIsStatic),
	"HasHeaderImage":           uint64(FlagHasHeaderImage),
	"NoTitle":                  uint64(FlagNoTitle),
	"ToClipboard":              uint64(FlagToClipboard),
	"IsPassword":               uint64(FlagIsPassword),
	"IsDuration":               uint64(FlagIsDuration),
	"IsOpaque":                 uint64(FlagIsOpaque),
	"IsActions":                uint64(FlagIsActions),
	"FrameIsTitled":            uint64(FlagFrameIsTitled),
	"IsGroup":                  uint64(FlagIsGroup),
	"HasFrame":                 uint64(FlagHasFrame),
	"SkipIndicator":            uint64(FlagSkipIndicator),
	"LongPress":                uint64(FlagLongPress),
	"Press":                    uint64(FlagPress),
	"DisableAutofill":          uint64(FlagDisableAutofill),
	"IsSearchable":             uint64(FlagIsSearchable),
	"IsUseInValue":             uint64(FlagIsUseInValue),
	"AllowEmptyAsZero":         uint64(FlagAllowEmptyAsZero),
	"ZeroIsBig":                uint64(FlagZeroIsBig),
	"IsForZDebugOnly":          uint64(FlagIsForZDebugOnly),
	"IsRebuildAllOnChange":     uint64(FlagIsRebuildAllOnChange),
	"IsURL":                    uint64(FlagIsURL),
	"IsDocumentation":          uint64(FlagIsDocumentation),
	"IsDownload":               uint64(FlagIsDownload),
	"IsLabelize":               uint64(FlagIsLabelize),
	"LabelizeWithDescriptions": uint64(FlagLabelizeWithDescriptions),
	"IsLockable":               uint64(FlagIsLockable),
	"HeaderLockable":           uint64(FlagHeaderLockable),
	"DontJustifyHeader":        uint64(FlagDontJustifyHeader),
	"CheckerCell":              uint64(FlagCheckerCell),
	"FutureInvalid":            uint64(FlagFutureInvalid),
	"PastInvalid":              uint64(FlagPastInvalid),
	"HasDefault":               uint64(FlagHasDefault),
	"IsOpen":                   uint64(FlagIsOpen),
	"FlagIsOpener":             uint64(FlagIsOpener),
	"FlagDontLabelize":         uint64(FlagDontLabelize),
	"FlagOmitZero":             uint64(FlagOmitZero),
}

var (
	callSetupWidgeter func(f *Field)             // callSetupWidgeter is called to set gui widgets registered for use in zui tags. It is dependent on a gui, so injected with this func variable.
	fieldEnums        = map[string]zdict.Items{} // fieldEnums stores registered enums used by enum/radio tag
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

func (f Field) HasAllFlags(flags FlagType) bool {
	return f.Flags&flags != flags
}

func (f *Field) SetFlag(flag FlagType) {
	f.Flags |= flag
}

func (f *Field) ClearFlag(flag FlagType) {
	f.Flags &= ^flag
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
	finfo, found := zreflect.FieldForName(structure, FlattenIfAnonymousOrZUITag, name)
	if !found {
		finfo.FieldIndex = -1
	}
	return finfo.ReflectValue, finfo.FieldIndex
}

func (f *Field) SetFromRValAndStructField(rval reflect.Value, sf reflect.StructField, index int, params FieldParameters) bool {
	tagPart, _ := sf.Tag.Lookup("zui")
	return f.SetFromRVal(rval, tagPart, sf.Name, sf.PkgPath, index, params)
}

func (f *Field) SetFromRVal(rval reflect.Value, zuiTagPart string, sfName, sfPkg string, index int, params FieldParameters) bool {
	f.Index = index
	//	f.ID = fieldNameToID(sf.Name)
	fTypeName := rval.Type().Name()
	f.Kind = zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	f.FieldName = sfName
	f.Alignment = zgeo.AlignmentNone
	f.UpdateSecs = -1
	f.SortSmallFirst = zbool.Unknown
	f.SetEdited = true
	f.Vertical = zbool.Unknown
	f.PackageName = rval.Type().PkgPath()
	var skipping bool
	// zlog.Info("Packagename:", f.PackageName, f.FieldName)
	// zlog.Info("Field:", f.ID)

	var keyVals []zstr.KeyValue
	if zuiTagPart != "" {
		var skip bool
		keyVals, skip = zreflect.TagKeyValuesFromString(zuiTagPart)
		if skip {
			return false
		}
	}
	for _, kv := range keyVals {
		// key := kv.Key
		// val := kv.Value
		barParts := strings.Split(kv.Value, "|")
		if kv.Key == "IN" {
			skipping = !zstr.SlicesIntersect(params.UseInValues, barParts)
			continue
		}
		if skipping {
			continue
		}
		n, floatErr := strconv.ParseFloat(kv.Value, 32)
		flag := zbool.FromString(kv.Value, false)
		switch kv.Key {
		case "search":
			f.Flags |= FlagIsSearchable
		case "password":
			f.Flags |= FlagIsPassword
			if kv.Value == "existing" {
				f.Flags |= FlagIsFixed
			}
		case "setedited":
			f.SetEdited = flag
		case "format":
			f.Format = kv.Value
		case "vertical":
			f.Vertical = zbool.True
		case "horizontal":
			f.Vertical = zbool.False
		case "align":
			f.Alignment = zgeo.AlignmentFromString(kv.Value)
			// zlog.Info("ALIGN:", f.Name, kv.Value, a)
		// case "cannil"
		// f.Flags |= flagAllowNil
		case "celljustify", "justify":
			if kv.Value == "" {
				f.Justify = f.Alignment
			} else {
				f.Justify = zgeo.AlignmentFromString(kv.Value)
			}
			if kv.Key == "celljustify" {
				f.SetFlag(FlagDontJustifyHeader)
			}
		case "wrap":
			f.Wrap = kv.Value
			if kv.Value == "" {
				f.Wrap = "tailtrunc"
			}
		case "name":
			f.Name = kv.Value
		case "title":
			f.Title = kv.Value
		case "header":
			f.Header = kv.Value
		case "prefix":
			f.Prefix = kv.Value
		case "suffix":
			f.Suffix = kv.Value
		case "url":
			f.Path = kv.Value
			f.Flags |= FlagIsURL
		case "doc":
			f.Path = kv.Value
			f.Flags |= FlagIsDocumentation
		case "usein":
			f.UseIn = barParts
		case "rebuild":
			f.Flags |= FlagIsRebuildAllOnChange
		case "sep":
			f.StringSep = kv.Value
			if kv.Value == "" {
				f.StringSep = " "
			}
		case "mod":
			f.WhenMods.FromString(kv.Value)
		case "hlockable":
			f.SetFlag(FlagHeaderLockable)
			if kv.Value == "start" {
				f.SetFlag(FlagIsStart)
			} else if kv.Value == "end" {
				f.SetFlag(FlagIsEnd)
			}
		case "lockable":
			f.SetFlag(FlagIsLockable)
		case "filter":
			f.Filters = barParts
		case "trans":
			f.Transformer = kv.Value
		case "count":
			f.SetFlag(FlagShowSliceCount | FlagIsStatic)
		case "isuseinval":
			f.Flags |= FlagIsUseInValue
		case "popup":
			f.Flags |= FlagShowPopup
		case "color":
			f.Colors = barParts
			if len(f.Colors) == 1 {
				f.Styling.FGColor.SetFromString(f.Colors[0])
			}
		case "bgcolor":
			f.Styling.BGColor.SetFromString(kv.Value)
		case "download":
			f.Flags |= FlagIsDownload
			f.Path = kv.Value
		case "zrpc":
			f.RPCCall = kv.Value
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
		case "optional":
			f.Flags |= FlagShowIfExtraSpace
			if floatErr == nil && n > 0 {
				f.MinWidth = n
			}
		case "widget":
			f.WidgetName = kv.Value
		case "descending", "ascending":
			if kv.Key == "ascending" {
				f.SortSmallFirst = zbool.True
			} else {
				f.SortSmallFirst = zbool.False
			}
			for _, part := range barParts {
				if part == "bigzero" {
					f.Flags |= FlagZeroIsBig
				} else {
					f.SortPriority, _ = strconv.Atoi(part)
				}
			}
		case "actions":
			f.Flags |= FlagIsActions
		case "noautofill":
			f.Flags |= FlagDisableAutofill
		case "size":
			f.Size, _ = zgeo.SizeFromString(kv.Value)
			if f.Size.IsNull() {
				f.Size = zgeo.SizeBoth(n)
			}
		case "marg":
			var err error
			f.Margin, err = zgeo.SizeFromString(kv.Value)
			zlog.OnError(err, kv.Value)
		case "minwidth":
			if floatErr == nil {
				f.MinWidth = n
			}
		case "spacing":
			if floatErr == nil {
				f.Styling.Spacing = n
			}
		case "storekey":
			f.ValueStoreKey = kv.Value
		case "default":
			f.Default = kv.Value
			f.SetFlag(FlagHasDefault)
		case "allowempty":
			f.Flags |= FlagAllowEmptyAsZero
		case "omitzero":
			f.Flags |= FlagOmitZero
		case "required":
			f.Required = kv.Value
			if kv.Value == "" {
				f.Required = RequiredSingleValue
			}
		case "zerotext":
			f.ZeroText = kv.Value
		case "maxtext":
			f.MaxText = kv.Value
		case "invalid":
			switch kv.Value {
			case "past":
				f.SetFlag(FlagPastInvalid)
			case "future":
				f.SetFlag(FlagFutureInvalid)
			default:
				zlog.Error("invalid: bad val:", kv.Value)
			}
		case "open":
			f.Flags |= FlagIsOpen
		case "static":
			if flag || kv.Value == "" {
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
		case "press":
			f.Flags |= FlagPress
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
			for _, part := range strings.Split(kv.Value, "|") {
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
			for _, part := range strings.Split(kv.Value, "|") {
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
						f.Styling.DropShadow.Blur = num
						got = true
					}
				}
				if !got && !f.Styling.DropShadow.Color.Valid {
					f.Styling.DropShadow.Color.SetFromString(part)
				}
			}

		case "font":
			var sign int
			for _, part := range strings.Split(kv.Value, "|") {
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

		case "path":
			f.Path = kv.Value
		case "off":
			f.OffImagePath = "images/" + kv.Value
		case "opener":
			f.Flags |= FlagIsOpen | FlagIsOpener | FlagNoTitle | FlagIsImage
			f.MinWidth = 20
			f.MaxWidth = 20
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
			if kv.Key == "image" {
				f.Size = size
				f.Flags |= FlagIsImage
				f.ImageFixedPath = path
			} else {
				f.Flags |= FlagHasHeaderImage
				f.HeaderSize = size
				f.HeaderImageFixedPath = path
			}
		case "radio":
			_, got := fieldEnums[kv.Value]
			if !got {
				zlog.Error("no such radio:", kv.Value, fieldEnums, f.FieldName, zdebug.CallingStackString())
			}
			f.Radio = kv.Value
		case "enum":
			if zstr.HasPrefix(kv.Value, "./", &f.LocalEnum) {
			} else {
				_, got := fieldEnums[kv.Value]
				if !got {
					zlog.Error("no such enum:", kv.Value, fieldEnums, f.FieldName, zdebug.CallingStackString())
				}
				f.Enum = kv.Value
			}
		case "notitle":
			f.Flags |= FlagNoTitle
		case "tip":
			f.Tooltip = kv.Value
		case "desc":
			f.Description = kv.Value
		case "immediate":
			f.UpdateSecs = 0
		case "upsecs":
			if floatErr == nil && n > 0 {
				f.UpdateSecs = n
			}
		case "checker":
			f.SetFlag(FlagCheckerCell)
		case "2clip":
			f.Flags |= FlagToClipboard
		case "labelize":
			f.Flags |= FlagIsLabelize
			if kv.Value == "withdesc" {
				f.Flags |= FlagLabelizeWithDescriptions
			}
		case "unlabled":
			f.SetFlag(FlagDontLabelize)
		case "button":
			f.Flags |= FlagIsButton | FlagPress
		case "ask":
			f.Ask = kv.Value
		case "enable":
			f.LocalEnable = kv.Value
		case "disable":
			if kv.Value != "" {
				f.LocalDisable = kv.Value
			} else {
				f.Disabled = true // not used yet
			}
		case "show":
			if kv.Value == "" {
				f.Visible = true
			} else {
				f.LocalShow = kv.Value
			}
		case "hide":
			if kv.Value == "" {
				f.Visible = false
			} else {
				f.LocalHide = kv.Value
			}
		case "placeholder":
			if kv.Value != "" {
				f.Placeholder = kv.Value
			} else {
				f.Placeholder = "$HAS$" // set to this special value to set to name once set
			}
		case "dur":
			f.Flags |= FlagIsDuration
		case "since":
			f.Flags |= FlagIsStatic | FlagIsDuration
		default:
			f.CustomFields[kv.Key] = kv.Value
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
	// zfloat.Maximize(&f.MaxWidth, f.MinWidth)
	if f.MaxWidth != 0 {
		zfloat.Minimize(&f.MinWidth, f.MaxWidth)
	}
	name := zstr.PadCamelCase(sfName, " ")
	name = zlocale.FirstToTitleCaseExcept(name, "")
	if f.Name == "" {
		f.Name = name
	}
	if f.Title == "" {
		f.Title = name
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
			if sfPkg == "time" && fTypeName == "Duration" {
				if f.Flags&flagTimeFlags == 0 { // if no flags set, set default h,m,s
					f.Flags |= flagTimeFlags
				}
				setDurationColumns(f)
			}
			if f.Enum == "" && f.Radio == "" && f.LocalEnum == "" {
				if f.MinWidth == 0 {
					f.MinWidth = 40
				}
				// if f.MaxWidth == 0 {
				// 	f.MaxWidth = 80
				// }
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
		if f.MinWidth == 0 && f.Flags&FlagIsButton == 0 && f.Enum == "" && f.Radio == "" && f.LocalEnum == "" {
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
			f.Columns = len(f.Format) + 2 // 2 more for breathing room
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
				f.MinWidth = 100
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
	// zlog.Info("Field:", f.Name, f.Columns)
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

func SetEnumIntRange[N ~int](name string, from, to N) {
	var items zdict.Items
	for i := from; i <= to; i++ {
		var item zdict.Item
		item.Name = fmt.Sprint(i)
		item.Value = i
		items = append(items, item)
	}
	fieldEnums[name] = items
}

func SetEnumItems(name string, nameValPairs ...any) {
	var items zdict.Items

	for i := 0; i < len(nameValPairs); i += 2 {
		var di zdict.Item
		di.Name = fmt.Sprint(nameValPairs[i])
		di.Value = nameValPairs[i+1]
		items = append(items, di)
	}
	fieldEnums[name] = items
}

func AppendEnumItem(enum, key string, value any) {
	e := fieldEnums[enum]
	var item zdict.Item
	item.Name = key
	item.Value = value
	fieldEnums[enum] = append(e, item)
}

func SetStringBasedEnum(name string, vals ...string) {
	var items zdict.Items
	for _, s := range vals {
		item := zdict.Item{Name: s, Value: s}
		items = append(items, item)
	}
	fieldEnums[name] = items
}

func SetAnyToEnum[S any](name string, vals ...S) {
	var items zdict.Items
	for _, v := range vals {
		s := fmt.Sprint(v)
		item := zdict.Item{Name: s, Value: v}
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
		field := zreflect.FieldForIndex(fi, zreflect.FlattenIfAnonymous, f.Index)
		a := field.ReflectValue.Interface()
		di := enum.FindValue(a)
		if di != nil {
			m[a] = di.Name
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
	// fmt.Printf("Sort FieldMap: %+v %+v\n", fieldMap, sortOrder)
	val := reflect.ValueOf(slice)
	// zlog.Info("SORT:", sortOrder, enumTitles, val.Len())
	// var count int
	sort.SliceStable(slice, func(i, j int) bool {
		// count++
		ei := val.Index(i).Addr().Interface()
		ej := val.Index(j).Addr().Interface()
		for _, s := range sortOrder {
			f := fieldMap[s.FieldName]
			iitem := zreflect.FieldForIndex(ei, zreflect.FlattenIfAnonymous, f.Index).ReflectValue
			jitem := zreflect.FieldForIndex(ej, zreflect.FlattenIfAnonymous, f.Index).ReflectValue
			sliceEnumNames := enumTitles[f.FieldName]
			if sliceEnumNames != nil {
				ni := sliceEnumNames[iitem.Interface()]
				nj := sliceEnumNames[jitem.Interface()]
				r := (zstr.CaselessCompare(ni, nj) < 0) == s.SmallFirst
				if ni == nj {
					// zlog.Info("sliceEnumNames same:", r, s.FieldName, i, ni, j, nj)
					continue
				}
				// zlog.Info("sliceEnumNames:", r, s.FieldName, i, ni, j, nj)
				return r
			}
			switch zreflect.KindFromReflectKindAndType(iitem.Kind(), iitem.Type()) {
			case zreflect.KindBool:
				ia := iitem.Interface().(bool)
				ja := jitem.Interface().(bool)
				if ia == ja {
					continue
				}
				return (ia == false) == s.SmallFirst

			case zreflect.KindInt:
				ia, ierr := zint.GetAny(iitem.Interface())
				ja, jerr := zint.GetAny(jitem.Interface())
				zlog.Assert(ierr == nil && jerr == nil, ierr, jerr)
				if ia == ja {
					continue
				}
				return (ia < ja) == s.SmallFirst
			case zreflect.KindFloat:
				ia, ierr := zfloat.GetAny(iitem.Interface())
				ja, jerr := zfloat.GetAny(jitem.Interface())
				zlog.Assert(ierr == nil && jerr == nil, ierr, jerr)
				if ia == ja {
					continue
				}
				return (ia < ja) == s.SmallFirst
			case zreflect.KindString:
				ia, got := iitem.Interface().(string)
				ja, _ := jitem.Interface().(string)
				if !got {
					ia = fmt.Sprint(iitem.Interface())
					ja = fmt.Sprint(iitem.Interface())
				}
				if ia == ja {
					continue
				}
				// zlog.Info("sort:", i, j, ia, "<", ja, "less:", zstr.CaselessCompare(ia, ja) < 0, s.SmallFirst)
				return (zstr.CaselessCompare(ia, ja) < 0) == s.SmallFirst
			case zreflect.KindTime:
				ia := iitem.Interface().(time.Time)
				ja := jitem.Interface().(time.Time)
				if ia == ja {
					continue
				}
				if f.HasFlag(FlagZeroIsBig) {
					if ia.IsZero() {
						ia = ztime.BigTime
					}
					if ja.IsZero() {
						ja = ztime.BigTime
					}
				}
				return ia.Before(ja) == s.SmallFirst
			default:
				continue
			}
		}
		// zlog.Fatal("No sort fields set for struct")
		return false
	})
	// zlog.Info("SORT TIME:", time.Since(start), count)
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

func FlattenIfAnonymousOrZUITag(f reflect.StructField) bool {
	if f.Anonymous {
		return true
	}
	keyVals, skip := zreflect.TagKeyValuesForKeyInStructField(&f, "zui")
	if len(keyVals) == 0 || skip {
		return false
	}
	_, i := zstr.KeyValuesFindForKey(keyVals, "flatten")
	if i != -1 {
		return true
	}
	return false
}

type FieldInfo struct {
	zreflect.FieldInfo
	Field *Field
}

func ForEachField(structure any, params FieldParameters, fields []Field, got func(each FieldInfo) bool) {
	if len(fields) == 0 {
		zreflect.ForEachField(structure, FlattenIfAnonymousOrZUITag, func(each zreflect.FieldInfo) bool {
			f := EmptyField
			if !f.SetFromRValAndStructField(each.ReflectValue, each.StructField, each.FieldIndex, params) {
				return true
			}
			// zlog.Info("UseIn:", f.Name, f.UseIn, params.UseInValues)
			fields = append(fields, f)
			return true
		})
	}
	zreflect.ForEachField(structure, FlattenIfAnonymousOrZUITag, func(each zreflect.FieldInfo) bool {
		f := findFieldWithIndex(&fields, each.FieldIndex)
		if f == nil {
			return true
		}
		if f.HasFlag(FlagIsUseInValue) {
			zstr.AddToSet(&params.UseInValues, fmt.Sprint(each.ReflectValue.Interface()))
		}
		return true
	})
	zreflect.ForEachField(structure, FlattenIfAnonymousOrZUITag, func(each zreflect.FieldInfo) bool {
		if zstr.IndexOf(each.StructField.Name, params.SkipFieldNames) != -1 {
			return true
		}
		f := findFieldWithIndex(&fields, each.FieldIndex)
		if f == nil {
			return true
		}
		if f.HasFlag(FlagIsForZDebugOnly) && !zui.DebugOwnerMode {
			return true
		}
		usePlain, useDollar := zslice.SplitWithFunc(f.UseIn, func(s string) bool {
			return strings.HasPrefix(s, "$")
		})
		hasPlain, hasDollar := zslice.SplitWithFunc(params.UseInValues, func(s string) bool {
			return strings.HasPrefix(s, "$")
		})
		if len(usePlain) != 0 && !zstr.SlicesIntersect(usePlain, hasPlain) {
			return true
		}
		if len(useDollar) != 0 && !zstr.SlicesIntersect(useDollar, hasDollar) {
			return true
		}
		var finfo FieldInfo
		finfo.FieldInfo = each
		finfo.Field = f
		return got(finfo)
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
	ForEachField(structPtr, FieldParameters{}, nil, func(each FieldInfo) bool {
		vals, _ := zreflect.TagValuesForKey(each.StructField.Tag, "zui")
		for _, part := range vals {
			if part == "indicator" {
				rval = each.ReflectValue
				field = each.Field
				got = true
			}
		}
		return true
	})
	return
}

func (f *Field) IsImageToggle() bool {
	return f.Flags&FlagIsImage != 0 && f.OffImagePath != ""
}

func getField(val reflect.Value, indent, desc string) string {
	var dstr string
	if desc != "" {
		dstr = " // " + desc
	}
	kind := zreflect.KindFromReflectKindAndType(val.Kind(), val.Type())
	switch kind {
	case zreflect.KindInt, zreflect.KindFloat, zreflect.KindBool, zreflect.KindString:
		return "<" + string(kind) + ">" + dstr + "\n"
	case zreflect.KindStruct:
		str := "{" + dstr + "\n"
		str += OutputJSONStructDescription(val.Interface(), indent+"  ")
		str += indent + "}\n"
		return str
	case zreflect.KindSlice:
		e := zslice.MakeAnElementOfSliceRValType(val)
		sliceKind := zreflect.KindFromReflectKindAndType(e.Kind(), e.Type())
		if sliceKind != zreflect.KindStruct {
			return "[ <" + string(sliceKind) + "*> ]" + dstr + "\n"
		}
		str := "[" + dstr + "\n"
		str += OutputJSONStructDescription(e.Interface(), indent+"  ")
		str += indent + "]\n"
		return str
	default:
		return "[unknown]" + dstr + "\n"
	}
}

// OutputJSONStructDescription outputs a json encoding of s, but with descriptions etc from zui tags
func OutputJSONStructDescription(s any, indent string) string {
	var str string
	str += indent + "struct {\n"
	zreflect.ForEachField(s, FlattenIfAnonymousOrZUITag, func(each zreflect.FieldInfo) bool {
		zuiKV, skip := zreflect.TagKeyValuesForKeyInStructField(&each.StructField, "zui")
		if skip {
			return true
		}
		fn := each.StructField.Name
		jsonVals, skip := zreflect.TagValuesForKey(each.StructField.Tag, "json")
		var desc string
		if !skip {
			if len(jsonVals) > 0 {
				fn = jsonVals[0]
			}
			val, got := zstr.KeyValuesValueForKey(zuiKV, "desc")
			if got {
				desc = val
			}
		}
		str += indent + `"` + fn + `": `
		str += getField(each.ReflectValue, indent+"  ", desc)
		return true
	})
	str += indent + "}\n"
	return str
}

func (f *Field) JoinSeparatedSlice(rval reflect.Value) string {
	var parts []string
	for i := 0; i < rval.Len(); i++ {
		v := rval.Index(i).Interface()
		parts = append(parts, fmt.Sprint(v))
	}
	return strings.Join(parts, f.StringSep)
}
