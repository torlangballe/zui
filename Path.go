package zgo

import (
	"math"
)

//  Created by Tor Langballe on /21/10/15.

type PathLineType int
type PathPartType int

const (
	Square PathLineType = iota
	Round
	Butt
)

const (
	PathMove PathPartType = iota
	PathLine
	PathQuadCurve
	PathCurve
	PathClose
)

type node struct {
	Type   PathPartType
	Points []Pos
}
type Path struct {
	nodes []node
}

func NewPath() *Path {
	return new(Path)
}

func (p *Path) Copy() *Path {
	n := NewPath()
	n.nodes = append(n.nodes, p.nodes...)
	return n
}

func NewRectPath(rect Rect, corner Size) *Path {
	p := NewPath()

	p.AddRect(rect, Size{})
	return p
}

func NewOvalPath(rect Rect) *Path {
	p := NewPath()
	p.AddOval(rect)
	return p
}

func (p *Path) Empty() {
	p.nodes = p.nodes[:]
}

func (p *Path) IsEmpty() bool {
	return len(p.nodes) == 0
}

func (p *Path) GetRect() Rect {
	if p.IsEmpty() {
		return Rect{}
	}
	var box Rect
	first := true
	p.ForEachPart(func(part node) {
		if first {
			box.Pos = part.Points[0]
		} else {
			box.UnionWithPos(part.Points[0])
		}
		switch part.Type {
		case PathQuadCurve:
			box.UnionWithPos(part.Points[1])
		case PathCurve:
			box.UnionWithPos(part.Points[1])
			box.UnionWithPos(part.Points[2])
		}
		first = false
	})
	return box
}

func (p *Path) AddOval(inrect Rect) {
}

func (p *Path) GetPos() Pos {
	l := len(p.nodes)
	if l != 0 {
		p := p.nodes[l].Points
		pl := len(p)
		if pl != 0 {
			return p[pl-1]
		}
	}
	return Pos{}
}

func (p *Path) MoveTo(pos Pos) {
	p.nodes = append(p.nodes, node{PathMove, []Pos{pos}})
}

func (p *Path) LineTo(pos Pos) {
	p.nodes = append(p.nodes, node{PathLine, []Pos{pos}})
}

func (p *Path) QuadCurveTo(a, b Pos) {
	p.nodes = append(p.nodes, node{PathQuadCurve, []Pos{a, b}})
}

func (p *Path) BezierTo(c1 Pos, c2 Pos, end Pos) {
	p.nodes = append(p.nodes, node{PathCurve, []Pos{c1, c2, end}})
}

func (p *Path) Close() {
	p.nodes = append(p.nodes, node{PathClose, []Pos{}})
}

func polarPoint(r float64, phi float64) Pos {
	s, c := math.Sincos(phi)
	return Pos{r * c, r * s}
}

func arcControlPoints(angle, delta float64) (Size, Size) {
	p0 := polarPoint(1, angle)
	p1 := polarPoint(1, angle+delta)
	n0 := Size{p0.Y, -p0.X} // rot 90
	n1 := Size{-p1.Y, p1.X} // ccw 90
	var s float64
	if math.Abs(n0.W+n1.W) > math.Abs(n0.H+n1.H) {
		s = (float64(math.Cos(angle+delta/2)*2) - p0.X - p1.X) * (4 / 3.0) / (n0.W + n1.W)
	} else {
		s = (float64(math.Sin(angle+delta/2)*2) - p0.Y - p1.Y) * (4 / 3.0) / (n0.H + n1.H)
	}
	return Size{p0.X + n0.W*s, p0.Y + n0.H*s}, Size{p1.X + n1.W*s, p1.Y + n1.H*s}
}

func (p *Path) ArcTo(rect Rect, degStart, degDelta float64, clockwise bool) {
	circleCenter := rect.Center()
	circleRadius := rect.Size.W / 2
	p0 := polarPoint(circleRadius, degStart).Plus(circleCenter)
	needLineTo := false
	if p.IsEmpty() || p.nodes[len(p.nodes)-1].Type == PathClose {
		p.LineTo(p0)
	} else {
		p.MoveTo(p0)
		needLineTo = true
	}
	if degDelta == 0 || circleRadius <= 0 {
		if needLineTo {
			p.LineTo(p0)
		}
		return
	}
	n := math.Ceil(math.Abs(degDelta) / (math.Pi / 2))
	rm := MatrixIdentity.RotatedAroundPos(circleCenter, degDelta/n)
	k0, k1 := arcControlPoints(degStart, degDelta/n)
	c0 := Pos{k0.W*circleRadius + circleCenter.X, k0.H*circleRadius + circleCenter.Y}
	c1 := Pos{k1.W*circleRadius + circleCenter.X, k1.H*circleRadius + circleCenter.Y}
	for i := 0; i < int(n); i++ {
		p0 = rm.MulPos(p0)
		p.BezierTo(c0, c1, p0)
		c0 = rm.MulPos(c0)
		c1 = rm.MulPos(c1)
	}
}

func (p *Path) Transformed(m *Matrix) (newPath *Path) {
	newPath = NewPath()
	for _, n := range p.nodes {
		nn := node{}
		for _, p := range n.Points {
			nn.Points = append(n.Points, m.MulPos(p))
		}
		newPath.nodes = append(newPath.nodes, nn)
	}
	return
}

func (p *Path) AddPath(addPath *Path, join bool, m *Matrix) {
	if m != nil {
		addPath = addPath.Transformed(m)
	}
	p.nodes = append(p.nodes, addPath.nodes...)
}

func (p *Path) Rotated(deg float64, origin *Pos) *Path {
	var pos = Pos{}
	if origin == nil {
		bounds := p.GetRect()
		pos = bounds.Center()
	} else {
		pos = *origin
	}
	angle := MathDegToRad(deg)
	m := MatrixIdentity.RotatedAroundPos(pos, angle)
	return p.Transformed(&m)
}

func (p *Path) ForEachPart(forPart func(part node)) {
	for _, ppt := range p.nodes {
		forPart(ppt)
	}
}

func (p *Path) AddRect(rect Rect, corner Size) {
	if rect.Size.W != 0 && rect.Size.H != 0 {
		p.MoveTo(rect.TopLeft())
		if corner.IsNull() || rect.Size.W == 0 || rect.Size.H == 0 {
			p.LineTo(rect.TopRight())
			p.LineTo(rect.BottomRight())
			p.LineTo(rect.BottomLeft())
			p.Close()
		} else {
			min := rect.Min()
			max := rect.Max()
			p.LineTo(Pos{max.X - corner.W, min.Y})
			p.QuadCurveTo(Pos{max.X, min.Y}, Pos{max.X, min.Y + corner.H})
			p.LineTo(Pos{max.X, max.Y - corner.H})
			p.QuadCurveTo(Pos{max.X, max.Y}, Pos{max.X - corner.H, max.Y})
			p.LineTo(Pos{min.X + corner.W, max.Y})
			p.QuadCurveTo(Pos{min.X, max.Y}, Pos{min.X, max.Y - corner.H})
			p.LineTo(Pos{min.X, min.Y + corner.H})
			p.QuadCurveTo(Pos{min.X, min.Y}, Pos{min.X + corner.W, min.Y})
			p.Close()
		}
	}
}
