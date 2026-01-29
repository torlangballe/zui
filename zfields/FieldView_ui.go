//go:build zui

package zfields

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zbutton"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcursor"
	"github.com/torlangballe/zui/zdocs"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zpresent"
	"github.com/torlangballe/zui/zradio"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/ztext"
	"github.com/torlangballe/zui/ztextinfo"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdebug"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zguiutil"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmap"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
	"github.com/torlangballe/zutil/zwords"
)

type FieldView struct {
	zcontainer.StackView
	ID             string
	ParentFV       *FieldView
	Fields         []Field
	data           interface{}
	dataHash       int64
	params         FieldViewParameters
	sliceItemIndex int
	lastCheckered  bool // toggled during build for each column
	frameTitle     string
}

type FieldViewParameters struct {
	Field
	FieldParameters
	BuildChildrenHidden       bool
	ImmediateEdit             bool                                 // ImmediateEdit forces immediate write-to-data when editing a field.
	MultiSliceEditInProgress  bool                                 // MultiSliceEditInProgress is on if the field represents editing multiple structs in a list. Checkboxes can be indeterminate etc.
	EditWithoutCallbacks      bool                                 // Set so not get edit/changed callbacks when editing. Example: Dialog box editing edits a copy, so no callbacks needed.
	IsEditOnNewStruct         bool                                 // IsEditOnNewStruct when an just-created struct is being edited. Menus can have a storage key-value to set last-used option then for example
	triggerHandlers           map[trigger]func(ap ActionPack) bool // triggerHandlers is a map of functions to call if an action occurs in this FieldView. Somewhat replacing ActionHandler
	CreateActionMenuItemsFunc func(sid string) []zmenu.MenuedOItem // If a field with FlagIsActions is found, this func is called to create the action items.
}

type ActionPack struct {
	FieldView *FieldView
	Field     *Field
	Action    ActionType
	RVal      reflect.Value
	View      *zview.View
}

// If a structure/slice used in FieldViews has this method, it is called when edited/changed etc.
// ap.Field may be nil if it's for the entire structure.
type ActionHandler interface {
	HandleAction(ap ActionPack) bool
}

// If a struct implements StructInitializer, it is initialized when a element is added to a slice in FieldSliceView
// Note: the method must be on a pointer, not value, as it changes the contents of the struct.
// TODO: Make default Add element in SliceGridView that uses this.
type StructInitializer interface {
	InitZFieldStruct()
}

type FieldViewOwner interface {
	GetFieldView() *FieldView
}

var (
	fieldViewEdited         = map[string]time.Time{}
	textFilters             = map[string]func(string) string{}
	EnableLog               zlog.Enabler
	enumEditHandlers        = map[string]func(item *zmenu.MenuedOItem, action zmenu.EditAction){}
	fieldToTextTransformers = map[string]func(val, structure any, label *zlabel.Label){}
)

var DefaultFieldViewParameters = FieldViewParameters{
	ImmediateEdit: true,
	Field: Field{
		Styling: zstyle.EmptyStyling,
	},
}

func init() {
	var dl DocumentationLink
	zreflect.DefaultTypeRegistrar.Register(dl, nil)
	zlog.RegisterEnabler("zfields.LogGUI", &EnableLog)

	RegisterTextFilter("$nowhite", func(s string) string {
		return zstr.WhitespaceRemover.Replace(s)
	})
	RegisterTextFilter("$trim", strings.TrimSpace)
	RegisterTextFilter("$lower", strings.ToLower)
	RegisterTextFilter("$upper", strings.ToLower)
	RegisterTextFilter("$uuid", zstr.CreateFilterFunc(zstr.IsRuneValidInUUID))
	RegisterTextFilter("$hex", zstr.CreateFilterFunc(zstr.IsRuneHex))
	RegisterTextFilter("$alpha", zstr.CreateFilterFunc(zstr.IsRuneASCIIAlpha))
	RegisterTextFilter("$num", zstr.CreateFilterFunc(zstr.IsRuneASCIINumeric))
	RegisterTextFilter("$alphanum", zstr.CreateFilterFunc(zstr.IsRuneASCIIAlphaNumeric))
	RegisterTextFilter("$headerkey", zstr.CreateFilterFunc(zhttp.IsRuneValidForHeaderKey))
	RegisterTextFilter("$ascii", zstr.CreateFilterFunc(zstr.IsRuneASCIIPrintable))

}

func RegisterFieldTransformer(name string, trans func(val, structure any, label *zlabel.Label)) {
	fieldToTextTransformers[name] = trans
}

func RegisterFieldTransformerText(name string, trans func(val, structure any) string) {
	fieldToTextTransformers[name] = func(val, s any, label *zlabel.Label) {
		out := trans(val, s)
		label.SetText(out)
	}
}

func RegisterTextFilter(name string, filter func(string) string) {
	textFilters[name] = filter
}

func GetTextFilter(name string) func(string) string {
	return textFilters[name]
}

func RegisterEnumEditHandler(enumName string, handler func(item *zmenu.MenuedOItem, action zmenu.EditAction)) {
	enumEditHandlers[enumName] = handler
}

func CallStructInitializer(a any) {
	i, _ := a.(StructInitializer)
	if i != nil {
		i.InitZFieldStruct()
	}
}

func (v *FieldView) Data() any {
	return v.data
}

func (v *FieldView) isRows() bool {
	return zstr.StringsContain(v.params.UseInValues, RowUseInSpecialName)
}

func (v *FieldView) IsSlice() bool {
	rval := reflect.ValueOf(v.data)
	if rval.Kind() == reflect.Map {
		return false
	}
	return rval.Elem().Kind() == reflect.Slice
}

func setFieldViewEdited(fv *FieldView) {
	fieldViewEdited[fv.Hierarchy()] = time.Now()
}

func (fv *FieldView) IsEditedRecently() bool {
	h := fv.Hierarchy()
	t, got := fieldViewEdited[h]
	if got {
		if time.Since(t) < time.Second*3 {
			return true
		}
		delete(fieldViewEdited, h)
	}
	return false
}

func (fv *FieldView) ClearEditedRecently() {
	h := fv.Hierarchy()
	delete(fieldViewEdited, h)
}

func maybeAddDocToHeader(f *Field, header *zcontainer.StackView) {
	if header != nil && f.HasFlag(FlagIsDocumentation) && f.Path != "" {
		help := zwidgets.DocumentationIconViewNew(f.Path)
		header.Add(help, zgeo.CenterRight)
	}
}

func makeFrameIfFlag(f *Field, fv *FieldView, overrideTitle string) (view zview.View, header *zcontainer.StackView) {
	if !f.HasFlag(FlagHasFrame) {
		return nil, nil
	}
	title := overrideTitle
	if title == "" && f.HasFlag(FlagFrameIsTitled) {
		title = f.TitleOrName()
	}
	fv.frameTitle = title
	frame := zcontainer.StackViewVert("frame")
	header = zguiutil.MakeStackATitledFrame(frame, title, f.Flags&FlagFrameTitledOnFrame != 0, f.Styling, f.Styling)
	maybeAddDocToHeader(f, header)
	frame.Add(fv, zgeo.TopLeft|zgeo.Expand)
	return frame, header
}

func fieldViewNew(id string, vertical bool, data any, params FieldViewParameters, marg zgeo.Size, parent *FieldView) *FieldView {
	v := &FieldView{}
	v.StackView.Init(v, vertical, id)
	v.SetChildrenAboveParent(true)

	v.data = data
	v.sliceItemIndex = -1
	zreflect.ForEachField(v.data, FlattenIfAnonymousOrZUITag, func(each zreflect.FieldInfo) bool {
		// zlog.Info("FVNew:", each.StructField)
		f := EmptyField

		if !f.SetFromRValAndStructField(each.ReflectValue, each.StructField, each.FieldIndex, params.FieldParameters) {
			return true
		}
		if params.ImmediateEdit {
			f.UpdateSecs = 0
		}
		callActionHandlerFunc(ActionPack{FieldView: v, Field: &f, Action: SetupFieldAction, RVal: each.ReflectValue.Addr(), View: nil})
		v.Fields = append(v.Fields, f)
		return true
	})
	if params.Field.Flags&FlagHasFrame == 0 {
		v.SetMargin(zgeo.RectFromMinMax(marg.Pos(), marg.Pos().Negative()))
	}
	v.params = params
	// zlog.Info("fieldViewNew:", v.ObjectName(), params.Use)

	v.ID = id
	v.ParentFV = parent

	return v
}

func (v *FieldView) Parameters() FieldViewParameters {
	return v.params
}

func (v *FieldView) Build(update bool) {
	a := zgeo.Left
	if v.Vertical {
		a |= zgeo.Top
	} else {
		a |= zgeo.VertCenter
	}
	v.BuildStack(v.ObjectName(), a, zgeo.SizeNull, true)
	if update {
		dontOverwriteEdited := false
		v.Update(nil, dontOverwriteEdited, false)
	}
}

func (v *FieldView) FindNamedViewOrInLabelized(name string) (view, label, labelStack zview.View) {
	for _, c := range (v.View.(zcontainer.ChildrenOwner)).GetChildren(false) {
		n := c.ObjectName()
		if n == name {
			return c, nil, nil
		}
		if strings.HasPrefix(n, "$labelize.") {
			s, _ := c.(*zcontainer.StackView)
			if s != nil {
				view, _ = s.FindViewWithName(name, false)
				if view != nil {
					label, _ = s.FindViewWithName("$labelize.label."+name, false)
					return view, label, s
				}
			}
		}
	}
	return nil, nil, nil
}

func (v *FieldView) updateShowEnableFromZeroer(isZero, isShow bool, toID string) {
	if v.params.MultiSliceEditInProgress {
		return
	}
	for _, f := range v.Fields {
		var id string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		if zstr.HasPrefix(local, "./", &id) && id == toID {
			// _, foundView := v.FindNamedViewOrInLabelized(f.FieldName)
			foundView, _, _ := v.FindNamedViewOrInLabelized(f.FieldName)
			if foundView == nil {
				continue
			}
			zero := isZero
			if neg {
				zero = !zero
			}
			if isShow {
				foundView.Show(!zero)
			} else {
				foundView.SetUsable(!zero)
			}
			continue
		}
	}
	//TODO: handle ../ and substruct/id style
}

func doShowEnableItem(rval reflect.Value, isShow bool, view zview.View, not bool) {
	zero := rval.IsZero()
	if not {
		zero = !zero
	}
	// zlog.Info("LocalDisable", view.Native().Hierarchy(), rval, isShow, zero)
	if isShow {
		view.Show(!zero)
	} else {
		view.SetUsable(!zero)
	}
}

func getLocalFromShowOrEnable(isShow bool, f *Field) (local string, neg bool) {
	if isShow {
		if f.LocalShow != "" {
			local = f.LocalShow
		} else {
			local = f.LocalHide
			neg = true
		}
	} else {
		if f.LocalEnable != "" {
			local = f.LocalEnable
		} else {
			local = f.LocalDisable
			neg = true
		}
	}
	return
}

func (v *FieldView) updateShowEnableOnView(view zview.View, isShow bool, toFieldName string) {
	if v.params.MultiSliceEditInProgress {
		return
	}
	for _, f := range v.Fields {
		if f.FieldName != toFieldName {
			continue
		}
		var prefix, fname string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		if zstr.HasPrefix(local, "./", &fname) {
			finfo, found := zreflect.FieldForName(v.data, FlattenIfAnonymousOrZUITag, fname)
			if found {
				doShowEnableItem(finfo.ReflectValue, isShow, view, neg)
			}
			continue
		}
		if zstr.SplitN(local, "/", &prefix, &fname) && prefix == f.FieldName {
			fv := viewToFieldView(view)
			if fv == nil {
				zlog.Error("updateShowOrEnable: not field view:", f.FieldName, local, v.ObjectName)
				return
			}
			finfo, found := zreflect.FieldForName(v.data, FlattenIfAnonymousOrZUITag, fname)
			if found {
				doShowEnableItem(finfo.ReflectValue, isShow, view, neg)
			}
			continue
		}
		if zstr.HasPrefix(local, "../", &fname) && v.ParentFV != nil {
			finfo, found := zreflect.FieldForName(v.ParentFV.Data(), FlattenIfAnonymousOrZUITag, fname)
			if found {
				doShowEnableItem(finfo.ReflectValue, isShow, view, neg)
			}
			continue
		}
	}
}

