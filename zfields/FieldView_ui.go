//go:build zui

package zfields

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zcheckbox"
	"github.com/torlangballe/zui/zclipboard"
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
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zguiutil"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlocale"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
	"github.com/torlangballe/zutil/zwords"
)

type FieldView struct {
	zcontainer.StackView
	ID       string
	parent   *FieldView
	Fields   []Field
	data     interface{}
	dataHash int64
	params   FieldViewParameters
}

type FieldViewParameters struct {
	Field
	FieldParameters
	ImmediateEdit            bool                                                                        // ImmediateEdit forces immediate write-to-data when editing a field.
	MultiSliceEditInProgress bool                                                                        // MultiSliceEditInProgress is on if the field represents editing multiple structs in a list. Checkboxes can be indeterminate etc.
	EditWithoutCallbacks     bool                                                                        // Set so not get edit/changed callbacks when editing. Example: Dialog box editing edits a copy, so no callbacks needed.
	triggerHandlers          map[trigger]func(fv *FieldView, f *Field, value any, view *zview.View) bool // triggerHandlers is a map of functions to call if an action occurs in this FieldView. Somewhat replacing ActionHandler
}

// If a structure/slice used in FieldViews has this method, it is called when edited/changed etc.
// The field may be nil if it's for the entire structure.
type ActionHandler interface {
	HandleAction(f *Field, action ActionType, view *zview.View) bool
}

// If a struct implements StructInitializer, it is initialized when a element is added to a slice in FieldSliceView
// Note: the method must be on a pointer, not value, as it changes the contents of the struct.
// TODO: Make default Add element in SliceGridView that uses this.
type StructInitializer interface {
	InitZFieldStruct()
}

var fieldViewEdited = map[string]time.Time{}

func FieldViewParametersDefault() (f FieldViewParameters) {
	f.ImmediateEdit = true
	f.Styling = zstyle.EmptyStyling
	f.Styling.Spacing = 2
	return f
}

// func (v *FieldView) ID() string {
// 	return v.id
// }

func (v *FieldView) Data() any {
	return v.data
}

func (v *FieldView) IsSlice() bool {
	return reflect.ValueOf(v.data).Elem().Kind() == reflect.Slice
}

func setFieldViewEdited(fv *FieldView) {
	fieldViewEdited[fv.Hierarchy()] = time.Now()
}

