//go:build zui

package zcontainer

import (
	"slices"

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zslice"
)

//  Created by Tor Langballe on /20/10/15.

type StackView struct {
	ContainerView
	spacing             float64
	Vertical            bool
	GridVerticalSpace   float64
	NoCalculatedMaxSize zgeo.BoolSize
	MaxColumns          int // If set, it wraps, creating a grid, assuming all cells are the same size
}

func StackViewNew(vertical bool, name string) *StackView {
	s := &StackView{}
	s.Init(s, vertical, name)
	return s
}

func (v *StackView) Init(view zview.View, vertical bool, name string) {
	v.ContainerView.Init(view, name)
	v.Vertical = vertical
	v.spacing = 6
	v.GridVerticalSpace = zfloat.Undefined
}

func StackViewVert(name string) *StackView {
	return StackViewNew(true, name)
}

func StackViewHor(name string) *StackView {
	return StackViewNew(false, name)
}

func (v *StackView) SetSpacing(spacing float64) *StackView {
	v.spacing = spacing
	return v
}

func (v *StackView) Spacing() float64 {
	return v.spacing
}

func (v *StackView) calculateMaxColumnSize(total zgeo.Size) zgeo.Size {
	var child zview.View
	var count int
	for _, c := range v.Cells {
		if c.Alignment == 0 || c.Free || c.View == nil {
			continue
		}
		child = c.View
		count++
	}
	if child == nil {
		return zgeo.SizeNull
	}
	childSize, _ := child.CalculatedSize(total)
	spacing := zgeo.SizeBoth(v.Spacing())
	y := (count + v.MaxColumns - 1) / v.MaxColumns
	x, y := zgeo.MaxCellsInSize(spacing, childSize, total)

	w := float64(x)*childSize.W + float64(x-1)*spacing.W
	h := float64(y)*childSize.H + float64(y-1)*spacing.H
	return zgeo.SizeD(w, h)
}

func (v *StackView) calculateGridSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	row, heights := v.getGridLayoutRow(total)
	vertical := false
	sl, _ := zgeo.LayoutGetCellsStackedSize(v.ObjectName(), vertical, v.spacing, row)
	// zlog.Info("calculateGridSize grid1:", v.Hierarchy(), row, sl)
	s.W = sl.W
	j := 0
	for _, vc := range v.Cells {
		if vc.Collapsed || vc.View == nil || vc.Free {
			continue
		}
		if vc.NotInGrid {
			notS, _ := vc.View.CalculatedSize(total)
			zfloat.Maximize(&s.W, notS.W)
		}
		s.H += vc.Margin.Size.H
		s.H += v.GridVerticalSpace
		j++
	}
	for _, h := range heights {
		s.H += h
	}
	s.Subtract(v.margin.Size)
	return s
}

func (v *StackView) CalculatedSize(total zgeo.Size) (s, max zgeo.Size) {
	// if v.ObjectName() == "block-stack" {
	// 	zlog.Info("CalcSize:", total, v.MinSize())
	// }
	var ws float64
	if len(v.Cells) != 0 {
		if v.MaxColumns > 0 {
			s := v.calculateMaxColumnSize(total)
			return s, s
		}
		if v.GridVerticalSpace != zfloat.Undefined {
			s := v.calculateGridSize(total)
			return s, s
		}
		lays := v.getLayoutCells(total)
		spacing := v.Spacing()
		if v.GridVerticalSpace != zfloat.Undefined {
			spacing = v.GridVerticalSpace
		}
		s, max = zgeo.LayoutGetCellsStackedSize(v.ObjectName(), v.Vertical, spacing, lays)
		zfloat.Maximize(&s.W, ws) // why for W only?
	}
	ms := v.Margin().Size.Negative()
	s.Add(ms)
	if v.NoCalculatedMaxSize.W {
		max.W = 0
	}
	if v.NoCalculatedMaxSize.H {
		max.H = 0
	}
	if max.W != 0 {
		max.W += ms.W
	}
	if max.H != 0 {
		max.H += ms.H
	}
	s.Maximize(v.MinSize())
	return s, max
}