func (v *FieldView) Update(data any, dontOverwriteEdited, forceUpdateOnFieldSlice bool) {
	if data != nil { // must be after fv.IsEditedRecently, or we set new data without update slice pointers and maybe more????
		v.data = data
	}
	// zlog.Info("FV.Update:", v.Hierarchy(), zlog.Full(v.data))
	recentEdit := (dontOverwriteEdited && v.IsEditedRecently())
	fh, _ := v.data.(ActionHandler)
	sview := v.View
	if fh != nil {
		fh.HandleAction(ActionPack{FieldView: v, Action: DataChangedActionPre, View: &sview})
	}
	ForEachField(v.data, v.params.FieldParameters, v.Fields, func(each FieldInfo) bool {
		if !recentEdit || each.Field.IsStatic() || each.StructField.Type.Kind() == reflect.Slice {
			v.updateField(each.FieldIndex, each.ReflectValue, each.StructField, dontOverwriteEdited, forceUpdateOnFieldSlice)
		}
		return true
	})
	// call general one with no id. Needs to be after above loop, so values set
	if fh != nil {
		fh.HandleAction(ActionPack{FieldView: v, Action: DataChangedAction, View: &sview})
	}
}

func (v *FieldView) updateField(index int, rval reflect.Value, sf reflect.StructField, dontOverwriteEdited, forceUpdateOnFieldSlice bool) bool {
	// zlog.Info("updateField:", v.Hierarchy(), sf.Name)
	var valStr string
	f := findFieldWithIndex(&v.Fields, index)
	if f == nil {
		return true
	}
	foundView, _, _ := v.FindNamedViewOrInLabelized(f.FieldName)
	if foundView == nil {
		return true
	}
	v.updateShowEnableOnView(foundView, true, foundView.ObjectName())
	v.updateShowEnableOnView(foundView, false, foundView.ObjectName())
	var called bool
	tri, _ := rval.Interface().(TriggerDataChangedTriggerer)
	if tri != nil {
		called = tri.HandleDataChange(v, f, rval.Addr().Interface(), &foundView)
	}
	if !called {
		// zlog.Info("Call trigger DataChangedAction:", v.Hierarchy(), f.Name, rval)
		ap := ActionPack{Field: f, Action: DataChangedAction, RVal: rval.Addr(), View: &foundView}
		called = v.callTriggerHandler(ap)
	}
	if !called {
		defer callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: DataChangedAction, RVal: rval.Addr(), View: &foundView})
	}
	if called {
		return true
	}
	if f.Transformer != "" && (f.IsStatic() || v.params.AllStatic) {
		label, _ := foundView.(*zlabel.Label)
		if label != nil {
			trans, got := fieldToTextTransformers[f.Transformer]
			if got {
				trans(rval.Interface(), v.data, label)
				return true
			}
		}
	}

	// zlog.Info("updateField3:", v.Hierarchy(), sf.Name)
	if f.WidgetName != "" && f.Kind != zreflect.KindSlice {
		w := widgeters[f.WidgetName]
		if w != nil {
			setter, _ := foundView.(zview.AnyValueSetter)
			if setter != nil {
				setter.SetValueWithAny(rval.Interface())
			}
			return true
		}
	}
	if f.StringSep == "" && f.Enum != "" && (f.IsStatic() || v.params.AllStatic) {
		enum := GetEnum(f.Enum)
		str := findNameOfEnumForRVal(rval, enum)
		str = f.Prefix + str + f.Suffix
		foundView.(*zlabel.Label).SetText(str)
		return true
	}

	menuType, _ := foundView.(zmenu.MenuType)
	if menuType == nil {
		o := zmenu.OwnerForView(foundView)
		if o != nil {
			menuType = o
		}
	}
	_, isGetter := rval.Interface().(zdict.ItemsGetter)
	if menuType != nil && (isGetter || f.Enum != "" || f.LocalEnum != "") { // && f.Kind != zreflect.KindSlice
		var enum zdict.Items
		if f.Enum != "" {
			enum, _ = fieldEnums[f.Enum]
			zslice.CopyTo(&enum, enum) // we make a copy of enum, or else global one is messed up
			// zlog.Info("updateMenu2:", v.Hierarchy(), sf.Name, enum)
		} else if f.LocalEnum != "" {
			ei, findex := FindLocalFieldWithFieldName(v.data, f.LocalEnum)
			zlog.Assert(findex != -1, f.Name, f.LocalEnum)
			var err error
			enum, err = getDictItemsFromSlice(ei, f)
			if err != nil {
				return false
			}
		} else {
			var err error
			enum, err = getDictItemsFromSlice(rval, f)
			if err != nil {
				return false
			}
		}
		if v.params.MultiSliceEditInProgress {
			var hasZero bool
			for _, item := range enum { // find if items has a zero item already:
				if item.Value == nil {
					hasZero = true
					break
				}
				rval := reflect.ValueOf(item)
				if rval.IsZero() {
					hasZero = true
					break
				}
			}
			// if no zero item, add one:
			if !hasZero {
				var item zdict.Item
				enum = append(zdict.Items{item}, enum...)
			}
		}
		menuType.UpdateItems(enum, rval.Interface(), f.Flags&FlagIsActions != 0)
		return true
	}
	updateToolTip(f, v.data, foundView)
	to, _ := foundView.(ztext.TextOwner)
	zuistringer, _ := rval.Interface().(UIStringer)
	if to != nil && zuistringer != nil {
		// if f.IsStatic() || v.params.AllStatic {
		to.SetText(zuistringer.ZUIString(f.HasFlag(FlagAllowEmptyAsZero)))
		return true
	}
	switch f.Kind {
	case zreflect.KindMap:
		if v.setCountString(f, foundView, rval) {
			return true
		} else {
			v.updateMapList(f, rval, foundView)
		}

	case zreflect.KindSlice:
		if f.StringSep != "" {
			v.updateSeparatedStringWithSlice(f, rval, foundView)
			break
		}
		if v.setCountString(f, foundView, rval) {
			return true
		}
		// zlog.Info("updateFieldSlice:", v.Hierarchy(), sf.Name)
		sv, _ := foundView.(*FieldSliceView)
		if sv == nil {
			return false
		}
		hash := zreflect.HashAnyToInt64(rval.Interface(), "")
		sameHash := (sv.dataHash == hash)
		// zlog.Info("FV.Update slice:", f.Name, v.Hierarchy(), zlog.CallingStackString())
		sv.dataHash = hash
		if !sameHash || forceUpdateOnFieldSlice {
			sv.UpdateSlice(f, rval.Addr().Interface())
		}

	case zreflect.KindTime:
		tv, _ := foundView.(*ztext.TextView)
		if tv != nil && tv.IsEditing() {
			break
		}
		if f.Flags&FlagIsDuration != 0 {
			v.updateSinceTime(foundView.(*zlabel.Label), f)
			break
		}
		valStr = getTimeString(rval, f)
		t := rval.Interface().(time.Time)
		if !t.IsZero() && f.HasFlag(FlagPastInvalid|FlagFutureInvalid) {
			since := time.Since(t)
			if f.HasFlag(FlagPastInvalid) && since > 0 || f.HasFlag(FlagFutureInvalid) && since < 0 {
				// zlog.Info("INVALID:", valStr, f.FieldName, t, since, f.HasFlag(FlagPastInvalid), f.HasFlag(FlagFutureInvalid))
				foundView.SetColor(zgeo.ColorRed)
			} else {
				foundView.SetColor(zstyle.DefaultFGColor())
			}
		}
		to := foundView.(ztext.TextOwner)
		str := f.Prefix + valStr + f.Suffix
		to.SetText(str)

	case zreflect.KindStruct:
		fv := viewToFieldView(foundView)
		if fv == nil {
			fv = findSubFieldView(foundView, "")
		}
		if fv == nil {
			break
		}
		fv.Update(rval.Addr().Interface(), dontOverwriteEdited, forceUpdateOnFieldSlice)

	case zreflect.KindBool:
		if f.Flags&FlagIsImage != 0 && f.IsImageToggle() && rval.Kind() == reflect.Bool {
			iv, _ := foundView.(*zimageview.ImageView) // it might be a button or something instead
			path := f.OffImagePath
			val, _ := v.fieldToDataItem(f, iv)
			on := val.Bool()
			if on {
				path = f.ImageFixedPath
			}
			iv.SetImage(nil, path, nil)
			break
		}
		rv, _ := foundView.(*zradio.RadioButton) // it might be a button or something instead
		if rv != nil {
			rval.SetBool(rv.Value())
			return true
		}
		cv, _ := foundView.(*zcheckbox.CheckBox) // it might be a button or something instead
		if cv != nil {
			b := zbool.ToBoolInd(rval.Interface().(bool))
			v := cv.Value()
			if v != b {
				cv.SetValue(b)
			}
		}

	case zreflect.KindInt, zreflect.KindFloat:
		var dur time.Duration
		var tip string
		valStr, tip, dur = getTextFromNumberishItem(rval, f)
		v.setText(f, valStr, foundView)
		label, _ := foundView.(*zlabel.Label)
		if label != nil {
			updateOldDurationColor(label, dur, f)
			if tip != "" {
				label.SetToolTip(tip)
			}
		}

	case zreflect.KindString, zreflect.KindFunc: // why KindFunc???
		valStr = rval.String()
		if f.Flags&FlagIsImage != 0 {
			path := ""
			if f.Kind == zreflect.KindString {
				path = valStr
			}
			if path != "" && strings.Contains(f.ImageFixedPath, "*") {
				path = strings.Replace(f.ImageFixedPath, "*", path, 1)
			} else if path == "" || f.Flags&FlagIsFixed != 0 {
				path = f.ImageFixedPath
			}
			io := foundView.(zimage.Owner)
			io.SetImage(nil, path, nil)
		} else {
			if f.IsStatic() || v.params.AllStatic {
				if f.Flags&FlagIsFixed != 0 {
					valStr = f.Name
				}
				// valStr = f.Prefix + valStr + f.Suffix // done in v.setString()
			}
			v.setText(f, valStr, foundView)
		}
	}
	return true
}

func (v *FieldView) setCountString(f *Field, foundView zview.View, rval reflect.Value) bool {
	if f.HasFlag(FlagShowSliceCount) {
		label, _ := foundView.(*zlabel.Label)
		str := strconv.Itoa(rval.Len())
		if str == "0" && f.HasFlag(FlagAllowEmptyAsZero) {
			str = f.ZeroText
		}
		label.SetText(str)
		v.maybeMakeLabelHandleFromClipboard(f, label, str, rval)
		return true
	}
	return false
}

func makeMapTextView(fv *FieldView, stackFV *FieldView, f *Field, str, name string) *ztext.TextView {
	var style ztext.Style
	tv := ztext.NewView(str, style, 20, 1)
	tv.SetObjectName(name)
	// f.SetFont(tv, nil)
	tv.SetPlaceholder(name)
	tv.UpdateSecs = 1.5
	tv.SetValueHandler("", func(edited bool) {
		if edited {
			updateMap(fv, stackFV, f)
		}
	})
	return tv
}

var inMapRows int