func IsFieldViewEditedRecently(fv *FieldView) bool {
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

func makeFrameIfFlag(f *Field, child zview.View) zview.View {
	if f.Flags&FlagHasFrame == 0 {
		return nil
	}
	var title string
	if f.Flags&FlagFrameIsTitled != 0 {
		title = f.TitleOrName()
	}
	frame := zcontainer.StackViewVert("frame")
	zwidgets.MakeStackATitledFrame(frame, title, f.Flags&FlagFrameTitledOnFrame != 0, f.Styling, f.Styling)
	frame.Add(child, zgeo.TopLeft)
	return frame
}

func fieldViewNew(id string, vertical bool, data any, params FieldViewParameters, marg zgeo.Size, parent *FieldView) *FieldView {
	v := &FieldView{}
	v.StackView.Init(v, vertical, id)
	v.SetSpacing(params.Styling.Spacing)
	v.data = data
	zreflect.ForEachField(v.data, true, func(index int, val reflect.Value, sf reflect.StructField) bool {
		f := EmptyField
		if !f.SetFromReflectValue(val, sf, index, params.ImmediateEdit) {
			return true
		}
		callActionHandlerFunc(v, &f, SetupFieldAction, val.Addr().Interface(), nil)
		v.Fields = append(v.Fields, f)
		return true
	})
	if params.Field.Flags&FlagHasFrame == 0 {
		v.SetMargin(zgeo.RectFromMinMax(marg.Pos(), marg.Pos().Negative()))
	}
	v.params = params
	v.ID = id
	v.parent = parent

	return v
}

func (v *FieldView) Build(update bool) {
	a := zgeo.Left
	if v.Vertical {
		a |= zgeo.Top
	} else {
		a |= zgeo.VertCenter
	}
	v.BuildStack(v.ObjectName(), a, zgeo.Size{}, true)
	if update {
		dontOverwriteEdited := false
		v.Update(nil, dontOverwriteEdited)
	}
}

func (v *FieldView) findNamedViewOrInLabelized(name string) (view, maybeLabel zview.View) {
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
	for _, f := range v.Fields {
		var id string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		if zstr.HasPrefix(local, "./", &id) && id == toID {
			_, foundView := v.findNamedViewOrInLabelized(f.FieldName)
			if foundView == nil {
				continue
			}
			if neg {
				isShow = !isShow
			}
			if isShow {
				foundView.Show(!isZero)
			} else {
				foundView.SetUsable(!isZero)
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
	for _, f := range v.Fields {
		if f.FieldName != toFieldName {
			continue
		}
		var prefix, fname string
		local, neg := getLocalFromShowOrEnable(isShow, &f)
		if zstr.HasPrefix(local, "./", &fname) {
			rval, _, findex := zreflect.FieldForName(v.data, true, fname)
			if findex != -1 {
				doShowEnableItem(rval, isShow, view, neg)
			}
			continue
		}
		if zstr.SplitN(local, "/", &prefix, &fname) && prefix == f.FieldName {
			fv := view.(*FieldView)
			if fv == nil {
				zlog.Error(nil, "updateShowOrEnable: not field view:", f.FieldName, local, v.ObjectName)
				return
			}
			rval, _, findex := zreflect.FieldForName(v.data, true, fname)
			if findex != -1 {
				doShowEnableItem(rval, isShow, view, neg)
			}
			continue
		}
		if zstr.HasPrefix(local, "../", &fname) && v.parent != nil {
			rval, _, findex := zreflect.FieldForName(v.parent.Data(), true, fname)
			if findex != -1 {
				doShowEnableItem(rval, isShow, view, neg)
			}
			continue
		}
	}
}

func (v *FieldView) Update(data any, dontOverwriteEdited bool) {
	if data != nil { // must be after IsFieldViewEditedRecently, or we set new data without update slice pointers and maybe more
		v.data = data
	}
	if dontOverwriteEdited && IsFieldViewEditedRecently(v) {
		zreflect.ForEachField(v.data, true, func(index int, rval reflect.Value, sf reflect.StructField) bool {
			if sf.Type.Kind() == reflect.Slice {
				v.updateField(index, rval, sf, dontOverwriteEdited) // even if edited recently, we call v.updateField on slices, to set new address of each slice in FieldView.data
			}
			return true
		})
		zlog.Info("FV No Update because recently edited", v.Hierarchy())
		return
	}
	fh, _ := v.data.(ActionHandler)
	sview := v.View
	if fh != nil {
		fh.HandleAction(nil, DataChangedActionPre, &sview)
	}
	zreflect.ForEachField(v.data, true, func(index int, rval reflect.Value, sf reflect.StructField) bool {
		v.updateField(index, rval, sf, dontOverwriteEdited)
		return true
	})
	// call general one with no id. Needs to be after above loop, so values set
	if fh != nil {
		fh.HandleAction(nil, DataChangedAction, &sview)
	}
}

func (v *FieldView) updateField(index int, rval reflect.Value, sf reflect.StructField, dontOverwriteEdited bool) bool {
	var valStr string
	f := findFieldWithIndex(&v.Fields, index)
	if f == nil {
		return true
	}
	foundView, flabelized := v.findNamedViewOrInLabelized(f.FieldName)
	if foundView == nil {
		return true
	}
	v.updateShowEnableOnView(flabelized, true, foundView.ObjectName())
	v.updateShowEnableOnView(flabelized, false, foundView.ObjectName())
	var called bool
	tri, _ := rval.Interface().(TriggerDataChangedTriggerer)
	if tri != nil {
		called = tri.HandleDataChange(v, f, rval.Addr().Interface(), &foundView)
	}
	if !called {
		called = v.callTriggerHandler(f, DataChangedAction, rval.Addr().Interface(), &foundView)
	}
	if !called {
		called = callActionHandlerFunc(v, f, DataChangedAction, rval.Addr().Interface(), &foundView)
	}
	if called {
		return true
	}
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
		} else {
			ei, findex := FindLocalFieldWithID(v.data, f.LocalEnum)
			zlog.Assert(findex != -1, f.Name, f.LocalEnum)
			enum = ei.Interface().(zdict.ItemsGetter).GetItems()
		}
		if v.params.ForceZeroOption {
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
			if !hasZero && rtype.Kind() != reflect.Invalid {
				item := zdict.Item{Name: "", Value: reflect.Zero(rtype)}
				enum = append(zdict.Items{item}, enum...)
			}
		}
		menuType.UpdateItems(enum, rval.Interface(), f.Flags&FlagIsActions != 0)
		return true
	}
	updateItemLocalToolTip(f, v.data, foundView)
	if f.IsStatic() || v.params.AllStatic {
		zuistringer, _ := rval.Interface().(UIStringer)
		if zuistringer != nil {
			label, _ := foundView.(*zlabel.Label)
			if label != nil {
				label.SetText(zuistringer.ZUIString())
				return true
			}
		}
	}
	switch f.Kind {
	case zreflect.KindSlice:
		sv, _ := foundView.(*FieldSliceView)
		if sv == nil {
			zlog.Info("UpdateSlice: not a *FieldSliceView:", v.Hierarchy())
			return false
		}
		hash := zstr.HashAnyToInt64(reflect.ValueOf(sv.data).Elem(), "")
		sameHash := (sv.dataHash == hash)
		sv.dataHash = hash
		if !sameHash {
			sv.UpdateSlice(rval.Addr().Interface())
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
		to := foundView.(ztext.LayoutOwner)
		to.SetText(valStr)

	case zreflect.KindStruct:
		fv, _ := foundView.(*FieldView)
		if fv == nil {
			fv = findSubFieldView(foundView, "")
		}
		if fv == nil {
			break
		}
		fv.Update(rval.Addr().Interface(), dontOverwriteEdited)

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
		valStr = getTextFromNumberishItem(rval, f)
		if f.IsStatic() || v.params.AllStatic {
			label, _ := foundView.(*zlabel.Label)
			if label != nil {
				label.SetText(valStr)
			}
			break
		}
		tv, _ := foundView.(*ztext.TextView)
		if tv != nil {
			if tv.IsEditing() {
				break
			}
			tv.SetText(valStr)
		}

	case zreflect.KindString, zreflect.KindFunc:
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
				label, _ := foundView.(*zlabel.Label)
				if label != nil {
					if f.Flags&FlagIsFixed != 0 {
						valStr = f.Name
					}
					label.SetText(valStr)
				}
			} else {
				tv, _ := foundView.(*ztext.TextView)
				if tv != nil {
					if tv.IsEditing() {
						break
					}
					tv.SetText(valStr)
				}
			}
		}
	}
	return true
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
	v := fieldViewNew(id, true, data, params, zgeo.Size{10, 10}, nil)
	return v
}

func (v *FieldView) Rebuild() {
	zlog.Info("FV.Rebuild:", v.data != nil, reflect.ValueOf(v.data).Kind())
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
		zlog.Error(nil, "CallFieldAction find view", fieldID)
		return
	}
	f := v.findFieldWithID(fieldID)
	if f == nil {
		zlog.Error(nil, "CallFieldAction find field", fieldID)
		return
	}
	callActionHandlerFunc(v, f, action, fieldValue, &view)
}