func (v *StackView) getLayoutCells(total zgeo.Size) (lays []zgeo.LayoutCell) {
	// if v.ObjectName() == "zheader" {
	// 	zlog.Info("getLayoutCells", v.ObjectName(), total)
	// }
	// var nonMax bool
	for _, c := range v.Cells {
		if c.Free || c.View == nil {
			continue
		}
		l := c.LayoutCell
		// zlog.Info("StackView getLayoutCells:", v.Hierarchy(), c.Collapsed, c.Name, c.View != nil)
		var max zgeo.Size
		l.OriginalSize, max = c.View.CalculatedSize(total)
		// if v.ObjectName() == "eventgrid" {
		// 	zlog.Info("StackView getLayoutCells:", c.View.ObjectName(), l.OriginalSize, max, l.MaxSize)
		// }
		if c.MaxSize.W == 0 && !v.NoCalculatedMaxSize.W && max.W > 0 {
			zfloat.Minimize(&l.MaxSize.W, max.W)
		}
		if c.MaxSize.H == 0 && !v.NoCalculatedMaxSize.H && max.H > 0 {
			zfloat.Minimize(&l.MaxSize.H, max.H)
		}
		l.Name = c.View.ObjectName()
		// if l.MaxSize.H == 0 {
		// 	nonMax = true
		// }
		// if l.MaxSize.H != 0 {
		// zlog.Info("StackView getLayoutCells:", v.ObjectName(), c.View.ObjectName(), l.OriginalSize)
		div, _ := c.View.(*DividerView)
		if div != nil {
			// zlog.Info("StackView getLayoutCells:", v.ObjectName(), div.Ratio)
			l.DividerRatio = div.Ratio
		}
		lays = append(lays, l)
	}
	// if v.ObjectName() == "zheader" {
	// 	zlog.Info("getLayoutCells end", v.ObjectName(), "nonMax:", nonMax)
	// }
	// zlog.Info("Layout Stack getCells end", v.ObjectName(), zlog.Full(lays))
	return
}

func (v *StackView) arrangeChildrenInGrid() {
	rect := v.LocalRect().Plus(v.margin)
	row, heights := v.getGridLayoutRow(rect.Size)
	j := 0
	r := rect
	for _, vc := range v.Cells {
		if vc.Collapsed || vc.View == nil || vc.Free {
			continue
		}
		r.Size.H = heights[j]
		if !vc.NotInGrid {
			if vc.Alignment&(zgeo.Top|zgeo.VertCenter) != 0 {
				r.Pos.Y += vc.Margin.Min().Y
				cellsOwner, _ := vc.View.(CellsOwner)
				zlog.Assert(cellsOwner != nil, v.Hierarchy())
				rowCells := slices.Clone(*cellsOwner.GetCells())
				zslice.DeleteFromFunc(&rowCells, func(c Cell) bool {
					return c.Collapsed || c.Free || c.View == nil
				})
				rowView := vc.View
				rowView.Native().SetRect(r)
				rbox := r
				mo, _ := rowView.(zview.MarginOwner)
				rbox.Pos = zgeo.Pos{}
				if mo != nil {
					rbox.Add(mo.Margin())
				}
				rects := zgeo.LayoutCellsInStack(v.ObjectName(), rbox, false, v.spacing, row)
				for i := range row {
					box := rects[i]
					if box.IsNull() {
						continue
					}
					zfloat.Maximize(&box.Size.H, heights[j])
					box = box.MovedInto(rbox) // TODO: Why do we need to do this? Cause must be in  LayoutCellsInStack?
					if i >= len(rowCells) {
						continue
					}
					cell := (rowCells)[i]
					s, _ := cell.View.CalculatedSize(box.Size)
					ar := box.AlignPro(s, cell.Alignment, cell.Margin, cell.MaxSize, cell.MinSize)
					ar = ar.Intersected(box)
					cell.View.SetRect(ar)
				}
			}
		} else {
			v.ArrangeChild(vc, r)
		}
		r.Pos.Y = r.Max().Y + v.GridVerticalSpace
		j++
	}
}

