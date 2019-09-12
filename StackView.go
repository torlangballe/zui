package zgo

import (
	"fmt"
	"math"
	"time"
)

//  Created by Tor Langballe on /20/10/15.

type StackView struct {
	ContainerView
	spacing  float64
	Vertical bool
}

func StackViewNew(vertical bool, alignment Alignment, elements ...interface{}) *StackView {
	s := &StackView{}
	s.ContainerView.init(s, "stack")
	s.AddElements(alignment, elements...)
	s.Vertical = vertical
	s.spacing = 6
	return s
}

func (s *StackView) Spacing(spacing float64) *StackView {
	s.spacing = spacing
	return s
}

func (s *StackView) GetSpacing() float64 {
	return s.spacing
}

func (s *StackView) getCellFitSizeInTotal(total Size, cell ContainerViewCell) Size {
	var tot = total.Minus(cell.Margin)
	if cell.Alignment&AlignmentHorCenter != 0 {
		tot.W -= cell.Margin.W
	}
	if cell.Alignment&AlignmentVertCenter != 0 {
		tot.H -= cell.Margin.H
	}
	return tot
}

func (s *StackView) GetCalculatedSize(total Size) Size {
	var size = s.nettoCalculateSize(total)
	size.Maximize(s.GetMinSize())
	return size
}

func (s *StackView) nettoCalculateSize(total Size) Size { // can force size calc without needed result
	var size = Size{}
	for _, c1 := range s.cells {
		if !c1.Collapsed && !c1.Free {
			tot := s.getCellFitSizeInTotal(total, c1)
			fs := c1.View.GetCalculatedSize(tot)
			//                var fs = zConvertViewSizeThatFitstToSize (c1.view!, sizeIn tot)
			var m = c1.Margin
			if c1.Alignment&AlignmentMarginIsOffset != 0 {
				m = Size{0, 0}
			}
			*size.VerticeP(s.Vertical) += fs.Vertice(s.Vertical) + m.Vertice(s.Vertical)
			*size.VerticeP(!s.Vertical) = math.Max(size.Vertice(!s.Vertical), fs.Vertice(!s.Vertical)-m.Vertice(!s.Vertical))
			*size.VerticeP(s.Vertical) += s.spacing
		}
	}
	size.Subtract(s.margin.Size)
	if len(s.cells) > 0 {
		*size.VerticeP(s.Vertical) -= s.spacing
	}
	*size.VerticeP(!s.Vertical) = math.Max(size.Vertice(!s.Vertical), s.GetMinSize().Vertice(!s.Vertical))
	return size
}

func (s *StackView) handleAlign(size Size, inRect Rect, a Alignment, cell ContainerViewCell) Rect {
	var vr = inRect.Align(size, a, cell.Margin, cell.MaxSize)
	// if cell.handleTransition != nil {
	//     let r := cell.handleTransition!(size, ZScreen.Orientation(), inRect, vr) {
	//         vr = r
	//     }
	// }
	return vr
}

// Function for setting better focus on Android, due to bug with ScrollViews.
// Does nothing elsewhere
func (s *StackView) ForceHorizontalFocusNavigation() {
}

func (v *StackView) Rect(rect Rect) View {
	v.CustomView.Rect(rect)
	v.ArrangeChildren(nil)
	return v
}

func addDiff(size *Size, vertical bool, diff *float64, count *int) {
	*(*size).VerticeP(vertical) += *diff / float64(*count)
}

