package zui

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"

	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /20/10/15.

var debugHints = true

type StackView struct {
	ContainerView
	spacing        float64
	Vertical       bool
	weightMinSizes []float64
}

func StackViewNew(vertical bool, name string) *StackView {
	s := &StackView{}
	s.ContainerView.Init(s, name)
	s.Vertical = vertical
	s.spacing = 6
	return s
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

func (v *StackView) calcWeightMins() {
	v.weightMinSizes = make([]float64, len(v.cells), len(v.cells))
	var weights []float64
	sizes := make([]float64, len(v.cells), len(v.cells))
	minWeight := -1.0
	for i, c := range v.cells {
		if !c.Collapsed && !c.Free && c.Weight > 0 {
			if minWeight != -1 && c.Weight < minWeight {
				minWeight = c.Weight
			}
			zfloat.AddToSet(c.Weight, &weights)
			sizes[i] = v.getCellSize(c, nil).Vertice(v.Vertical)
		}
	}
	for _, w := range weights {
		max := 0.0
		for i, c := range v.cells {
			if !c.Collapsed && !c.Free && c.Weight == w {
				zfloat.Maximize(&max, sizes[i])
			}
		}
		for i, c := range v.cells {
			if !c.Collapsed && !c.Free && c.Weight == w {
				v.weightMinSizes[i] = max
			}
		}
	}
}

func (v *StackView) CalculatedSize(total zgeo.Size) zgeo.Size {
	//	v.calcWeightMins()
	// for i, c := range v.cells {
	// 	if !c.Collapsed && !c.Free && c.Weight > 0 {
	// 		zlog.Info("weight size:", c.View.ObjectName(), v.weightMinSizes[i])
	// 	}
	// }
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
	return size
}

func (v *StackView) handleAlign(size zgeo.Size, inRect zgeo.Rect, a zgeo.Alignment, cell ContainerViewCell) (zgeo.Rect, zgeo.Rect) {
	max := cell.MaxSize
	var box = inRect.Align(size, a, zgeo.Size{}, max)
	// var box = inRect.Align(size, a, cell.Margin, max)

	var vr zgeo.Rect
	// if v.ObjectName() == "header" {
	// 	zlog.Info("handleAlign:", box, cell.View.ObjectName(), inRect, size, a, max, vr, cell.Margin)
	// }
	if cell.Alignment.Only(v.Vertical)&zgeo.Shrink != 0 {
		s := cell.View.CalculatedSize(inRect.Size)
		zlog.Fatal(nil, "strange align")
		// zlog.Info("handleAlign", s, cell.Alignment, cell.Margin, max)
		vr = box.Align(s, cell.Alignment, cell.Margin, max)
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

// func (v *StackView) SetRect(rect Rect) View {
// 	v.CustomView.SetRect(rect)
// 	v.ArrangeChildren(nil)
// 	return v
// }

func addDiff(size *zgeo.Size, maxSize float64, vertical bool, diff *float64, count *int) {
	d := math.Floor(*diff / float64(*count))
	if maxSize != 0 {
		zfloat.Minimize(&d, math.Max(maxSize-size.W, 0))
		*diff -= d
		(*count)--
	}
	// zlog.Info("addDiff:", size.W, d, maxSize)
	*(*size).VerticeP(vertical) += d
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
	size.Maximize(c.MinSize)
	if c.MaxSize.W != 0 {
		zfloat.Minimize(&size.W, c.MaxSize.W)
	}
	if c.MaxSize.H != 0 {
		zfloat.Minimize(&size.H, c.MaxSize.H)
	}
	m := calcMarginAdd(c)
	*size.VerticeP(v.Vertical) += m.Vertice(v.Vertical)
	// if weightIndex != nil {
	// 	len := v.weightMinSizes[*weightIndex]
	// 	if len != 0 {
	// 		*size.VerticeP(v.Vertical) = len
	// 	}
	// }
	// zlog.Info("get cell size2:", v.ObjectName(), ":", m, c.View.ObjectName(), size, c.Margin, c.MinSize, c.MaxSize)
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

	// zlog.Info("*********** Stack.ArrangeChildren:", v.ObjectName())
	var r = v.Rect()
	r.Pos = zgeo.Pos{} // translate to 0,0 cause children are in parent

	if v.layoutHandler != nil {
		v.layoutHandler.HandleBeforeLayout()
	}
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

	diff := r.Size.Vertice(v.Vertical) - cs
	//	zlog.Info("DIFF:", v.ObjectName(), diff, r.Size, cs)
	var lastNoFreeIndex = -1
	for _, useMaxSize := range []bool{true, false} {
		for i, c3 := range v.cells {
			if !c3.Collapsed && !c3.Free && (c3.MaxSize.W != 0) == useMaxSize {
				lastNoFreeIndex = i
				size := v.getCellSize(c3, &i)
				if decs > 0 && c3.Alignment&ashrink != 0 && diff != 0.0 {
					//addDiff(&size, c3.MaxSize.W, v.Vertical, &diff, &decs)
				} else if incs > 0 && c3.Alignment&aexpand != 0 && diff != 0.0 {
					// zlog.Info("addDiff:", c3.View.ObjectName(), size.W, diff, r.Size.W)
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
				// fmt.Printf("cellsides: %d %s %s %v %p\n", i, v.ObjectName(), c4.View.ObjectName(), sizes[c4.View], c4.View)
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
				ct, _ := c4.View.(ContainerType)
				if ct != nil {
					ct.ArrangeChildren(nil)
				} else {
					//! (c4.view as? ZCustomView)?.HandleAfterLayout()
				}
			} else {
				centerDim += sizes[c4.View].Vertice(v.Vertical)
				if !firstCenter {
					centerDim += v.spacing
				}
				firstCenter = false
			}
		}
	}
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
	// Centered children:
	for _, c5 := range v.cells {
		if !c5.Collapsed && c5.Alignment&amid != 0 && !c5.Free { // .reversed()
			a := c5.Alignment.Subtracted(amid|zgeo.Expand) | aless
			box, vr := v.handleAlign(sizes[c5.View], r, a, c5)
			if onlyChild == nil || *onlyChild == c5.View {
				c5.View.SetRect(vr)
			}
			// zlog.Info("cellmid:", v.ObjectName(), c5.View.ObjectName(), c5.Alignment, vr, "s:", sizes[c5.View], r, "get:", c5.View.Rect(), c5.Margin, c5.MaxSize)
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
