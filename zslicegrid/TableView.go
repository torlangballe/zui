//go:build zui

package zslicegrid

import (
	"strings"

	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zheader"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zslice"
	"github.com/torlangballe/zutil/zstr"
)

const (
	AddHeader      OptionType = LastBaseOption
	AddBarInHeader OptionType = LastBaseOption << 1
)

// TableView is a SliceGridView which creates rows from structs using the zfields package.
// See zfields for details on how to tag struct fields with `zui:"xxx"` for styling.
// if the AddHeader option is set, it adds header on top using zheader.HeaderView.
type TableView[S zstr.StrIDer] struct {
	SliceGridView[S]
	Header          *zheader.HeaderView // Optional header based on S struct
	ColumnMargin    float64             // Margin between columns
	RowInset        float64             // inset on far left and right
	FieldParameters zfields.FieldViewParameters
	fields          []zfields.Field // the fields in an S struct used to generate columns for the table
}

func TableViewNew[S zstr.StrIDer](s *[]S, name string, options OptionType) *TableView[S] {
	v := &TableView[S]{}
	v.Init(v, s, name, options)
	return v
}

func (v *TableView[S]) Init(view zview.View, s *[]S, storeName string, options OptionType) {
	v.SliceGridView.Init(view, s, storeName, options)
	v.Grid.MaxColumns = 1
	v.Grid.SetMargin(zgeo.Rect{})
	v.ColumnMargin = 5
	v.RowInset = 7 //RowInset not used yet, should be Grid margin, but calculated OnReady
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
	cell.Margin.H = -1 // this seems to make no 1-space beteen header and table
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
		if v.options&AddBarInHeader != 0 {
			v.RemoveChild(v.Bar)
			v.Bar.SetMargin(zgeo.RectFromXY2(6, 2, -6, -3))
		}
		v.SliceGridView.AddAdvanced(v.Header, zgeo.Left|zgeo.Top|zgeo.HorExpand, zgeo.SizeNull, zgeo.SizeNull, index, false)
		if v.Bar != nil && options&AddBarInHeader != 0 {
			v.Header.Add(v.Bar, zgeo.CenterRight).Free = true
			v.Bar.SetZIndex(200)
		}
	}
	// zlog.Info("TableInit", v.ObjectName(), v.SliceGridView.Grid.CreateCellFunc != nil)
}

func (v *TableView[S]) ArrangeChildren() {
	// zlog.Info("TV ArrangeChildren", v.Hierarchy(), v.Rect())
	// defer zlog.Info("TV ArrangeChildren Done", v.Hierarchy(), v.Rect())
	v.SliceGridView.ArrangeChildren()
	if v.Header == nil {
		return
	}
	freeOnly := true
	v.Header.ArrangeAdvanced(freeOnly)
	if v.Header != nil {
		// zlog.Info("TV: ArrangeChildren", v.Header != nil, v.Grid.CellCount())
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
			v.Header.FitToRowStack(&fv.StackView)
		}
	}
}

func (v *TableView[S]) ReadyToShow(beforeWindow bool) {
	v.SliceGridView.ReadyToShow(beforeWindow)
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
	}
	v.Grid.UpdateCellFunc = func(grid *zgridlist.GridListView, id string) {
		fv := grid.CellView(id).(*zfields.FieldView)
		zlog.Assert(fv != nil)
		fv.Update(v.StructForID(id), true, false)
		// fv.ArrangeChildren()
	}
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