func buildMapRow(parent, stackFV *FieldView, i int, key string, mval reflect.Value, fixed bool, f *Field) (zview.View, Field) { // v *FieldView,
	inMapRows++
	defer func() {
		inMapRows--
	}()
	var mf Field
	if strings.Contains(key, ":") {
		n, fname, typeName, tags, err := zreflect.ValueFromTypeFormatSuffixedName(key, mval.Interface())
		if !zlog.OnError(err, key, mval.Interface()) && n != nil {
			mval = reflect.ValueOf(n)
			var pkg, field string
			zstr.SplitN(typeName, ".", &pkg, &field)
			mf.SetFromRVal(mval, tags, field, pkg, 0, FieldParameters{})
			key = fname
		}
	}
	if mf.Format == "" {
		lkey := strings.ToLower(key)
		for _, suf := range []string{"bps", "bits/s", "b/s", "bits/sec"} {
			if strings.HasSuffix(lkey, suf) {
				mf.Format = HumanFormat
				break
			}
		}
	}
	mf.Name = zlocale.FirstToTitleCaseExcept(key, "")
	mf.FieldName = key
	if mf.Alignment == zgeo.AlignmentNone {
		mf.Alignment = zgeo.CenterLeft | zgeo.HorExpand
	}
	if f.HasFlag(FlagToClipboard) {
		mf.SetFlag(FlagToClipboard)
	}
	if f.HasFlag(FlagIsFixed) {
		mf.SetFlag(FlagIsFixed)
	}
	mf.MaxWidth = 600
	if fixed {
		if f.IsStatic() {
			mf.SetFlag(FlagIsStatic)
		}
		mf.SetFlag(FlagIsLabelize)
		a := zgeo.CenterLeft
		text := fmt.Sprint(mval.Interface())
		lineFeeds := strings.Count(text, "\n")
		if lineFeeds > 0 {
			mf.Rows = min(4, lineFeeds+1)
		} else {
			// mf.Wrap = ztextinfo.WrapTailTruncate.String()
		}
		if mval.Kind() == reflect.Interface {
			mval = mval.Elem()
		}
		if zhttp.StringStartsWithHTTPX(mval.String()) {
			mf.SetFlag(FlagIsURL)
		}
		// zlog.Info(f.Name, "map buildItem:", mf.Name, mval.Type())
		mf.Vertical = zbool.False
		view := stackFV.buildItem(&mf, mval, i, a, zgeo.Size{}, true)
		if f.IsStatic() {
			view.Native().SetUsable(true)
			setter, _ := view.(zview.InteractiveSetter)
			label, isLabel := view.(*zlabel.Label)
			_, isStack := view.(zcontainer.ChildrenOwner)
			if setter != nil && !isLabel && !isStack {
				setter.SetInteractive(false)
			}
			if label != nil {
				label.SetMaxLines(0)
			}
		}
		return view, mf
	}
	hor := zcontainer.StackViewHor("map-row")
	hor.SetSpacing(14)
	hor.SetMargin(zgeo.RectFromXY2(4, 0, 0, 0))
	stackFV.Add(hor, zgeo.CenterLeft|zgeo.HorExpand)

	vkey := makeMapTextView(parent, stackFV, f, key, "key")
	if len(f.Filters) >= 2 {
		vkey.FilterFunc = getFilterFuncFromFilterNames(f.Filters[:1], f)
	}
	hor.Add(vkey, zgeo.CenterLeft, zgeo.SizeD(0, 0))

	vval := makeMapTextView(parent, stackFV, f, fmt.Sprint(mval.Interface()), "value")

	if len(f.Filters) != 0 {
		i := 0
		if len(f.Filters) > 1 {
			i = 1
		}
		vval.FilterFunc = getFilterFuncFromFilterNames(f.Filters[i:], f)
	}
	hor.Add(vval, zgeo.CenterRight|zgeo.HorExpand, zgeo.SizeD(8, 0))
	return hor, mf
}

func updateMap(fv *FieldView, stackFV *FieldView, f *Field) {
	data := fv.parentsDataForMe()
	finfo := zreflect.FieldForIndex(data, FlattenIfAnonymousOrZUITag, f.Index)
	rval := finfo.ReflectValue
	if rval.IsNil() {
		m := reflect.MakeMap(rval.Type())
		rval.Set(m)
	} else {
		for _, rkey := range rval.MapKeys() {
			rval.SetMapIndex(rkey, reflect.Value{})
		}
	}
	for _, row := range stackFV.GetChildren(false) {
		vkey, _ := zcontainer.ContainerOwnerFindViewWithName(row, "key", false)
		key := vkey.(*ztext.TextView).Text()
		vvalue, _ := zcontainer.ContainerOwnerFindViewWithName(row, "value", false)
		zlog.Assert(vkey != nil && vvalue != nil, vkey != nil, vvalue != nil)
		value := vvalue.(*ztext.TextView).Text()
		if key == "" && value == "" {
			continue
		}
		valType := rval.Type().Elem()
		e := reflect.New(valType)
		zstr.SetStringToAny(e.Interface(), value)
		rval.SetMapIndex(reflect.ValueOf(key), e.Elem())
		zlog.Info("map add:", rval.Interface())
	}
	var view zview.View = stackFV
	ap := ActionPack{Field: f, Action: EditedAction, RVal: rval, View: &view, FieldView: stackFV.ParentFV}
	stackFV.callTriggerHandler(ap)
	callActionHandlerFunc(ap)
}

func (v *FieldView) BuildMapList(rval reflect.Value, f *Field, frameTitle string) zview.View {
	return BuildMapList(rval, f, frameTitle, v.params, v.Spacing(), v)
}

func BuildMapList(rval reflect.Value, f *Field, frameTitle string, params FieldViewParameters, spacing float64, parent *FieldView) zview.View {
	var outView zview.View
	// params := v.params
	params.triggerHandlers = zmap.EmptyOf(params.triggerHandlers)
	if f.IsStatic() {
		params.AllStatic = true
	}
	stackFV := fieldViewNew(f.FieldName+".FV", true, rval.Interface(), params, zgeo.Size{}, parent)
	outView = stackFV
	stackFV.GridVerticalSpace = math.Max(6, spacing)
	stackFV.SetSpacing(f.Styling.SpacingOrMax(12))
	stackFV.params.SetFlag(FlagIsLabelize)
	fixed := f.HasFlag(FlagIsFixed)
	f.SetFlag(FlagHasFrame)
	frame, header := makeFrameIfFlag(f, stackFV, frameTitle)
	if frame != nil {
		outView = frame
	}
	if !fixed {
		frameContainer := frame.(*zcontainer.StackView)
		add := makeButton("plus", "gray")
		if header != nil {
			header.Add(add, zgeo.CenterRight)
		} else {
			m := stackFV.Margin()
			m.SetMinY(m.Pos.Y + 7)
			stackFV.SetMargin(m)
			frameContainer.SetMinSize(zgeo.SizeD(200, 40))
			frameContainer.Add(add, zgeo.TopRight, zgeo.SizeD(-5, -9)).Free = true
		}
		add.SetPressedHandler("", zkeyboard.ModifierNone, func() {
			i := rval.Len()
			str := ""
			mval := reflect.ValueOf(str)
			buildMapRow(parent, stackFV, i, "", mval, false, f)
			toModalWindowRootOnly := false
			zcontainer.ArrangeChildrenAtRootContainer(parent, toModalWindowRootOnly)
		})
	}
	var keys []reflect.Value
	iter := rval.MapRange()
	for iter.Next() {
		keys = append(keys, iter.Key())
	}
	sort.Slice(keys, func(i, j int) bool {
		return zstr.SmartCompare(fmt.Sprint(keys[i]), fmt.Sprint(keys[j]))
	})
	i := 0
	for _, _mkey := range keys {
		mkey := _mkey
		mval := rval.MapIndex(mkey)
		key := fmt.Sprint(mkey)
		// if key != "" {
		// zlog.Info("zfields.BuildMapList:", mkey)
		view, mf := buildMapRow(parent, stackFV, i, key, mval, fixed, f)
		check, _ := view.(*zcheckbox.CheckBox)
		if check != nil {
			check.SetValueHandler("zfields.mapCheck", func(edited bool) {
				ron := reflect.ValueOf(check.On())
				rval.SetMapIndex(mkey, ron)
				var cview zview.View = check
				ap := ActionPack{Field: &mf, Action: EditedAction, RVal: ron, View: &cview}
				stackFV.callTriggerHandler(ap)
				// zlog.Info("check in map:", key, check.On(), zlog.Full(rval.Interface()))
			})
			// }
		}
		i++
	}
	return outView
}

func (v *FieldView) updateMapList(f *Field, rval reflect.Value, foundView zview.View) {
	// zlog.Info("update map!")
}

func (v *FieldView) updateSeparatedStringWithSlice(f *Field, rval reflect.Value, foundView zview.View) {
	str := f.JoinSeparatedSlice(rval)
	v.setText(f, str, foundView)
}

func (v *FieldView) setText(f *Field, valStr string, foundView zview.View) {
	if f.IsStatic() || v.params.AllStatic {
		label, _ := foundView.(*zlabel.Label)
		if label != nil {
			label.SetText(f.Prefix + valStr + f.Suffix)
		}
		return
	}
	tv, _ := foundView.(*ztext.TextView)
	if tv != nil {
		if tv.IsEditing() {
			return
		}
		tv.SetText(valStr)
	}
}

func findSubFieldView(view zview.View, optionalID string) (fv *FieldView) {
	zcontainer.ViewRangeChildren(view, true, true, func(view zview.View) bool {
		fvf := viewToFieldView(view)
		if fvf != nil {
			if optionalID == "" || fvf.ID == optionalID {
				fv = fvf
				return false
			}
		}
		return true
	})
	return
}

func FieldViewNew(id string, data any, params FieldViewParameters) *FieldView {
	v := fieldViewNew(id, true, data, params, zgeo.SizeD(10, 10), nil)
	return v
}

// func (v *FieldView) ArrangeChildren() {
// 	v.StackView.ArrangeChildren()
// }

func (v *FieldView) Rebuild() {
	// zlog.Info("FV.Rebuild:", v.data != nil, reflect.ValueOf(v.data).Kind())
	fview := FieldViewNew(v.ID, v.data, v.params)
	fview.Build(true)
	rep, _ := v.Parent().View.(zview.ChildReplacer)
	if rep != nil {
		rep.ReplaceChild(v, fview)
	}
	toModalWindowRootOnly := false
	zcontainer.ArrangeChildrenAtRootContainer(v, toModalWindowRootOnly)
}

func (v *FieldView) CallFieldAction(fieldID string, action ActionType, fieldValue interface{}) {
	view, _ := v.FindViewWithName(fieldID, false)
	if view == nil {
		zlog.Error("CallFieldAction find view", fieldID)
		return
	}
	f := v.FindFieldWithFieldName(fieldID)
	if f == nil {
		zlog.Error("CallFieldAction find field", fieldID)
		return
	}
	callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: action, RVal: reflect.ValueOf(fieldValue), View: &view})
}

func callActionHandlerFunc(ap ActionPack) bool {
	if ap.Action == EditedAction {
		if ap.Field.SetEdited {
			setFieldViewEdited(ap.FieldView)
		}
		if ap.FieldView.params.EditWithoutCallbacks {
			return true
		}
	}
	direct := (ap.Action == CreateFieldViewAction || ap.Action == SetupFieldAction)
	fh, _ := ap.FieldView.data.(ActionHandler)
	var result bool
	if fh != nil {
		result = fh.HandleAction(ap)
	}
	// zlog.Info("callActionHandlerFunc1", ap.Field.Name, ap.Action)
	if ap.View != nil && *ap.View != nil {
		first := true
		n := (*ap.View).Native()
		for n != nil {
			parent := n.Parent()
			if parent != nil {
				fv := viewToFieldView(parent.View)
				if fv != nil {
					if fv.IsSlice() {
						zlog.Info("SetDataHash:", n.Hierarchy())
						fv.dataHash = zreflect.HashAnyToInt64(reflect.ValueOf(fv.data).Elem().Interface(), "")
					}
					if !first {
						fh2, _ := fv.data.(ActionHandler)
						if fh2 != nil {
							fh2.HandleAction(ActionPack{FieldView: ap.FieldView, Action: ap.Action, RVal: ap.RVal, View: &parent.View})
						}
					}
					first = false
				}
			}
			n = parent
		}
	}
	// zlog.Info("callActionHandlerFunc", ap.Field.Name, ap.Action)
	if !result {
		var fieldAddress interface{}
		if !direct {
			// changed := false
			sv := reflect.ValueOf(ap.FieldView.data)
			if sv.Kind() == reflect.Ptr || sv.CanAddr() {
				// Here we run thru the possiblly new struct again, and find the item with same id as field
				finfo, found := zreflect.FieldForName(ap.FieldView.data, FlattenIfAnonymousOrZUITag, ap.Field.FieldName)
				if found {
					// changed = true
					if finfo.ReflectValue.CanAddr() {
						fieldAddress = finfo.ReflectValue.Addr().Interface()
					}
				}
			}
			// if !changed {
			// 	zlog.Info("NOOT!!!", ap.Field.FN(), ap.Action, ap.FieldView.data != nil, sv.Kind() == reflect.Ptr, sv.CanAddr())
			// 	// zlog.Fatal("Not CHANGED!", ap.Field.FN())
			// }
		}
		if !ap.RVal.IsZero() {
			aih, _ := ap.RVal.Interface().(ActionHandler)
			if aih == nil && fieldAddress != nil {
				aih, _ = fieldAddress.(ActionHandler)
			}
			if aih != nil {
				result = aih.HandleAction(ap)
			}
		}
	}
	return result
}

func (v *FieldView) FindFieldWithFieldName(fn string) *Field {
	for i, f := range v.Fields {
		if f.FieldName == fn {
			return &v.Fields[i]
		}
	}
	return nil
}

