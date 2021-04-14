// +build zui

package zui

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /20/10/15.

var NewStack bool

type StackView struct {
	NewStack bool
	ContainerView
	spacing  float64
	Vertical bool
}

func StackViewNew(vertical bool, name string) *StackView {
	s := &StackView{}
	s.Init(s, vertical, name)
	return s
}

func (v *StackView) Init(view View, vertical bool, name string) {
	v.NewStack = NewStack
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

// TODO: Cache sizes more, CalculatedSize will recalculate sub-childens size many times
func (v *StackView) CalculatedSize(total zgeo.Size) zgeo.Size {
	var size = zgeo.Size{}
	for i, c := range v.cells {
		if !c.Collapsed && !c.Free {
			fs := v.getCellSize(c, &i)
			m := calcMarginAdd(c)
			// zlog.Info("calcsize:", c.View.ObjectName(), fs, m)
			*size.VerticeP(v.Vertical) += fs.Vertice(v.Vertical)
			//			zmath.Maximize(size.VerticeP(!v.Vertical), fs.Vertice(!v.Vertical)-m.Vertice(!v.Vertical))
			zfloat.Maximize(size.VerticeP(!v.Vertical), fs.Vertice(!v.Vertical)+m.Vertice(!v.Vertical))
			*size.VerticeP(v.Vertical) += v.spacing
		}
	}
	size.Subtract(v.margin.Size)
	if len(v.cells) > 0 {
		*size.VerticeP(v.Vertical) -= v.spacing
	}
	zfloat.Maximize(size.VerticeP(!v.Vertical), v.MinSize().Vertice(!v.Vertical))
	size.Maximize(v.MinSize())
	// zlog.Info("calcsized stack size:", v.ObjectName(), size)
	return size
}

func (v *StackView) handleAlign(size zgeo.Size, inRect zgeo.Rect, a zgeo.Alignment, cell ContainerViewCell) (zgeo.Rect, zgeo.Rect) {
	max := cell.MaxSize
	var box = inRect.AlignPro(size, a, zgeo.Size{}, max, zgeo.Size{})
	// zlog.Info("handleAlign:", box, cell.View.ObjectName(), inRect, size, a, max, cell.Margin)
	// var box = inRect.Align(size, a, cell.Margin, max)

	var vr zgeo.Rect
	// if v.ObjectName() == "header" {
	// }
	if cell.Alignment.Only(v.Vertical)&zgeo.Shrink != 0 {
		s := cell.View.CalculatedSize(inRect.Size)
		//!!		zlog.Fatal(nil, "strange align")
		// zlog.Info("handleAlign", s, cell.Alignment, cell.Margin, max)
		vr = box.AlignPro(s, cell.Alignment, cell.Margin, max, zgeo.Size{})
	} else {
		vr = box
		if true { //cell.Alignment.Only(v.Vertical)&zgeo.Center != 0 {
			vr = box.AlignmentTransform(cell.Margin, cell.Alignment)
			if v.ObjectName() == "bar" {
				// zlog.Info("AFTER expand:", cell.View.ObjectName(), vr, box)
			}
			box.UnionWith(vr)
		} else {
			zlog.Info(nil, "NOT shrink expanding:", cell.View.ObjectName())
		}
		// zlog.Info("handleAlign expin:", box, cell.View.ObjectName(), vr, cell.Margin, cell.Alignment)
	}
	// zlog.Info("handleAlign:", box, cell.View.ObjectName(), inRect, size, a, max, vr, cell.Margin)
	return box, vr
}

// Function for setting better focus on Android, due to bug with ScrollViews.
// Does nothing elsewhere
func (v *StackView) ForceHorizontalFocusNavigation() {
}

func addDiff(size *zgeo.Size, maxSize float64, vertical bool, diff *float64, count *int) {
	d := math.Floor(*diff / float64(*count))
	if maxSize != 0 {
		zfloat.Minimize(&d, math.Max(maxSize-size.W, 0))
		*diff -= d
		(*count)--
	}
	// old := *size
	*(*size).VerticeP(vertical) += d
	// zlog.Info("addDiffIn:", old, "=>", size.W, d, maxSize)
}

func calcMarginAdd(c ContainerViewCell) zgeo.Size {
	var m = c.Margin
	if c.Alignment&zgeo.MarginIsOffset != 0 {
		m = zgeo.Size{0, 0}
	} else {
		if c.Alignment&zgeo.HorCenter != 0 {
			m.W *= 2
		}
		if c.Alignment&zgeo.VertCenter != 0 {
			m.H *= 2
		}
	}
	return m
}

func (v *StackView) getCellSize(c ContainerViewCell, weightIndex *int) zgeo.Size {
	//	tot := v.getCellFitSizeInTotal(total, c)
	var size = c.View.CalculatedSize(zgeo.Size{})
	// zlog.Info("get cell size1:", v.ObjectName(), c.View.ObjectName(), size)
	vert := v.Vertical
	size.Maximize(c.MinSize)
	if c.MaxSize.W != 0 {
		zfloat.Minimize(&size.W, c.MaxSize.W)
	}
	if c.MaxSize.H != 0 {
		zfloat.Minimize(&size.H, c.MaxSize.H)
	}
	if c.ExpandFromMinSize && c.MinSize.Vertice(vert) != 0 {
		*size.VerticeP(vert) = c.MinSize.Vertice(vert)
	}
	m := calcMarginAdd(c)
	*size.VerticeP(vert) += m.Vertice(vert)
	// zlog.Info("get cell size2:", c.MaxSize, c.MinSize, v.ObjectName(), c.View.ObjectName(), size)
	return size
}

func (v *StackView) ArrangeChildren(onlyChild *View) {
	var incs = 0
	var decs = 0
	var sizes = map[View]zgeo.Size{}
	var ashrink = zgeo.HorShrink
	var aexpand = zgeo.HorExpand
	var aless = zgeo.Left
	var amore = zgeo.Right
	var amid = zgeo.HorCenter | zgeo.MarginIsOffset

	if v.layoutHandler != nil {
		v.layoutHandler.HandleBeforeLayout()
	}
	// v.NewStack = true
	// zlog.Info("*********** Stack.ArrangeChildren:", v.ObjectName(), v.Rect(), len(v.cells))
	if v.NewStack {
		zlog.Assert(onlyChild == nil) // going away...
		rm := v.LocalRect().Plus(v.Margin())
		var lays []zgeo.LayoutCell
		for _, c := range v.cells {
			l := c.LayoutCell
			l.OriginalSize = c.View.CalculatedSize(rm.Size)
			// zlog.Info("STv OSize:", c.View.ObjectName(), l.OriginalSize, rm.Size)
			l.Name = c.View.ObjectName()
			lays = append(lays, l)
		}
		rects := zgeo.LayoutCellsInStack(rm, v.Vertical, v.spacing, lays)
		// for i, r := range rects {
		// 	zlog.Info("R:", i, v.cells[i].View.ObjectName(), r)
		// }
		for i, c := range v.cells {
			r := rects[i]
			// zlog.Info(i, "LAYOUT SETRECT:", r, c.Alignment, c.View.ObjectName())
			if !r.IsNull() {
				c.View.SetRect(r)
			}
		}
		return
	}
	var r = v.Rect()
	r.Pos = zgeo.Pos{} // translate to 0,0 cause children are in parent

	if v.Vertical {
		ashrink = zgeo.VertShrink
		aexpand = zgeo.VertExpand
		aless = zgeo.Top
		amore = zgeo.Bottom
		amid = zgeo.VertCenter
	}
	for _, c2 := range v.cells {
		if c2.Alignment&zgeo.Horizontal == 0 || c2.Alignment&zgeo.Vertical == 0 {
			zlog.Error(nil, "\n\nStack Align: No vertical or horizontal component:", c2.View.ObjectName(), c2.Alignment, "\n\n")
			return
		}
		if !c2.Free {
			if c2.Collapsed {
				// causes crash:
				// if c2.View != nil && ViewGetNative(c2.View).Parent() != nil {
				// 	v.CustomView.RemoveChild(c2.View)
				// }
				//				v.RemoveChild(c2.View) // why would we do this? Should be done already
			} else {
				if c2.Alignment&ashrink != 0 {
					decs++
				}
				if c2.Alignment&aexpand != 0 {
					incs++
				}
			}
		}
		cv, got := c2.View.(*ContainerView)
		if got && v.layoutHandler != nil {
			cv.layoutHandler.HandleBeforeLayout()
		}
	}
	r.Add(v.margin)
	for _, c1 := range v.cells {
		if c1.Free {
			v.arrangeChild(c1, r)
		}
	}
	cn := r.Center().Vertice(v.Vertical)
	var cs = v.CalculatedSize(zgeo.Size{}).Vertice(v.Vertical)
	cs += v.margin.Size.Vertice(v.Vertical)

	//	zlog.Info("Arrange:", v.ObjectName(), cs, r)
	diff := r.Size.Vertice(v.Vertical) - cs
	var lastNoFreeIndex = -1
	for _, useMaxSize := range []bool{true, false} {
		for i, c3 := range v.cells {
			if !c3.Collapsed && !c3.Free && (c3.MaxSize.W != 0) == useMaxSize {
				//				zlog.Info("DIFF3:", incs, v.ObjectName(), diff, r.Size, cs)
				lastNoFreeIndex = i
				size := v.getCellSize(c3, &i)
				if decs > 0 && c3.Alignment&ashrink != 0 && diff != 0.0 {
					//addDiff(&size, c3.MaxSize.W, v.Vertical, &diff, &decs)
				} else if incs > 0 && c3.Alignment&aexpand != 0 && diff != 0.0 && c3.Alignment&zgeo.Proportional == 0 {
					// if v.ObjectName() == "header" {
					//					zlog.Info("addDiff:", c3.View.ObjectName(), size.W, diff, r.Size.W, c3.Alignment)
					// }
					addDiff(&size, c3.MaxSize.W, v.Vertical, &diff, &incs)
				}
				// zlog.Info("cellsize:", v.ObjectName(), c3.MinSize.W, c3.MaxSize.W, c3.View.ObjectName(), size, c3.Alignment)
				sizes[c3.View] = size
			}
		}
	}
	var centerDim = 0.0
	var firstCenter = true

	// Not-centered children:
	for i, c4 := range v.cells {
		if !c4.Collapsed && !c4.Free {
			if (c4.Alignment & (amore | aless)) != 0 {
				var a = c4.Alignment
				if i != lastNoFreeIndex {
					a = a.Subtracted(zgeo.Expand.Only(v.Vertical))
				}
				box, vr := v.handleAlign(sizes[c4.View], r, a, c4)
				if onlyChild == nil || *onlyChild == c4.View {
					c4.View.SetRect(vr)
					// zlog.Info("cellsides:", v.ObjectName(), c4.View.ObjectName(), c4.Alignment, vr, "s:", sizes[c4.View], r, "get:", c4.View.Rect())
				}
				if c4.Alignment&aless != 0 {
					m := math.Max(r.Min().Vertice(v.Vertical), box.Max().Vertice(v.Vertical)+v.spacing)
					if v.Vertical {
						r.SetMinY(m)
					} else {
						r.SetMinX(m)
					}
				}
				if c4.Alignment&amore != 0 {
					m := math.Min(r.Max().Vertice(v.Vertical), box.Pos.Vertice(v.Vertical)-v.spacing)
					if v.Vertical {
						r.SetMaxY(m)
					} else {
						r.SetMaxX(m)
					}
				}
				// zlog.Info("cellsides:", v.ObjectName(), c4.View.ObjectName(), r)
				// ct, _ := c4.View.(ContainerType)
				// if ct != nil {
				// 	ct.ArrangeChildren(nil)
				// } else {
				// 	//! (c4.view as? ZCustomView)?.HandleAfterLayout()
				// }
			} else {
				centerDim += sizes[c4.View].Vertice(v.Vertical)
				if !firstCenter {
					centerDim += v.spacing
				}
				firstCenter = false
			}
		}
	}
	//	zlog.Info("Arrange2:", v.ObjectName(), r, v.Vertical, "cn:", cn, "cedim:", centerDim)

	if v.Vertical {
		r.SetMinY(math.Max(r.Min().Y, cn-centerDim/2))
	} else {
		r.SetMinX(math.Max(r.Min().X, cn-centerDim/2))
	}
	if v.Vertical {
		r.SetMaxY(math.Min(r.Max().Y, cn+centerDim/2))
	} else {
		r.SetMaxX(math.Min(r.Max().X, cn+centerDim/2))
	}
	//	Centered children:
	for _, c5 := range v.cells {
		if !c5.Collapsed && c5.Alignment&amid != 0 && !c5.Free { // .reversed()
			a := c5.Alignment.Subtracted(amid|zgeo.Expand) | aless
			// a := c5.Alignment | aless // need this for other case, how to make work?
			box, vr := v.handleAlign(sizes[c5.View], r, a, c5)
			if onlyChild == nil || *onlyChild == c5.View {
				c5.View.SetRect(vr)
			}
			// zlog.Info("cellmid:", v.ObjectName(), c5.View.ObjectName(), c5.View.CalculatedSize(zgeo.Size{}), c5.Alignment, vr, "s:", sizes[c5.View], r, "get:", c5.View.Rect(), c5.Margin, c5.MaxSize)
			*r.Pos.VerticeP(v.Vertical) = box.Max().Vertice(v.Vertical) + v.spacing
			ct, _ := c5.View.(ContainerType)
			if ct != nil {
				ct.ArrangeChildren(nil)
			} else {
				//!          (c5.view as? ZCustomView)?.HandleAfterLayout()
			}
		}
	}

	//        HandleAfterLayout()
}
