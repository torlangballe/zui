package zgo

import "strings"

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

func (a Alignment) stringStorage() string {
	var parts []string
	if a&AlignmentLeft != 0 {
		parts = append(parts, "left")
	}
	if a&AlignmentHorCenter != 0 {
		parts = append(parts, "horcenter")
	}
	if a&AlignmentRight != 0 {
		parts = append(parts, "right")
	}
	if a&AlignmentTop != 0 {
		parts = append(parts, "top")
	}
	if a&AlignmentVertCenter != 0 {
		parts = append(parts, "vertcenter")
	}
	if a&AlignmentBottom != 0 {
		parts = append(parts, "bottom")
	}
	if a&AlignmentHorExpand != 0 {
		parts = append(parts, "horexpand")
	}
	if a&AlignmentVertExpand != 0 {
		parts = append(parts, "vertexpand")
	}
	if a&AlignmentHorShrink != 0 {
		parts = append(parts, "horshrink")
	}
	if a&AlignmentVertShrink != 0 {
		parts = append(parts, "vertshrink")
	}
	if a&AlignmentHorOut != 0 {
		parts = append(parts, "horout")
	}
	if a&AlignmentVertOut != 0 {
		parts = append(parts, "vertout")
	}
	if a&AlignmentNonProp != 0 {
		parts = append(parts, "nonprop")
	}
	if a&AlignmentHorJustify != 0 {
		parts = append(parts, "horjustify")
	}
	return strings.Join(parts, " ")
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
		switch s {
		case "left":
			a = a | AlignmentLeft
		case "horcenter":
			a = a | AlignmentHorCenter
		case "right":
			a = a | AlignmentRight
		case "top":
			a = a | AlignmentTop
		case "vertcenter":
			a = a | AlignmentVertCenter
		case "bottom":
			a = a | AlignmentBottom
		case "horexpand":
			a = a | AlignmentHorExpand
		case "vertexpand":
			a = a | AlignmentVertExpand
		case "horshrink":
			a = a | AlignmentHorShrink
		case "vertshrink":
			a = a | AlignmentVertShrink
		case "horout":
			a = a | AlignmentHorOut
		case "vertout":
			a = a | AlignmentVertOut
		case "nonprop":
			a = a | AlignmentNonProp
		case "horjustify":
			a = a | AlignmentHorJustify
		default:
			break
		}
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