func (fv *FieldView) makeButton(rval reflect.Value, f *Field) zview.View {
	format := f.Format
	if format == "" {
		format = "%v"
	}
	textCol := zstyle.DefaultFGColor()
	var color string
	if len(f.Colors) > 0 {
		color = f.Colors[0]
		if len(f.Colors) > 1 {
			textCol = zgeo.ColorFromString(f.Colors[1])
		}
		if !textCol.Valid {
			fg := zgeo.ColorFromString(color)
			if fg.Valid {
				textCol = fg.ContrastingGray()
			}
		}
	}
	name := f.Title
	if f.Name != "" || fv.params.Field.HasFlag(FlagIsLabelize) {
		name = f.Name
	}
	s := zgeo.SizeD(20, 24)
	if f.Height != 0 {
		s.H = f.Height
	}
	var view zview.View
	if color != "" {
		button := zshape.ImageButtonViewNew(name, color, s, zgeo.SizeNull)
		button.SetTextColor(textCol)
		button.SetSpacing(0)
		view = button
	} else {
		button := zbutton.New(name)
		button.SetColor(zgeo.ColorBlack)
		view = button
	}

	if f.HasFlag(FlagIsURL) {
		view.Native().SetPressedHandler("", zkeyboard.ModifierNone, func() {
			surl := ReplaceDoubleSquiggliesWithFields(fv, f, f.Path)
			zwindow.GetMain().SetLocation(surl)
		})
	}
	return view
}

func maybeAskBeforeAction(f *Field, action func()) {
	if f.Ask == "" {
		go action()
		return
	}
	zalert.Ask(f.Ask, func(ok bool) {
		if ok {
			go action()
		}
	})
}

func (v *FieldView) makeMenuedOwner(static, isSlice, isEdit bool, rval reflect.Value, f *Field, items zdict.Items) zview.View {
	isImage := (f.ImageFixedPath != "")
	shape := zshape.TypeRoundRect
	if isImage {
		shape = zshape.TypeNone
	}
	menuOwner := zmenu.NewMenuedOwner()
	menuOwner.IsStatic = static
	menuOwner.IsMultiple = isSlice

	if isEdit {
		zlog.Assert(!isSlice)
		menuOwner.EditFunc = enumEditHandlers[f.Enum]
	}
	if v.params.IsEditOnNewStruct {
		menuOwner.StoreKey = f.ValueStoreKey
	}
	menuOwner.SetTitle = true
	for _, format := range strings.Split(f.Format, "|") {
		if menuOwner.TitleIsAll == " " {
			menuOwner.TitleIsAll = format
		}
		switch format {
		case "all":
			menuOwner.TitleIsAll = " "
		case "%d":
			menuOwner.GetTitleFunc = func(icount int) string { return strconv.Itoa(icount) }
		case `title`:
			menuOwner.TitleIsValueIfOne = true
		default:
			menuOwner.PluralableWord = format
		}
	}
	menuOwner.ShowZeroAsEmpty = f.HasFlag(FlagAllowEmptyAsZero)
	mItems := zmenu.MOItemsFromZDictItemsAndValues(items, rval.Interface(), f.Flags&FlagIsActions != 0)

	menu := zmenu.MenuOwningButtonCreate(menuOwner, mItems, shape)
	if isImage {
		menu.SetImage(nil, true, f.Size, f.ImageFixedPath, zgeo.SizeNull, nil)
	} else {
		if len(f.Colors) != 0 {
			menu.SetColor(zgeo.ColorFromString(f.Colors[0]))
		}
	}
	var view zview.View
	view = menu
	// SetAuthenticationIDAsDefaultForSliceViewview := menu
	if menuOwner.IsStatic {
		menuOwner.StaticSelectedHandlerFunc = func(id string) {
			idVal := reflect.ValueOf(id)
			callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: PressedAction, RVal: idVal, View: &view})
		}
	}
	menuOwner.SelectedHandlerFunc = func(edited bool) {
		sel := menuOwner.SelectedItem()
		if sel != nil {
			kind := reflect.ValueOf(sel.Value).Kind()
			if menuOwner.IsStatic {
				if kind != reflect.Ptr && kind != reflect.Struct {
					callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: PressedAction, RVal: rval, View: &view})
				}
			} else {
				if menuOwner.IsMultiple {
					allSelected := menuOwner.SelectedItems()
					slicePtr := rval.Addr().Interface()
					zslice.Empty(slicePtr)
					for _, item := range allSelected {
						zslice.AddAtEnd(slicePtr, item.Value)
					}
					return
				}
				if sel.Value != nil {
					val, _ := v.fieldToDataItem(f, menu)
					val.Set(reflect.ValueOf(sel.Value))
				}
			}
		}

		ap := ActionPack{Field: f, Action: EditedAction, RVal: rval, View: &view}
		v.callTriggerHandler(ap)
		// zlog.Info("MV.Call action:", f.Name, rval)
		callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: EditedAction, RVal: rval, View: &view})
	}
	menuOwner.ClosedFunc = func() {
		if menuOwner.IsMultiple {
			zlog.Assert(isSlice)
			zslice.Empty(rval.Addr().Interface())
			for _, mi := range menuOwner.SelectedItems() {
				zslice.AddAtEnd(rval.Addr().Interface(), mi.Value)
			}
		}
	}
	return menu
}

func (v *FieldView) makeSimpleMenu(rval reflect.Value, f *Field, items zdict.Items) zview.View {
	name := f.Name + "Menu"
	if v.params.IsEditOnNewStruct && f.ValueStoreKey != "" && !v.isRows() {
		name = "key:" + f.ValueStoreKey
	}
	menu := zmenu.NewView(name, items, rval.Interface())
	menu.RowFormat = f.Format
	menu.SetMaxWidth(f.MaxWidth)
	var view zview.View
	view = menu
	menu.SetSelectedHandler(func(edited bool) {
		val, _ := v.fieldToDataItem(f, menu)
		isZero := true
		if val.IsValid() {
			isZero = val.IsZero()
		}
		action := DataChangedAction
		if edited {
			action = EditedAction
		}
		// zlog.Info("Menu Edited", v.Hierarchy(), f.Name, isZero, val, menu.CurrentValue())
		v.updateShowEnableFromZeroer(isZero, true, menu.ObjectName())
		v.updateShowEnableFromZeroer(isZero, false, menu.ObjectName())
		ap := ActionPack{Field: f, Action: action, RVal: rval, View: &view}
		v.callTriggerHandler(ap)
		callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: action, RVal: rval, View: &view})
	})
	return view
}

func (v *FieldView) makeMenu(rval reflect.Value, f *Field, items zdict.Items) zview.View {
	static := f.IsStatic() //|| v.params.AllStatic
	isSlice := rval.Kind() == reflect.Slice
	// zlog.Info("makeMenu", f.Name, f.IsStatic(), v.params.AllStatic, isSlice, rval.Kind(), len(items))
	_, isEdit := enumEditHandlers[f.Enum]
	if static || isSlice || isEdit {
		return v.makeMenuedOwner(static, isSlice, isEdit, rval, f, items)
	}
	return v.makeSimpleMenu(rval, f, items)
}

func getTimeString(rval reflect.Value, f *Field) string {
	var str string
	t := rval.Interface().(time.Time)
	if f.Flags&FlagAllowEmptyAsZero != 0 {
		if t.IsZero() {
			return f.ZeroText
		}
	}
	if t == ztime.BigTime {
		return f.MaxText
	}
	format := f.Format
	secs := (f.Flags&FlagHasSeconds != 0)
	if format == "" {
		format = "15:04"
		if secs {
			format += ":05"
		}
		if zlocale.IsDisplayServerTime.Get() {
			format += "-07"
		}
		format += " 02-Jan-06"
	}
	if format == "nice" {
		subSecs := -1
		if f.HasFlag(FlagHasSeconds) {
			subSecs = 0
		}
		if f.FractionDecimals != 0 {
			subSecs = f.FractionDecimals
		}
		str = ztime.GetNiceSubSecs(t, subSecs)
	} else {
		t = ztime.GetTimeWithServerLocation(t)
		str = t.Format(format)
	}
	return str
}

func getTextFromNumberishItem(rval reflect.Value, f *Field) (text, tip string, dur time.Duration) {

	if f.Flags&FlagAllowEmptyAsZero != 0 {
		if rval.IsZero() {
			return f.ZeroText, "", 0
		}
	}
	stringer, got := rval.Interface().(UIStringer)
	if got {
		return stringer.ZUIString(f.HasFlag(FlagAllowEmptyAsZero)), "", 0
	}
	zkind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	isDurTime := zkind == zreflect.KindTime && f.Flags&FlagIsDuration != 0
	if zkind == zreflect.KindTime && !isDurTime {
		return getTimeString(rval, f), "", 0
	}
	if isDurTime || f.PackageName == "time" && rval.Type().Name() == "Duration" {
		var dur time.Duration
		if isDurTime {
			dur = time.Since(rval.Interface().(time.Time))
		} else {
			dur = time.Duration(rval.Int())
			// zlog.Info("DurTime", dur, f.Flags&FlagHasSeconds != 0)
		}
		var str string
		if dur != 0 && dur < time.Second {
			str = fmt.Sprintf("%dms", dur/time.Millisecond)
		} else {
			str = ztime.GetDurationAsHMSString(dur, f.HasFlag(FlagHasHours), f.HasFlag(FlagHasMinutes), f.HasFlag(FlagHasSeconds), f.FractionDecimals)
		}
		// zlog.Info("DurTime", dur, str, f.HasFlag(FlagHasHours))
		return str, "", dur
	}
	format := f.Format
	significant := f.FractionDecimals
	if significant == 0 {
		significant = -1
	}
	if f.Kind == zreflect.KindFloat {
		n, err := zfloat.FromAnyFloat(rval.Interface())
		if err != nil {
			return fmt.Sprint(rval.Interface()), "", 0
		}
		if n == 0 && f.Flags&FlagAllowEmptyAsZero != 0 {
			return f.ZeroText, "", 0
		}
		if f.MaxText != "" && n == math.MaxFloat32 { // should be MaxFloat64?
			return f.MaxText, "", 0
		}
		str := zwords.NiceFloat(n, significant)
		return str, "", 0
	} else {
		b, err := zint.FromAnyInt(rval.Interface())
		if err != nil {
			return fmt.Sprint(rval.Interface()), "", 0
		}
		if b == 0 && f.Flags&FlagAllowEmptyAsZero != 0 {
			return f.ZeroText, "", 0
		}
		if f.MaxText != "" && b == math.MaxInt64 {
			return f.MaxText, "", 0
		}
		tip = fmt.Sprint(b)
		if significant <= 0 && (format == MemoryFormat || format == StorageFormat || format == BPSFormat) {
			significant = 1
		}
		switch format {
		case MemoryFormat:
			return zwords.GetMemoryString(b, "", significant), tip, 0
		case StorageFormat:
			return zwords.GetStorageSizeString(b, "", significant), tip, 0
		case BPSFormat:
			return zwords.GetBandwidthString(b, "", significant), tip, 0
		case HumanFormat:
			return zint.MakeHumanFriendly(b), "", 0
		}
	}
	if format == "" {
		format = "%v"
	}
	// zlog.Info("FInt:", ni, err, f.FieldName, rval.Interface())
	return fmt.Sprintf(format, rval.Interface()), "", 0
}

func (v *FieldView) maybeMakeLabelHandleFromClipboard(f *Field, label *zlabel.Label, str string, rval reflect.Value) {
	if v.isRows() || f.Flags&FlagFromClipboard == 0 {
		return
	}
	add := "➕"
	if !strings.Contains(str, add) {
		label.SetText(str + " " + add)
	}
	label.SetPressedHandler("zfromclip", 0, func() {
		zlog.Info("FromClipboard pressed for field:", f.FieldName)
		view := label.View
		ap := ActionPack{Field: f, Action: FromClipboardAction, RVal: rval, View: &view}
		v.invokeAction(ap)
	})
}