func callActionHandlerFunc(v *FieldView, f *Field, action ActionType, fieldValue interface{}, view *zview.View) bool {
	if action == EditedAction {
		if f.SetEdited {
			setFieldViewEdited(v)
		}
		if v.params.EditWithoutCallbacks {
			return true
		}
	}
	direct := (action == CreateFieldViewAction || action == SetupFieldAction)
	fh, _ := v.data.(ActionHandler)
	var result bool
	if fh != nil {
		result = fh.HandleAction(f, action, view)
	}
	if view != nil && *view != nil {
		first := true
		n := (*view).Native()
		for n != nil {
			parent := n.Parent()
			if parent != nil {
				fv, _ := parent.View.(*FieldView)
				if fv != nil {
					if fv.IsSlice() {
						fv.dataHash = zstr.HashAnyToInt64(reflect.ValueOf(fv.data).Elem(), "")
					}
					if !first {
						fh2, _ := fv.data.(ActionHandler)
						if fh2 != nil {
							fh2.HandleAction(nil, action, &parent.View)
						}
					}
					first = false
				}
			}
			n = parent
		}
	}

	if !result {
		var fieldAddress interface{}
		if !direct {
			changed := false
			sv := reflect.ValueOf(v.data)
			if sv.Kind() == reflect.Ptr || sv.CanAddr() {
				// Here we run thru the possiblly new struct again, and find the item with same id as field
				fieldVal, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
				if found {
					changed = true
					fieldValue = fieldVal.Interface()
					if fieldVal.CanAddr() {
						fieldAddress = fieldVal.Addr().Interface()
					}
				}
			}
			if !changed {
				zlog.Info("NOOT!!!", f.Name, action, v.data != nil)
				zlog.Fatal(nil, "Not CHANGED!", f.Name)
			}
		}
		aih, _ := fieldValue.(ActionHandler)
		if aih == nil && fieldAddress != nil {
			aih, _ = fieldAddress.(ActionHandler)
		}
		if aih != nil {
			result = aih.HandleAction(f, action, view)
		}
	}
	return result
}

