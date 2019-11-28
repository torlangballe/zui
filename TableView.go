package zgo

import (
	"fmt"
	"math"
	"reflect"

	"github.com/torlangballe/zutil/zmath"
	"github.com/torlangballe/zutil/zreflect"
)

type TableView struct {
	StackView
	List          *ListView
	Header        *StackView
	ColumnMargin  float64
	RowInset      float64
	DefaultHeight float64

	GetRowCount   func() int
	GetRowHeight  func(i int) float64
	GetRow        func(i int) interface{}
	HeaderPressed func(id string, i int)
	RowUpdated    func(edited bool, i int) bool

	fields []field
}

func tableGetSliceFromPonter(structure interface{}) reflect.Value {
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
	v.DefaultHeight = 34
	unnestAnon := true
	recursive := false

	rval := tableGetSliceFromPonter(inStruct)
	if !rval.IsNil() {
		v.GetRowCount = func() int {
			return tableGetSliceFromPonter(inStruct).Len()
		}
		v.GetRow = func(i int) interface{} {
			val := tableGetSliceFromPonter(inStruct)
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
		var f field
		if f.makeFromReflectItem(item, i) {
			v.fields = append(v.fields, f)
		}
	}
	if header {
		v.Header = StackViewNew(false, 0, "header")
		v.Header.BGColor(ColorBlue)
		v.Header.Spacing(0)
		v.AddElements(AlignmentLeft|AlignmentTop|AlignmentHorExpand|AlignmentNonProp, v.Header)
	}
	v.List = ListViewNew(name + ".list")
	v.List.MinSize(Size{50, 50})
	v.List.RowColors = []Color{ColorNewGray(0.97, 1), ColorNewGray(0.85, 1)}
	v.List.HandleScrolledToRows = func(y float64, first, last int) {
		v.ArrangeChildren(nil)
	}
	v.AddElements(AlignmentLeft|AlignmentTop|AlignmentExpand|AlignmentNonProp, v.List)

	v.GetRowHeight = func(i int) float64 { // default height
		return 50
	}
	v.List.CreateRow = func(rowSize Size, i int) View {
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

func (v *TableView) ReadyToShow() {
	for i, f := range v.fields {
		if f.Height == 0 {
			v.fields[i].Height = v.GetRowHeight(i) - 6
		}
		s := Size{f.MinWidth, 28}
		cell := ContainerViewCell{}
		exp := AlignmentNone
		if f.Kind == zreflect.KindString && f.Enum == nil {
			exp = AlignmentHorExpand
		}
		t := ""
		if f.Flags&fieldsNoHeader == 0 {
			t = f.Title
			if t == "" {
				t = f.Name
			}
		}
		cell.Alignment = AlignmentLeft | AlignmentVertCenter | exp
		button := ButtonNew(t, "grayHeader", s, Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
		//		button.Text(f.name)
		cell.View = button
		zmath.Maximize(&v.fields[i].MinWidth, button.GetCalculatedSize(Size{}).W)
		if f.MaxWidth != 0 {
			cell.MaxSize.W = math.Max(f.MaxWidth, v.fields[i].MinWidth)

		}
		v.Header.AddCell(cell, -1)
	}
}

func (v *TableView) Reload() {
	v.List.ReloadData()
}

func (v *TableView) Margin(r Rect) *TableView {
	v.List.ScrollView.Margin = r
	return v
}

func (v *TableView) SetStructureList(list interface{}) {
	vs := reflect.ValueOf(list)
	v.GetRowCount = func() int {
		return vs.Len()
	}
	v.GetRow = func(i int) interface{} {
		if vs.Len() != 0 {
			return vs.Index(i).Addr().Interface()
		}
		return nil
	}
}

func (v *TableView) FlashRow() {

}

func createRow(v *TableView, rowSize Size, i int) View {
	name := fmt.Sprintf("row %d", i)
	rowStack := StackViewNew(false, AlignmentNone, name)
	rowStack.Spacing(0)
	rowStack.CanFocus(true)
	rowStack.Margin(RectMake(v.RowInset, 0, -v.RowInset, 0))
	rowStruct := v.GetRow(i)
	useWidth := (v.Header != nil)
	fieldsBuildStack(nil, rowStack, rowStruct, nil, &v.fields, AlignmentCenter, Size{v.ColumnMargin, 0}, useWidth, v.RowInset, i, func(i int) {
		rowStruct := v.GetRow(i)
		FieldsCopyBack(rowStruct, v.fields, rowStack, true)
		if v.RowUpdated != nil {
			edited := true
			if v.RowUpdated(edited, i) {
				fieldsUpdateStack(rowStack, rowStruct, &v.fields)
			}
		}
	})
	edited := false
	v.RowUpdated(edited, i)
	fieldsUpdateStack(rowStack, v.GetRow(i), &v.fields)
	return rowStack
}
