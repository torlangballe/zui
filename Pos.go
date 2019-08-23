package zgo

import (
	"math"

	"github.com/torlangballe/zutil/zmath"
)

type Pos struct {
	X float64
	Y float64
}

func (p Pos) Vertice(vertical bool) float64 {
	if vertical {
		return p.Y
	}
	return p.X
}

func (p *Pos) VerticeP(vertical bool) *float64 {
	if vertical {
		return &p.Y
	}
	return &p.X
}

func (p *Pos) SetOne(vertical bool, v float64) {
	if vertical {
		p.Y = v
	}
	p.X = v
}

func (p Pos) Size() Size {
	return Size{p.X, p.Y}
}

func (p *Pos) Set(x, y float64) {
	*p = Pos{x, y}
}

func (p *Pos) SetF(x, y float32) {
	*p = Pos{float64(x), float64(y)}
}

func (p *Pos) Swap() {
	*p = Pos{p.Y, p.X}
}

func (p Pos) Max(a Pos) Pos {
	return Pos{math.Max(p.X, a.X), math.Max(p.Y, a.Y)}
}

func (p Pos) Min(a Pos) Pos {
	return Pos{math.Min(p.X, a.X), math.Min(p.Y, a.Y)}
}

func (p Pos) GetRot90CW() Pos    { return Pos{p.Y, -p.X} }
func (p Pos) Dot(a Pos) float64  { return p.X*a.X + p.Y*a.Y }
func (p Pos) Length() float64    { return math.Sqrt(p.X*p.X + p.Y*p.Y) }
func (p Pos) IsNull() bool       { return p.X == 0.0 && p.Y == 0.0 }
func (p Pos) GetNormalized() Pos { return p.DividedByD(p.Length()) }
func (p Pos) Sign() Pos          { return Pos{zmath.Sign(p.X), zmath.Sign(p.Y)} }
func (p Pos) Negative() Pos {
	return Pos{-p.X, -p.Y}
}
func (p Pos) Abs() Pos {
	return Pos{math.Abs(p.X), math.Abs(p.Y)}
}
func (p Pos) IsSameDirection(pos Pos) bool {
	if p == pos {
		return true
	}
	if zmath.Sign(pos.X) != zmath.Sign(p.X) || zmath.Sign(pos.Y) != zmath.Sign(p.Y) {
		return false
	}
	if pos.Y == 0.0 {
		return p.Y == 0.0
	}
	if p.Y == 0.0 {
		return p.Y == 0.0
	}
	if p.X/p.Y == pos.X/pos.Y {
		return true
	}
	return false
}

func (p Pos) RotatedCCW(angle float64) Pos {
	s := math.Sin(angle)
	c := math.Cos(angle)

	return Pos{p.X*c - p.Y*s, p.X*s + p.Y*c}
}

func (p *Pos) MultiplyD(a float64) {
	p.X *= a
	p.Y *= a
}

func (p Pos) PlusD(a float64) Pos      { return Pos{p.X + a, p.Y + a} }
func (p Pos) MinusD(a float64) Pos     { return Pos{p.X - a, p.Y - a} }
func (p Pos) TimesD(a float64) Pos     { return Pos{p.X * a, p.Y * a} }
func (p Pos) DividedByD(a float64) Pos { return Pos{p.X / a, p.Y / a} }
func (p Pos) Plus(a Pos) Pos           { return Pos{p.X + a.X, p.Y + a.Y} }
func (p Pos) Minus(a Pos) Pos          { return Pos{p.X - a.X, p.Y - a.Y} }
func (p Pos) Times(a Pos) Pos          { return Pos{p.X * a.X, p.Y * a.Y} }
func (p Pos) DividedBy(a Pos) Pos      { return Pos{p.X / a.X, p.Y / a.Y} }
func (p Pos) AddedSize(s Size) Pos     { return Pos{p.X + s.W, p.Y + s.H} }
func (p Pos) Equals(a Pos) bool        { return p.X == a.X && p.Y == a.Y }
func (p *Pos) Add(a Pos)               { p.X += a.X; p.Y += a.Y }
func (p *Pos) Subtract(a Pos)          { p.X -= a.X; p.Y -= a.Y }

type FPos struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

func (p FPos) Pos() Pos {
	return Pos{float64(p.X), float64(p.Y)}
}

// itterates through positions, making vector between them, optionally closing
func ForVectors(positions []Pos, close bool, handle func(s Pos, v Pos) bool) {
	var i = 0

	for i < len(positions) {
		s := positions[i]
		e := Pos{}
		if i == len(positions)-1 {
			if close {
				e = positions[0].Minus(s)
			} else {
				break
			}
		} else {
			e = positions[i+1]
		}
		if !handle(s, e.Minus(s)) {
			break
		}
		i++
	}
}

func GetTPositionInPosPath(path []Pos, t float64, close bool) Pos {
	var length = 0.0
	var resultPos = Pos{}

	if t <= 0 {
		return path[0]
	}
	ForVectors(path, close, func(s, v Pos) bool {
		length += v.Length()
		return true
	})
	if t >= 1 {
		if close {
			return path[0]
		}
		return path[len(path)-1]
	}

	tlen := t * length
	length = 0.0

	ForVectors(path, close, func(s, v Pos) bool {
		vlen := v.Length()
		l := length + vlen
		if l >= tlen {
			ldiff := tlen - length
			f := ldiff / vlen
			resultPos = s.Plus(v.TimesD(f))
			return false
		}
		length = l
		return true
	})

	return resultPos
}

func (p Pos) Copy() Pos {
	return p
}
