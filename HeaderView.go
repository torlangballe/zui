package zui

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
)

type HeaderView struct {
	StackView
}

func HeaderViewNew() *HeaderView {
	v := &HeaderView{}
	v.StackView.init(v, "header")
	v.Vertical = false
	v.SetSpacing(6)

	return v
}

func (v *HeaderView) Populate(fields []Field, height float64, pressed func(id string)) {
	for i, f := range fields {
		if f.Height == 0 {
			fields[i].Height = height - 6
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
				iv.SetObjectName(f.ID + ".image")
				button.Add(zgeo.Center, iv)
			}
		}
		//		button.Text(f.name)
		cell.View = button
		if pressed != nil {
			id := f.ID // nned to get actual ID here, not just f.ID (f is pointer)
			button.SetPressedHandler(func() {
				pressed(id)
			})
		}
		zfloat.Maximize(&fields[i].MinWidth, button.GetCalculatedSize(zgeo.Size{}).W)
		if f.MaxWidth != 0 {
			cell.MaxSize.W = math.Max(f.MaxWidth, f.MinWidth)
		}
		// if f.MinWidth != 0 {
		// 	cell.MinSize.W = math.Max(f.MinWidth, fields[i].MinWidth)
		// }
		v.AddCell(cell, -1)
	}
}

func (v *HeaderView) FitToRowStack(stack *StackView, marg float64) {
	children := stack.GetChildren()
	x := 0.0
	w := stack.Rect().Size.W
	for i := range children {
		var e float64
		if i < len(children)-1 {
			e = children[i+1].Rect().Pos.X
			e -= marg
		} else {
			e = w
		}
		hv := v.cells[i].View
		hr := hv.Rect()
		hr.Pos.X = x
		hr.SetMaxX(e)
		x = e
		hv.SetRect(hr)
		// fmt.Println("TABLE View rect item:", child.GetObjectName(), hv.Rect())
	}
}
