package zgo

import (
	"fmt"

	"github.com/torlangballe/zutil/zreflect"
)

type field struct {
	name      string
	width     float64
	maxWidth  float64
	kind      zreflect.TypeKind
	alignment Alignment
	format    string
}

type TableView struct {
	StackView
	list          *ListView
	fields        []field
	Header        *StackView
	RowColors     []Color
	ColumnSpacing float64

	GetRowCount  func() int
	GetRowHeight func(i int) float64
	GetRow       func(i int) interface{}
}

func TableViewNew(name string, header bool, structure interface{}) *TableView {
	v := &TableView{}
	v.StackView.init(v, name)

	v.ColumnSpacing = 8
	v.RowColors = []Color{ColorNewGray(0.97, 1), ColorNew(0.9, 1, 1, 1)}
	froot, err := zreflect.ItterateStruct(structure, true)
	if err != nil {
		panic(err)
	}
	for _, item := range froot.Children {
		v.fields = append(v.fields, makeFieldFromReflectItem(item))
	}
	if header {
		v.Header = StackViewNew(false, 0)
		v.AddElements(AlignmentLeft|AlignmentVertCenter|AlignmentHorExpand, v.Header)
	}
	v.list = ListViewNew(name + ".list")
	v.list.MinSize(Size{50, 50})
	v.AddElements(AlignmentLeft|AlignmentTop|AlignmentExpand|AlignmentNonProp, v.list)

	v.list.GetRowCount = func() int {
		return v.GetRowCount()
	}
	v.list.GetRowHeight = func(i int) float64 {
		return v.GetRowHeight(i)
	}
	v.GetRowHeight = func(i int) float64 { // default height
		return 50
	}
	v.list.CreateRow = func(rowSize Size, i int) View {
		rowStack := StackViewNew(false, AlignmentLeft|AlignmentVertCenter|AlignmentHorExpand)
		rowStack.Spacing(v.ColumnSpacing)
		rowStack.ObjectName(fmt.Sprintf("row %d", i))
		rowData := v.GetRow(i)
		col := ColorGray
		if len(v.RowColors) != 0 {
			col = v.RowColors[i%len(v.RowColors)]
		}
		rowStack.BGColor(col)

		rootData, err := zreflect.ItterateStruct(rowData, true)
		if err != nil {
			panic(err)
		}
		for j, item := range rootData.Children {
			f := v.fields[j]
			exp := AlignmentNone
			var v View
			switch f.kind {
			case zreflect.KindBool:
				b := BoolIndMake(item.Value.Interface().(bool))
				fmt.Println("Bool:", b, item.TypeName)
				v = SwitchNew(b)

			case zreflect.KindInt:
				if item.TypeName == "BoolInd" {
					v = SwitchNew(BoolInd(item.Value.Int()))
				} else {
					v = makeLabel(f, item)
					exp = AlignmentHorExpand
				}

			case zreflect.KindFloat, zreflect.KindString, zreflect.KindTime:
				v = makeLabel(f, item)
				exp = AlignmentHorExpand

			default:
				panic(fmt.Sprint("bad type: ", f.kind))
			}
			v.ObjectName(fmt.Sprintf("row:%d item:%d", i, j))
			cell := ContainerViewCell{}
			cell.Alignment = AlignmentLeft | AlignmentVertCenter | exp
			cell.MaxSize.W = f.maxWidth
			cell.View = v
			rowStack.AddCell(cell, -1)
		}
		return rowStack
	}

	return v
}

func makeLabel(f field, item zreflect.Item) *Label {
	format := f.format
	if format == "" {
		format = "%v"
	}
	return LabelNew(fmt.Sprintf(format, item.Value.Interface()))
}
func makeFieldFromReflectItem(item zreflect.Item) field {
	var f field
	f.name = item.FieldName
	f.kind = item.Kind
	f.alignment = AlignmentLeft
	switch item.Kind {
	case zreflect.KindBool:
		f.width = 20
		f.maxWidth = 20
	case zreflect.KindFloat:
		f.width = 44
		f.maxWidth = 60
	case zreflect.KindInt:
		f.width = 40
		f.maxWidth = 60
	case zreflect.KindString:
		f.width = 100
	case zreflect.KindTime:
		f.width = 80
		f.maxWidth = 80
	}
	return f
}