func (v *FieldView) makeText(rval reflect.Value, f *Field, noUpdate bool) zview.View {
	str, tip, _ := getTextFromNumberishItem(rval, f)
	if f.IsStatic() || v.params.AllStatic {
		var label *zlabel.Label
		// zlog.Info("LABEL1:", f.FieldName, f.Rows, str)
		if f.HasFlag(FlagIsDocumentation) {
			surl := ReplaceDoubleSquiggliesWithFields(v, f, f.Path)
			label = zlabel.New(str)
			ztextinfo.SetTextDecoration(label, ztextinfo.DecorationUnderlined)
			label.SetPressedHandler("zfield.DocPressed", zkeyboard.ModifierNone, func() {
				go zwidgets.DocumentationViewPresent(surl, false)
			})
		} else {
			surl := ReplaceDoubleSquiggliesWithFields(v, f, f.Path)
			isLink := f.HasFlag(FlagIsURL) && !v.isRows()
			if isLink {
				if surl == "" {
					surl = rval.String()
				}
			} else {
				ug, is := rval.Interface().(zstr.URLGetter)
				if is {
					isLink = true
					surl = ug.GetURL()
				}
			}
			str = f.Prefix + str + f.Suffix
			if isLink {
				label = zlabel.NewLink(str, surl, true)
			} else {
				label = zlabel.New(str)
			}
			v.maybeMakeLabelHandleFromClipboard(f, label, str, rval)
		}
		setDocumentationLink(label, rval)

		if tip != "" {
			label.SetToolTip(tip)
		}
		if f.Wrap == ztextinfo.WrapTailTruncate.String() || v.isRows() {
			// label.SetWrap(ztextinfo.WrapTailTruncate)
		}
		label.Columns = f.Columns
		if !v.isRows() {
			label.SetMaxLines(f.Rows)
		}
		if f.MaxWidth != 0 {
			label.SetMaxWidth(f.MaxWidth)
		}
		if f.Flags&FlagIsDuration != 0 {
			v.updateSinceTime(label, f) // we should really not do getTextFromNumberishItem above if we do this
		}
		j := f.Justify
		if j == zgeo.AlignmentNone {
			j = f.Alignment & (zgeo.Left | zgeo.HorCenter | zgeo.Right)
			if j == zgeo.AlignmentNone {
				j = zgeo.Left
			}
		}
		f.SetFont(label, nil)
		label.SetTextAlignment(j)
		if f.Rows <= 1 || inMapRows > 0 {
			// label.SetWrap(ztextinfo.WrapTailTruncate)
		}
		if !v.isRows() && f.Flags&FlagToClipboard != 0 {
			label.SetPressWithModifierToClipboard(zkeyboard.ModifierNone)
		}
		label.SetPressWithModifierToClipboard(zkeyboard.ModifierAlt)

		return label
	}
	var style ztext.Style
	cols := f.Columns
	if cols == 0 {
		cols = 20
	}
	zkind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	if f.Flags&FlagIsPassword != 0 {
		style.KeyboardType = zkeyboard.TypePassword
		if f.HasFlag(FlagIsFixed) {
			style.IsExistingPassword = true
		}
	} else if zkind == zreflect.KindInt {
		style.KeyboardType = zkeyboard.TypeInteger
	} else if zkind == zreflect.KindFloat {
		style.KeyboardType = zkeyboard.TypeFloat
	}
	if f.Flags&FlagDisableAutofill != 0 {
		style.DisableAutoComplete = true
	}
	if f.Format == "email" {
		style.KeyboardType = zkeyboard.TypeEmailAddress
	}
	tv := ztext.NewView(str, style, cols, f.Rows)
	tv.SetObjectName(f.FieldName)
	f.SetFont(tv, nil)
	tv.UpdateSecs = f.UpdateSecs
	if !noUpdate && tv.UpdateSecs == -1 {
		tv.UpdateSecs = 4
	}
	if len(f.Filters) > 0 {
		tv.FilterFunc = getFilterFuncFromFilterNames(f.Filters, f)
	}
	placeHolder := f.Placeholder
	if v.params.MultiSliceEditInProgress {
		placeHolder = "multiple, edit to change all"
	}
	tv.SetPlaceholder(placeHolder)
	tv.SetValueHandler("zfields.Filter", func(edited bool) {
		v.fieldHandleValueChanged(f, edited, tv.View)
	})
	if tip != "" {
		tv.SetToolTip(tip)
	}
	return tv
}

func setDocumentationLink(label *zlabel.Label, rval reflect.Value) {
	link, isDoc := rval.Interface().(DocumentationLink)
	if !isDoc {
		return
	}
	str := string(link)
	str = zstr.HeadUntil(str, "#")
	if zstr.HasSuffix(str, ".md", &str) {
		str = strings.Replace(str, "_", " ", -1)
		str = zstr.TitledWords(str)
		label.SetText(str)
	}
	label.SetFont(label.Font().NewWithStyle(zgeo.FontStyleBold))
	label.SetTextUnderline(true)
	label.SetCursor(zcursor.Pointer)
	col := label.Color().Mixed(zgeo.ColorBlue, 0.3)
	label.SetColor(col)
	// zlog.Info("DOC:", str, link)
	label.SetPressedHandler("", 0, func() {
		zwidgets.DocumentationViewPresent(string(link), zwidgets.DocumentationViewDefaultModal)
	})
}

func getFilterFuncFromFilterNames(names []string, f *Field) func(string) string {
	var funcs []func(string) string
	for _, fname := range names {
		fn := GetTextFilter(fname)
		if fn == nil {
			zlog.Error("No registered text filter for:", fname, f.FieldName)
			continue
		}
		funcs = append(funcs, fn)
	}
	return func(s string) string {
		for _, fn := range funcs {
			s = fn(s)
		}
		return s
	}
}

func (v *FieldView) invokeAction(ap ActionPack) {
	if v.callTriggerHandler(ap) {
		return
	}
	ap.FieldView = v
	callActionHandlerFunc(ap)
}

func (v *FieldView) fieldHandleValueChanged(f *Field, edited bool, view zview.View) {
	if !edited {
		return
	}
	rval, err := v.fieldToDataItem(f, view)
	if zlog.OnError(err) {
		return
	}
	action := DataChangedAction
	if edited {
		action = EditedAction
	}
	ap := ActionPack{Field: f, Action: action, RVal: rval, View: &view}
	v.invokeAction(ap)
}

func (v *FieldView) makeCheckbox(f *Field, b zbool.BoolInd) zview.View {
	cv := zcheckbox.New(b)
	cv.SetObjectName(f.FieldName)
	// if reflect.ValueOf(v.data).Kind() == reflect.Map {
	// 	return cv
	// }
	cv.SetValueHandler("zfields.checkBox", func(edited bool) {
		action := DataChangedAction
		if edited {
			action = EditedAction
		}
		val, _ := v.fieldToDataItem(f, cv)
		v.updateShowEnableFromZeroer(val.IsZero(), true, cv.ObjectName())
		v.updateShowEnableFromZeroer(val.IsZero(), false, cv.ObjectName())
		view := zview.View(cv)
		ap := ActionPack{Field: f, Action: action, RVal: reflect.ValueOf(cv.On()), View: &cv.View}
		if v.callTriggerHandler(ap) {
			return
		}
		callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: action, RVal: val, View: &view})
	})
	if !v.params.Field.HasFlag(FlagIsLabelize) && !v.isRows() {
		title := f.TitleOrName()
		if f.HasFlag(FlagNoTitle) {
			title = ""
		}
		_, stack := zcheckbox.Labelize(cv, title)
		return stack
	}
	if f.IsStatic() {
		cv.SetUsable(false)
	}
	return cv
}

func (v *FieldView) makeImage(rval reflect.Value, f *Field) zview.View {
	iv := zimageview.NewWithCachedPath("", f.Size)
	iv.DownsampleImages = true
	iv.SetMinSize(f.Size)
	iv.SetObjectName(f.FieldName)
	iv.OpaqueDraw = (f.Flags&FlagIsOpaque != 0)
	if f.Styling.FGColor.Valid {
		iv.EmptyColor = f.Styling.FGColor
	}
	if f.HasFlag(FlagIsURL) {
		iv.SetPressedHandler("", zkeyboard.ModifierNone, func() {
			surl := ReplaceDoubleSquiggliesWithFields(v, f, f.Path)
			zwindow.GetMain().SetLocation(surl)
		})
		return iv
	}
	if f.IsImageToggle() && rval.Kind() == reflect.Bool {
		iv.SetPressedHandler("", zkeyboard.ModifierNone, func() {
			val, _ := v.fieldToDataItem(f, iv)
			on := val.Bool()
			on = !on
			val.SetBool(on)
			v.Update(nil, false, false)
			ap := ActionPack{FieldView: v, Field: f, Action: EditedAction, RVal: val, View: &iv.View} // Set FieldView as used in callActionHandlerFunc
			v.callTriggerHandler(ap)
			callActionHandlerFunc(ap)
		})
		return iv
	}
	iv.SetPressedHandler("zfield.Image.CallAction", zkeyboard.ModifierNone, func() {
		maybeAskBeforeAction(f, func() {
			ap := ActionPack{Field: f, Action: PressedAction, RVal: reflect.ValueOf(f.FieldName), View: &iv.View}
			v.callTriggerHandler(ap)
		})
	})
	return iv
}

func setColorFromField(view zview.View, f *Field) {
	if f.Styling.FGColor.Valid {
		col := f.Styling.FGColor
		view.SetColor(col)
		return
	}
	_, is := view.(ztext.TextOwner)
	if !is {
		return
	}
	_, is = view.(*zshape.ShapeView)
	if is {
		return // ShapeView is a TextOwner, this is all a bit of a hack
	}
	view.Native().SetColor(zstyle.DefaultFGColor())
}

func (v *FieldView) updateOldTimeColor(label *zlabel.Label, f *Field) {
	finfo, found := zreflect.FieldForName(v.data, FlattenIfAnonymousOrZUITag, f.FieldName)
	if found {
		t := finfo.ReflectValue.Interface().(time.Time)
		if !t.IsZero() {
			updateOldDurationColor(label, time.Since(t), f)
		}
	}
}

func updateOldDurationColor(label *zlabel.Label, dur time.Duration, f *Field) {
	if f.OldSecs != 0 && ztime.DurSeconds(dur) > float64(f.OldSecs) {
		// zlog.Info("updateOldDurationColor1", f.Name, dur, f.OldSecs)
		label.SetColor(zgeo.ColorRed)
	} else {
		setColorFromField(label, f)
	}
}

func (v *FieldView) updateSinceTime(label *zlabel.Label, f *Field) {
	if zdebug.IsInTests { // if in unit-tests, we don't show anything as it would change
		label.SetText("")
		return
	}
	sv := reflect.ValueOf(v.data)
	zlog.Assert(sv.Kind() == reflect.Ptr || sv.CanAddr())
	finfo, found := zreflect.FieldForName(v.data, FlattenIfAnonymousOrZUITag, f.FieldName)
	if found {
		var str string
		t := finfo.ReflectValue.Interface().(time.Time)
		tooBig := false
		if !t.IsZero() {
			since := time.Since(t)
			str, tooBig = ztime.GetDurationString(since, f.Flags&FlagHasSeconds != 0, f.Flags&FlagHasMinutes != 0, f.Flags&FlagHasHours != 0, f.FractionDecimals)
		}
		// inter := val.Interface()
		// if val.CanAddr() {
		// 	inter = val.Addr()
		// }
		if tooBig {
			label.SetText("●")
			label.SetColor(zgeo.ColorRed)
		} else {
			label.SetText(str)
			setColorFromField(label, f)
		}
	}
}

func getColumnsForTime(f *Field) int {
	var c int
	for _, flag := range []FlagType{FlagHasSeconds, FlagHasMinutes, FlagHasHours, FlagHasDays, FlagHasMonths, FlagHasYears} {
		if f.Flags&flag != 0 {
			c += 3
		}
	}
	return c - 1
}

func getDictItemsFromSlice(slice reflect.Value, f *Field) (zdict.Items, error) {
	ditems, got := slice.Interface().(zdict.Items)
	if got {
		return ditems, nil
	}
	getter, _ := slice.Interface().(zdict.ItemsGetter)
	items := zdict.ItemsFromRowGetterSlice(slice.Interface())
	if getter == nil && items == nil {
		items = &zdict.Items{}
		for i := 0; i < slice.Len(); i++ {
			item := slice.Index(i)
			*items = append(*items, zdict.Item{Value: item.Interface(), Name: fmt.Sprint(item.Interface())})
		}
	}
	if items == nil {
		return getter.GetItems(), nil
	}
	return *items, nil
}

