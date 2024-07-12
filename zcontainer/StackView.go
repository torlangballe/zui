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
	v.GridVerticalSpace = zgeo.UndefValue
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

func (v *StackView) calculateGridSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	row, heights := v.getGridLayoutRow(total)
	// zlog.Info("calculateGridSize grid1:", v.Hierarchy(), row)
	vertical := false
	sl, _ := zgeo.LayoutGetCellsStackedSize(v.ObjectName(), vertical, v.spacing, row)
	s.W = sl.W
	j := 0
	for _, vc := range v.Cells {
		if vc.Collapsed || vc.View == nil || vc.Free {
			continue
		}
		s.H += vc.Margin.H
		if vc.Alignment&zgeo.VertCenter != 0 {
			s.H += vc.Margin.H
		}
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
	var ws float64
	if len(v.Cells) != 0 {
		if v.GridVerticalSpace != zgeo.UndefValue {
			s := v.calculateGridSize(total)
			return s, s
		}
		lays := v.getLayoutCells(total)
		spacing := v.Spacing()
		if v.GridVerticalSpace != zgeo.UndefValue {
			spacing = v.GridVerticalSpace
		}
		s, max = zgeo.LayoutGetCellsStackedSize(v.ObjectName(), v.Vertical, spacing, lays)
		// if v.ObjectName() == "zheader" {
		// 	zlog.Info("sv.CalculatedSize:", s, max)
		// }
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
		if !v.NoCalculatedMaxSize.W && max.W > 0 {
			zfloat.Minimize(&l.MaxSize.W, max.W)
		}
		if !v.NoCalculatedMaxSize.H && max.H > 0 {
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
			l.Divider = div.Delta
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
	// if v.ObjectName() == "License" {
	// 	zlog.Info("arrangeChildrenInGrid:", v.ObjectName(), rect)
	// }
	row, heights := v.getGridLayoutRow(rect.Size)
	j := 0
	r := rect
	for _, vc := range v.Cells {
		if vc.Collapsed || vc.View == nil || vc.Free {
			continue
		}
		if vc.Alignment&(zgeo.Top+zgeo.VertCenter) != 0 {
			r.Pos.Y += vc.Margin.H
		}
		r.Size.H = heights[j]
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
			if i >= len(rowCells) {
				continue
			}
			cell := (rowCells)[i]
			s, _ := cell.View.CalculatedSize(box.Size)
			ar := box.AlignPro(s, cell.Alignment, cell.Margin, cell.MaxSize, cell.MinSize)
			ar = ar.Intersected(box)
			cell.View.SetRect(ar)
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
				// zlog.Info("getGridLayoutRow:", j, i, l.OriginalSize.H)
			} else {
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
	if v.GridVerticalSpace != zgeo.UndefValue {
		zlog.Assert(v.Vertical, v.Hierarchy(), v.GridVerticalSpace)
		v.arrangeChildrenInGrid()
		for _, c := range v.Cells {
			if c.View != nil && !c.Free && !c.Collapsed {
				co, _ := c.View.(CellsOwner)
				ca, _ := c.View.(Arranger)
				if co != nil && ca != nil {
					// zlog.Info(Debug, "SV ArrangeChild1:", len(*co.GetCells()))
					for _, c := range *co.GetCells() {
						// zlog.Info(Debug, "SV ArrangeChild:", c.View.Native().ObjectName(), c.Free)
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
	// if v.ObjectName() == "eventgrid" {
	// 	zlog.Info("stack get layout", v.Hierarchy(), rm)
	// 	for i, l := range lays {
	// 		zlog.Info("stack layout:", v.ObjectName(), i, l.Name, l.OriginalSize, l.MaxSize, l.Alignment)
	// 	}
	// }
	rects := zgeo.LayoutCellsInStack(v.ObjectName(), rm, v.Vertical, v.spacing, lays)
	// if v.ObjectName() == "zheader" {
	// 	zlog.Info("StackView ArrangeChildren:", v.ObjectName(), rm, rects)
	// }
	// zlog.ProfileLog("did layout")
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
			// if v.ObjectName() == "eventgrid" {
			// 	zlog.Info("arrangeChildren:", rm, v.ObjectName(), c.View.ObjectName(), r)
			// }
			c.View.SetRect(r)
			if c.ShowIfExtraSpace != 0 && !c.View.Native().IsShown() {
				c.View.Show(true)
			}
		} else {
			if c.View != nil {
				c.View.Show(false)
			}
		}
		j++
	}
}

func MakeLinkedStack(surl, name string, add zview.View) *StackView {
	v := StackViewHor("#type:a") // #type is hack that sets html-type of stack element
	v.SetAboveParent(true)
	v.SetMargin(zgeo.RectFromXY2(3, 0, -3, 0))
	v.MakeLink(surl, name)
	v.Add(add, zgeo.TopLeft, zgeo.SizeNull)
	return v
}
