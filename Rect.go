package zgo

import (
	"math"
)

type Rect struct {
	Pos  Pos  `json:"pos"`
	Size Size `json:"size"`
}

func RectMake(x0, y0, x1, y1 float64) Rect {
	r := Rect{}
	r.Pos.X = x0
	r.Pos.Y = y0
	r.Size.W = x1 - x0
	r.Size.H = y1 - y0
	return r
}

func RectFromXYWH(x, y, w, h float64) Rect {
	return Rect{Pos{x, y}, Size{w, h}}
}

func RectFromMinMax(min, max Pos) Rect {
	return Rect{min, max.Minus(min).Size()}
}

var RectNull Rect

func (r Rect) IsNull() bool {
	return r.Pos.X == 0 && r.Pos.Y == 0 && r.Size.W == 0 && r.Size.H == 0
}

func (r Rect) TopLeft() Pos     { return r.Min() }
func (r Rect) TopRight() Pos    { return Pos{r.Max().X, r.Min().Y} }
func (r Rect) BottomLeft() Pos  { return Pos{r.Min().X, r.Max().Y} }
func (r Rect) BottomRight() Pos { return r.Max() }

func (r Rect) Max() Pos {
	return Pos{r.Pos.X + r.Size.W, r.Pos.Y + r.Size.H}
}

func (r *Rect) SetMax(max Pos) {
	r.Size.W = max.X - r.Pos.X
	r.Size.H = max.Y - r.Pos.Y
}

// func (r *Rect) SetMaxAsPos(max Pos) {
// 	r.Size.W = max.X - r.Pos.X
// 	r.Size.H = max.Y - r.Pos.Y
// }

func (r Rect) Min() Pos {
	return r.Pos
}
func (r *Rect) SetMin(min Pos) {
	r.Size.W += (r.Pos.X - min.X)
	r.Size.H += (r.Pos.Y - min.Y)
	r.Pos = min
}

func (r *Rect) SetMaxX(x float64) {
	r.Size.W = x - r.Pos.X
}
func (r *Rect) SetMaxY(y float64) {
	r.Size.H = y - r.Pos.Y
}
func (r *Rect) SetMinX(x float64) {
	r.Size.W += (r.Pos.X - x)
	r.Pos.X = x
}
func (r *Rect) SetMinY(y float64) {
	r.Size.H += (r.Pos.Y - y)
	r.Pos.Y = y
}

func (r Rect) Center() Pos {
	return r.Pos.Plus(r.Size.DividedByD(2).Pos())
}
func (r *Rect) SetCenter(c Pos) {
	r.Pos = c.Minus(r.Size.Pos().DividedByD(2))
}

func MergeAll(rects []Rect) []Rect {
	var merged = true
	var rold = rects
	for merged {
		var rnew []Rect
		merged = false
		for i, r := range rold {
			var used = false
			for j := i + 1; j < len(rold); j++ {
				if r.Overlaps(rold[j].ExpandedD(4)) {
					var n = rects[i]
					n.UnionWith(rold[j])
					rnew = append(rnew, n)
					merged = true
					used = true
				}
			}
			if !used {
				rnew = append(rnew, r)
			}
		}
		rold = rnew
	}
	return rold
}