func (v *FieldView) makeRadioButtonGroup(f *Field, rval reflect.Value) zview.View {
	stack := zcontainer.StackViewVert(f.FieldName)
	// stack.SetSpacing(44)
	stack.SetObjectName(f.FieldName)
	stack.GridVerticalSpace = 16
	enum, got := fieldEnums[f.Radio]
	if rval.IsZero() && f.HasFlag(FlagHasDefault) && f.Default != "" {
		zstr.SetStringToAny(rval.Addr().Interface(), f.Default)
	}
	if f.ValueStoreKey != "" {
		val, got := zkeyvalue.DefaultStore.GetItemAsAny(f.ValueStoreKey)
		if got {
			rval.Set(reflect.ValueOf(val))
		}
	}
	zlog.Assert(got, f.Radio)
	for _, e := range enum {
		equal := reflect.DeepEqual(rval.Interface(), e.Value)
		b := zradio.NewButton(equal, fmt.Sprint(e.Value), f.FieldName)
		b.SetValueHandler("zfields", func(edited bool) {
			frval, err := v.fieldToDataItem(f, stack)
			if err == nil {
				return
			}
			action := DataChangedAction
			if edited {
				action = EditedAction
				if f.ValueStoreKey != "" {
					zkeyvalue.DefaultStore.SetItem(f.ValueStoreKey, frval.Interface(), true)
				}
			}
			ap := ActionPack{FieldView: v, Field: f, Action: action, RVal: frval, View: &b.View}
			if v.callTriggerHandler(ap) {
				return
			}
			callActionHandlerFunc(ap)
		})
		_, row, _, _ := zguiutil.Labelize(b, e.Name, "", f.MinWidth, zgeo.CenterLeft, f.Description)
		stack.Add(row, zgeo.CenterLeft)
	}
	return stack
}

func (v *FieldView) createActionMenu(f *Field, sid string) zview.View {
	size := zgeo.SizeD(18, 18)
	if !f.Size.IsNull() {
		size = f.Size
	}
	actions := zimageview.NewWithCachedPath("images/zcore/gear.png", size)
	actions.MixColorForDarkMode = zgeo.ColorNewGray(0.5, 1)
	actions.DownsampleImages = true
	menu := zmenu.NewMenuedOwner()
	menu.Build(actions, nil) // do we need to do this?
	menu.CreateItemsFunc = func() []zmenu.MenuedOItem {
		return v.params.CreateActionMenuItemsFunc(sid)
	}
	actions.ShortCutHandler = menu
	return actions
}

func (v *FieldView) createSpecialView(rval reflect.Value, f *Field) (view zview.View, skip bool) {
	if f.Transformer != "" {
		return v.makeText(rval, f, false), false
	}
	if f.Flags&FlagIsButton != 0 {
		if v.params.HideStatic {
			return nil, true
		}
		return v.makeButton(rval, f), false
	}
	if f.HasFlag(FlagIsActions) && rval.Kind() == reflect.Bool {
		if v.isRows() {
			// zlog.Info("CreateSpecial action", zlog.Pointer(v), v.Hierarchy(), f.Name, v.params.CreateActionMenuItemsFunc != nil)
			zlog.Assert(v.params.CreateActionMenuItemsFunc != nil, f.Name)
			sget, _ := v.data.(zstr.StrIDer)
			zlog.Assert(sget != nil, reflect.TypeOf(v.data))
			return v.createActionMenu(f, sget.GetStrID()), false
		} else {
			return nil, true
		}
	}
	if !rval.IsValid() {
		// zlog.Info("createSpecialView: not valid", f.FieldName)
		return nil, true
	}
	if f.Radio != "" {
		return v.makeRadioButtonGroup(f, rval), false
	}
	// zlog.Info("createSpecialView:", f.FieldName, rval.IsZero(), rval.Type())

	// if !rval.IsZero() {
	stype := rval.Type().String()
	fcreate := creators[stype]
	if fcreate != nil {
		view = fcreate(v, f, rval.Interface())
		return view, false
	}
	// }
	if f.WidgetName != "" && rval.Kind() != reflect.Slice {
		w := widgeters[f.WidgetName]
		if w != nil {
			widgetView := w.Create(v, f)
			changer, _ := widgetView.(zview.ValueHandler)
			if changer != nil {
				changer.SetValueHandler("zfields.widgetChanged", func(edited bool) {
					v.fieldHandleValueChanged(f, edited, widgetView)
				})
			}
			return widgetView, false
		}
	}
	ap := ActionPack{Field: f, Action: CreateFieldViewAction, RVal: reflect.ValueOf(nil), View: &view}
	if v.callTriggerHandler(ap) {
		return
	}
	if rval.CanAddr() {
		callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: CreateFieldViewAction, RVal: rval.Addr(), View: &view}) // this sees if actual ITEM is a field handler
		if view != nil {
			return
		}
	}
	if f.LocalEnum != "" {
		ei, findex := FindLocalFieldWithFieldName(v.data, f.LocalEnum)
		if zlog.ErrorIf(findex == -1, v.Hierarchy(), f.Name, f.LocalEnum) {
			return nil, true
		}
		enum, err := getDictItemsFromSlice(ei, f)
		if err != nil {
			return nil, false
		}
		for i := range enum {
			if f.Flags&FlagAllowEmptyAsZero != 0 {
				if enum[i].Value != nil && reflect.ValueOf(enum[i].Value).IsZero() {
					enum[i].Name = f.ZeroText
				}
			}
		}
		menu := v.makeMenu(rval, f, enum)
		if menu == nil {
			zlog.Error("no local enum for", f.LocalEnum)
			return nil, true
		}
		return menu, false
	}
	if f.Enum != "" {
		if f.IsStatic() || v.params.AllStatic {
			enum := GetEnum(f.Enum)
			str := findNameOfEnumForRVal(rval, enum)
			view = v.makeText(reflect.ValueOf(str), f, true)
			tv := view.(*zlabel.Label)
			tv.SetFont(tv.Font().NewWithStyle(zgeo.FontStyleBold))
		} else {
			view = v.makeMenu(rval, f, nil)
		}
		return view, false
	}
	_, got := rval.Interface().(UIStringer)
	if got && (f.IsStatic() || v.params.AllStatic) {
		return v.makeText(rval, f, false), false
	}
	if f.HasFlag(FlagShowSliceCount) {
		zlog.Assert(rval.Kind() == reflect.Slice || rval.Kind() == reflect.Map, rval.Kind(), f.Name)
		zlog.Assert(f.IsStatic(), f.Name)
		return v.makeText(rval, f, false), false
	}
	return nil, false
}

func findNameOfEnumForRVal(rval reflect.Value, enum zdict.Items) string {
	a := rval.Interface()
	item := enum.FindValue(a)
	if item == nil {
		return ""
	}
	return item.Name
}

func (v *FieldView) BuildStack(name string, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) {
	// zlog.Info("FV.BuildStack", v.ID, v.Hierarchy())
	if v.params.Styling.Spacing != zfloat.Undefined {
		v.SetSpacing(v.params.Styling.Spacing)
	}
	if v.params.Field.HasFlag(FlagIsLabelize) {
		v.GridVerticalSpace = v.Spacing()
		v.SetSpacing(math.Max(12, v.Spacing()))
	}
	zlog.Assert(reflect.ValueOf(v.data).Kind() == reflect.Ptr, name, reflect.ValueOf(v.data).Kind())

	v.lastCheckered = false
	ForEachField(v.data, v.params.FieldParameters, v.Fields, func(each FieldInfo) bool {
		if !each.ReflectValue.IsValid() {
			return true
		}
		v.buildItem(each.Field, each.ReflectValue, each.FieldIndex, defaultAlign, cellMargin, useMinWidth)
		return true
	})
}