func (v *FieldView) findFieldWithID(id string) *Field {
	for i, f := range v.Fields {
		if f.ID == id {
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
	name := f.Name
	if f.Title != "" && fv.params.LabelizeWidth == 0 {
		name = f.Title
	}
	s := zgeo.Size{20, 24}
	if f.Height != 0 {
		s.H = f.Height
	}
	button := zshape.ImageButtonViewNew(name, color, s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
	button.SetTextColor(zgeo.ColorBlack)
	button.TextXMargin = 0
	return button
}

func (v *FieldView) makeMenu(rval reflect.Value, f *Field, items zdict.Items) zview.View {
	var view zview.View
	static := f.IsStatic() || v.params.AllStatic
	isSlice := rval.Kind() == reflect.Slice
	if static || isSlice {
		multi := isSlice
		isImage := (f.ImageFixedPath != "")
		shape := zshape.TypeRoundRect
		if isImage {
			shape = zshape.TypeNone
		}
		menuOwner := zmenu.NewMenuedOwner()
		menuOwner.IsStatic = static
		menuOwner.IsMultiple = multi
		menuOwner.StoreKey = f.ValueStoreKey
		for _, format := range strings.Split(f.Format, "|") {
			if menuOwner.TitleIsAll == " " {
				menuOwner.TitleIsAll = format
			}
			switch format {
			case "all":
				menuOwner.TitleIsAll = " "
			case "%d":
				menuOwner.GetTitleFunc = func(icount int) string { return strconv.Itoa(icount) }
			case `%n`:
				menuOwner.TitleIsValueIfOne = true
			default:
				menuOwner.PluralableWord = format
			}
		}
		mItems := zmenu.MOItemsFromZDictItemsAndValues(items, nil, f.Flags&FlagIsActions != 0)
		menu := zmenu.MenuOwningButtonCreate(menuOwner, mItems, shape)
		if isImage {
			menu.SetImage(nil, f.ImageFixedPath, nil)
			menu.ImageMaxSize = f.Size
		} else {
			if len(f.Colors) != 0 {
				menu.SetColor(zgeo.ColorFromString(f.Colors[0]))
			}
		}
		menu.Ratio = 0.3
		view = menu
		menuOwner.SelectedHandlerFunc = func() {
			if menuOwner.IsStatic {
				sel := menuOwner.SelectedItem()
				if sel != nil {
					kind := reflect.ValueOf(sel.Value).Kind()
					if kind != reflect.Ptr && kind != reflect.Struct {
						nf := *f
						nf.ActionValue = sel.Value
						callActionHandlerFunc(v, &nf, PressedAction, rval.Interface(), &view)
					}
				}
			}
			v.callTriggerHandler(f, EditedAction, rval.Interface(), &view)
			callActionHandlerFunc(v, f, EditedAction, rval.Interface(), &view)
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
		menu := zmenu.NewView(f.Name+"Menu", items, rval.Interface())
		menu.SetMaxWidth(f.MaxWidth)
		view = menu
		menu.SetSelectedHandler(func() {
			v.fieldToDataItem(f, menu)
			v.callTriggerHandler(f, EditedAction, rval.Interface(), &view)
			callActionHandlerFunc(v, f, EditedAction, rval.Interface(), &view)
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
			format = "15:04:03"
		}
		if zlocale.DisplayServerTime.Get() {
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

func getTextFromNumberishItem(rval reflect.Value, f *Field) string {
	if f.Flags&FlagZeroIsEmpty != 0 {
		if rval.IsZero() {
			return ""
		}
	}
	stringer, got := rval.Interface().(UIStringer)
	if got {
		return stringer.ZUIString()
	}
	zkind := zreflect.KindFromReflectKindAndType(rval.Kind(), rval.Type())
	isDurTime := zkind == zreflect.KindTime && f.Flags&FlagIsDuration != 0
	if zkind == zreflect.KindTime && !isDurTime {
		return getTimeString(rval, f)
	}
	if isDurTime || f.PackageName == "time" && rval.Type().Name() == "Duration" {
		var t float64
		if isDurTime {
			t = ztime.Since(rval.Interface().(time.Time))
		} else {
			t = ztime.DurSeconds(time.Duration(rval.Int()))
		}
		return ztime.GetSecsAsHMSString(t, f.Flags&FlagHasSeconds != 0, 0)
	}
	format := f.Format
	significant := f.Columns
	switch format {
	case "memory":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			return zwords.GetMemoryString(b, "", significant)
		}
	case "storage":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			return zwords.GetStorageSizeString(b, "", significant)
		}
	case "bps":
		b, err := zint.GetAny(rval.Interface())
		if err == nil {
			return zwords.GetBandwidthString(b, "", significant)
		}
	case "":
		format = "%v"
	}
	return fmt.Sprintf(format, rval.Interface())
}

func (v *FieldView) makeText(rval reflect.Value, f *Field, noUpdate bool) zview.View {
	str := getTextFromNumberishItem(rval, f)
	if f.IsStatic() || v.params.AllStatic {
		label := zlabel.New(str)
		label.SetMaxLines(f.Rows)
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
			label.SetPressedHandler(func() {
				text := label.Text()
				zclipboard.SetString(text)
				label.SetText("📋 " + text)
				ztimer.StartIn(0.6, func() {
					label.SetText(text)
				})
			})
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
	tv.SetPlaceholder(f.Placeholder)
	tv.SetChangedHandler(func() {
		v.fieldHandleEdited(f, tv.View)
	})
	return tv
}

func (v *FieldView) fieldHandleEdited(f *Field, view zview.View) {
	rval, err := v.fieldToDataItem(f, view)
	if zlog.OnError(err) {
		return
	}
	if v.callTriggerHandler(f, EditedAction, rval.Interface(), &view) {
		return
	}
	callActionHandlerFunc(v, f, EditedAction, rval.Interface(), &view)
}

func (v *FieldView) makeCheckbox(f *Field, b zbool.BoolInd) zview.View {
	cv := zcheckbox.New(b)
	cv.SetObjectName(f.FieldName)
	cv.SetValueHandler(func() {
		val, _ := v.fieldToDataItem(f, cv)
		v.updateShowEnableFromZeroer(val.IsZero(), true, cv.ObjectName())
		v.updateShowEnableFromZeroer(val.IsZero(), false, cv.ObjectName())
		view := zview.View(cv)
		if v.callTriggerHandler(f, EditedAction, cv.On(), &cv.View) {
			return
		}
		callActionHandlerFunc(v, f, EditedAction, val.Interface(), &view)
	})
	if v.params.LabelizeWidth == 0 && !zstr.StringsContain(v.params.UseInValues, "$row") {
		_, stack := zcheckbox.Labelize(cv, f.TitleOrName())
		return stack
	}
	return cv
}

func (v *FieldView) makeImage(rval reflect.Value, f *Field) zview.View {
	iv := zimageview.New(nil, "", f.Size)
	iv.DownsampleImages = true
	iv.SetMinSize(f.Size)
	iv.SetObjectName(f.FieldName)
	iv.OpaqueDraw = (f.Flags&FlagIsOpaque != 0)
	if f.Styling.FGColor.Valid {
		iv.EmptyColor = f.Styling.FGColor
	}
	if f.IsImageToggle() && rval.Kind() == reflect.Bool {
		iv.SetPressedHandler(func() {
			val, _ := v.fieldToDataItem(f, iv)
			on := val.Bool()
			on = !on
			val.SetBool(on)
			v.Update(nil, false)
			v.callTriggerHandler(f, EditedAction, on, &iv.View)
			callActionHandlerFunc(v, f, EditedAction, val.Interface(), &iv.View)
		})
	} else {
		iv.SetPressedHandler(func() {
			v.callTriggerHandler(f, PressedAction, v.ID, &iv.View)
		})
	}
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
	val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
	if found {
		t := val.Interface().(time.Time)
		if ztime.Since(t) > float64(f.OldSecs) {
			label.SetColor(zgeo.ColorRed)
		} else {
			setColorFromField(label, f)
		}
	}
}

func (v *FieldView) updateSinceTime(label *zlabel.Label, f *Field) {
	if zlog.IsInTests { // if in unit-tests, we don't show anything as it would change
		label.SetText("")
		return
	}
	sv := reflect.ValueOf(v.data)
	zlog.Assert(sv.Kind() == reflect.Ptr || sv.CanAddr())
	val, found := zreflect.FindFieldWithNameInStruct(f.FieldName, v.data, true)
	if found {
		var str string
		t := val.Interface().(time.Time)
		tooBig := false
		if !t.IsZero() {
			since := time.Since(t)
			str, tooBig = ztime.GetDurationString(since, f.Flags&FlagHasSeconds != 0, f.Flags&FlagHasMinutes != 0, f.Flags&FlagHasHours != 0, f.FractionDecimals)
		}
		inter := val.Interface()
		if val.CanAddr() {
			inter = val.Addr()
		}
		if tooBig {
			label.SetText("●")
			label.SetColor(zgeo.ColorRed)
		} else {
			label.SetText(str)
			setColorFromField(label, f)
		}
		callActionHandlerFunc(v, f, DataChangedAction, inter, &label.View)
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
			changer, _ := widgetView.(zview.ChangedReporter)
			if changer != nil {
				changer.SetChangedHandler(func() {
					v.fieldHandleEdited(f, widgetView)
				})
			}
			return widgetView, false
		}
	}
	if v.callTriggerHandler(f, CreateFieldViewAction, nil, &view) {
		return
	}
	callActionHandlerFunc(v, f, CreateFieldViewAction, rval.Addr().Interface(), &view) // this sees if actual ITEM is a field handler
	if view != nil {
		return
	}
	if f.LocalEnum != "" {
		ei, findex := FindLocalFieldWithID(v.data, f.LocalEnum)
		if zlog.ErrorIf(findex == -1, v.Hierarchy(), f.Name, f.LocalEnum) {
			return nil, true
		}
		getter, _ := ei.Interface().(zdict.ItemsGetter)
		if zlog.ErrorIf(getter == nil, "field isn't enum, not ItemGetter type", f.Name, f.LocalEnum) {
			return nil, true
		}
		enum := getter.GetItems()
		for i := range enum {
			if f.Flags&FlagZeroIsEmpty != 0 {
				if enum[i].Value != nil && reflect.ValueOf(enum[i].Value).IsZero() {
					enum[i].Name = ""
				}
			}
		}

		menu := v.makeMenu(rval, f, enum)
		if menu == nil {
			zlog.Error(nil, "no local enum for", f.LocalEnum)
			return nil, true
		}
		return menu, false
	}
	if f.Enum != "" {
		enum, _ := fieldEnums[f.Enum]
		zlog.Assert(enum != nil, f.Enum, f.FieldName)
		if rval.IsZero() {
			if !v.params.MultiSliceEditInProgress {
				if len(enum) > 0 {
					rval.Set(reflect.ValueOf(enum[0].Value))
				}
			} else {
				var zero bool
				for i := range enum {
					if enum[i].Value != nil && reflect.ValueOf(enum[i].Value).IsZero() {
						zero = true
						break
					}
				}
				if !zero {
					enum = append(enum, zdict.Item{"", rval.Interface()})
				}
			}
		}
		view = v.makeMenu(rval, f, enum)
		return view, false
	}
	_, got := rval.Interface().(UIStringer)
	if got && (f.IsStatic() || v.params.AllStatic) {
		return v.makeText(rval, f, false), false
	}
	return nil, false
}

func (v *FieldView) BuildStack(name string, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) {
	zlog.Assert(reflect.ValueOf(v.data).Kind() == reflect.Ptr, name, v.data, reflect.ValueOf(v.data).Kind())
	ForEachField(v.data, v.params.FieldParameters, v.Fields, func(index int, f *Field, val reflect.Value, sf reflect.StructField) {
		v.buildItem(f, val, index, defaultAlign, cellMargin, useMinWidth)
	})
}

func (v *FieldView) buildItem(f *Field, rval reflect.Value, index int, defaultAlign zgeo.Alignment, cellMargin zgeo.Size, useMinWidth bool) {
	labelizeWidth := v.params.LabelizeWidth
	parentFV := ParentFieldView(v)
	if parentFV != nil && v.params.LabelizeWidth == 0 {
		labelizeWidth = parentFV.params.LabelizeWidth
	}
	exp := zgeo.AlignmentNone
	if v.params.HideStatic && f.IsStatic() {
		return
	}
	view, skip := v.createSpecialView(rval, f)
	if skip {
		return
	}
	if view == nil {
		switch f.Kind {
		case zreflect.KindStruct:
			vert := true
			if !f.Vertical.IsUnknown() {
				vert = f.Vertical.Bool()
			}
			params := v.params
			params.Field.MergeInField(f)
			fieldView := fieldViewNew(f.FieldName, vert, rval.Addr().Interface(), params, zgeo.Size{}, v)
			view = makeFrameIfFlag(f, fieldView)
			if view == nil {
				view = fieldView
			}
			fieldView.BuildStack(f.FieldName, zgeo.TopLeft, zgeo.Size{}, true)

		case zreflect.KindBool:
			if f.Flags&FlagIsImage != 0 && f.IsImageToggle() && rval.Kind() == reflect.Bool {
				view = v.makeImage(rval, f)
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

		case zreflect.KindSlice:
			getter, _ := rval.Interface().(zdict.ItemsGetter)
			if getter != nil {
				menu := v.makeMenu(rval, f, getter.GetItems())
				view = menu
				break
			}
			if v.params.MultiSliceEditInProgress {
				return
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
						nlabel := view.(*zlabel.Label)
						if f.Flags&FlagIsDuration != 0 {
							v.updateSinceTime(nlabel, f)
						} else {
							v.updateOldTime(nlabel, f)
						}
						return true
					})
					v.AddOnRemoveFunc(timer.Stop)
					view.Native().AddOnRemoveFunc(timer.Stop)
				} else {
					if f.Format == "nice" {
						timer := ztimer.StartAt(ztime.OnTheNextHour(time.Now()), func() {
							nlabel := view.(*zlabel.Label)
							str := getTextFromNumberishItem(rval, f)
							nlabel.SetText(str)
						})
						v.AddOnRemoveFunc(timer.Stop)
					}
				}
			}

		default:
			panic(fmt.Sprintln("buildStack bad type:", f.Name, f.Kind))
		}
	}
	zlog.Assert(view != nil)
	pt, _ := view.(zview.Pressable)
	if view != nil && pt != nil && pt.PressedHandler != nil {
		ph := pt.PressedHandler()
		nowItem := rval // store item in nowItem so closures below uses right item
		pt.SetPressedHandler(func() {
			if !callActionHandlerFunc(v, f, PressedAction, nowItem.Interface(), &view) && ph != nil {
				ph()
			}
		})
		if f.Flags&FlagLongPress != 0 {
			lph := pt.LongPressedHandler()
			pt.SetLongPressedHandler(func() {
				if !callActionHandlerFunc(v, f, LongPressedAction, nowItem.Interface(), &view) && lph != nil {
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
		view.SetBGColor(zgeo.ColorNew(1, 0.9, 0.9, 1))
	}
	callActionHandlerFunc(v, f, CreatedViewAction, rval.Addr().Interface(), &view)
	if f.Download != "" {
		surl := zstr.ReplaceAllCapturesWithoutMatchFunc(zstr.InDoubleSquigglyBracketsRegex, f.Download, func(fieldName string, index int) string {
			a, findex := FindLocalFieldWithID(v.data, fieldName)
			if findex == -1 {
				zlog.Error(nil, "field download", f.Download, ":", "field not found in struct:", fieldName)
				return ""
			}
			return fmt.Sprint(a.Interface())
		})
		link := zcontainer.MakeLinkedStack(surl, "", view)
		view = link
	}

	cell := &zcontainer.Cell{}
	def := defaultAlign
	all := zgeo.Left | zgeo.HorCenter | zgeo.Right
	if f.Alignment&all != 0 {
		def &= ^all
	}
	cell.Margin = cellMargin
	cell.Alignment = def | exp | f.Alignment
	if labelizeWidth != 0 || f.LabelizeWidth < 0 {
		var lstack *zcontainer.StackView
		title := f.Name
		if f.Title != "" {
			title = f.Title
		}
		if f.Flags&FlagNoTitle != 0 {
			title = ""
		}
		_, lstack, cell = zguiutil.Labelize(view, title, labelizeWidth, cell.Alignment)
		v.Add(lstack, zgeo.HorExpand|zgeo.Left|zgeo.Top)
	}
	if useMinWidth {
		cell.MinSize.W = f.MinWidth
	}
	cell.MaxSize.W = f.MaxWidth
	if labelizeWidth == 0 {
		cell.View = view
		v.AddCell(*cell, -1)
	}
}

func updateItemLocalToolTip(f *Field, structure any, view zview.View) {
	var tipField, tip string
	if zstr.HasPrefix(f.Tooltip, "./", &tipField) {
		ei, _, findex := zreflect.FieldForName(structure, true, tipField)
		if findex != -1 {
			tip = fmt.Sprint(ei.Interface())
		} else { // can't use tip == "" to check, since field might just be empty
			zlog.Error(nil, "updateItemLocalToolTip: no local field for tip", f.Name, tipField)
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
		foundView, _ := v.findNamedViewOrInLabelized(f.FieldName)
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
	rval, _ := zreflect.FieldForIndex(v.data, true, f.Index)

	if f.WidgetName != "" && f.Kind != zreflect.KindSlice {
		w := widgeters[f.WidgetName]
		if w != nil {
			getter, _ := view.(zview.AnyValueGetter)
			if getter != nil {
				val := getter.ValueAsAny()
				rval.Set(reflect.ValueOf(val))
				value = rval
			}
			return
		}
	}
	if f.Enum != "" || f.LocalEnum != "" {
		mv, _ := view.(*zmenu.MenuView)
		if mv != nil {
			iface := mv.CurrentValue()
			vo := reflect.ValueOf(iface)
			// zlog.Debug(iface, f.Name, iface == nil)
			if iface == nil {
				vo = reflect.Zero(rval.Type())
			}
			rval.Set(vo)
		}
		return
	}

	switch f.Kind {
	case zreflect.KindBool:
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
				zlog.Fatal(nil, "Should be checkbox", view, reflect.TypeOf(view))
			}
		}
		b, _ := rval.Addr().Interface().(*bool)
		if b != nil {
			*b = bv.Value().Bool()
		}
		bi, _ := rval.Addr().Interface().(*zbool.BoolInd)
		if bi != nil {
			*bi = bv.Value()
		}

	case zreflect.KindInt:
		if rval.Type().Name() == "BoolInd" {
			bv, _ := view.(*zcheckbox.CheckBox)
			*rval.Addr().Interface().(*zbool.BoolInd) = bv.Value()
		} else {
			tv, _ := view.(*ztext.TextView)
			if f.Flags&FlagZeroIsEmpty != 0 {
				if tv.Text() == "" {
					rval.SetZero()
					v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())
					break
				}
			}
			str := tv.Text()
			if f.PackageName == "time" && rval.Type().Name() == "Duration" {
				var secs float64
				secs, err = ztime.GetSecsFromHMSString(str, f.Flags&FlagHasHours != 0, f.Flags&FlagHasMinutes != 0, f.Flags&FlagHasSeconds != 0)
				if err != nil {
					break
				}
				d := rval.Addr().Interface().(*time.Duration)
				if d != nil {
					*d = ztime.SecondsDur(secs)
				}
				value = rval
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
			zint.SetAny(rval.Addr().Interface(), i64)
			v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())
		}

	case zreflect.KindFloat:
		tv, _ := view.(*ztext.TextView)
		text := tv.Text()
		var f64 float64
		if f.Flags&FlagZeroIsEmpty != 0 {
			if text == "" {
				rval.SetZero()
				v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())
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
		zfloat.SetAny(rval.Addr().Interface(), f64)
		v.updateShowEnableFromZeroer(rval.IsZero(), false, tv.ObjectName())

	case zreflect.KindTime:
		break

	case zreflect.KindString:
		if (!f.IsStatic() && !v.params.AllStatic) && f.Flags&FlagIsImage == 0 {
			tv, _ := view.(*ztext.TextView)
			if tv == nil {
				zlog.Fatal(nil, "Copy Back string not TV:", f.Name)
			}
			text := tv.Text()
			str := rval.Addr().Interface().(*string)
			*str = text
		}

	case zreflect.KindFunc:
		break

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

	default:
		// zlog.Info("fieldToDataItem for slice:", f.Name)
		break
	}
	value = reflect.ValueOf(rval.Addr().Interface()).Elem() //.Interface()
	return
}

func ParentFieldView(view zview.View) *FieldView {
	for _, nv := range view.Native().AllParents() {
		fv, _ := nv.View.(*FieldView)
		if fv != nil {
			return fv
		}
	}
	return nil
}

func PresentOKCancelStruct[S any](structPtr *S, params FieldViewParameters, title string, att zpresent.Attributes, done func(ok bool) (close bool)) {
	slice := []S{*structPtr}
	PresentOKCancelStructSlice(&slice, params, title, att, func(ok bool) (close bool) {
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