func (r Rect) Expanded(e Size) Rect {
	return Rect{r.Pos.Minus(e.Pos()), r.Size.Plus(e.TimesD(2))}
}
func (r Rect) ExpandedD(n float64) Rect { return r.Expanded(Size{n, n}) }
func (r Rect) Centered(center Pos) Rect { return Rect{center.Minus(r.Size.Pos().DividedByD(2)), r.Size} }
func (r Rect) Overlaps(rect Rect) bool {
	min := r.Min()
	max := r.Max()
	rmin := rect.Min()
	rmax := rect.Max()
	return rmin.X < max.X && rmin.Y < max.Y && rmax.X > min.X && rmax.Y > min.Y
}
func (r Rect) Contains(pos Pos) bool {
	min := r.Min()
	max := r.Max()
	return pos.X >= min.X && pos.X <= max.X && pos.Y >= min.Y && pos.Y <= max.Y
}
func (r Rect) Align(s Size, align Alignment, marg Size, maxSize Size) Rect {
	var x float64
	var y float64
	var scalex float64
	var scaley float64

	var wa = float64(s.W)
	var wf = float64(r.Size.W)

	if align&AlignmentMarginIsOffset == 0 {
		wf -= float64(marg.W)
		if align&AlignmentHorCenter != 0 {
			wf -= float64(marg.W)
		}
	}
	//        }
	var ha = float64(s.H)
	var hf = float64(r.Size.H)
	//        if (align & (AlignmentVertShrink|AlignmentVertExpand)) {
	if align&AlignmentMarginIsOffset == 0 {
		hf -= float64(marg.H * 2.0)
	}
	if align == AlignmentScaleToFitProportionally {
		xratio := wf / wa
		yratio := hf / ha
		var ns = r.Size
		if xratio != 1.0 || yratio != 1.0 {
			if xratio > yratio {
				ns = Size{wf, ha * xratio}
			} else {
				ns = Size{wa * yratio, hf}
			}
		}
		return Rect{Size: (ns)}.Centered(r.Center())
	}
	if align&AlignmentHorExpand != 0 && align&AlignmentVertExpand != 0 {
		if align&AlignmentNonProp != 0 {
			wa = wf
			ha = hf
		} else {
			DebugAssert(align&AlignmentHorOut != 0)
			scalex = wf / wa
			scaley = hf / ha
			if scalex > 1 || scaley > 1 {
				if scalex < scaley {
					wa = wf
					ha *= scalex
				} else {
					ha = hf
					wa *= scaley
				}
			}
		}
	} else if align&AlignmentNonProp != 0 {
		if align&AlignmentHorExpand != 0 && wa < wf {
			wa = wf
		} else if align&AlignmentVertExpand != 0 && ha < hf {
			ha = hf
		}
	}
	if align&AlignmentHorShrink != 0 && align&AlignmentVertShrink != 0 && align&AlignmentNonProp == 0 {
		scalex = wf / wa
		scaley = hf / ha
		if align&AlignmentHorOut != 0 && align&AlignmentVertOut != 0 {
			if scalex < 1 || scaley < 1 {
				if scalex > scaley {
					wa = wf
					ha *= scalex
				} else {
					ha = hf
					wa *= scaley
				}
			}
		} else {
			if scalex < 1 || scaley < 1 {
				if scalex < scaley {
					wa = wf
					ha *= scalex
				} else {
					ha = hf
					wa *= scaley
				}
			}
		}
	} else if align&AlignmentHorShrink != 0 && wa > wf {
		wa = wf
	}
	//  else
	if align&AlignmentVertShrink != 0 && ha > hf {
		ha = hf
	}

	if maxSize.W != 0.0 {
		wa = math.Min(wa, float64(maxSize.W))
	}
	if maxSize.H != 0.0 {
		ha = math.Min(ha, float64(maxSize.H))
	}
	if align&AlignmentHorOut != 0 {
		if align&AlignmentLeft != 0 {
			x = float64(r.Pos.X - marg.W - s.W)
		} else if align&AlignmentHorCenter != 0 {
			//                x = float64(Pos.X) - wa / 2.0
			x = float64(r.Pos.X) + (wf-wa)/2.0
		} else {
			x = float64(r.Max().X + marg.W)
		}
	} else {
		if align&AlignmentLeft != 0 {
			x = float64(r.Pos.X + marg.W)
		} else if align&AlignmentRight != 0 {
			x = float64(r.Max().X) - wa - float64(marg.W)
		} else {
			x = float64(r.Pos.X)
			if align&AlignmentMarginIsOffset == 0 {
				x += float64(marg.W)
			}
			x = x + (wf-wa)/2.0
			if align&AlignmentMarginIsOffset != 0 {
				x += float64(marg.W)
			}
		}
	}

	if align&AlignmentVertOut != 0 {
		if align&AlignmentTop != 0 {
			y = float64(r.Pos.Y-marg.H) - ha
		} else if align&AlignmentVertCenter != 0 {
			y = float64(r.Pos.Y) + (hf-ha)/2.0
		} else {
			y = float64(r.Pos.Y + marg.H)
		}
	} else {
		if align&AlignmentTop != 0 {
			y = float64(r.Pos.Y + marg.H)
		} else if align&AlignmentBottom != 0 {
			y = float64(r.Max().Y) - ha - float64(marg.H)
		} else {
			y = float64(r.Pos.Y)
			if align&AlignmentMarginIsOffset == 0 {
				y += float64(marg.H)
			}
			y = y + math.Max(0.0, hf-ha)/2.0
			if align&AlignmentMarginIsOffset != 0 {
				y += float64(marg.H)
			}
		}
	}
	return Rect{Pos{x, y}, Size{wa, ha}}
}