func (v *FieldView) buildItem(f *Field, rval reflect.Value, index int, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) zview.View {
	if f.WhenMods != zkeyboard.ModifierNone {
		if zkeyboard.ModifiersAtPress != f.WhenMods {
			return nil
		}
	}
	if rval.Kind() == reflect.Interface {
		rval = rval.Elem()
	}
	// 	zlog.Info("BuildItem:", f.Name, f.Size, f.Flags, f.ImageFixedPath)
	if !f.Margin.IsNull() {
		cellMargin = f.Margin
	}
	// labelizeWidth := v.params.LabelizeWidth
	// parentFV := ParentFieldView(v)
	// if parentFV != nil && v.params.LabelizeWidth == 0 {
	// 	labelizeWidth = parentFV.params.LabelizeWidth
	// }
	exp := zgeo.AlignmentNone

	if f.IsStatic() {
		if v.params.HideStatic {
			return nil
		}
		if f.HasFlag(FlagOmitZero) && rval.IsZero() {
			return nil
		}
	}
	view, skip := v.createSpecialView(rval, f)
	if skip {
		return nil
	}
	kind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	if view == nil {
		switch kind {
		case zreflect.KindStruct:
			vert := true
			if !f.Vertical.IsUnknown() {
				vert = f.Vertical.Bool()
			}
			params := v.params
			params.Field.MergeInField(f)
			if f.IsStatic() {
				params.AllStatic = true
			}
			fieldView := fieldViewNew(f.FieldName, vert, rval.Addr().Interface(), params, zgeo.SizeNull, v)
			view, _ = makeFrameIfFlag(f, fieldView, "")
			if view == nil {
				view = fieldView
			}
			fieldView.BuildStack(f.FieldName, zgeo.TopLeft, zgeo.SizeNull, true)

		case zreflect.KindBool:
			if f.Flags&FlagIsImage != 0 && f.IsImageToggle() && rval.Kind() == reflect.Bool {
				view = v.makeImage(rval, f)
				break
			}
			exp = zgeo.AlignmentNone
			vbool := rval.Interface().(bool)
			b := zbool.ToBoolInd(vbool)
			view = v.makeCheckbox(f, b)

		case zreflect.KindInt:
			if rval.Type().Name() == "BoolInd" {
				exp = zgeo.HorShrink
				view = v.makeCheckbox(f, zbool.BoolInd(rval.Int()))
			} else {
				noUpdate := true
				view = v.makeText(rval, f, noUpdate)
			}

		case zreflect.KindFloat:
			noUpdate := true
			view = v.makeText(rval, f, noUpdate)

		case zreflect.KindString:
			if f.Flags&FlagIsImage != 0 {
				view = v.makeImage(rval, f)
			} else {
				if (f.MaxWidth != f.MinWidth || f.MaxWidth != 0) && f.Flags&FlagIsButton == 0 {
					exp = zgeo.HorExpand
				}
				view = v.makeText(rval, f, false)
			}

		case zreflect.KindMap:
			view = v.BuildMapList(rval, f, "")

		case zreflect.KindSlice:
			if !f.HasFlag(FlagIsGroup) || v.isRows() {
				if f.StringSep != "" {
					noUpdate := true
					rv := reflect.ValueOf("")
					view = v.makeText(rv, f, noUpdate)
					break
				}
				getter, _ := rval.Interface().(zdict.ItemsGetter)
				if getter != nil {
					menu := v.makeMenu(rval, f, getter.GetItems())
					view = menu
					break
				}
				items := zdict.ItemsFromRowGetterSlice(rval.Interface())
				if items != nil {
					menu := v.makeMenu(rval, f, *items)
					view = menu
					break
				}
			}
			if v.params.MultiSliceEditInProgress {
				return nil
			}
			if f.Alignment != zgeo.AlignmentNone {
				exp = zgeo.Expand
			} else {
				exp = zgeo.AlignmentNone
			}
			params := v.params
			params.Field.MergeInField(f)
			// zlog.Info("NewSlice", f.Name, rval.Type(), rval.Kind(), f.Flags, zdebug.CallingStackString())
			if rval.Kind() != reflect.Pointer && rval.CanAddr() {
				rval = rval.Addr()
			}
			fieldCopy := *f
			fieldCopy.ClearFlag(FlagShowIfExtraSpace) // we don't want sub-views to be built with this ... and probably other things...
			view = v.NewSliceView(rval.Interface(), &fieldCopy)

		case zreflect.KindTime:
			columns := f.Columns
			if columns == 0 {
				columns = getColumnsForTime(f)
			}
			noUpdate := true
			view = v.makeText(rval, f, noUpdate)
			if f.IsStatic() || v.params.AllStatic {
				label := view.(*zlabel.Label)
				label.Columns = columns
				if f.Flags&FlagIsDuration != 0 || f.OldSecs != 0 {
					timer := ztimer.RepeatForever(1, func() {
						nlabel := view.(*zlabel.Label)
						if f.Flags&FlagIsDuration != 0 {
							v.updateSinceTime(nlabel, f)
						} else {
							v.updateOldTimeColor(nlabel, f)
						}
					})
					// v.AddOnRemoveFunc(timer.Stop)
					view.Native().AddOnRemoveFunc(func() {
						timer.Stop()
					})
				} else {
					if f.Format == "nice" {
						timer := ztimer.StartAt(ztime.OnThisHour(time.Now(), 1), func() {
							repeater := ztimer.RepeatForever(60*60, func() {
								nlabel := view.(*zlabel.Label)
								str, tip, _ := getTextFromNumberishItem(rval, f)
								nlabel.SetText(str)
								if tip != "" {
									nlabel.SetToolTip(tip)
								}
							})
							v.AddOnRemoveFunc(repeater.Stop)
						})
						v.AddOnRemoveFunc(timer.Stop)
					}
				}
			}

		case zreflect.KindPointer:
			if !rval.IsNil() && rval.Interface() != nil && zlog.Pointer(rval.Interface()) != "0x0" {
				return v.buildItem(f, rval.Elem(), index, defaultAlign, cellMargin, useMinWidth)
			}
			return nil

		case zreflect.KindInterface:
			zlog.Info("buildItem interface:", rval.Elem().Type())
		default:
			panic(fmt.Sprintln("buildItem bad type:", f.Name, kind))
		}
	}
	zlog.Assert(view != nil)
	if f.HasFlag(f.Flags&FlagLongPress | FlagPress | FlagShowPopup) {
		nowItem := rval // store item in nowItem so closures below uses right item
		if f.HasFlag(FlagShowPopup) {
			view.Native().SetPressedDownHandler("zfields.ShowPopup", 0, func() bool {
				return v.popupContent(view, f)
			})
		} else {
			if f.Flags&FlagPress != 0 {
				view.Native().SetPressedHandler("zfield.CallAction", zkeyboard.ModifierNone, func() {
					if f.RPCCall != "" {
						maybeAskBeforeAction(f, func() {
							var reply string
							a := reflect.ValueOf(v.data).Elem().Interface()
							err := zrpc.MainClient.Call(f.RPCCall, a, &reply)
							if err != nil {
								zalert.ShowError(err)
							}
							if reply != "" {
								zalert.Show(reply)
							}
						})
						return
					}
					maybeAskBeforeAction(f, func() {
						callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: PressedAction, RVal: nowItem, View: &view})
					})
				})
			}
			if f.Flags&FlagLongPress != 0 {
				view.Native().SetLongPressedHandler("zfield.CallAction", zkeyboard.ModifierNone, func() {
					maybeAskBeforeAction(f, func() {
						callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: LongPressedAction, RVal: nowItem, View: &view})
					})
				})
			}
		}
	}
	updateToolTip(f, v.data, view)
	if !f.Styling.DropShadow.Delta.IsNull() {
		nv := view.Native()
		nv.SetDropShadow(f.Styling.DropShadow)
	}
	view.SetObjectName(f.FieldName)
	setColorFromField(view, f)
	if f.Styling.BGColor.Valid {
		view.SetBGColor(f.Styling.BGColor)
	} else if f.HasFlag(FlagIsForZDebugOnly) {
		view.SetBGColor(zstyle.DebugBackgroundColor)
	} else if f.HasFlag(FlagCheckerCell) {
		if !v.lastCheckered {
			lum := float32(0)
			if zstyle.Dark {
				lum = 1
			}
			view.SetBGColor(zgeo.ColorNewGray(lum, 0.05))
		}
		v.lastCheckered = !v.lastCheckered
	}
	if rval.CanAddr() {
		callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: CreatedViewAction, RVal: rval.Addr(), View: &view})
	}
	if f.Path != "" {
		path := f.Path
		zstr.HasPrefix(path, "./", &path)
		path = ReplaceDoubleSquiggliesWithFields(v, f, path)
		if f.HasFlag(FlagIsDownload) {
			view.Native().SetPressedHandler("zfield.Download", zkeyboard.ModifierNone, func() {
				zview.DownloadURI(path, "")
			})
		}
	}
	_, isLabel := view.(*zlabel.Label)
	if isLabel && (f.MaxWidth != f.MinWidth || f.MaxWidth != 0) {
		exp = zgeo.HorExpand
	}
	cell := &zcontainer.Cell{}
	def := defaultAlign
	all := zgeo.Left | zgeo.HorCenter | zgeo.Right
	if f.Alignment&all != 0 {
		def &= ^all
	}
	cell.Alignment = def | exp | f.Alignment
	cell.Margin = zgeo.RectMarginForSizeAndAlign(cellMargin, cell.Alignment)

	var lastShowIfExtraSpace float64
	if f.HasFlag(FlagShowIfExtraSpace) && f.MinWidth > 0 {
		for _, c := range v.Cells {
			if c.ShowIfExtraSpace != 0 {
				lastShowIfExtraSpace = c.ShowIfExtraSpace
			}
		}
		cell.ShowIfExtraSpace = lastShowIfExtraSpace + f.MinWidth
		// zlog.Info("FlagShowIfExtraSpace:", v.ObjectName(), f.FieldName, cell.ShowIfExtraSpace, f.MinWidth)
	}
	// doLabelize := (labelizeWidth != 0 || f.LabelizeWidth < 0) && !f.HasFlag(FlagNoLabel)
	// zlog.Info("CELLMARGIN:", f.Name, cellMargin, cell.Alignment)
	var lstack *zcontainer.StackView
	isLabelize := (v.params.Field.HasFlag(FlagIsLabelize) && !f.HasFlag(FlagDontLabelize))
	if isLabelize {
		title := f.TitleOrName()
		if f.HasFlag(FlagNoTitle) {
			title = ""
		}
		var desc string
		if v.params.Field.HasFlag(FlagLabelizeWithDescriptions) {
			desc = f.Description
			if desc == "" {
				desc = " " // we force an empty description so grid handles easy
			}
		}
		if view.ObjectName() == "" {
			view.SetObjectName(title)
		}
		var label *zlabel.Label
		a := cell.Alignment
		a &= ^zgeo.Vertical
		a |= (f.Alignment & zgeo.Vertical)
		if a&zgeo.Vertical == 0 {
			a |= zgeo.VertCenter
		}
		if !v.params.MultiSliceEditInProgress && f.Required != "" {
			title += "*"
		}
		label, lstack, cell, _ = zguiutil.Labelize(view, title, f.FieldName, 0, a, desc)
		_, is := view.(*FieldSliceView)
		if is {
			cell.Alignment |= zgeo.Expand
		} else {
			cell.Alignment |= zgeo.HorShrink
		}
		if f.HasFlag(FlagIsLockable) {
			if !zlog.ErrorIf(view.ObjectName() == "", f.FieldName) {
				lock := zguiutil.CreateLockIconForView(view)
				mr := zgeo.RectFromXY2(0, 2, 2, 0)
				lstack.AddAdvanced(lock, zgeo.CenterRight, mr, zgeo.Size{}, -1, true).RelativeToName = view.ObjectName()
				// zlog.Info("Lock relative:", view.ObjectName(), len(lstack.GetChildren(true)))
			}
		}
		if f.HasFlag(FlagIsForZDebugOnly) {
			label.SetBGColor(zstyle.DebugBackgroundColor)
			label.SetCorner(4)
		}
		updateToolTip(f, v.data, lstack)
		v.Add(lstack, zgeo.HorExpand|zgeo.Left|zgeo.Top)
	}
	if useMinWidth {
		cell.MinSize.W = f.MinWidth
	}
	cell.MaxSize.W = f.MaxWidth
	if v.params.BuildChildrenHidden {
		view.Show(false)
	}
	if !isLabelize {
		cell.View = view
		if v.params.Field.HasFlag(FlagIsLabelize) {
			cell.NotInGrid = true
		}
		v.AddCell(*cell, -1)
	}
	return view
}

func (fv *FieldView) freshRValue(fieldName string) reflect.Value {
	finfo, found := zreflect.FieldForName(fv.data, FlattenIfAnonymousOrZUITag, fieldName)
	zlog.Assert(found, fieldName)
	return finfo.ReflectValue
}

func (fv *FieldView) popupContent(target zview.View, f *Field) bool {
	rval := fv.freshRValue(f.FieldName)
	a := rval.Interface()
	str := fmt.Sprint(a)
	zs, _ := a.(UIStringer)
	if zs != nil {
		str = zs.ZUIString(f.HasFlag(FlagAllowEmptyAsZero))
	}
	d, _ := a.(zstr.Describer)
	if d != nil {
		desc := d.GetDescription()
		str = zstr.Concat("\n", str, desc)
	}
	// zlog.Info("popupContent:", str)
	if str == "" {
		return false
	}
	att := zpresent.AttributesDefault()
	att.Alignment = zgeo.TopLeft // | zgeo.HorOut
	att.PlaceOverMargin = zgeo.SizeD(-8, -4)
	stack := zcontainer.StackViewVert("popup")
	stack.SetMarginS(zgeo.SizeD(14, 10))
	stack.SetBGColor(zgeo.ColorWhite)
	label := zlabel.New(str)
	label.SetMaxLines(0)
	label.SetColor(target.Native().Color())
	stack.Add(label, zgeo.Center|zgeo.Expand)
	zpresent.PopupView(stack, target, att)
	return true
}

func ReplaceDoubleSquiggliesWithFields(v *FieldView, f *Field, str string) string {
	out := zstr.ReplaceAllCapturesFunc(zstr.InDoubleSquigglyBracketsRegex, str, zstr.RegWithoutMatch, func(fieldName string, index int) string {
		a, findex := FindLocalFieldWithFieldName(v.data, fieldName)
		if findex == -1 {
			zlog.Error("field download", str, ":", "field not found in struct:", fieldName)
			return ""
		}
		return fmt.Sprint(a.Interface())
	})
	return out
}

func updateToolTip(f *Field, structure any, view zview.View) {
	var tipField, tip string
	if zstr.HasPrefix(f.Tooltip, "./", &tipField) {
		// ei, _, findex := zreflect.FieldForName(structure, FlattenIfAnonymousOrZUITag, tipField)
		finfo, found := zreflect.FieldForName(structure, FlattenIfAnonymousOrZUITag, tipField)
		if found {
			tip = fmt.Sprint(finfo.ReflectValue.Interface())
		} else { // can't use tip == "" to check, since field might just be empty
			zlog.Error("updateToolTip: no local field for tip", f.Name, tipField)
		}
	} else if f.Tooltip != "" {
		tip = f.Tooltip
		end := zstr.TailUntilWithRest(tip, "   ", &tip)
		if zstr.HasPrefix(end, "[", &end) && zstr.HasSuffix(end, "]", &end) {
			end = end + zstr.UTFPostModifierForRoundRect
			tip = strings.TrimRight(tip, " ") + "    " + end
		}
	}
	if tip != "" {
		view.Native().SetToolTip(tip)
	}
}

func (v *FieldView) ToData(showError bool) (err error) {
	for _, f := range v.Fields {
		foundView, _, _ := v.FindNamedViewOrInLabelized(f.FieldName)
		if foundView == nil {
			continue
		}
		_, e := v.fieldToDataItem(&f, foundView)
		if e != nil {
			if err == nil {
				err = e
			}
		}
	}
	if showError && err != nil {
		zalert.ShowError(err)
	}
	return
}

func (v *FieldView) parentsDataForMe() any {
	if v.ParentFV == nil {
		return v.data
	}
	data := v.ParentFV.parentsDataForMe()
	if data == nil {
		return v.data
	}
	dataKind := reflect.ValueOf(data).Kind()
	if v.sliceItemIndex != -1 && dataKind == reflect.Slice {
		ritem := reflect.ValueOf(data).Elem().Index(v.sliceItemIndex)
		data = ritem.Addr().Interface()
		return data
	}
	finfo := zreflect.FieldForIndex(data, FlattenIfAnonymousOrZUITag, v.params.Field.Index)
	if finfo.FieldIndex == -1 {
		return v.data
	}
	rval := finfo.ReflectValue
	if rval.CanAddr() {
		rval = rval.Addr()
	}
	return rval.Interface()
}

func (v *FieldView) getRadioGroupValue(f *Field, view zview.View) (any, error) {
	var found any
	var err error
	enum := fieldEnums[f.Radio]
	zcontainer.ViewRangeChildren(view, true, false, func(v zview.View) bool {
		rb, _ := v.(*zradio.RadioButton)
		if rb != nil && rb.Value() {
			// zlog.Info("getRadioGroupValue:", f.Name, f.Radio, rb.ObjectName(), enum)
			for _, e := range enum {
				if rb.ObjectName() == fmt.Sprint(e.Value) {
					found = e.Value
					return false
				}
			}
			err = zlog.Error("Not found", rb.ObjectName)
		}
		return true
	})
	return found, err
}

