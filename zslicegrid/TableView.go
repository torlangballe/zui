//go:build zui

package zslicegrid

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zheader"
	"github.com/torlangballe/zui/zshape"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztime"
	"github.com/torlangballe/zutil/ztimer"
)

// TableView is a SliceGridView which creates rows from structs using the zfields package.
// See zfields for details on how to tag struct fields with `zui:"xxx"` for styling.
// if the AddHeader option is set, it adds header on top using zheader.HeaderView.
type TableView[S zstr.StrIDer] struct {
	SliceGridView[S]
	Header               *zheader.HeaderView // Optional header based on S struct
	ColumnMargin         float64             // Margin between columns
	RowInset             float64             // inset on far left and right
	FieldParameters      zfields.FieldViewParameters
	AfterLockPressedFunc func(fieldName string, didLock bool)
	fields               []zfields.Field // the fields in an S struct used to generate columns for the table
	lockedFieldValues    map[string]any  // this is map of FieldName to list of values in the field that need to equal row's or its filtered out
	hasFitHeaderToRows   bool
}

func TableViewNew[S zstr.StrIDer](s *[]S, storeName string, options OptionType) *TableView[S] {
	v := &TableView[S]{}
	v.Init(v, s, storeName, options)
	return v
}

func (v *TableView[S]) Init(view zview.View, s *[]S, storeName string, options OptionType) {
	v.SliceGridView.Init(view, s, storeName, options)
	v.Grid.MaxColumns = 1
	v.Grid.SetMargin(zgeo.Rect{})
	v.ColumnMargin = 5
	v.RowInset = 7 //RowInset not used yet, should be Grid margin, but calculated OnReady
	v.lockedFieldValues = map[string]any{}
	// v.HeaderHeight = 28
	v.FieldParameters = zfields.FieldViewParametersDefault()
	v.FieldParameters.AllStatic = true
	v.FieldParameters.UseInValues = []string{zfields.RowUseInSpecialName}
	v.FieldParameters.AddTrigger("*", zfields.EditedAction, func(ap zfields.ActionPack) bool {
		if v.StoreChangedItemFunc != nil {
			go v.StoreChangedItemFunc(*(ap.FieldView.Data().(*S)), true)
		}
		return false
	})

	cell, _ := v.FindCellWithView(v.Grid)
	cell.Margin.H = 1 // this seems to make no 1-space beteen header and table
	v.Grid.CreateCellFunc = func(grid *zgridlist.GridListView, id string) zview.View {
		r := v.createRow(id)
		return r
	}
	// zlog.Info("TABLE INIT:", v.Hierarchy(), v.Grid.CreateCellFunc != nil, zlog.Pointer(v.Grid))
	if v.options&AddHeader != 0 {
		v.Header = zheader.NewView(v.ObjectName() + ".header")
		index := 0
		if v.options&AddBar != 0 && v.options&AddBarInHeader == 0 {
			index = 1
		}
		v.SliceGridView.AddAdvanced(v.Header, zgeo.Left|zgeo.Top|zgeo.HorExpand, zgeo.SizeNull, zgeo.SizeNull, index, false)
	}
	v.Grid.HandleSelectionChangedFunc = func() {
		v.UpdateWidgets()
		selCount := len(v.Grid.SelectedIDs())
		for _, c := range v.Header.GetChildren(false) {
			h, _ := c.(*zshape.ImageButtonView)
			if h == nil {
				continue
			}
			fieldName := h.ObjectName()
			_, isLocked := v.lockedFieldValues[fieldName]
			// zlog.Info("HandleSelectionChangedFunc show:", h.ObjectName(), hasSelected, isLocked)
			canLock := selCount > 0
			if !isLocked {
				f, _ := v.findField(fieldName)
				if f.HasFlag(zfields.FlagIsTimeBoundry) {
					canLock = (selCount == 1)
				}
			}
			zheader.ShowLock(h, canLock || isLocked)
		}
	}
}

func (v *TableView[S]) findField(fieldName string) (*zfields.Field, int) {
	for i, f := range v.fields {
		if f.FieldName == fieldName {
			return &v.fields[i], i
		}
	}
	return nil, -1
}

