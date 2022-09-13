// TableView is a SliceGridView which creates rows from structs using the zfields package.
// See zfields for details on how to tag struct fields with `zui:"xxx"` for styling.
// if the AddHeader option is set, it adds header on top using zheader.HeaderView.
//
package zslicegrid

import (
	"strings"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zfields"
	"github.com/torlangballe/zui/zgridlist"
	"github.com/torlangballe/zui/zheader"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zstr"
)

const (
	AddHeader      OptionType = 128
	AddBarInHeader OptionType = 256
)

type TableView[S zstr.StrIDer] struct {
	SliceGridView[S]
	Header       *zheader.HeaderView // Optional header based on S struct
	ColumnMargin float64             // Margin between columns
	RowInset     float64             // inset on far left and right
	// DefaultHeight         float64
	HeaderHeight          float64
	HeaderPressedFunc     func(fieldID string) // triggered if user presses in header. fieldID is zfield-based id of field header column is based on
	HeaderLongPressedFunc func(fieldID string) // Like HeaderPressedFunc
	FieldParameters       zfields.FieldViewParameters
	fields                []zfields.Field // the fields in an S struct used to generate columns for the table
	addFlags              OptionType
}

func TableViewNew[S zstr.StrIDer](s *[]S, name string, addFlags OptionType) *TableView[S] {
	v := &TableView[S]{}
	v.Init(v, s, name, addFlags)
	return v
}

func (v *TableView[S]) Init(view zview.View, s *[]S, storeName string, addFlags OptionType) {
	if addFlags&AddBarInHeader != 0 {
		zlog.Assert(addFlags&AddHeader != 0)
		v.Bar = zcontainer.StackViewHor("bar")
		v.Bar.SetSpacing(4)
	}
	v.SliceGridView.Init(view, s, storeName, addFlags)
	v.Grid.MaxColumns = 1
	v.Grid.SetMargin(zgeo.Rect{})
	v.SetSpacing(0)
	v.ColumnMargin = 5
	v.RowInset = 7
	v.HeaderHeight = 28
	v.FieldParameters = zfields.FieldViewParametersDefault()
	if addFlags&AddEditDelete != 0 {
		v.Grid.MultiSelectable = true
	}
	v.FieldParameters.AllTextStatic = true
	v.addFlags = addFlags
	// v.DefaultHeight = 30
	cell, _ := v.FindCellWithView(v.Grid)
	cell.Margin.H = -1
	v.Grid.CreateCellFunc = func(grid *zgridlist.GridListView, id string) zview.View {
		r := v.createRow(id)
		return r
	}
	if v.addFlags&AddHeader != 0 {
		v.Header = zheader.NewView(v.ObjectName() + ".header")
		index := 0
		if v.addFlags&AddBar != 0 {
			index = 1
		}
		v.SliceGridView.AddAdvanced(v.Header, zgeo.Left|zgeo.Top|zgeo.HorExpand, zgeo.Size{}, zgeo.Size{}, index, false)
		if v.Bar != nil && addFlags&AddBarInHeader != 0 {
			v.Header.Add(v.Bar, zgeo.CenterRight).Free = true
			v.Bar.SetZIndex(200)
		}
	}
	// zlog.Info("TableInit", v.ObjectName(), v.SliceGridView.Grid.CreateCellFunc != nil)
}

