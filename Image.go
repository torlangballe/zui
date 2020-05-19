package zui

import (
	"strconv"

	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zgeo"
)

//  Created by Tor Langballe on /20/10/15.

type Image struct {
	imageBase
	scale   int
	Path    string
	loading bool
}

type ImageOwner interface {
	GetImage() *Image
}

func MakeImageFromDrawFunction(size zgeo.Size, scale float32, draw func(size zgeo.Size, canvas Canvas)) *Image {
	return nil
}

func (i *Image) ForPixels(got func(pos zgeo.Pos, color zgeo.Color)) {
}

func (i *Image) CapInsetsCorner(c zgeo.Size) *Image {
	r := zgeo.RectFromMinMax(c.Pos(), c.Pos().Negative())
	return i.CapInsets(r)
}

func imageGetScaleFromPath(path string) int {
	var n string
	_, _, m, _ := zfile.Split(path)
	if zstr.SplitN(m, "@", &n, &m) {
		if zstr.HasSuffix(m, "x", &m) {
			scale, err := strconv.ParseInt(m, 10, 32)
			if err == nil && scale >= 1 && scale <= 3 {
				return int(scale)
			}
		}
	}
	return 1
}
