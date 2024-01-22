//go:build zui

package zcontainer

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /20/10/15.

type StackView struct {
	ContainerView
	spacing           float64
	Vertical          bool
	GridVerticalSpace float64
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

func (v *StackView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	var ws float64
	if len(v.Cells) != 0 {
		if v.GridVerticalSpace != 0 {
			sv := v.Cells[0].View.(*StackView) // we assume it's a stack
			rows, _ := v.getGridLayoutRows(total)
			row0 := rows[0]
			wlays := sv.getLayoutCells(total)
			zlog.Assert(len(row0) == len(wlays) && len(row0) > 0, len(row0), len(wlays))
			for _, row := range rows {
				for i, col := range row {
					zfloat.Maximize(&wlays[i].MinSize.W, col.MinSize.W)
					zfloat.Maximize(&wlays[i].MaxSize.W, col.MaxSize.W)
					zfloat.Maximize(&wlays[i].OriginalSize.W, col.OriginalSize.W)
					// zlog.Info("CalcGrid:", j, wlays[i].Name, wlays[i].MinSize, wlays[i].MaxSize, wlays[i].OriginalSize)
				}
			}
			ws = zgeo.LayoutGetCellsStackedSize(v.ObjectName(), false, v.Spacing(), wlays).W
		}
		lays := v.getLayoutCells(total)
		spacing := v.Spacing()
		if v.GridVerticalSpace != 0 {
			spacing = v.GridVerticalSpace
		}
		s = zgeo.LayoutGetCellsStackedSize(v.ObjectName(), v.Vertical, spacing, lays)
		zfloat.Maximize(&s.W, ws)
	}
	s.Subtract(v.Margin().Size)
	s.MaximizeNonZero(v.MinSize())
	return s
}

func (v *StackView) getLayoutCells(total zgeo.Size) (lays []zgeo.LayoutCell) {
	// zlog.Info("Layout Stack getCells start", v.ObjectName())
	for _, c := range v.Cells {
		if c.Free {
			continue
		}
		l := c.LayoutCell
		l.OriginalSize = c.View.CalculatedSize(total)
		// zlog.Info("StackView getLayoutCells:", v.ObjectName(), c.View.ObjectName(), l.OriginalSize)
		l.Name = c.View.ObjectName()
		div, _ := c.View.(*DividerView)
		if div != nil {
			l.Divider = div.Delta
		}
		lays = append(lays, l)
	}
	return
}

func (v *StackView) arrangeChildrenInGrid() {
	rect := v.LocalRect()
	rows, rowStackCells := v.getGridLayoutRows(rect.Size)
	// zlog.Info("arrangeChildrenInGrid:", v.LocalRect().Size)
	y := rect.Min().Y
	for j, row := range rows {
		rowHeight := zgeo.LayoutGetCellsStackedSize(v.ObjectName(), false, v.spacing, row).H
		layoutRowRect := zgeo.RectFromXYWH(0, 0, rect.Size.W, rowHeight)
		rects := zgeo.LayoutCellsInStack(v.ObjectName(), layoutRowRect, false, v.spacing, row)
		// zlog.Info("Rects:", j, rects, layoutRowRect)
		rowCells := rowStackCells[j]
		rowView := v.Cells[j].View
		placeRowRect := zgeo.RectFromXYWH(0, y, rect.Size.W, rowHeight)
		//cv, _ := rowView.(*zcustom.CustomView)
		rowView.Native().SetRect(placeRowRect)
		var maxHeight float64
		for i := range row {
			box := rects[i]
			if box.IsNull() {
				continue
			}
			zfloat.Maximize(&maxHeight, box.Size.H)
			if i >= len(rowCells) {
				continue
			}
			cell := rowCells[i]
			ar := box.AlignPro(row[i].OriginalSize, cell.Alignment, cell.Margin, cell.MaxSize, cell.MinSize)
			ar = ar.Intersected(box)
			cell.View.SetRect(ar)
		}
		y += maxHeight + v.GridVerticalSpace
	}
}

func (v *StackView) getGridLayoutRows(total zgeo.Size) (rows [][]zgeo.LayoutCell, rowStackCells [][]Cell) {
	for j := 0; j < len(v.Cells); j++ {
		vertCell := v.Cells[j]
		if vertCell.Collapsed || vertCell.View == nil {
			continue
		}
		cellsOwner, _ := vertCell.View.(CellsOwner)
		zlog.Assert(cellsOwner != nil, v.Hierarchy())
		rowCells := cellsOwner.GetCells()
		rowStackCells = append(rowStackCells, *rowCells)
		var row []zgeo.LayoutCell
		for _, rc := range *rowCells {
			l := rc.LayoutCell
			l.Alignment = zgeo.CenterLeft
			l.OriginalSize = rc.View.CalculatedSize(total)
			l.Alignment = zgeo.CenterLeft
			l.Name = rc.View.ObjectName()
			if rc.Alignment&zgeo.VertExpand != 0 {
				l.Alignment |= zgeo.VertExpand
			}
			row = append(row, l)
		}
		rows = append(rows, row)
	}
	for j := 1; j < len(rows); j++ {
		for i := range rows[j] {
			zfloat.Maximize(&rows[0][i].OriginalSize.W, rows[j][i].OriginalSize.W)
			zfloat.Maximize(&rows[0][i].MaxSize.W, rows[j][i].MaxSize.W)
			if rows[j][i].MinSize.W != 0 {
				zfloat.Maximize(&rows[0][i].MinSize.W, rows[j][i].MinSize.W)
			}
		}
	}
	// for i := range rows[0] {
	// 	zlog.Info("NewCellHMax2:", i, rows[0][i].OriginalSize.W, rows[0][i].Name)
	// }
	for j := 1; j < len(rows); j++ {
		for i := range rows[j] {
			// zlog.Info("NewCellHMax:", j, i, rows[j][i].OriginalSize.H, rows[j][i].Name)
			rows[j][i].OriginalSize.W = rows[0][i].OriginalSize.W
			rows[j][i].MinSize.W = rows[0][i].MinSize.W
			rows[j][i].MaxSize.W = rows[0][i].MaxSize.W
		}
	}
	for j := range rows {
		var maxOH, maxMH, minH float64
		for i := range rows[j] {
			// zlog.Info("NewCellHMax2:", j, i, rows[j][i].OriginalSize.H, rows[j][i].Name)
			zfloat.Maximize(&maxOH, rows[j][i].OriginalSize.H)
			zfloat.Maximize(&maxMH, rows[j][i].MaxSize.H)
			zfloat.Maximize(&minH, rows[j][i].MinSize.H)
		}
		for i := range rows[j] {
			// TODO: Won't expand vertically, need to fix! OriginalSize should not change, as we need to place it with it's actualy size, but a new variable creating a box to place it in should be max'ed below:
			//rows[j][i].OriginalSize.H = maxOH
			rows[j][i].MaxSize.H = maxMH
			rows[j][i].MinSize.H = minH
		}
	}
	return rows, rowStackCells
}

func (v *StackView) ArrangeChildren() {
	// zlog.Info("*********** Stack.ArrangeChildren:", v.Hierarchy(), v.Rect(), len(v.Cells), v.GridVerticalSpace)
	// zlog.PushProfile(v.ObjectName())
	if v.GridVerticalSpace != 0 {
		zlog.Assert(v.Vertical, v.Hierarchy(), v.GridVerticalSpace)
		v.arrangeChildrenInGrid()
		return
	}
	layouter, _ := v.View.(zview.Layouter)
	if layouter != nil {
		layouter.HandleBeforeLayout()
	}
	rm := v.LocalRect().Plus(v.Margin())
	lays := v.getLayoutCells(rm.Size)
	rects := zgeo.LayoutCellsInStack(v.ObjectName(), rm, v.Vertical, v.spacing, lays)
	// zlog.ProfileLog("did layout")
	j := 0
	for _, c := range v.Cells {
		if c.Free {
			v.ArrangeChild(c, rm)
			continue
		}
		r := rects[j]
		// 	zlog.Info("Stack.ArrangeChild:", v.Hierarchy(), c.View.ObjectName(), r)
		if !r.IsNull() {
			c.View.SetRect(r)
		}
		j++
	}
}

func MakeLinkedStack(surl, name string, add zview.View) *StackView {
	v := StackViewHor("#type:a") // #type is hack that sets html-type of stack element
	v.SetAboveParent(true)
	v.SetMargin(zgeo.RectFromXY2(3, 0, -3, 0))
	v.MakeLink(surl, name)
	v.Add(add, zgeo.TopLeft, zgeo.Size{})
	return v
}
