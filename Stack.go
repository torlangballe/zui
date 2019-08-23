package zgo

import (
	"math"
	"time"
)

//  Created by Tor Langballe on /20/10/15.

type StackView struct {
	ContainerView
	space    float64
	Vertical bool
}

func Stack(vertical bool, alignment Alignment, elements ...interface{}) *StackView {
	s := &StackView{}
	s.ContainerView = *ContainerViewNew(s)
	s.SetElements(alignment, elements...)
	s.Vertical = vertical
	s.space = 6
	return s
}

func (s *StackView) Space(space float64) *StackView {
	s.space = space
	return s
}

func (s *StackView) GetSpace() float64 {
	return s.space
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
			*size.VerticeP(s.Vertical) += s.space
		}
	}
	size.Subtract(s.margin.Size)
	if len(s.cells) > 0 {
		*size.VerticeP(s.Vertical) -= s.space
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
				zRemoveViewFromSuper(c2.View.GetView(), false)
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
	var r = s.GetView().GetRect()
	r.Pos = Pos{} // translate to 0,0 cause children are in parent
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
			// var size = zConvertViewSizeThatFitstToSize(c3.View.GetView(), tot)
			var size = c3.View.GetCalculatedSize(tot)
			if decs > 0 && c3.Alignment&ashrink != 0 && diff != 0.0 {
				*size.VerticeP(s.Vertical) += diff / float64(decs)
			} else if incs > 0 && c3.Alignment&aexpand != 0 && diff != 0.0 {
				*size.VerticeP(s.Vertical) += diff / float64(incs)
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
//				DebugPrint("cell2 ", c4.Alignment, r, vr, c4.View.GetObjectName(), sizes[c4.View])
				if onlyChild == nil || *onlyChild == c4.View {
					zViewSetRect(c4.View, vr, true)
				}
				if c4.Alignment&aless != 0 {
					m := math.Max(r.Min().Vertice(s.Vertical), vr.Max().Vertice(s.Vertical)+s.space)
					if s.Vertical {
						r.SetMinY(m)
					} else {
						r.SetMinX(m)
					}
				}
				if c4.Alignment&amore != 0 {
					m := math.Min(r.Max().Vertice(s.Vertical), vr.Pos.Vertice(s.Vertical)-s.space)
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
					centerDim += s.space
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
				zViewSetRect(c5.View, vr, true)
			}
			//                ZDebug.Print("alignm ", (c5.view as! ZView).objectName, vr)
			*r.Pos.VerticeP(s.Vertical) = vr.Max().Vertice(s.Vertical) + s.space
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
	return Stack(true, alignment, elements...)
}

func StackColumns(alignment Alignment, elements ...interface{}) *StackView {
	return Stack(false, alignment, elements...)
}

/*
open class ZColumnStack   ZStackView {
    var vstack  ZStackView? = nil
    var max Int = 0

    init(max Int, horSpace float64) {
        self.max = max
        super.init(name "zcolumnstack")
        space = horSpace
        vertical = false
    }

    // #swift-only
    required public init?(coder aDecoder  NSCoder) { fatalError("init(coder )") }
    // #end

    @discardableResult override open func Add( view  ZNativeView, align Alignment, marg  Size  = Size (), maxSize Size  = Size (), index Int = -1, free Bool = false)  Int {
        if vstack == nil || vstack!.cells.count == max {
            vstack = ZVStackView(space space)
            return super.Add(vstack!, align AlignmentLeft | AlignmentBottom, marg Size (), maxSize Size (), index -1, free false)
        }
        return vstack!.Add(view, align align, marg marg, maxSize maxSize, index index, free free) // need all args specified for kotlin super call
    }
}



*/