func (r *Rect) MoveInto(rect Rect) {
	r.Pos.X = math.Max(r.Pos.X, rect.Pos.X)
	r.Pos.Y = math.Max(r.Pos.Y, rect.Pos.Y)
	min := r.Max().Min(rect.Max())
	r.Pos.Add(min.Minus(r.Max()))
}

/* #kotlin-raw:
   fun copy() : Rect {
       var r = Rect()
       r.pos = Pos.copy()
       r.size = Size.copy()
       return r
   }
*/

// Dummy function, but if translated(?) to kotlin needed to actually copy rect, not make reference
func (r Rect) Copy() Rect {
	return r
}

func (r *Rect) UnionWith(rect Rect) {
	if !rect.IsNull() {
		if r.IsNull() {
			r.Pos = rect.Pos.Copy()
			r.Size = rect.Size.Copy()
		} else {
			min := r.Min()
			max := r.Max()
			rmin := rect.Min()
			rmax := rect.Max()
			if rmin.X < min.X {
				r.SetMinX(rmin.X)
			}
			if rmin.Y < min.Y {
				r.SetMinY(rmin.Y)
			}
			if rmax.X > max.X {
				r.SetMaxX(rmax.X)
			}
			if rmax.Y > max.Y {
				r.SetMaxY(rmax.Y)
			}
		}
	}
}

func (r *Rect) UnionWithPos(pos Pos) {
	min := r.Min()
	max := r.Max()
	if pos.X > max.X {
		r.SetMaxX(pos.X)
	}
	if pos.Y > max.Y {
		r.SetMaxY(pos.Y)
	}
	if pos.X < min.X {
		r.SetMinX(pos.X)
	}
	if pos.Y < min.Y {
		r.SetMinY(pos.Y)
	}
}

func (r Rect) Plus(a Rect) Rect  { return RectFromMinMax(r.Pos.Plus(a.Pos), r.Max().Plus(a.Max())) }
func (r Rect) Minus(a Rect) Rect { return RectFromMinMax(r.Pos.Minus(a.Pos), r.Max().Minus(a.Max())) }
func (r Rect) DividedBy(a Size) Rect {
	return RectFromMinMax(r.Min().DividedBy(a.Pos()), r.Max().DividedBy(a.Pos()))
}
func (r Rect) TimesD(d float64) Rect {
	return RectFromMinMax(r.Min().TimesD(d), r.Max().TimesD(d))
}

func (r *Rect) Add(a Rect)     { r.Pos.Add(a.Pos); r.SetMax(r.Max().Plus(a.Max())) }
func (r *Rect) AddPos(a Pos)   { r.Pos.Add(a) }
func (r *Rect) Subtract(a Pos) { r.Pos.Subtract(a) }

func centerToRect(center Pos, radius float64, radiusy float64) Rect {
	var s = Size{radius, radius}
	if radiusy != 0 {
		s.H = radiusy
	}
	return Rect{center.Minus(s.Pos()), s.TimesD(2.0)}
}
