package zgo

import "math"

type Matrix struct {
	A, B, C, D, Tx, Ty float64
}

var IdentityMatrix = Matrix{1, 0, 0, 1, 0, 0}

/*
func SM(w, h Num) *Matrix {
	return &Matrix{sz.W, 0, 0, sz.H, 0, 0}
}

func TM(dx, dy Num) *Matrix {
	return &Matrix{1, 0, 0, 1, dx, dy}
}
*/

func ScaleMatrix(sz Size) *Matrix {
	return &Matrix{sz.W, 0, 0, sz.H, 0, 0}
}

func TranslateMatrix(delta Size) *Matrix {
	return &Matrix{1, 0, 0, 1, delta.W, delta.H}
}

func RotateMatrix(angle float64) *Matrix {
	s, c := math.Sincos(angle)
	return &Matrix{c, s, -s, c, 0, 0}
}

func (m Matrix) MulPos(pt Pos) Pos {
	return Pos{m.A*pt.X + m.C*pt.Y + m.Tx, m.B*pt.X + m.D*pt.Y + m.Ty}
}

func (m Matrix) MulSize(sz Size) Size {
	return Size{m.A*sz.W + m.C*sz.H, m.B*sz.W + m.D*sz.H}
}

func (m Matrix) TransformRect(rect Rect) Rect {
	var r Rect
	r.SetMin(m.MulPos(rect.Min()))
	r.SetMax(m.MulPos(rect.Max()))
	return r
}

func (m Matrix) Multiplied(a Matrix) Matrix {
	m.A, m.B, m.C, m.D, m.Tx, m.Ty =
		a.A*m.A+a.B*m.C, a.A*m.B+a.B*m.D,
		a.C*m.A+a.D*m.C, a.C*m.B+a.D*m.D,
		a.Tx*m.A+a.Ty*m.C+m.Tx, a.Tx*m.B+a.Ty*m.D+m.Ty
	return m
}

func (m Matrix) Rotated(angle float64) Matrix {
	s, c := math.Sincos(angle)
	sv, cv := s, c
	m.A, m.B, m.C, m.D = cv*m.A+sv*m.C, cv*m.B+sv*m.D, cv*m.C-sv*m.A, cv*m.D-sv*m.B
	return m
}

func (m Matrix) Scaled(sz Size) Matrix {
	m.A *= sz.W
	m.B *= sz.W
	m.C *= sz.H
	m.D *= sz.H
	return m
}

func (m Matrix) Translated(delta Size) Matrix {
	m.Tx += delta.W*m.A + delta.H*m.C
	m.Ty += delta.W*m.B + delta.H*m.D
	return m
}

func (m Matrix) TranslatedByPos(delta Pos) Matrix {
	return m.Translated(delta.Size())
}

// Det calculates the determinant of the matrix
func (m Matrix) det() float64 {
	return m.A*m.D - m.C*m.B
}

func (m Matrix) Inverted() (Matrix, bool) {
	det := m.det()
	if det == 0 {
		return Matrix{}, false
	}
	return Matrix{
		m.D / det, -m.B / det,
		-m.C / det, m.A / det,
		(m.Ty*m.C - m.Tx*m.D) / det, (m.Tx*m.B - m.Ty*m.A) / det,
	}, true
}
