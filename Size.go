package zgo

import (
	"math"
)

type Size struct {
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// SizeF creates a Size from float64 w and h
func SizeF(w, h float32) Size {
	return Size{float64(w), float64(h)}
}

// SizeI creates a Size from integer w and h
func SizeI(w, h int) Size {
	return Size{float64(w), float64(h)}
}

// Pos converts a size to a Pos
func (s Size) Pos() Pos {
	return Pos{s.W, s.H}
}

//IsNull returns true if S and W are zero
func (s Size) IsNull() bool {
	return s.W == 0 && s.H == 0
}

// Vertice returns the non-vertical s.W or vertical s.H
func (s Size) Vertice(vertical bool) float64 {
	if vertical {
		return s.H
	}
	return s.W
}

// VerticeP returns a pointer to the non-vertical s.W or vertical s.H
func (s *Size) VerticeP(vertical bool) *float64 {
	if vertical {
		return &s.H
	}
	return &s.W
}

// Max returns the greater of W and H
func (s Size) Max() float64 {
	return math.Max(s.W, s.H)
}

// Min returns the lesser of W and H
func (s Size) Min() float64 {
	return math.Min(s.W, s.H)
}

// EqualSided returns a Size where W and H are largest of the two
func (s Size) EqualSided() Size {
	m := s.Max()
	return Size{m, m}
}

// Area returns the product of W, H (WxH)
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

func (s *Size) Add(a Size) {
	s.W += a.W
	s.H += a.H
}

func (s *Size) MultiplyD(a float64) {
	s.W *= a
	s.H *= a
}

func (s *Size) MultiplyF(a float32) {
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
func (s Size) MinusD(a float64) Size     { return Size{s.W - a, s.H - a} }
func (s Size) Times(a Size) Size         { return Size{s.W * a.W, s.H * a.H} }
func (s Size) TimesD(a float64) Size     { return Size{s.W * a, s.H * a} }
func (s Size) DividedBy(a Size) Size     { return Size{s.W / a.W, s.H / a.H} }
func (s Size) DividedByD(a float64) Size { return Size{s.W / a, s.H / a} }

func (s *Size) Subtract(a Size) { s.W -= a.W; s.H -= a.H }

func (s Size) Copy() Size {
	return s
}

func (s Size) ScaledInto(in Size) Size {
	if s.W == 0 || s.H == 0 || in.W == 0 || in.H == 0 {
		return Size{}
	}
	f := in.DividedBy(s)
	min := f.Min()
	scaled := s.TimesD(min)

	return scaled
}
