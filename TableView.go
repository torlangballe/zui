package zui

import (
	"fmt"
	"math"
	"reflect"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
)

type TableView struct {
	StackView
	List          *ListView
	Header        *StackView
	ColumnMargin  float64
	RowInset      float64
	DefaultHeight float64

	GetRowCount    func() int
	GetRowHeight   func(i int) float64
	CreateRow      func(i int) interface{}
	RowUpdated     func(edited bool, i int, rowView *StackView) bool
	RowDataUpdated func(i int)
	HeaderPressed  func(id string)

	fields []field
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
	v.DefaultHeight = 34
	unnestAnon := true
	recursive := false

	rval := tableGetSliceFromPointer(inStruct)
	if !rval.IsNil() {
		v.GetRowCount = func() int {
			return tableGetSliceFromPointer(inStruct).Len()
		}
		v.CreateRow = func(i int) interface{} {
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
		var f field
		if f.makeFromReflectItem(item, i) {
			v.fields = append(v.fields, f)
		}
	}
	if header {
		v.Header = StackNewHor("header")
		v.Header.Spacing(0)
		v.Add(zgeo.Left|zgeo.Top|zgeo.HorExpand, v.Header)
	}
	v.List = ListViewNew(name + ".list")
	v.List.MinSize(zgeo.Size{50, 50})
	v.List.RowColors = []zgeo.Color{zgeo.ColorNewGray(0.97, 1), zgeo.ColorNewGray(0.85, 1)}
	v.List.HandleScrolledToRows = func(y float64, first, last int) {
		v.ArrangeChildren(nil)
	}
	v.Add(zgeo.Left|zgeo.Top|zgeo.Expand, v.List)

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
		children := stack.GetChildren()
		for i, child := range children {
			cr := child.GetRect()
			hv := v.Header.cells[i].View
			hr := hv.GetRect()
			hr.Pos.X = cr.Pos.X
			hr.Size.W = cr.Size.W
			hv.SetRect(hr)
			fmt.Println("TABLE View rect item:", child.GetObjectName(), hv.GetRect())
		}
	}
	return v
}

func (v *TableView) ReadyToShow() {
	for i, f := range v.fields {
		if f.Height == 0 {
			v.fields[i].Height = v.GetRowHeight(i) - 6
		}
		s := zgeo.Size{f.MinWidth, 28}
		cell := ContainerViewCell{}
		exp := zgeo.AlignmentNone
		if f.Kind == zreflect.KindString && f.Enum == nil {
			exp = zgeo.HorExpand
		}
		t := ""
		if f.Flags&(fieldHasHeaderImage|fieldsNoHeader) == 0 {
			t = f.Title
			if t == "" {
				t = f.Name
			}
		}
		cell.Alignment = zgeo.Left | zgeo.VertCenter | exp

		button := ButtonNew(t, "grayHeader", s, zgeo.Size{}) //ShapeViewNew(ShapeViewTypeRoundRect, s)
		if f.Flags&fieldHasHeaderImage != 0 {
			if f.FixedPath == "" {
				zlog.Error(nil, "no image path for header image field", f.Name)
			} else {
				iv := ImageViewNew(f.FixedPath, f.Size)
				iv.ObjectName(f.ID + ".image")
				button.Add(zgeo.Center, iv)
			}
		}
		//		button.Text(f.name)
		cell.View = button
		if v.HeaderPressed != nil {
			id := f.ID // nned to get actual ID here, not just f.ID (f is pointer)
			button.PressedHandler(func() {
				v.HeaderPressed(id)
			})
		}
		zfloat.Maximize(&v.fields[i].MinWidth, button.GetCalculatedSize(zgeo.Size{}).W)
		if f.MaxWidth != 0 {
			cell.MaxSize.W = math.Max(f.MaxWidth, f.MinWidth)
		}
		// if f.MinWidth != 0 {
		// 	cell.MinSize.W = math.Max(f.MinWidth, v.fields[i].MinWidth)
		// }
		v.Header.AddCell(cell, -1)
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
	v.CreateRow = func(i int) interface{} {
		if vs.Len() != 0 {
			return vs.Index(i).Addr().Interface()
		}
		return nil
	}
}

func (v *TableView) FlashRow() {

}

func createRow(v *TableView, rowSize zgeo.Size, i int) View {
	name := fmt.Sprintf("row %d", i)
	rowStack := StackNewHor(name)
	rowStack.Spacing(0)
	rowStack.CanFocus(true)
	rowStack.SetMargin(zgeo.RectMake(v.RowInset, 0, -v.RowInset, 0))
	rowStruct := v.CreateRow(i)
	useWidth := true //(v.Header != nil)
	fieldsBuildStack(nil, rowStack, rowStruct, nil, &v.fields, zgeo.Center, zgeo.Size{v.ColumnMargin, 0}, useWidth, v.RowInset, i, func(i int) {
		fmt.Println("createRow:", i, name)
		rowStruct := v.CreateRow(i)
		FieldsCopyBack(rowStruct, v.fields, rowStack, true)
		if v.RowUpdated != nil {
			edited := true
			if v.RowUpdated(edited, i, rowStack) {
				fieldsUpdateStack(rowStack, rowStruct, &v.fields)
			}
		}
	})
	edited := false
	v.RowUpdated(edited, i, rowStack)
	fieldsUpdateStack(rowStack, v.CreateRow(i), &v.fields)
	return rowStack
}
