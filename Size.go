package zgo

import (
	"math"
)

type Size struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

func (s Size) Pos() Pos     { return Pos{s.W, s.H} }
func (s Size) IsNull() bool { return s.W == 0 && s.H == 0 }

func (s Size) Subscript(vertical bool) float64 {
	if vertical {
		return s.H
	}
	return s.W
}

func (s Size) Max() float64 {
	return math.Max(s.W, s.H)
}

func (s Size) Min() float64 {
	return math.Max(s.W, s.H)
}

func (s Size) EqualSided() Size {
	m := s.Max()
	return Size{m, m}
}

func (s Size) Area() float64 {
	return s.W * s.H
}

func (s *Size) Maximize(a Size) {
	s.W = math.Max(s.W, a.W)
	s.H = math.Max(s.H, a.H)
}

func (s *Size) Minimize(a Size) {
	s.W = math.Min(s.W, a.W)
	s.H = math.Min(s.H, a.H)
}

func (s *Size) MultiplyD(a float64) {
	s.W *= a
	s.H *= a
}

func (s Size) MultiplyF(a float32) {
	s.W *= float64(a)
	s.H *= float64(a)
}

func (s Size) Negative() Size {
	return Size{-s.W, -s.H}
}

func (s Size) Equals(a Size) bool {
	return s.W == a.W && s.H == a.H
}

func (s Size) Plus(a Size) Size          { return Size{s.W + a.W, s.H + a.H} }
func (s Size) Minus(a Size) Size         { return Size{s.W - a.W, s.H - a.H} }
func (s Size) Times(a Size) Size         { return Size{s.W * a.W, s.H * a.H} }
func (s Size) TimesD(a float64) Size     { return Size{s.W * a, s.H * a} }
func (s Size) DividedBy(a Size) Size     { return Size{s.W / a.W, s.H / a.H} }
func (s Size) DividedByD(a float64) Size { return Size{s.W / a, s.H / a} }

func (s Size) Copy() Size {
	return s
}
