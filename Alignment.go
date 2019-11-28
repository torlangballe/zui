package zgo

import (
	"sort"
	"strings"
)

type Alignment uint64

const (
	AlignmentNone                     = Alignment(0)
	AlignmentLeft                     = Alignment(1)
	AlignmentHorCenter                = Alignment(2)
	AlignmentRight                    = Alignment(4)
	AlignmentTop                      = Alignment(8)
	AlignmentVertCenter               = Alignment(16)
	AlignmentBottom                   = Alignment(32)
	AlignmentHorExpand                = Alignment(64)
	AlignmentVertExpand               = Alignment(128)
	AlignmentHorShrink                = Alignment(256)
	AlignmentVertShrink               = Alignment(512)
	AlignmentHorOut                   = Alignment(1024)
	AlignmentVertOut                  = Alignment(2048)
	AlignmentNonProp                  = Alignment(4096)
	AlignmentHorJustify               = Alignment(8192)
	AlignmentMarginIsOffset           = Alignment(16384)
	AlignmentScaleToFitProportionally = Alignment(32768)

	AlignmentCenter     = AlignmentHorCenter | AlignmentVertCenter
	AlignmentExpand     = AlignmentHorExpand | AlignmentVertExpand
	AlignmentShrink     = AlignmentHorShrink | AlignmentVertShrink
	AlignmentHorScale   = AlignmentHorExpand | AlignmentHorShrink
	AlignmentVertScale  = AlignmentVertExpand | AlignmentVertShrink
	AlignmentScale      = AlignmentHorScale | AlignmentVertScale
	AlignmentOut        = AlignmentHorOut | AlignmentVertOut
	AlignmentVertical   = AlignmentTop | AlignmentVertCenter | AlignmentBottom | AlignmentVertExpand | AlignmentVertShrink | AlignmentVertOut
	AlignmentHorizontal = AlignmentLeft | AlignmentHorCenter | AlignmentRight | AlignmentHorExpand | AlignmentHorShrink | AlignmentHorOut
)

var alignmentNames = map[string]Alignment{
	"none":                     AlignmentNone,
	"left":                     AlignmentLeft,
	"horCenter":                AlignmentHorCenter,
	"right":                    AlignmentRight,
	"top":                      AlignmentTop,
	"vertCenter":               AlignmentVertCenter,
	"bottom":                   AlignmentBottom,
	"horExpand":                AlignmentHorExpand,
	"vertExpand":               AlignmentVertExpand,
	"horShrink":                AlignmentHorShrink,
	"vertShrink":               AlignmentVertShrink,
	"horOut":                   AlignmentHorOut,
	"vertOut":                  AlignmentVertOut,
	"nonProp":                  AlignmentNonProp,
	"horJustify":               AlignmentHorJustify,
	"marginIsOffset":           AlignmentMarginIsOffset,
	"scaleToFitProportionally": AlignmentScaleToFitProportionally,
}

func AlignmentFromVector(fromVector Pos) Alignment {
	//        a.init(rawValue rawFromVector(fromVector))
	return AlignmentNone
}

func (a Alignment) FlippedVertical() Alignment {
	var r = a
	r.AndWith(AlignmentHorizontal)
	if a&AlignmentTop != 0 {
		r.UnionWith(AlignmentBottom)
	}
	if a&AlignmentBottom != 0 {
		r.UnionWith(AlignmentTop)
	}
	return r
}
func (a Alignment) FlippedHorizontal() Alignment {
	var r = a
	r.AndWith(AlignmentVertical)
	if a&AlignmentLeft != 0 {
		r.UnionWith(AlignmentRight)
	}
	if a&AlignmentRight != 0 {
		r.UnionWith(AlignmentLeft)
	}
	return r
}
func (a Alignment) Subtracted(sub Alignment) Alignment {
	return Alignment(a & Alignment(MathBitwiseInvert(uint64(sub))))
}

func (a Alignment) Only(vertical bool) Alignment {
	if vertical {
		return a.Subtracted(AlignmentHorizontal | AlignmentHorExpand | AlignmentHorShrink | AlignmentHorOut)
	}
	return a.Subtracted(AlignmentVertical | AlignmentVertExpand | AlignmentVertShrink | AlignmentVertOut)
}

func (a Alignment) String() string {
	var array []string

	center := (a&AlignmentCenter == AlignmentCenter)
	if center {
		array = append(array, "center")
	}
	for k, v := range alignmentNames {
		if center && v&AlignmentCenter != 0 {
			continue
		}
		if a&v != 0 {
			array = append(array, k)
		}
	}
	sort.Strings(array)

	return strings.Join(array, "|")
}

func AlignmentFromString(str string) Alignment {
	var a Alignment
	for _, s := range strings.Split(str, "|") {
		if s == "center" {
			a |= AlignmentCenter
		} else {
			a |= alignmentNames[s]
		}
	}
	return a
}

func (a Alignment) UnionWith(b Alignment) Alignment {
	return a | b
}

func (a Alignment) AndWith(b Alignment) Alignment {
	return a & b
}

func stringToRaw(str string) uint64 {
	var a = Alignment(0)
	for _, s := range strings.Split(str, " ") {
		a |= alignmentNames[s]
	}
	return uint64(a)
}

func rawFromVector(vector Pos) uint64 {
	var raw = Alignment(0)
	var angle = MathPosToAngleDeg(vector)
	if angle < 0 {
		angle += 360
	}
	if angle < 45*0.5 {
		raw = AlignmentRight
	} else if angle < 45*1.5 {
		raw = AlignmentRight | AlignmentTop
	} else if angle < 45*2.5 {
		raw = AlignmentTop
	} else if angle < 45*3.5 {
		raw = AlignmentTop | AlignmentLeft
	} else if angle < 45*4.5 {
		raw = AlignmentLeft
	} else if angle < 45*5.5 {
		raw = AlignmentLeft | AlignmentBottom
	} else if angle < 45*6.5 {
		raw = AlignmentBottom
	} else if angle < 45*7.5 {
		raw = AlignmentBottom | AlignmentRight
	} else {
		raw = AlignmentRight
	}
	return uint64(raw)
}