func (v *TableView[S]) ArrangeChildren() {
	zlog.Info("TV ArrangeChildren1", v.Hierarchy(), v.Rect())
	// defer zlog.Info("TV ArrangeChildren Done", v.Hierarchy(), v.Rect())
	var sw float64
	if v.Grid.HasSize() {
		sw = v.Grid.Rect().Size.W
	}
	v.SliceGridView.ArrangeChildren()
	if v.Header == nil || (v.Grid.HasSize() && v.Grid.Rect().Size.W == sw) {
		return
	}
	zlog.Info("TV ArrangeChildren", v.ObjectName(), v.Rect(), v.Grid.CellCountFunc())
	freeOnly := true
	v.Header.ArrangeAdvanced(freeOnly)
	var fv *zfields.FieldView
	if v.Grid.CellCountFunc() > 0 {
		view := v.Grid.AnyChildView()
		if view != nil {
			// zlog.Info("TV: ArrangeChildren FitHEader")
			fv = view.(*zfields.FieldView)
		}
	} else { // no rows, make an empty one to fit header with
		var s S
		view := v.createRowFromStruct(&s, zstr.GenerateRandomHexBytes(10))
		fv = view.(*zfields.FieldView)
		fv.SetRect(v.LocalRect())
		fv.ArrangeChildren()
	}
	if fv != nil {
		v.hasFitHeaderToRows = true
		v.Header.FitToRowStack(&fv.StackView)
	}
}

func (v *TableView[S]) ReadyToShow(beforeWindow bool) {
	if !beforeWindow {
		// for i, c := range v.Cells {
		// 	zlog.Info("Table", i, c.View.ObjectName(), c.Alignment, c.Margin)
		// }
		return
	}
	s := zslice.MakeAnElementOfSliceType(v.slicePtr)
	zfields.ForEachField(s, v.FieldParameters.FieldParameters, nil, func(each zfields.FieldInfo) bool {
		v.fields = append(v.fields, *each.Field)
		return true
	})
	if v.options&AddHeader != 0 {
		v.SortFunc = func(s []S) {
			// zlog.Info("SORT TABLE:", v.Hierarchy())
			zfields.SortSliceWithFields(s, v.fields, v.Header.SortOrder)
		}
		v.Header.SortingPressedFunc = func() {
			v.SortFunc(*v.slicePtr)
			v.UpdateViewFunc()
			// for i, s := range *v.slice {
			// 	fmt.Printf("Sorted: %d %+v\n", i, s)
			// }
		}
		headers := makeHeaderFields(v.fields)
		v.Header.Populate(headers)
		if v.options&AddBarInHeader != 0 {
			right := v.Header.RightColumn()
			zlog.Assert(right != nil)
			right.Add(v.Bar, zgeo.CenterRight)
		}
	}
	v.Grid.UpdateCellFunc = func(grid *zgridlist.GridListView, id string) {
		fv := grid.CellView(id).(*zfields.FieldView)
		zlog.Assert(fv != nil)
		fv.Update(v.StructForID(id), true, false)
		// fv.ArrangeChildren()
	}
	v.SliceGridView.ReadyToShow(beforeWindow)
}

func (v *TableView[S]) createRow(id string) zview.View {
	// zlog.Info("createRow", id)
	s := v.StructForID(id)
	view := v.createRowFromStruct(s, id)
	view.Native().SetSelectable(false)
	if v.options&AllowEdit != 0 {
		view.Native().SetDoublePressedHandler(func() {
			v.EditItemIDs([]string{id}, false, nil)
		})
	}
	return view
}