func (v *FieldView) fieldToDataItem(f *Field, view zview.View) (value reflect.Value, err error) {
	if f.IsStatic() {
		return
	}
	data := v.parentsDataForMe()
	finfo := zreflect.FieldForIndex(data, FlattenIfAnonymousOrZUITag, f.Index)

	if f.WidgetName != "" {
		w := widgeters[f.WidgetName]
		if w != nil {
			getter, _ := view.(zview.AnyValueGetter)
			if getter != nil {
				val := getter.ValueAsAny()
				finfo.ReflectValue.Set(reflect.ValueOf(val))
				value = finfo.ReflectValue
			}
			return
		}
	}
	if f.Radio != "" {
		val, e := v.getRadioGroupValue(f, view)
		if e == nil {
			value = reflect.ValueOf(val)
			finfo.ReflectValue.Set(value)
		}
		err = e
		return
	}
	if f.Enum != "" || f.LocalEnum != "" {
		mv, _ := view.(*zmenu.MenuView)
		if mv != nil {
			iface := mv.CurrentValue()
			vo := reflect.ValueOf(iface)
			// zlog.Info("fieldToDataItem1:", iface, f.Name, vo.IsValid(), iface == nil)
			if iface == nil {
				vo = reflect.Zero(finfo.ReflectValue.Type())
			}
			finfo.ReflectValue.Set(vo)
		}
		return finfo.ReflectValue, nil
	}

	switch f.Kind {
	case zreflect.KindBool:
		if f.HasFlag(FlagIsButton) {
			return
		}
		_, got := view.(*zimageview.ImageView)
		if got {
			break
		}
		bv, _ := view.(*zcheckbox.CheckBox)
		if bv == nil {
			zcontainer.ViewRangeChildren(view, false, false, func(v zview.View) bool {
				bv, _ = v.(*zcheckbox.CheckBox)
				return bv == nil
			})
			if bv == nil {
				zlog.Fatal("Should be checkbox", view.Native().Hierarchy(), reflect.TypeOf(view))
			}
		}
		b, _ := finfo.ReflectValue.Addr().Interface().(*bool)
		if b != nil {
			*b = bv.Value().Bool()
		}
		bi, _ := finfo.ReflectValue.Addr().Interface().(*zbool.BoolInd)
		if bi != nil {
			*bi = bv.Value()
		}

	case zreflect.KindInt:
		if finfo.ReflectValue.Type().Name() == "BoolInd" {
			bv, _ := view.(*zcheckbox.CheckBox)
			*finfo.ReflectValue.Addr().Interface().(*zbool.BoolInd) = bv.Value()
		} else {
			tv, _ := view.(*ztext.TextView)
			if f.Flags&FlagAllowEmptyAsZero != 0 {
				if tv.Text() == "" {
					finfo.ReflectValue.SetZero()
					v.updateShowEnableFromZeroer(finfo.ReflectValue.IsZero(), false, tv.ObjectName())
					break
				}
			}
			str := tv.Text()
			if f.PackageName == "time" && finfo.ReflectValue.Type().Name() == "Duration" {
				var secs float64
				secs, err = ztime.GetSecsFromHMSString(str, f.Flags&FlagHasHours != 0, f.Flags&FlagHasMinutes != 0, f.Flags&FlagHasSeconds != 0)
				if err != nil {
					break
				}
				d := finfo.ReflectValue.Addr().Interface().(*time.Duration)
				if d != nil {
					*d = ztime.SecondsDur(secs)
				}
				value = finfo.ReflectValue
				return
			}
			var i64 int64
			if str != "" {
				i64, err = strconv.ParseInt(str, 10, 64)
				if err != nil {
					err = zlog.NewError("Error parsing", f.Name, "in", v.ObjectName())
					break
				}
			}
			zint.SetAny(finfo.ReflectValue.Addr().Interface(), i64)
			v.updateShowEnableFromZeroer(finfo.ReflectValue.IsZero(), false, tv.ObjectName())
		}

	case zreflect.KindFloat:
		tv, _ := view.(*ztext.TextView)
		text := tv.Text()
		var f64 float64
		if f.Flags&FlagAllowEmptyAsZero != 0 {
			if text == "" {
				finfo.ReflectValue.SetZero()
				v.updateShowEnableFromZeroer(finfo.ReflectValue.IsZero(), false, tv.ObjectName())
				break
			}
		}
		f64, err = strconv.ParseFloat(text, 64)
		if err != nil {
			err = zlog.NewError("Error parsing", f.Name, "in", v.ObjectName())
		}
		if err != nil {
			break
		}
		zfloat.SetAny(finfo.ReflectValue.Addr().Interface(), f64)
		v.updateShowEnableFromZeroer(finfo.ReflectValue.IsZero(), false, tv.ObjectName())

	case zreflect.KindTime:
		break

	case zreflect.KindString:
		if (!f.IsStatic() && !v.params.AllStatic) && f.Flags&FlagIsImage == 0 {
			tv, _ := view.(*ztext.TextView)
			if tv == nil {
				zlog.Fatal("Copy Back string not TV:", f.Name)
			}
			text := tv.Text()
			str := finfo.ReflectValue.Addr().Interface().(*string)
			*str = text
			v.updateShowEnableFromZeroer(finfo.ReflectValue.IsZero(), false, tv.ObjectName())
		}

	case zreflect.KindFunc:
		break

	case zreflect.KindSlice:
		if f.StringSep != "" {
			separatedStringToData(f.StringSep, view, finfo.ReflectValue)
		}

	case zreflect.KindStruct:
		zcontainer.ViewRangeChildren(v, true, true, func(view zview.View) bool {
			if view.ObjectName() == f.FieldName {
				fv := viewToFieldView(view)
				if fv != nil {
					cerr := fv.ToData(false)
					if cerr != nil {
						err = cerr
					}
					return false
				}
			}
			return true
		})
		break

	case zreflect.KindMap:
		if !f.HasFlag(FlagIsFixed) {
			return
		}
		// iter := rval.MapRange()
		// for iter.Next() {
		// 	mkey := iter.Key()
		// 	mval := iter.Value()
		// 	zlog.Info("fieldToDataItem map:", mkey, mval)
		// }

	default:
		// zlog.Info("fieldToDataItem for slice:", f.Name)
		break
	}
	value = reflect.ValueOf(finfo.ReflectValue.Addr().Interface()).Elem()
	return
}

func separatedStringToData(sep string, view zview.View, rval reflect.Value) {
	tv, _ := view.(*ztext.TextView)
	seps := sep
	if sep != " " {
		seps += " "
	}
	text := strings.Trim(tv.Text(), seps)
	// a := rval.Interface()
	e := reflect.New(rval.Type().Elem()).Elem()
	zslice.Empty(rval.Addr().Interface())
	if text == "" {
		return
	}
	for _, part := range strings.Split(text, sep) {
		switch zreflect.KindFromReflectKindAndType(e.Kind(), e.Type()) {
		case zreflect.KindString:
			e.SetString(part)
		case zreflect.KindInt:
			n, err := strconv.ParseInt(part, 10, 64)
			if zlog.OnError(err, part) {
				break
			}
			e.SetInt(n)
		case zreflect.KindFloat:
			n, err := strconv.ParseFloat(part, 64)
			if zlog.OnError(err, part) {
				break
			}
			e.SetFloat(n)
		default:
			return
		}
		zslice.AddAtEnd(rval.Addr().Interface(), e.Interface())
	}
}

func viewToFieldView(view zview.View) *FieldView {
	fv, _ := view.(*FieldView)
	if fv != nil {
		return fv
	}
	fo, _ := view.(FieldViewOwner)
	if fo != nil {
		return fo.GetFieldView()
	}
	return nil
}

// ParentFieldView returns the closest FieldView parent to view
func ParentFieldView(view zview.View) *FieldView {
	var got *FieldView
	for _, nv := range view.Native().AllParents() {
		fv := viewToFieldView(nv.View)
		if fv != nil {
			got = fv
		}
	}
	return got
}

func EditStruct[S any](structPtr *S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	EditOrViewStruct(structPtr, false, params, title, att, done)
}

func ViewStruct[S any](structPtr *S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	EditOrViewStruct(structPtr, true, params, title, att, done)
}

func EditOrViewStruct[S any](structPtr *S, isReadOnly bool, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	slice := []S{*structPtr}
	EditOrViewStructSlice(&slice, isReadOnly, params, title, att, func(ok bool) (close bool) {
		// zlog.Info("PresentOKCancelStruct:", ok, slice[0])
		if ok {
			*structPtr = slice[0]
		}
		return done(ok)
	})
}

func (v *FieldView) CreateStoreKeyForField(f *Field, name string) string {
	h := v.ObjectName()
	if f.ValueStoreKey != "" {
		h = f.ValueStoreKey
	}
	return zstr.Concat("/", h, name)
}

func (v *FieldView) OpenGUIFromPathParts(parts []zdocs.PathPart) bool {
	fieldName := parts[0].PathStub
	child, _, _ := v.FindNamedViewOrInLabelized(fieldName)
	if child != nil {
		if len(parts) == 1 {
			old := v.BGColor()
			child.SetBGColor(zgeo.ColorYellow)
			child.Native().Focus(true)
			ztimer.StartIn(1.2, func() {
				child.SetBGColor(old)
			})
			return true
		}
		o, _ := child.(zdocs.GUIPartOpener)
		if o != nil {
			return o.OpenGUIFromPathParts(parts[1:])
		}
	}
	return false
}

func (v *FieldView) GetSearchableItems(currentPath []zdocs.PathPart) []zdocs.SearchableItem {
	var parts []zdocs.SearchableItem
	// zlog.Info("FV.GetSearchableItems", v.Hierarchy(), v.frameTitle)
	path := currentPath
	if v.frameTitle != "" {
		item := zdocs.MakeSearchableItem(currentPath, zdocs.StaticField, "", "", v.frameTitle)
		parts = append(parts, item)
		path = zdocs.AddedPath(currentPath, zdocs.StaticField, v.frameTitle, v.frameTitle)
	}
	ForEachField(v.data, v.params.FieldParameters, v.Fields, func(each FieldInfo) bool {
		view, label, _ := v.FindNamedViewOrInLabelized(each.Field.FieldName)
		var labelText string
		if label != nil && label.Native().IsSearchable() {
			labelText = each.Field.TitleOrName()
			item := zdocs.MakeSearchableItem(path, zdocs.StaticField, "", "", labelText)
			parts = append(parts, item)
		}
		if view != nil && view.Native().IsSearchable() {
			o, _ := view.(ztext.TextOwner)
			// zlog.Info("FV.GetSearchableItems", v.Hierarchy(), each.Field.FieldName, o != nil, reflect.TypeOf(view))
			if o != nil {
				text := o.Text()
				if text != "" && text != labelText {
					stub := each.Field.FieldName
					titleOrName := each.Field.TitleOrName()
					if titleOrName == text {
						stub = ""
						titleOrName = ""
					}
					item := zdocs.MakeSearchableItem(path, zdocs.ValueField, titleOrName, stub, text)
					parts = append(parts, item)
				}
			} else {
				sig, _ := view.(zdocs.SearchableItemsGetter)
				if sig != nil {
					subPath := zdocs.AddedPath(path, zdocs.StaticField, each.Field.TitleOrName(), each.Field.FieldName)
					items := sig.GetSearchableItems(subPath)
					parts = append(parts, items...)
				}
			}
		}
		return true
	})
	return parts
}

func MakeSearchableFieldItems(path []zdocs.PathPart, f *Field) []zdocs.SearchableItem {
	var parts []zdocs.SearchableItem
	item := zdocs.MakeSearchableItem(path, zdocs.StaticField, "", "", f.Name)
	parts = append(parts, item)
	subPath := zdocs.AddedPath(path, zdocs.StaticField, f.Name, f.FieldName)

	if f.Tooltip != "" && !strings.HasPrefix(f.Tooltip, ".") {
		item := zdocs.MakeSearchableItem(subPath, zdocs.StaticField, "Tip", "Tip", f.Tooltip)
		parts = append(parts, item)
	}
	if f.Description != "" {
		item := zdocs.MakeSearchableItem(subPath, zdocs.StaticField, "Desc", "Desc", f.Description)
		parts = append(parts, item)
	}
	return parts
}