func (s *StackView) ArrangeChildren(onlyChild *View) {
	for s.isLoading() {
		time.Sleep(time.Millisecond * 2)
	}
	var incs = 0
	var decs = 0
	var sizes = map[View]Size{}
	var ashrink = AlignmentHorShrink
	var aexpand = AlignmentHorExpand
	var aless = AlignmentLeft
	var amore = AlignmentRight
	var amid = AlignmentHorCenter | AlignmentMarginIsOffset

	var r = s.GetRect()
	r.Pos = Pos{} // translate to 0,0 cause children are in parent

	if s.layoutHandler != nil {
		s.layoutHandler.HandleBeforeLayout()
	}
	if s.Vertical {
		ashrink = AlignmentVertShrink
		aexpand = AlignmentVertExpand
		aless = AlignmentTop
		amore = AlignmentBottom
		amid = AlignmentVertCenter
	}
	for _, c2 := range s.cells {
		if !c2.Free {
			if c2.Collapsed {
				s.RemoveChild(c2.View)
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
		if got && s.layoutHandler != nil {
			cv.layoutHandler.HandleBeforeLayout()
		}
	}
	r.Add(s.margin)
	for _, c1 := range s.cells {
		if c1.Free {
			s.arrangeChild(c1, r)
		}
	}
	cn := r.Center().Vertice(s.Vertical)
	var cs = s.GetCalculatedSize(r.Size).Vertice(s.Vertical)
	cs += s.margin.Size.Vertice(s.Vertical) // subtracts Margin, since we've already indented for that
	diff := r.Size.Vertice(s.Vertical) - cs
	var lastNoFreeIndex = -1
	for i, c3 := range s.cells {
		if !c3.Collapsed && !c3.Free {
			lastNoFreeIndex = i
			tot := s.getCellFitSizeInTotal(r.Size, c3)
			var size = c3.View.GetCalculatedSize(tot)
			fmt.Println("size add:", c3.View, size)
			if decs > 0 && c3.Alignment&ashrink != 0 && diff != 0.0 {
				addDiff(&size, s.Vertical, &diff, &decs)
			} else if incs > 0 && c3.Alignment&aexpand != 0 && diff != 0.0 {
				addDiff(&size, s.Vertical, &diff, &incs)
			}
			sizes[c3.View] = size
		}
	}
	var centerDim = 0.0
	var firstCenter = true

	for i, c4 := range s.cells {
		if !c4.Collapsed && !c4.Free {
			if (c4.Alignment & (amore | aless)) != 0 {
				var a = c4.Alignment
				if i != lastNoFreeIndex {
					a = a.Subtracted(AlignmentExpand.Only(s.Vertical))
				}
				vr := s.handleAlign(sizes[c4.View], r, a, c4)
				DebugPrint("cell2 ", i, c4.View.GetObjectName(), sizes[c4.View])
				if onlyChild == nil || *onlyChild == c4.View {
					c4.View.Rect(vr)
				}
				if c4.Alignment&aless != 0 {
					m := math.Max(r.Min().Vertice(s.Vertical), vr.Max().Vertice(s.Vertical)+s.spacing)
					if s.Vertical {
						r.SetMinY(m)
					} else {
						r.SetMinX(m)
					}
				}
				if c4.Alignment&amore != 0 {
					m := math.Min(r.Max().Vertice(s.Vertical), vr.Pos.Vertice(s.Vertical)-s.spacing)
					if s.Vertical {
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
				centerDim += sizes[c4.View].Vertice(s.Vertical)
				if !firstCenter {
					centerDim += s.spacing
				}
				firstCenter = false
			}
		}
	}
	if s.Vertical {
		r.SetMinY(math.Max(r.Min().Y, cn-centerDim/2))
	} else {
		r.SetMinX(math.Max(r.Min().X, cn-centerDim/2))
	}
	if s.Vertical {
		r.SetMaxY(math.Min(r.Max().Y, cn+centerDim/2))
	} else {
		r.SetMaxX(math.Min(r.Max().X, cn+centerDim/2))
	}
	for _, c5 := range s.cells {
		if !c5.Collapsed && c5.Alignment&amid != 0 && !c5.Free { // .reversed()
			a := c5.Alignment.Subtracted(amid) | aless
			vr := s.handleAlign(sizes[c5.View], r, a, c5)
			if onlyChild == nil || *onlyChild == c5.View {
				c5.View.Rect(vr)
			}
			*r.Pos.VerticeP(s.Vertical) = vr.Max().Vertice(s.Vertical) + s.spacing
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

func StackRows(alignment Alignment, elements ...interface{}) *StackView {
	return StackViewNew(true, alignment, elements...)
}

func StackColumns(alignment Alignment, elements ...interface{}) *StackView {
	return StackViewNew(false, alignment, elements...)
}
