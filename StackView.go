package zgo

import (
	"math"

	"github.com/torlangballe/zutil/zfloat"

	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmath"
)

//  Created by Tor Langballe on /20/10/15.

type StackView struct {
	ContainerView
	spacing        float64
	Vertical       bool
	weightMinSizes []float64
}

func StackViewNew(vertical bool, alignment Alignment, name string, elements ...interface{}) *StackView {
	s := &StackView{}
	s.ContainerView.init(s, name)
	s.AddElements(alignment, elements...)
	s.Vertical = vertical
	s.spacing = 6
	return s
}

func (v *StackView) Spacing(spacing float64) *StackView {
	v.spacing = spacing
	return v
}

func (v *StackView) GetSpacing() float64 {
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
			zfloat.AddIntFloat64ToSet(c.Weight, &weights)
			sizes[i] = v.getCellSize(c, nil).Vertice(v.Vertical)
		}
	}
	for _, w := range weights {
		max := 0.0
		for i, c := range v.cells {
			if !c.Collapsed && !c.Free && c.Weight == w {
				zmath.Maximize(&max, sizes[i])
			}
		}
		for i, c := range v.cells {
			if !c.Collapsed && !c.Free && c.Weight == w {
				v.weightMinSizes[i] = max
			}
		}
	}
}

func (v *StackView) GetCalculatedSize(total Size) Size {
	v.calcWeightMins()
	// for i, c := range v.cells {
	// 	if !c.Collapsed && !c.Free && c.Weight > 0 {
	// 		fmt.Println("weight size:", c.View.GetObjectName(), v.weightMinSizes[i])
	// 	}
	// }
	var size = Size{}
	for i, c := range v.cells {
		if !c.Collapsed && !c.Free {
			fs := v.getCellSize(c, &i)
			m := calcMarginAdd(c)
			//			fmt.Println("calcsize:", c.View.GetObjectName(), fs, m)
			*size.VerticeP(v.Vertical) += fs.Vertice(v.Vertical)
			//			zmath.Maximize(size.VerticeP(!v.Vertical), fs.Vertice(!v.Vertical)-m.Vertice(!v.Vertical))
			zmath.Maximize(size.VerticeP(!v.Vertical), fs.Vertice(!v.Vertical)+m.Vertice(!v.Vertical))
			*size.VerticeP(v.Vertical) += v.spacing
		}
	}
	size.Subtract(v.margin.Size)
	if len(v.cells) > 0 {
		*size.VerticeP(v.Vertical) -= v.spacing
	}
	zmath.Maximize(size.VerticeP(!v.Vertical), v.GetMinSize().Vertice(!v.Vertical))
	size.Maximize(v.GetMinSize())
	return size
}

func (v *StackView) handleAlign(size Size, inRect Rect, a Alignment, cell ContainerViewCell) (Rect, Rect) {
	max := cell.MaxSize
	// if max.W != 0 {
	// 	zmath.Maximize(&max.W, size.W)
	// }
	// if max.H != 0 {
	// 	zmath.Maximize(&max.H, size.H)
	// }
	var box = inRect.Align(size, a, Size{}, max)
	var vr Rect
	if cell.Alignment.Only(v.Vertical)&AlignmentShrink != 0 {
		s := cell.View.GetCalculatedSize(inRect.Size)
		vr = box.Align(s, cell.Alignment, cell.Margin, max)
	} else {
		vr = box.Expanded(cell.Margin.Negative())
	}
	// fmt.Println("handleAlign:", box, cell.View.GetObjectName(), inRect, size, a, max, vr, cell.Margin)
	return box, vr
}

// Function for setting better focus on Android, due to bug with ScrollViews.
// Does nothing elsewhere
func (v *StackView) ForceHorizontalFocusNavigation() {
}

// func (v *StackView) Rect(rect Rect) View {
// 	v.CustomView.Rect(rect)
// 	v.ArrangeChildren(nil)
// 	return v
// }

func addDiff(size *Size, maxSize float64, vertical bool, diff *float64, count *int) {
	d := math.Floor(*diff / float64(*count))
	if maxSize != 0 {
		zmath.Minimize(&d, math.Max(maxSize-size.W, 0))
		*diff -= d
		(*count)--
	}
	// fmt.Println("addDiff:", size.W, d, maxSize)
	*(*size).VerticeP(vertical) += d
}

func calcMarginAdd(c ContainerViewCell) Size {
	var m = c.Margin
	if c.Alignment&AlignmentMarginIsOffset != 0 {
		m = Size{0, 0}
	} else {
		if c.Alignment&AlignmentHorCenter != 0 {
			m.W *= 2
		}
		if c.Alignment&AlignmentVertCenter != 0 {
			m.H *= 2
		}
	}
	return m
}

