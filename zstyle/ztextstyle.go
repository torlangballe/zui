package zstyle

import (
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
)

type TextStyle struct {
	Size  float64
	Name  string
	Style zgeo.FontStyle
	Color zgeo.Color
	Gap   float64
}

type Name string
type Size float64
type SizeInc float64
type Gap float64
type Control int

var Start Control = 1
var NoOp Control = 2 // NoOp does nothing. Useful for maybe adding a parameter

var EmptyTextStyle = TextStyle{
	Gap: zfloat.Undefined,
}

func (t *TextStyle) Set(from TextStyle) {
	if from.Size != 0 {
		t.Size = from.Size
	}
	if from.Name != "" {
		t.Name = from.Name
	}
	if from.Style != zgeo.FontStyleUndef {
		t.Style = from.Style
	}
	if from.Color.Valid {
		t.Color = from.Color
	}
	if from.Gap != zfloat.Undefined {
		t.Gap = from.Gap
	}
}