func (v *TableView[S]) ArrangeChildren() {
	// zlog.Info("TV ArrangeChildren", v.Rect())
	v.SliceGridView.ArrangeChildren()
	if v.Header == nil {
		return
	}
	freeOnly := true
	v.Header.ArrangeAdvanced(freeOnly)
	if v.Header != nil {
		// zlog.Info("TV: ArrangeChildren", v.Header != nil, v.Grid.CellCount())
		if v.Grid.CellCountFunc() > 0 {
			view := v.Grid.AnyChildView()
			if view != nil {
				fv := view.(*zfields.FieldView)
				v.Header.FitToRowStack(&fv.StackView, v.ColumnMargin)
			}
		} else { // no rows, make an empty one to fit header with
			var sss S
			view := v.createRowFromStruct(&sss, zstr.GenerateRandomHexBytes(10))
			fv := view.(*zfields.FieldView)
			fv.SetRect(v.LocalRect())
			fv.ArrangeChildren()
			v.Header.FitToRowStack(&fv.StackView, v.ColumnMargin)
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
	options := zreflect.Options{UnnestAnonymous: true, MakeSliceElementIfNone: true}
	froot, err := zreflect.ItterateStruct(v.slicePtr, options)
	if err != nil {
		panic(err)
	}
	for i, item := range froot.Children {
		f := zfields.EmptyField
		immediateEdit := false
		if f.SetFromReflectItem(v.slicePtr, item, i, immediateEdit) {
			if zstr.IndexOf(f.FieldName, v.FieldParameters.SkipFieldNames) != -1 {
				continue
			}
			v.fields = append(v.fields, f)
		}
	}
	if v.addFlags&AddHeader != 0 {
		v.SortFunc = func(s []S) {
			zfields.SortSliceWithFields(s, v.fields, v.Header.SortOrder)
		}
		v.Header.SortingPressedFunc = func() {
			v.SortFunc(*v.slicePtr)
			v.UpdateViewFunc()
			// for i, s := range *v.slice {
			// 	fmt.Printf("Sorted: %d %+v\n", i, s)
			// }
		}
		v.Grid.UpdateCellFunc = func(grid *zgridlist.GridListView, id string) {
			// zlog.Info("UpdateCell:", id)
			fv := grid.CellView(id).(*zfields.FieldView)
			zlog.Assert(fv != nil)
			fv.Update(v.StructForID(id), true)
		}
		headers := makeHeaderFields(v.fields, v.HeaderHeight)
		v.Header.Populate(headers)
		v.Header.HeaderPressedFunc = v.HeaderPressedFunc
		v.Header.HeaderLongPressedFunc = v.HeaderLongPressedFunc
	}
}

func (v *TableView[S]) createRow(id string) zview.View {
	s := v.StructForID(id)
	return v.createRowFromStruct(s, id)
}

func (v *TableView[S]) createRowFromStruct(s *S, id string) zview.View {
	name := "row " + id
	params := v.FieldParameters
	params.ImmediateEdit = false
	params.Styling.Spacing = 0
	params.AllTextStatic = (v.Grid.Selectable || v.Grid.MultiSelectable)
	fv := zfields.FieldViewNew(id, s, params)
	fv.Vertical = false
	fv.Fields = v.fields
	fv.SetSpacing(0)
	// fv.SetCanFocus(true)
	fv.SetMargin(zgeo.Rect{})
	useWidth := true //(v.Header != nil)
	fv.BuildStack(name, zgeo.CenterLeft, zgeo.Size{v.ColumnMargin, 0}, useWidth)
	dontOverwriteEdited := false
	fv.Update(nil, dontOverwriteEdited)
	return fv
}

func makeHeaderFields(fields []zfields.Field, height float64) []zheader.Header {
	var headers []zheader.Header
	for _, f := range fields {
		var h zheader.Header
		h.Height = f.Height
		h.ID = f.ID
		h.Align = zgeo.Left | zgeo.VertCenter
		h.Justify = f.Justify
		if f.Kind == zreflect.KindString && f.Enum == "" {
			h.Align |= zgeo.HorExpand
		}
		if f.Height == 0 {
			h.Height = height - 6
		}
		if f.Flags&zfields.FlagHasHeaderImage != 0 {
			h.ImageSize = f.HeaderSize
			if h.ImageSize.IsNull() {
				h.ImageSize = zgeo.SizeBoth(height - 8)
			}
			h.ImagePath = f.HeaderImageFixedPath
		}
		if f.Flags&(zfields.FlagHasHeaderImage|zfields.FlagNoTitle) == 0 {
			h.Title = f.Title
			if h.Title == "" {
				h.Title = f.Name
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