func (v *StackView) getGridLayoutRow(total zgeo.Size) (row []zgeo.LayoutCell, heights []float64) {
	firstRow := true
	// zlog.Info("getGridLayoutRow1:", len(v.Cells))
	for _, vc := range v.Cells {
		if vc.Collapsed || vc.View == nil || vc.Free {
			continue
		}
		if vc.NotInGrid {
			s, _ := vc.View.CalculatedSize(total)
			heights = append(heights, s.H)
			continue
		}
		// zlog.Info("getGridLayoutRow2:", j)
		cellsOwner, _ := vc.View.(CellsOwner)
		zlog.Assert(cellsOwner != nil, v.Hierarchy())
		rowCells := slices.Clone(*cellsOwner.GetCells())
		zslice.DeleteFromFunc(&rowCells, func(c Cell) bool {
			return c.Collapsed || c.Free || c.View == nil
		})
		i := 0
		var height float64
		// zlog.Info("getGridLayoutRow:", v.Hierarchy(), j, vc.View.ObjectName(), i, len(rowCells), len(row))
		for _, rc := range rowCells {
			l := rc.LayoutCell
			var max zgeo.Size
			l.OriginalSize, max = rc.View.CalculatedSize(total)
			if firstRow {
				l.Alignment = rc.Alignment
				l.Name = rc.View.ObjectName()
				l.MaxSize = max
				l.MinSize = rc.MinSize
				row = append(row, l)
			} else {
				// zlog.Info("getGridLayoutRow:", i, l.Name, len(row))
				row[i].OriginalSize.Maximize(l.OriginalSize)
				if rc.MaxSize.W == 0 {
					row[i].MaxSize.W = 0
				} else if row[i].MaxSize.W != 0 {
					zfloat.Maximize(&row[i].MaxSize.W, rc.MaxSize.W)
				}
				zfloat.Maximize(&row[i].MaxSize.H, rc.MaxSize.H)
				row[i].MinSize.Maximize(rc.MinSize)
				// zlog.Info("getGridLayoutRow:", j, i, l.OriginalSize.H)
			}
			zfloat.Maximize(&height, l.OriginalSize.H)
			i++
		}
		heights = append(heights, height)
		firstRow = false
	}
	return row, heights
}

func (v *StackView) ArrangeChildren() {
	// zlog.Info("*********** Stack.ArrangeChildren:", v.Hierarchy(), v.Rect(), len(v.Cells), reflect.TypeOf(v))
	// zlog.PushProfile(v.ObjectName())
	rm := v.LocalRect().Plus(v.Margin())
	if v.GridVerticalSpace != zfloat.Undefined {
		zlog.Assert(v.Vertical, v.Hierarchy(), v.GridVerticalSpace)
		v.arrangeChildrenInGrid()
		for _, c := range v.Cells {
			if c.View != nil && !c.Free && !c.Collapsed {
				co, _ := c.View.(CellsOwner)
				ca, _ := c.View.(ChildArranger)
				if co != nil && ca != nil {
					for _, c := range *co.GetCells() {
						if c.Free && c.View != nil && !c.Collapsed {
							ca.ArrangeChild(c, rm)
						}
					}
				}
			}
		}
		return
	}
	layouter, _ := v.View.(zview.Layouter)
	if layouter != nil {
		layouter.HandleBeforeLayout()
	}
	lays := v.getLayoutCells(rm.Size)
	// if v.ObjectName() == "bar" {
	// 	zlog.Info("arrangeChildren1:", rm, v.Margin(), v.ObjectName())
	// }
	rects := zgeo.LayoutCellsInStack(v.ObjectName(), rm, v.Vertical, v.spacing, lays)
	j := 0
	for _, c := range v.Cells {
		if c.View == nil {
			continue
		}
		if c.Free {
			v.ArrangeChild(c, rm)
			continue
		}
		r := rects[j]
		// 	zlog.Info("Stack.ArrangeChild:", v.Hierarchy(), c.View.ObjectName(), r)
		if !r.IsNull() {
			// if v.ObjectName() == "hor-inf" {
			// 	zlog.Info("arrangeChildren:", rm, v.ObjectName(), c.View.ObjectName(), r, lays[j])
			// }
			c.View.SetRect(r)
			if c.ShowIfExtraSpace != 0 && !c.View.Native().IsShown() {
				c.View.Show(true)
			}
		} else {
			if c.View != nil && !c.Collapsed && c.Alignment != zgeo.AlignmentNone {
				// zlog.Info("Hide null rect:", c.View.ObjectName(), "in:", v.Hierarchy())
				c.View.Show(false)
			}
		}
		j++
	}
}

func MakeLinkedStack(surl, name string, add zview.View) *StackView {
	v := StackViewHor(name + "#type:a") // #type is hack that sets html-type of stack element
	v.SetChildrenAboveParent(true)
	v.SetMargin(zgeo.RectFromXY2(3, 0, -3, 0))
	v.MakeLink(surl, name)
	v.Add(add, zgeo.TopLeft, zgeo.SizeNull)
	return v
}