func (v *TableView[S]) createRowFromStruct(s *S, id string) zview.View {
	params := v.FieldParameters
	params.ImmediateEdit = false
	params.Styling.Spacing = 0
	params.AllStatic = (v.Grid.Selectable || v.Grid.MultiSelectable)

	fv := zfields.FieldViewNew(id, s, params)
	fv.Vertical = false
	fv.Fields = v.fields
	fv.SetSpacing(0)
	// fv.SetCanFocus(true)
	fv.SetMargin(zgeo.Rect{})
	useWidth := true //(v.Header != nil)
	name := "row " + id
	fv.BuildStack(name, zgeo.CenterLeft, zgeo.SizeD(v.ColumnMargin, 0), useWidth)
	// dontOverwriteEdited := false
	// fv.Update(nil, dontOverwriteEdited, false)
	// v.Grid.ClearDirtyRow(id) // we clear dirty as we did update above so ArrangeChild will work better
	return fv
}

func makeHeaderFields(fields []zfields.Field) []zheader.Header {
	var headers []zheader.Header
	for _, f := range fields {
		var h zheader.Header
		h.FieldName = f.FieldName
		h.Align = zgeo.Left | zgeo.VertCenter
		if !f.HasFlag(zfields.FlagDontJustifyHeader) {
			h.Justify = f.Justify
		}
		// zlog.Info("makeHeaderFields:", h.FieldName, f.HasFlag(zfields.FlagHeaderLockable))
		if f.HasFlag(zfields.FlagHeaderLockable) {
			h.Lockable = true
		}
		if f.Kind == zreflect.KindString && f.Enum == "" {
			h.Align |= zgeo.HorExpand
		}
		if f.Flags&zfields.FlagHasHeaderImage != 0 {
			h.ImageSize = f.HeaderSize
			if h.ImageSize.IsNull() {
				h.ImageSize = zgeo.SizeBoth(20)
			}
			h.ImagePath = f.HeaderImageFixedPath
		}
		if f.Flags&(zfields.FlagHasHeaderImage|zfields.FlagNoTitle) == 0 {
			if f.Header != "" {
				h.Title = f.Header
			} else {
				h.Title = f.TitleOrName()
			}
		}
		if f.Tooltip != "" && !strings.HasPrefix(f.Tooltip, ".") {
			h.Tip = f.Tooltip
		}
		h.SortSmallFirst = f.SortSmallFirst
		h.SortPriority = f.SortPriority
		headers = append(headers, h)
	}
	return headers
}

func (v *TableView[S]) FilterRowWithZFields(row *S, fieldNames []string) bool {
	var all []string
	zlog.Assert(len(v.fields) != 0)
	for _, fieldName := range fieldNames {
		f, _ := v.findField(fieldName)
		if zlog.ErrorIf(f == nil, fieldName, all) {
			continue
		}
		var rowFieldVal string
		canLock := f.HasFlag(zfields.FlagHeaderLockable)
		finfo, found := zreflect.FieldForName(row, zfields.FlattenIfAnonymousOrZUITag, fieldName)
		if canLock || v.currentLowerCaseSearchText != "" {
			zlog.Assert(found, fieldName)
			rowFieldVal = fmt.Sprint(finfo.ReflectValue.Interface())
		}
		if canLock {
			val, has := v.lockedFieldValues[fieldName]
			// if fieldName == "End" {
			// 	zlog.Info("Filter:", has, val, finfo.ReflectValue.Interface())
			// }
			if has {
				// zlog.Info("FilterRowWithZFieldInfo has:", fieldName)
				if f.HasFlag(zfields.FlagIsTimeBoundry) {
					filterTime := val.(time.Time)
					rowTime, gotTime := finfo.ReflectValue.Interface().(time.Time)
					// zlog.Info("Has Time Filter:", filterTime, rowTime, gotTime, filterTime.Sub(rowTime))
					if gotTime {
						if f.HasFlag(zfields.FlagIsStart) && rowTime.Before(filterTime) {
							return false
						}
						if f.HasFlag(zfields.FlagIsEnd) && rowTime.After(filterTime) {
							return false
						}
						// zlog.Info("OK Time:", filterTime, rowTime, filterTime.Sub(rowTime))
					}
					continue
				}
				strVals, got := val.([]string)
				if zlog.ErrorIf(!got, reflect.TypeOf(val)) {
					continue
				}
				equal := zstr.StringsContain(strVals, rowFieldVal)
				// zlog.Info("FilterRowWithZFieldInfo lock:", fieldName, rowFieldVal, strVals, equal)
				if !equal {
					return false
				}
			}
		}
		if v.currentLowerCaseSearchText != "" {
			if !strings.Contains(strings.ToLower(rowFieldVal), v.currentLowerCaseSearchText) {
				return false
			}
		}
	}
	return true
}

