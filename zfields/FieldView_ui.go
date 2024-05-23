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
	"github.com/torlangballe/zui/zaudio"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zmenu"
	"github.com/torlangballe/zui/zpresent"
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
	"github.com/torlangballe/zutil/zint"
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
}

type FieldViewParameters struct {
	Field
	FieldParameters
	BuildChildrenHidden      bool
	ImmediateEdit            bool                                 // ImmediateEdit forces immediate write-to-data when editing a field.
	MultiSliceEditInProgress bool                                 // MultiSliceEditInProgress is on if the field represents editing multiple structs in a list. Checkboxes can be indeterminate etc.
	EditWithoutCallbacks     bool                                 // Set so not get edit/changed callbacks when editing. Example: Dialog box editing edits a copy, so no callbacks needed.
	IsEditOnNewStruct        bool                                 // IsEditOnNewStruct when an just-created struct is being edited. Menus can have a storage key-value to set last-used option then for example
	triggerHandlers          map[trigger]func(ap ActionPack) bool // triggerHandlers is a map of functions to call if an action occurs in this FieldView. Somewhat replacing ActionHandler
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

var (
	fieldViewEdited = map[string]time.Time{}
	textFilters     = map[string]func(string) string{}
	EnableLog       zlog.Enabler
)

func init() {
	zlog.RegisterEnabler("zfields.LogGUI", &EnableLog)

	RegisterTextFilter("$nowhite", func(s string) string {
		return zstr.WhitespaceRemover.Replace(s)
	})
	RegisterTextFilter("$lower", strings.ToLower)
	RegisterTextFilter("$upper", strings.ToLower)
	RegisterTextFilter("$uuid", zstr.CreateFilterFunc(zstr.IsRuneValidInUUID))
	RegisterTextFilter("$hex", zstr.CreateFilterFunc(zstr.IsRuneHex))
	RegisterTextFilter("$alpha", zstr.CreateFilterFunc(zstr.IsRuneASCIIAlpha))
	RegisterTextFilter("$num", zstr.CreateFilterFunc(zstr.IsRuneASCIINumeric))
	RegisterTextFilter("$alphanum", zstr.CreateFilterFunc(zstr.IsRuneASCIIAlphaNumeric))
}

func RegisterTextFilter(name string, filter func(string) string) {
	textFilters[name] = filter
}

func GetTextFilter(name string) func(string) string {
	return textFilters[name]
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

func (v *FieldView) IsSlice() bool {
	return reflect.ValueOf(v.data).Elem().Kind() == reflect.Slice
}

func setFieldViewEdited(fv *FieldView) {
	fieldViewEdited[fv.Hierarchy()] = time.Now()
}

func FieldViewParametersDefault() (f FieldViewParameters) {
	f.ImmediateEdit = true
	f.Styling = zstyle.EmptyStyling
	f.Styling.Spacing = 2
	return f
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

func makeFrameIfFlag(f *Field, fv *FieldView) zview.View {
	if f.Flags&FlagHasFrame == 0 {
		return nil
	}
	var title string
	if f.Flags&FlagFrameIsTitled != 0 {
		title = f.TitleOrName()
	}
	frame := zcontainer.StackViewVert("frame")
	zguiutil.MakeStackATitledFrame(frame, title, f.Flags&FlagFrameTitledOnFrame != 0, f.Styling, f.Styling)
	frame.Add(fv, zgeo.TopLeft)
	return frame
}

func fieldViewNew(id string, vertical bool, data any, params FieldViewParameters, marg zgeo.Size, parent *FieldView) *FieldView {
	v := &FieldView{}
	v.StackView.Init(v, vertical, id)

	v.SetSpacing(params.Styling.Spacing)

	v.data = data
	v.sliceItemIndex = -1
	zreflect.ForEachField(v.data, FlattenIfAnonymousOrZUITag, func(each zreflect.FieldInfo) bool {
		f := EmptyField
		if !f.SetFromReflectValue(each.ReflectValue, each.StructField, each.FieldIndex, params.FieldParameters) {
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
	v.ID = id
	v.ParentFV = parent

	return v
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

func (v *FieldView) FindNamedViewOrInLabelized(name string) (view, maybeLabel zview.View) {
	for _, c := range (v.View.(zcontainer.ChildrenOwner)).GetChildren(false) {
		n := c.ObjectName()
		if n == name {
			return c, c
		}
		if strings.HasPrefix(n, "$labelize.") {
			s, _ := c.(*zcontainer.StackView)
			if s != nil {
				v, _ := s.FindViewWithName(name, true)
				if v != nil {
					return v, c
				}
			}
		}
	}
	return nil, nil
}

func (v *FieldView) updateShowEnableFromZeroer(isZero, isShow bool, toID string) {
	if v.params.MultiSliceEditInProgress {
		return
	}
	for _, f := range v.Fields {
		var id string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		if zstr.HasPrefix(local, "./", &id) && id == toID {
			_, foundView := v.FindNamedViewOrInLabelized(f.FieldName)
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
			fv := view.(*FieldView)
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
		// zlog.Info(EnableLog, "FV Update field:", each.Field.Name, recentEdit, each.Field.IsStatic(), each.StructField.Type.Kind()) // EnableLog,
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
	foundView, flabelized := v.FindNamedViewOrInLabelized(f.FieldName)
	if foundView == nil {
		return true
	}
	// zlog.Info("updateField2:", v.Hierarchy(), sf.Name)
	v.updateShowEnableOnView(flabelized, true, foundView.ObjectName())
	v.updateShowEnableOnView(flabelized, false, foundView.ObjectName())
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
	if f.Enum != "" && (f.IsStatic() || v.params.AllStatic) {
		enum := GetEnum(f.Enum)
		str := findNameOfEnumForRVal(rval, enum)
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
	if menuType != nil && ((f.Enum != "") || f.LocalEnum != "") { // && f.Kind != zreflect.KindSlice
		var enum zdict.Items
		if f.Enum != "" {
			enum, _ = fieldEnums[f.Enum]
			// zlog.Info("updateMenu2:", v.Hierarchy(), sf.Name, enum)
		} else {
			ei, findex := FindLocalFieldWithFieldName(v.data, f.LocalEnum)
			zlog.Assert(findex != -1, f.Name, f.LocalEnum)
			var err error
			enum, err = getDictItemsFromSlice(ei, f)
			if err != nil {
				return false
			}
		}
		if v.params.MultiSliceEditInProgress {
			var rtype reflect.Type
			var hasZero bool
			for _, item := range enum {
				if item.Value != nil {
					rval := reflect.ValueOf(item)
					rtype = rval.Type()
					if rval.IsZero() {
						hasZero = true
						break
					}
				}
			}
			if !hasZero && rtype != nil && rtype.Kind() != reflect.Invalid {
				item := zdict.Item{Name: "", Value: reflect.Zero(rtype)}
				enum = append(zdict.Items{item}, enum...)
			}
		}
		menuType.UpdateItems(enum, rval.Interface(), f.Flags&FlagIsActions != 0)
		return true
	}
	updateItemLocalToolTip(f, v.data, foundView)
	to, _ := foundView.(ztext.TextOwner)
	zuistringer, _ := rval.Interface().(UIStringer)
	if to != nil && zuistringer != nil {
		// if f.IsStatic() || v.params.AllStatic {
		to.SetText(zuistringer.ZUIString())
		return true
	}
	switch f.Kind {
	case zreflect.KindMap:
		if f.HasFlag(FlagShowSliceCount) {
			label, _ := foundView.(*zlabel.Label)
			label.SetText(strconv.Itoa(rval.Len()))
			return true
		} else {
			v.updateMapList(rval, foundView)
		}

	case zreflect.KindSlice:
		if f.StringSep != "" {
			v.updateSeparatedStringWithSlice(f, rval, foundView)
			break
		}
		if f.HasFlag(FlagShowSliceCount) {
			label, _ := foundView.(*zlabel.Label)
			label.SetText(strconv.Itoa(rval.Len()))
			return true
		}

		// zlog.Info("updateFieldSlice:", v.Hierarchy(), sf.Name)
		sv, _ := foundView.(*FieldSliceView)
		if sv == nil {
			zlog.Error("UpdateSlice: not a *FieldSliceView:", f.Name, v.Hierarchy(), reflect.TypeOf(foundView))
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
		to := foundView.(ztext.TextOwner)
		to.SetText(valStr)

	case zreflect.KindStruct:
		fv, _ := foundView.(*FieldView)
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
		valStr, dur = getTextFromNumberishItem(rval, f)
		v.setText(f, valStr, foundView)
		label, _ := foundView.(*zlabel.Label)
		if label != nil {
			updateOldDuration(label, dur, f)
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
			if f.IsStatic() && f.Flags&FlagIsFixed != 0 {
				valStr = f.Name
			}
			v.setText(f, valStr, foundView)
		}
	}
	return true
}

func (v *FieldView) buildMapList(rval reflect.Value, f *Field) zview.View {
	var outView zview.View
	params := v.params
	params.triggerHandlers = zmap.EmptyOf(params.triggerHandlers)
	stackFV := fieldViewNew(f.FieldName, true, rval.Interface(), params, zgeo.Size{}, v)
	outView = stackFV
	stackFV.GridVerticalSpace = math.Max(6, v.Spacing())
	stackFV.SetSpacing(f.Styling.SpacingOrMax(12))
	stackFV.params.SetFlag(FlagIsLabelize)
	frame := makeFrameIfFlag(f, stackFV)
	if frame != nil {
		outView = frame
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
		var mf Field
		mf.Kind = zreflect.KindMap
		mf.Name = zlocale.FirstToTitleCaseExcept(key, "")
		mf.FieldName = key
		mf.SetFlag(FlagIsLabelize)
		a := zgeo.Left
		view := stackFV.buildItem(&mf, mval, i, a, zgeo.Size{}, true)
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
		}
		i++
	}
	return outView
}

func (v *FieldView) updateMapList(rval reflect.Value, foundView zview.View) {

}

func (v *FieldView) updateSeparatedStringWithSlice(f *Field, rval reflect.Value, foundView zview.View) {
	var parts []string

	for i := 0; i < rval.Len(); i++ {
		v := rval.Index(i).Interface()
		parts = append(parts, fmt.Sprint(v))
	}
	str := strings.Join(parts, f.StringSep)
	v.setText(f, str, foundView)
}

func (v *FieldView) setText(f *Field, valStr string, foundView zview.View) {
	if f.IsStatic() || v.params.AllStatic {
		label, _ := foundView.(*zlabel.Label)
		if label != nil {
			label.SetText(valStr)
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
		f, _ := view.(*FieldView)
		if f != nil {
			if optionalID == "" || f.ID == optionalID {
				fv = f
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

func (v *FieldView) Rebuild() {
	// zlog.Info("FV.Rebuild:", v.data != nil, reflect.ValueOf(v.data).Kind())
	fview := FieldViewNew(v.ID, v.data, v.params)
	fview.Build(true)
	rep, _ := v.Parent().View.(zview.ChildReplacer)
	if rep != nil {
		rep.ReplaceChild(v, fview)
	}
	zcontainer.ArrangeChildrenAtRootContainer(v)
}

func (v *FieldView) CallFieldAction(fieldID string, action ActionType, fieldValue interface{}) {
	view, _ := v.FindViewWithName(fieldID, false)
	if view == nil {
		zlog.Error("CallFieldAction find view", fieldID)
		return
	}
	f := v.findFieldWithFieldName(fieldID)
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
				fv, _ := parent.View.(*FieldView)
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
			changed := false
			sv := reflect.ValueOf(ap.FieldView.data)
			if sv.Kind() == reflect.Ptr || sv.CanAddr() {
				// Here we run thru the possiblly new struct again, and find the item with same id as field
				finfo, found := zreflect.FieldForName(ap.FieldView.data, FlattenIfAnonymousOrZUITag, ap.Field.FieldName)
				if found {
					changed = true
					if finfo.ReflectValue.CanAddr() {
						fieldAddress = finfo.ReflectValue.Addr().Interface()
					}
				}
			}
			if !changed {
				zlog.Info("NOOT!!!", ap.Field.FN(), ap.Action, ap.FieldView.data != nil)
				zlog.Fatal("Not CHANGED!", ap.Field.FN())
			}
		}
		aih, _ := ap.RVal.Interface().(ActionHandler)
		if aih == nil && fieldAddress != nil {
			aih, _ = fieldAddress.(ActionHandler)
		}
		if aih != nil {
			result = aih.HandleAction(ap)
		}
	}
	return result
}

func (v *FieldView) findFieldWithFieldName(fn string) *Field {
	for i, f := range v.Fields {
		if f.FieldName == fn {
			return &v.Fields[i]
		}
	}
	return nil
}

func (fv *FieldView) makeButton(rval reflect.Value, f *Field) *zshape.ImageButtonView {
	format := f.Format
	if format == "" {
		format = "%v"
	}
	color := "gray"
	if len(f.Colors) > 0 {
		color = f.Colors[0]
	}
	var textCol zgeo.Color
	if len(f.Colors) > 1 {
		textCol = zgeo.ColorFromString(f.Colors[1])
	}
	if !textCol.Valid {
		fg := zgeo.ColorFromString(color)
		if fg.Valid {
			textCol = fg.ContrastingGray()
		} else {
			textCol = zgeo.ColorBlack
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
	button := zshape.ImageButtonViewNew(name, color, s, zgeo.SizeNull)
	button.SetTextColor(textCol)
	button.TextXMargin = 0
	if f.HasFlag(FlagIsURL) {
		button.SetPressedHandler(func() {
			surl := replaceDoubleSquiggliesWithFields(fv, f, f.Path)
			zwindow.GetMain().SetLocation(surl)
		})
	}
	if f.RPCCall != "" {
		button.SetPressedHandler(func() {
			go func() {
				var reply string
				a := reflect.ValueOf(fv.data).Elem().Interface()
				err := zrpc.MainClient.Call(f.RPCCall, a, &reply)
				if err != nil {
					zalert.ShowError(err)
				}
				if reply != "" {
					zalert.Show(reply)
				}
			}()
		})
	}
	return button
}

func (v *FieldView) makeMenu(rval reflect.Value, f *Field, items zdict.Items) zview.View {
	var view zview.View
	static := f.IsStatic() //|| v.params.AllStatic
	isSlice := rval.Kind() == reflect.Slice
	// zlog.Info("makeMenu", f.Name, f.IsStatic(), v.params.AllStatic, isSlice, rval.Kind(), len(items))
	if static || isSlice {
		// multi := isSlice
		isImage := (f.ImageFixedPath != "")
		shape := zshape.TypeRoundRect
		if isImage {
			shape = zshape.TypeNone
		}
		menuOwner := zmenu.NewMenuedOwner()
		menuOwner.IsStatic = static
		menuOwner.IsMultiple = isSlice
		if f.HasFlag(FlagIsEdit) {
			menuOwner.AddValueFunc = func() any {
				zlog.Assert(isSlice)
				e := zslice.MakeAnElementOfSliceRValType(rval)
				return e
			}
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
		mItems := zmenu.MOItemsFromZDictItemsAndValues(items, rval.Interface(), f.Flags&FlagIsActions != 0)

		menu := zmenu.MenuOwningButtonCreate(menuOwner, mItems, shape)
		if isImage {
			menu.SetImage(nil, true, f.ImageFixedPath, nil)
			menu.ImageMaxSize = f.Size
		} else {
			if len(f.Colors) != 0 {
				menu.SetColor(zgeo.ColorFromString(f.Colors[0]))
			}
		}
		menu.Ratio = 0.3
		view = menu
		menuOwner.SelectedHandlerFunc = func() {
			sel := menuOwner.SelectedItem()
			if sel != nil {
				kind := reflect.ValueOf(sel.Value).Kind()
				if menuOwner.IsStatic {
					if kind != reflect.Ptr && kind != reflect.Struct {
						callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: PressedAction, RVal: rval, View: &view})
					}
				} else {
					zlog.Info("Here!", menuOwner.IsMultiple, rval, rval.Type())
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
	} else {
		name := f.Name + "Menu"
		if v.params.IsEditOnNewStruct && f.ValueStoreKey != "" && !zstr.StringsContain(v.params.UseInValues, RowUseInSpecialName) {
			name = "key:" + f.ValueStoreKey
		}
		menu := zmenu.NewView(name, items, rval.Interface())
		menu.SetMaxWidth(f.MaxWidth)
		view = menu
		menu.SetSelectedHandler(func() {
			val, _ := v.fieldToDataItem(f, menu)
			isZero := true
			if val.IsValid() {
				isZero = val.IsZero()
			}
			// zlog.Info("Menu Edited", v.Hierarchy(), f.Name, isZero, val, menu.CurrentValue())
			v.updateShowEnableFromZeroer(isZero, true, menu.ObjectName())
			v.updateShowEnableFromZeroer(isZero, false, menu.ObjectName())
			ap := ActionPack{Field: f, Action: EditedAction, RVal: rval, View: &view}
			v.callTriggerHandler(ap)
			callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: EditedAction, RVal: rval, View: &view})
		})
	}
	return view
}

func getTimeString(rval reflect.Value, f *Field) string {
	var str string
	t := rval.Interface().(time.Time)
	if t.IsZero() {
		return ""
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
		str = ztime.GetNice(t, f.Flags&FlagHasSeconds != 0)
	} else {
		t = ztime.GetTimeWithServerLocation(t)
		str = t.Format(format)
	}
	return str
}

func getTextFromNumberishItem(rval reflect.Value, f *Field) (string, time.Duration) {
	if f.Flags&FlagAllowEmptyAsZero != 0 {
		if rval.IsZero() {
			return "", 0
		}
	}
	stringer, got := rval.Interface().(UIStringer)
	if got {
		return stringer.ZUIString(), 0
	}
	zkind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	isDurTime := zkind == zreflect.KindTime && f.Flags&FlagIsDuration != 0
	if zkind == zreflect.KindTime && !isDurTime {
		return getTimeString(rval, f), 0
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
		// zlog.Info("DurTime", dur, t, f.HasFlag(FlagHasHours))
		return str, dur
	}
	format := f.Format
	significant := f.Columns
	switch format {
	case "memory":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			if b == 0 && f.Flags&FlagAllowEmptyAsZero != 0 {
				return "", 0
			}
			return zwords.GetMemoryString(b, "", significant), 0
		}
	case "storage":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			if b == 0 && f.Flags&FlagAllowEmptyAsZero != 0 {
				return "", 0
			}
			return zwords.GetStorageSizeString(b, "", significant), 0
		}
	case "bps":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			if b == 0 && f.Flags&FlagAllowEmptyAsZero != 0 {
				return "", 0
			}
			return zwords.GetBandwidthString(b, "", significant), 0
		}
	case "":
		format = "%v"
	}
	return fmt.Sprintf(format, rval.Interface()), 0
}

func (v *FieldView) makeText(rval reflect.Value, f *Field, noUpdate bool) zview.View {
	var str string
	if f.IsStatic() || v.params.AllStatic {
		var label *zlabel.Label
		isLink := f.HasFlag(FlagIsURL)
		if !isLink {
			label = zlabel.New("")
		}
		if isLink || f.HasFlag(FlagIsDocumentation) {
			str, _ := getTextFromNumberishItem(rval, f)
			surl := str
			if f.Path != "" {
				surl = f.Path
			}
			if isLink {
				label = zlabel.NewLink(str, surl)
			} else {
				// zlog.Info("DOC:", surl, ztextinfo.DecorationUnderlined)
				ztext.SetTextDecoration(&label.NativeView, ztextinfo.DecorationUnderlined)
				label.SetPressedHandler(func() {
					go zwidgets.DocumentationViewPresent(surl, false)
				})
			}
		}
		label.SetMaxLines(f.Rows)
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
		label.SetWrap(ztextinfo.WrapTailTruncate)
		if f.Flags&FlagToClipboard != 0 {
			ztext.MakeViewPressToClipboard(label)
		}
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
		var funcs []func(string) string
		for _, fname := range f.Filters {
			fn := GetTextFilter(fname)
			if fn == nil {
				zlog.Error("No registerd text filter for:", fname, f.FieldName)
				continue
			}
			funcs = append(funcs, fn)
		}
		tv.FilterFunc = func(s string) string {
			for _, fn := range funcs {
				s = fn(s)
			}
			return s
		}
	}
	tv.SetPlaceholder(f.Placeholder)
	tv.SetValueHandler("zfields.Filter", func(edited bool) {
		// zlog.Info("Changed:", tv.Text())
		v.fieldHandleValueChanged(f, edited, tv.View)
	})
	return tv
}

func (v *FieldView) fieldHandleValueChanged(f *Field, edited bool, view zview.View) {
	rval, err := v.fieldToDataItem(f, view)
	if zlog.OnError(err) {
		return
	}
	ap := ActionPack{Field: f, Action: EditedAction, RVal: rval, View: &view}
	if v.callTriggerHandler(ap) {
		return
	}
	callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: EditedAction, RVal: rval, View: &view})
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
	if !v.params.Field.HasFlag(FlagIsLabelize) && !zstr.StringsContain(v.params.UseInValues, RowUseInSpecialName) {
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
		iv.SetPressedHandler(func() {
			surl := replaceDoubleSquiggliesWithFields(v, f, f.Path)
			zwindow.GetMain().SetLocation(surl)
		})
		return iv
	}
	if f.IsImageToggle() && rval.Kind() == reflect.Bool {
		iv.SetPressedHandler(func() {
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
	iv.SetPressedHandler(func() {
		ap := ActionPack{Field: f, Action: PressedAction, RVal: reflect.ValueOf(f.FieldName), View: &iv.View}
		v.callTriggerHandler(ap)
	})
	return iv
}

func setColorFromField(view zview.View, f *Field) {
	col := zstyle.DefaultFGColor()
	if f.Styling.FGColor.Valid {
		col = f.Styling.FGColor
	}
	view.SetColor(col)
}

func (v *FieldView) updateOldTime(label *zlabel.Label, f *Field) {
	finfo, found := zreflect.FieldForName(v.data, FlattenIfAnonymousOrZUITag, f.FieldName)
	if found {
		t := finfo.ReflectValue.Interface().(time.Time)
		updateOldDuration(label, time.Since(t), f)
	}
}

func updateOldDuration(label *zlabel.Label, dur time.Duration, f *Field) {
	if f.OldSecs != 0 && ztime.DurSeconds(dur) > float64(f.OldSecs) {
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
		return zdict.Items{}, zlog.Error("field isn't enum, not Item(s)Getter type", f.Name, f.LocalEnum)
	}
	if items == nil {
		return getter.GetItems(), nil
	}
	return *items, nil
}

func (v *FieldView) createSpecialView(rval reflect.Value, f *Field) (view zview.View, skip bool) {
	if f.Flags&FlagIsButton != 0 {
		if v.params.HideStatic {
			return nil, true
		}
		return v.makeButton(rval, f), false
	}
	if f.WidgetName != "" && rval.Kind() != reflect.Slice {
		w := widgeters[f.WidgetName]
		if w != nil {
			widgetView := w.Create(f)
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
					enum[i].Name = ""
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
	// zlog.Info("FV BuildStack:", name, v.Hierarchy())
	if v.params.Field.HasFlag(FlagIsLabelize) {
		v.GridVerticalSpace = math.Max(6, v.Spacing())
		v.SetSpacing(math.Max(12, v.Spacing()))
	}
	zlog.Assert(reflect.ValueOf(v.data).Kind() == reflect.Ptr, name, reflect.ValueOf(v.data).Kind())

	ForEachField(v.data, v.params.FieldParameters, v.Fields, func(each FieldInfo) bool {
		v.buildItem(each.Field, each.ReflectValue, each.FieldIndex, defaultAlign, cellMargin, useMinWidth)
		return true
	})
}

func (v *FieldView) buildItem(f *Field, rval reflect.Value, index int, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) zview.View {
	// zlog.Info("BuildItem:", f.Name, rval.Interface(), index)
	if !f.Margin.IsNull() {
		cellMargin = f.Margin
	}
	// labelizeWidth := v.params.LabelizeWidth
	// parentFV := ParentFieldView(v)
	// if parentFV != nil && v.params.LabelizeWidth == 0 {
	// 	labelizeWidth = parentFV.params.LabelizeWidth
	// }
	exp := zgeo.AlignmentNone
	if v.params.HideStatic && f.IsStatic() {
		return nil
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
			fieldView := fieldViewNew(f.FieldName, vert, rval.Addr().Interface(), params, zgeo.SizeNull, v)
			view = makeFrameIfFlag(f, fieldView)
			if view == nil {
				view = fieldView
			}
			fieldView.BuildStack(f.FieldName, zgeo.TopLeft, zgeo.SizeNull, true)

		case zreflect.KindBool:
			if f.Flags&FlagIsImage != 0 && f.IsImageToggle() && rval.Kind() == reflect.Bool {
				view = v.makeImage(rval, f)
				break
			}
			if f.HasFlag(FlagIsAudio) {
				s := f.Size
				if s.IsNull() {
					s = zgeo.SizeBoth(20)
				}
				path := replaceDoubleSquiggliesWithFields(v, f, f.Path)
				av := zaudio.NewAudioIconView(s, path)
				av.SetObjectName(f.FieldName)
				view = av
				break
			}
			b := zbool.ToBoolInd(rval.Interface().(bool))
			exp = zgeo.AlignmentNone
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
			view = v.buildMapList(rval, f)

		case zreflect.KindSlice:
			if !f.HasFlag(FlagIsGroup) || zstr.StringsContain(v.params.UseInValues, RowUseInSpecialName) {
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
			view = v.NewSliceView(rval.Addr().Interface(), f)

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
					timer := ztimer.RepeatNow(1, func() bool {
						// zlog.Info("Repeat time update", zlog.Pointer(nat), nat.Hierarchy())
						nlabel := view.(*zlabel.Label)
						if f.Flags&FlagIsDuration != 0 {
							v.updateSinceTime(nlabel, f)
						} else {
							v.updateOldTime(nlabel, f)
						}
						return true
					})
					// v.AddOnRemoveFunc(timer.Stop)
					view.Native().AddOnRemoveFunc(timer.Stop)
				} else {
					if f.Format == "nice" {
						timer := ztimer.StartAt(ztime.OnTheNextHour(time.Now()), func() {
							nlabel := view.(*zlabel.Label)
							str, _ := getTextFromNumberishItem(rval, f)
							nlabel.SetText(str)
						})
						v.AddOnRemoveFunc(timer.Stop)
					}
				}
			}

		default:
			panic(fmt.Sprintln("buildStack bad type:", f.Name, kind))
		}
	}
	zlog.Assert(view != nil)
	pt, _ := view.(zview.Pressable)
	if pt != nil {
		nowItem := rval           // store item in nowItem so closures below uses right item
		ph := pt.PressedHandler() // get old handler before we set it to override
		if f.HasFlag(FlagShowPopup) {
			pt.SetPressedDownHandler(func() {
				v.popupContent(view, f)
			})
		} else {
			pt.SetPressedHandler(func() {
				if !callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: PressedAction, RVal: nowItem, View: &view}) && ph != nil {
					ph()
				}
			})
		}
		if f.Flags&FlagLongPress != 0 {
			lph := pt.LongPressedHandler() // get old handler before we set it to override
			pt.SetLongPressedHandler(func() {
				if !callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: LongPressedAction, RVal: nowItem, View: &view}) && lph != nil {
					lph()
				}
			})
		}
	}
	updateItemLocalToolTip(f, v.data, view)
	if !f.Styling.DropShadow.Delta.IsNull() {
		nv := view.Native()
		nv.SetDropShadow(f.Styling.DropShadow)
	}
	view.SetObjectName(f.FieldName)
	if f.Styling.FGColor.Valid {
		view.SetColor(f.Styling.FGColor)
	}
	if f.Styling.BGColor.Valid {
		view.SetBGColor(f.Styling.BGColor)
	} else if f.HasFlag(FlagIsForZDebugOnly) {
		view.SetBGColor(zstyle.DebugBackgroundColor)
	}
	if rval.CanAddr() {
		callActionHandlerFunc(ActionPack{FieldView: v, Field: f, Action: CreatedViewAction, RVal: rval.Addr(), View: &view})
	}
	if f.Path != "" {
		path := replaceDoubleSquiggliesWithFields(v, f, f.Path)
		p := view.(zview.Pressable)
		if p != nil {
			if f.HasFlag(FlagIsDownload) {
				if !f.HasFlag(FlagIsAudio) {
					p.SetPressedHandler(func() {
						zview.DownloadURI(path, "")
					})
				}
			}
		}
	}
	cell := &zcontainer.Cell{}
	def := defaultAlign
	all := zgeo.Left | zgeo.HorCenter | zgeo.Right
	if f.Alignment&all != 0 {
		def &= ^all
	}
	cell.Margin = cellMargin
	cell.Alignment = def | exp | f.Alignment
	// doLabelize := (labelizeWidth != 0 || f.LabelizeWidth < 0) && !f.HasFlag(FlagNoLabel)
	// zlog.Info("CELLMARGIN:", f.Name, cellMargin, cell.Alignment)
	var lstack *zcontainer.StackView
	if v.params.Field.HasFlag(FlagIsLabelize) {
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
		// zlog.Info("LAB:", title, view.ObjectName(), f.FieldName)
		var label *zlabel.Label
		label, lstack, cell, _ = zguiutil.Labelize(view, title, 0, cell.Alignment, desc)
		if f.HasFlag(FlagLockable) {
			if !zlog.ErrorIf(view.ObjectName() == "", f.FieldName) {
				lock := zguiutil.CreateLockIconForView(view)
				lstack.AddAdvanced(lock, zgeo.CenterRight, zgeo.SizeD(-7, 7), zgeo.Size{}, -1, true).RelativeToName = view.ObjectName()
				// zlog.Info("Lock relative:", view.ObjectName(), len(lstack.GetChildren(true)))
			}
		}
		if f.HasFlag(FlagIsForZDebugOnly) {
			label.SetBGColor(zstyle.DebugBackgroundColor)
			label.SetCorner(4)
		}
		updateItemLocalToolTip(f, v.data, lstack)
		v.Add(lstack, zgeo.HorExpand|zgeo.Left|zgeo.Top)
	}
	if useMinWidth {
		cell.MinSize.W = f.MinWidth
	}
	cell.MaxSize.W = f.MaxWidth
	if v.params.BuildChildrenHidden {
		view.Show(false)
	}
	if !v.params.Field.HasFlag(FlagIsLabelize) {
		cell.View = view
		v.AddCell(*cell, -1)
	}
	return view
}

func (fv *FieldView) freshRValue(fieldName string) reflect.Value {
	finfo, found := zreflect.FieldForName(fv.data, FlattenIfAnonymousOrZUITag, fieldName)
	zlog.Assert(found, fieldName)
	return finfo.ReflectValue
}

func (fv *FieldView) popupContent(target zview.View, f *Field) {
	rval := fv.freshRValue(f.FieldName)
	a := rval.Interface()
	str := fmt.Sprint(a)
	zs, _ := a.(UIStringer)
	if zs != nil {
		str = zs.ZUIString()
	}
	// zlog.Info("popupContent:", str)
	if str == "" {
		return
	}
	att := zpresent.AttributesNew()
	att.Alignment = zgeo.TopLeft // | zgeo.HorOut
	att.PlaceOverMargin = zgeo.SizeD(-8, -4)
	stack := zcontainer.StackViewVert("popup")
	stack.SetMarginS(zgeo.SizeD(14, 10))
	stack.SetBGColor(zgeo.ColorWhite)
	label := zlabel.New(str)
	label.SetColor(target.Native().Color())
	stack.Add(label, zgeo.Center|zgeo.Expand)
	zpresent.PopupView(stack, target, att)
}

func replaceDoubleSquiggliesWithFields(v *FieldView, f *Field, str string) string {
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

func updateItemLocalToolTip(f *Field, structure any, view zview.View) {
	var tipField, tip string
	if zstr.HasPrefix(f.Tooltip, "./", &tipField) {
		// ei, _, findex := zreflect.FieldForName(structure, FlattenIfAnonymousOrZUITag, tipField)
		finfo, found := zreflect.FieldForName(structure, FlattenIfAnonymousOrZUITag, tipField)
		if found {
			tip = fmt.Sprint(finfo.ReflectValue.Interface())
		} else { // can't use tip == "" to check, since field might just be empty
			zlog.Error("updateItemLocalToolTip: no local field for tip", f.Name, tipField)
		}
	} else if f.Tooltip != "" {
		tip = f.Tooltip
	}
	if tip != "" {
		view.Native().SetToolTip(tip)
	}
}

func (v *FieldView) ToData(showError bool) (err error) {
	for _, f := range v.Fields {
		foundView, _ := v.FindNamedViewOrInLabelized(f.FieldName)
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

func (v *FieldView) fieldToDataItem(f *Field, view zview.View) (value reflect.Value, err error) {
	if f.IsStatic() {
		return
	}
	data := v.data
	if v.sliceItemIndex != -1 {
		zlog.Assert(v.params.Kind == zreflect.KindSlice)
		ritem := reflect.ValueOf(v.ParentFV.data).Elem().Index(v.sliceItemIndex)
		data = ritem.Addr().Interface()
	}
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
			// zlog.Info("fieldToData:", f.Name, text)
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
				fv, _ := view.(*FieldView)
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
		fv := findSubFieldView(view, "")
		zlog.Info("fieldToDataItem map:", view.Native().Hierarchy(), reflect.TypeOf(view), fv != nil, f.Name)
		return
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

// ParentFieldView returns the closest FieldView parent to view
func ParentFieldView(view zview.View) *FieldView {
	var got *FieldView
	for _, nv := range view.Native().AllParents() {
		fv, _ := nv.View.(*FieldView)
		if fv != nil {
			got = fv
		}
	}
	return got
}

func PresentOKCancelStruct[S any](structPtr *S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	slice := []S{*structPtr}
	PresentOKCancelStructSlice(&slice, params, title, att, func(ok bool) (close bool) {
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
