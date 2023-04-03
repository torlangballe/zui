//go:build zui

package zcontainer

import (
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on March 2023

type GridView struct {
	ContainerView
	spacing     zgeo.Size
	Columns     int
	addToColumn int
	table       map[zgeo.IPos]zview.View
	addPos      zgeo.IPos
}

func GridViewNew(name string, cols int) *GridView {
	v := &GridView{}
	v.Init(v, name, cols)
	return v
}

func (v *GridView) Init(view zview.View, name string, cols int) {
	// zlog.Info("GridView Init")
	v.ContainerView.Init(view, name)
	v.Columns = cols
	v.spacing = zgeo.Size{6, 4}
	v.addToColumn = -1
	v.table = map[zgeo.IPos]zview.View{}
}

func (v *GridView) SetAddColumn(c int) {
	// zlog.Info("SetAddColumn", c)
	v.addToColumn = c
	v.addPos.X = c
	for y := 0; ; y++ {
		if v.table[zgeo.IPos{c, y}] == nil {
			v.addPos.Y = y
			return
		}
	}
}

func (v *GridView) SetSpacing(spacing zgeo.Size) *GridView {
	v.spacing = spacing
	return v
}

func (v *GridView) Spacing() zgeo.Size {
	return v.spacing
}

func (v *GridView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	rows := v.getLayoutRows(zgeo.Rect{Size: total})
	if len(rows) == 0 {
		return v.Margin().Size.Negative()
	}
	for j := 0; j < len(rows); j++ {
		rowSize := zgeo.LayoutGetCellsStackedSize(v.ObjectName(), false, v.spacing.W, rows[j])
		zfloat.Maximize(&s.W, rowSize.W)
		s.H += rowSize.H
		if j != 0 {
			s.H += v.spacing.H
		}
	}
	s.Subtract(v.Margin().Size)
	// zlog.Info("GridView size:", s)
	return s
}

func (v *GridView) GetCell(x, y int) *Cell {
	view := v.table[zgeo.IPos{x, y}]
	if view == nil {
		return nil
	}
	_, i := v.FindCellWithView((view))
	if i == -1 {
		return nil
	}
	return &v.Cells[i]
}

func (v *GridView) RowCount() int {
	var count int
	for pos, _ := range v.table {
		zint.Maximize(&count, pos.Y+1)
	}
	return count
}

func (v *GridView) addCell(rect zgeo.Rect, rows *[][]zgeo.LayoutCell, i, j int) {
	if len(*rows) <= j {
		for k := len(*rows); k <= j; k++ {
			*rows = append(*rows, make([]zgeo.LayoutCell, v.Columns))
		}
	}
	row := (*rows)[j]
	cell := v.GetCell(i, j)
	// zlog.Info("addCell:", i, j, cell != nil)
	var l zgeo.LayoutCell
	if cell != nil {
		l = cell.LayoutCell
		l.OriginalSize = cell.View.CalculatedSize(rect.Size)
		// zlog.Info("Size:", i, j, cell.View.ObjectName(), l.OriginalSize)
		l.Name = cell.View.ObjectName()
	}
	l.Alignment = zgeo.CenterLeft // we force a simple layout and align in cells later
	// zlog.Info("OrgSize:", l.Name, l.OriginalSize)
	row[i] = l
}

func (v *GridView) getLayoutRows(rect zgeo.Rect) [][]zgeo.LayoutCell {
	var rows [][]zgeo.LayoutCell
	rowCount := v.RowCount()
	for j := 0; j < rowCount; j++ {
		for i := 0; i < v.Columns; i++ {
			v.addCell(rect, &rows, i, j)
		}
	}
	for j := 1; j < len(rows); j++ {
		for i := range rows[j] {
			zfloat.Maximize(&rows[0][i].OriginalSize.W, rows[j][i].OriginalSize.W)
			zfloat.Maximize(&rows[0][i].MaxSize.W, rows[j][i].MaxSize.W)
			if rows[j][i].MinSize.W != 0 {
				zfloat.Minimize(&rows[0][i].MinSize.W, rows[j][i].MinSize.W)
			}
		}
	}
	// for i := range rows[0] {
	// 	zlog.Info("NewCellHMax2:", i, rows[0][i].OriginalSize.H, rows[0][i].Name)
	// }
	for j := 1; j < len(rows); j++ {
		for i := range rows[j] {
			// zlog.Info("NewCellHMax:", j, i, rows[j][i].OriginalSize.H, rows[j][i].Name)
			rows[j][i].OriginalSize.W = rows[0][i].OriginalSize.W
			// zlog.Info("NewCellHMax2:", j, i, rows[j][i].OriginalSize.H, rows[j][i].Name)
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
			rows[j][i].OriginalSize.H = maxOH
			rows[j][i].MaxSize.H = maxMH
			rows[j][i].MinSize.H = minH
		}
	}
	return rows
}

func (v *GridView) ArrangeChildren() {
	// zlog.Info("*********** GridView.ArrangeChildren:", v.Hierarchy(), v.Rect(), len(v.Cells))
	// zlog.PushProfile(v.ObjectName())
	// zlog.Info("GridArrangeCh:", v.LocalRect().Size)
	rect := v.LocalRect()
	rows := v.getLayoutRows(rect)
	// for j := range rows {
	// 	for i, c := range rows[j] {
	// 		zlog.Info("RowCell:", j, i, c.OriginalSize, c.Name)
	// 	}
	// }
	for j, row := range rows {
		r := rect
		r.Size.H = zgeo.LayoutGetCellsStackedSize(v.ObjectName(), false, v.spacing.W, row).H
		rects := zgeo.LayoutCellsInStack(v.ObjectName(), r, false, v.spacing.W, row)
		y := rect.Min().Y
		// zlog.Info("Layout1:", r)
		for i := range row {
			r := rects[i]
			cell := v.GetCell(i, j)
			if r.IsNull() || cell == nil {
				continue
			}
			zfloat.Maximize(&y, r.Max().Y)
			ar := r.AlignPro(r.Size, cell.Alignment, cell.Margin, cell.MaxSize, cell.MinSize)
			// zlog.Info("Layout:", cell.Alignment, row[i].OriginalSize, j, i, row[i].Name, rects[i], ar)
			cell.View.SetRect(ar)
		}
		rect.SetMinY(r.Max().Y + v.spacing.H)
	}
}

func (v *GridView) AddCell(cell Cell, index int) (cvs *Cell) {
	zlog.Assert(index == -1)
	// name := "nil"
	// if cell.View != nil {
	// 	name = cell.View.ObjectName()
	// }
	// zlog.Info("GridView.AddCell", v.addPos, name)
	v.table[v.addPos] = cell.View
	if v.addToColumn != -1 {
		v.addPos.Y++
	} else {
		v.addPos.X++
		if v.addPos.X >= v.Columns {
			v.addPos.X = 0
			v.addPos.Y++
		}
	}
	return v.ContainerView.AddCell(cell, index)
}