// LockColumn sets or clears the lock value for a field.
// Set clearSkipCache to call ClearFilterSkipCache() on first field locked to clear the filter skip cache.
// Set updateTable to call UpdateViewFunc() on last field locked.
func (v *TableView[S]) LockColumn(fieldName string, setLocked bool, lockVal any, clearSkipCache, updateTable bool) {
	if clearSkipCache {
		v.ClearFilterSkipCache()
	}
	v.Grid.RestoreTopSelectedRowOnNextLayout = true
	headerButton := v.Header.ColumnView(fieldName)
	label, lock := zheader.GetLockViews(headerButton)
	defer func() {
		lock.Expose()
		headerButton.ArrangeChildren()
		if updateTable {
			ztimer.StartIn(0.1, func() {
				v.UpdateViewFunc()
			})
		}
	}()
	if !setLocked {
		delete(v.lockedFieldValues, fieldName)
		label.SetText("")
		label.SetToolTip("")
		lock.TintColor.Valid = false
		lock.SetToolTip("press to lock on selected rows")
		v.Grid.HandleSelectionChangedFunc()
		return
	}
	lock.TintColor = zgeo.ColorYellow.WithOpacity(0.2)
	v.lockedFieldValues[fieldName] = lockVal
	f, _ := v.findField(fieldName)
	zlog.Assert(f != nil, fieldName)
	t, _ := lockVal.(time.Time)
	// zlog.Info("LockColumn:", fieldName, t)
	label.Show(true)
	lock.Show(true)
	lock.SetToolTip("press to unlock this field")
	if !t.IsZero() {
		t = t.Local()
		emojii := ztime.TimeToNearestEmojii(t)
		tip := ">="
		if f.HasFlag(zfields.FlagIsEnd) {
			tip = "<="
		}
		tip += " " + ztime.GetNice(t, false)
		label.SetText(string(emojii))
		label.SetToolTip(tip)
		return
	}
	uniqueVals, _ := lockVal.([]string)
	var tip string
	for i, u := range uniqueVals {
		if i != 0 {
			tip += "\n"
		}
		tip += zstr.TruncatedFromEnd(u, 60, "…")
	}
	label.SetText(strconv.Itoa(len(uniqueVals)))
	label.SetToolTip(tip)
}

func (v *TableView[S]) HandleLockPressedWithZField(fieldName string) {
	var uniqueVals []string
	clearCache := true
	updateTable := true
	_, has := v.lockedFieldValues[fieldName]
	if has {
		v.LockColumn(fieldName, false, nil, clearCache, updateTable)
		if v.AfterLockPressedFunc != nil {
			v.AfterLockPressedFunc(fieldName, false)
		}
		return
	}
	f, _ := v.findField(fieldName)
	zlog.Assert(f != nil, fieldName)
	for _, sid := range v.Grid.SelectedIDs() {
		s := v.StructForID(sid)
		finfo, found := zreflect.FieldForName(s, zfields.FlattenIfAnonymousOrZUITag, fieldName)
		zlog.Assert(found, fieldName)
		if f.HasFlag(zfields.FlagIsTimeBoundry) {
			t, _ := finfo.ReflectValue.Interface().(time.Time)
			v.LockColumn(fieldName, true, t, clearCache, updateTable)
			if v.AfterLockPressedFunc != nil {
				v.AfterLockPressedFunc(fieldName, true)
			}
			return
		}
		str := fmt.Sprint(finfo.ReflectValue.Interface())
		zstr.AddToSet(&uniqueVals, str)
	}
	v.LockColumn(fieldName, true, uniqueVals, clearCache, updateTable)
	if v.AfterLockPressedFunc != nil {
		v.AfterLockPressedFunc(fieldName, true)
	}
}
