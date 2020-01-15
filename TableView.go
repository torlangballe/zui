package zui

import (
	"fmt"
	"reflect"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zreflect"
)

type TableView struct {
	StackView
	List          *ListView
	Header        *HeaderView
	ColumnMargin  float64
	RowInset      float64
	DefaultHeight float64
	HeaderHeight  float64

	GetRowCount  func() int
	GetRowHeight func(i int) float64
	GetRowData   func(i int) interface{}
	RowUpdated   func(edited bool, i int, rowView *StackView) bool
	//	RowDataUpdated func(i int)
	HeaderPressed func(id string)
	CellPressed   func(i int, id string)

	fields []Field
}

func tableGetSliceFromPointer(structure interface{}) reflect.Value {
	rval := reflect.ValueOf(structure)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
		if rval.Kind() == reflect.Slice {
			return rval
		}
	}
	var n *int
	return reflect.ValueOf(n)
}

func TableViewNew(name string, header bool, inStruct interface{}) *TableView {
	var structure interface{}

	v := &TableView{}
	v.StackView.init(v, name)
	v.Vertical = true
	v.ColumnMargin = 3
	v.RowInset = 7
	v.HeaderHeight = 28
	v.DefaultHeight = 34
	unnestAnon := true
	recursive := false

	rval := tableGetSliceFromPointer(inStruct)
	if !rval.IsNil() {
		v.GetRowCount = func() int {
			return tableGetSliceFromPointer(inStruct).Len()
		}
		v.GetRowData = func(i int) interface{} {
			val := tableGetSliceFromPointer(inStruct)
			if val.Len() != 0 {
				return val.Index(i).Addr().Interface()
			}
			return nil
		}
		if rval.Len() == 0 {
			structure = reflect.New(rval.Type().Elem()).Interface()
		} else {
			structure = rval.Index(0).Addr().Interface()
		}
	}

	froot, err := zreflect.ItterateStruct(structure, unnestAnon, recursive)
	if err != nil {
		panic(err)
	}
	for i, item := range froot.Children {
		var f Field
		if f.makeFromReflectItem(item, i) {
			v.fields = append(v.fields, f)
		}
	}
	if header {
		v.Header = HeaderViewNew()
		v.Add(zgeo.Left|zgeo.Top|zgeo.HorExpand, v.Header)
	}
	v.List = ListViewNew(name + ".list")
	v.List.SetMinSize(zgeo.Size{50, 50})
	v.List.RowColors = []zgeo.Color{zgeo.ColorNewGray(0.97, 1), zgeo.ColorNewGray(0.85, 1)}
	v.List.HandleScrolledToRows = func(y float64, first, last int) {
		v.ArrangeChildren(nil)
	}
	v.Add(zgeo.Left|zgeo.Top|zgeo.Expand, v.List)
	if !rval.IsNil() {
		v.List.RowUpdater = func(i int) {
			v.FlushDataToRow(i)
		}
	}
	v.GetRowHeight = func(i int) float64 { // default height
		return 50
	}
	v.List.CreateRow = func(rowSize zgeo.Size, i int) View {
		return createRow(v, rowSize, i)
	}
	v.List.GetRowHeight = func(i int) float64 {
		return v.GetRowHeight(i)
	}
	v.GetRowHeight = func(i int) float64 {
		return v.DefaultHeight
	}
	v.List.GetRowCount = func() int {
		return v.GetRowCount()
	}

	return v
}

func (v *TableView) SetRect(rect zgeo.Rect) View {
	v.StackView.SetRect(rect)
	if v.GetRowCount() > 0 && v.Header != nil {
		stack := v.List.GetVisibleRowViewFromIndex(0).(*StackView)
		v.Header.FitToRowStack(stack, v.ColumnMargin)
	}
	return v
}

func (v *TableView) ReadyToShow() {
	if v.Header != nil {
		v.Header.Populate(v.fields, v.HeaderHeight, v.HeaderPressed)
	}
}

func (v *TableView) Reload() {
	v.List.ReloadData()
}

func (v *TableView) Margin(r zgeo.Rect) *TableView {
	v.List.ScrollView.Margin = r
	return v
}

func (v *TableView) SetStructureList(list interface{}) {
	vs := reflect.ValueOf(list)
	v.GetRowCount = func() int {
		return vs.Len()
	}
	v.GetRowData = func(i int) interface{} {
		if vs.Len() != 0 {
			return vs.Index(i).Addr().Interface()
		}
		return nil
	}
}

func (v *TableView) FlashRow() {

}

func (v *TableView) FlushDataToRow(i int) {
	rowStack := v.List.GetVisibleRowViewFromIndex(i).(*StackView)
	if rowStack != nil {
		rowStruct := v.GetRowData(i)
		fieldsUpdateStack(rowStack, rowStruct, &v.fields)
	}
}

func createRow(v *TableView, rowSize zgeo.Size, i int) View {
	name := fmt.Sprintf("row %d", i)
	rowStack := StackNewHor(name)
	rowStack.SetSpacing(0)
	rowStack.CanFocus(true)
	rowStack.SetMargin(zgeo.RectMake(v.RowInset, 0, -v.RowInset, 0))
	rowStruct := v.GetRowData(i)
	useWidth := true //(v.Header != nil)
	fieldsBuildStack(nil, rowStack, rowStruct, nil, &v.fields, zgeo.Center, zgeo.Size{v.ColumnMargin, 0}, useWidth, v.RowInset, i, func(i int, a FieldActionType, id string) {
		rowStruct := v.GetRowData(i)
		FieldsCopyBack(rowStruct, v.fields, rowStack, true)
		switch a {
		case FieldUpdateAction:
			if v.RowUpdated != nil {
				edited := true
				if v.RowUpdated(edited, i, rowStack) {
					fieldsUpdateStack(rowStack, rowStruct, &v.fields)
				}
			}
		case FieldPressedAction:
			if v.CellPressed != nil {
				v.CellPressed(i, id)
			}
		}
	})
	edited := false
	v.RowUpdated(edited, i, rowStack)
	fieldsUpdateStack(rowStack, v.GetRowData(i), &v.fields)
	return rowStack
}