func (v *StackView) getCellSize(c ContainerViewCell, weightIndex *int) Size {
	//	tot := v.getCellFitSizeInTotal(total, c)
	var size = c.View.GetCalculatedSize(Size{})
	// fmt.Println("get cell size:", c.View.GetObjectName(), size.W)
	m := calcMarginAdd(c)
	*size.VerticeP(!v.Vertical) += m.Vertice(!v.Vertical)
	if c.MinSize.W != 0 {
		zmath.Maximize(&size.W, c.MinSize.W)
	}
	if c.MinSize.H != 0 {
		zmath.Maximize(&size.H, c.MinSize.H)
	}
	if weightIndex != nil {
		len := v.weightMinSizes[*weightIndex]
		if len != 0 {
			*size.VerticeP(v.Vertical) = len
		}
	}
	return size
}

func (v *StackView) ArrangeChildren(onlyChild *View) {
	var incs = 0
	var decs = 0
	var sizes = map[View]Size{}
	var ashrink = AlignmentHorShrink
	var aexpand = AlignmentHorExpand
	var aless = AlignmentLeft
	var amore = AlignmentRight
	var amid = AlignmentHorCenter | AlignmentMarginIsOffset

	// fmt.Println("Stack ArrangeChildren:", v.GetObjectName())
	var r = v.GetRect()
	r.Pos = Pos{} // translate to 0,0 cause children are in parent

	if v.layoutHandler != nil {
		v.layoutHandler.HandleBeforeLayout()
	}
	if v.Vertical {
		ashrink = AlignmentVertShrink
		aexpand = AlignmentVertExpand
		aless = AlignmentTop
		amore = AlignmentBottom
		amid = AlignmentVertCenter
	}
	for _, c2 := range v.cells {
		if c2.Alignment&AlignmentHorizontal == 0 || c2.Alignment&AlignmentVertical == 0 {
			zlog.Error(nil, "\n\nStack Align: No vertical or horizontal component:", c2.View.GetObjectName(), c2.Alignment, "\n\n")
			return
		}
		if !c2.Free {
			if c2.Collapsed {
				v.RemoveChild(c2.View)
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
	var cs = v.GetCalculatedSize(Size{}).Vertice(v.Vertical)
	cs += v.margin.Size.Vertice(v.Vertical)

	diff := r.Size.Vertice(v.Vertical) - cs
	var lastNoFreeIndex = -1
	for _, useMaxSize := range []bool{true, false} {
		for i, c3 := range v.cells {
			if !c3.Collapsed && !c3.Free && (c3.MaxSize.W != 0) == useMaxSize {
				lastNoFreeIndex = i
				size := v.getCellSize(c3, &i)
				if decs > 0 && c3.Alignment&ashrink != 0 && diff != 0.0 {
					//addDiff(&size, c3.MaxSize.W, v.Vertical, &diff, &decs)
				} else if incs > 0 && c3.Alignment&aexpand != 0 && diff != 0.0 {
					addDiff(&size, c3.MaxSize.W, v.Vertical, &diff, &incs)
				}
				//				fmt.Println("cellsize:", c3.MaxSize.W, c3.View.GetObjectName(), size, c3.Alignment)
				sizes[c3.View] = size
			}
		}
	}
	var centerDim = 0.0
	var firstCenter = true

	// Not-centered children:
	for i, c4 := range v.cells {
		// fmt.Println("cell:", c4.View.GetObjectName(), c4.Alignment, c4.Collapsed, c4.Free)
		if !c4.Collapsed && !c4.Free {
			if (c4.Alignment & (amore | aless)) != 0 {
				var a = c4.Alignment
				if i != lastNoFreeIndex {
					a = a.Subtracted(AlignmentExpand.Only(v.Vertical))
				}
				box, vr := v.handleAlign(sizes[c4.View], r, a, c4)
				if onlyChild == nil || *onlyChild == c4.View {
					c4.View.Rect(vr)
					// fmt.Println("cellsides:", c4.View.GetObjectName(), c4.Alignment, vr, "s:", sizes[c4.View], r, "get:", c4.View.GetRect())
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
				cv, got := c4.View.(*ContainerView)
				if got {
					cv.ArrangeChildren(nil)
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
			a := c5.Alignment.Subtracted(amid) | aless
			box, vr := v.handleAlign(sizes[c5.View], r, a, c5)
			if onlyChild == nil || *onlyChild == c5.View {
				// fmt.Println("cellmid:", c5.View.GetObjectName(), vr)
				c5.View.Rect(vr)
			}
			*r.Pos.VerticeP(v.Vertical) = box.Max().Vertice(v.Vertical) + v.spacing
			cv, got := c5.View.(*ContainerView)
			if got {
				cv.ArrangeChildren(nil)
			} else {
				//!          (c5.view as? ZCustomView)?.HandleAfterLayout()
			}
		}
	}

	//        HandleAfterLayout()
}

func StackRows(alignment Alignment, name string, elements ...interface{}) *StackView {
	return StackViewNew(true, alignment, name, elements...)
}

func StackColumns(alignment Alignment, name string, elements ...interface{}) *StackView {
	return StackViewNew(false, alignment, name, elements...)
}
