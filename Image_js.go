package zgo

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
)

type imageBase struct {
	size      zgeo.Size `json:"size"`
	capInsets zgeo.Rect `json:"capInsets"`
	hasAlpha  bool      `json:"hasAlpha"`
	imageJS   js.Value
}

func ImageFromPath(path string, got func()) *Image {
	if path == "" {
		if got != nil {
			got()
		}
		return nil
	}
	i := Image{}
	i.load(path, func() {
		i.loading = false
		if got != nil {
			got()
		}
	})
	return &i
}

func (i *Image) load(path string, done func()) {
	if !strings.HasPrefix(path, "http:") && !strings.HasPrefix(path, "https:") {
		path = "images/" + path
	}
	i.path = path
	i.loading = true
	i.scale = imageGetScaleFromPath(path)
	imageF := js.Global().Get("Image")
	i.imageJS = imageF.New()
	i.imageJS.Set("onload", js.FuncOf(func(js.Value, []js.Value) interface{} {
		i.loading = false
		i.size.W = i.imageJS.Get("width").Float()
		i.size.H = i.imageJS.Get("height").Float()
		if done != nil {
			done()
		}
		return nil
	}))

	i.imageJS.Set("src", path)
}

func (i *Image) Colored(color zgeo.Color, size zgeo.Size) *Image {
	//	rect := NewRect(0, 0, size.W, size.H)
	return i
}

func (i *Image) Size() zgeo.Size {
	return i.size.DividedByD(float64(i.scale))
}

func (i *Image) CapInsets(capInsets zgeo.Rect) *Image {
	i.capInsets = capInsets
	return i
}

func (i *Image) GetCapInsets() zgeo.Rect {
	return i.capInsets
}

func (i *Image) HasAlpha() bool {
	return i.hasAlpha
}

func (i *Image) TintedWithColor(color zgeo.Color) *Image {
	return i
}

func (i *Image) GetScaledInSize(size zgeo.Size, proportional bool) *Image {
	var vsize = size
	if proportional {
		vsize = zgeo.Rect{Size: size}.Align(i.Size(), zgeo.AlignmentCenter|zgeo.AlignmentShrink|zgeo.AlignmentScaleToFitProportionally, zgeo.Size{0, 0}, zgeo.Size{0, 0}).Size
	}
	width := int(vsize.W) / int(i.scale)
	height := int(vsize.H) / int(i.scale)
	fmt.Println("GetScaledInSize not made yet:", width, height)
	return nil
}

func (i *Image) GetCropped(crop zgeo.Rect) *Image {
	return i
}

func (i *Image) GetLeftRightFlipped() *Image {
	return i
}

func (i *Image) Normalized() *Image {
	return i
}

func (i *Image) GetRotationAdjusted(flip bool) *Image { // adjustForScreenOrientation:bool
	return i
}

func (i *Image) Rotated(deg float64, around *zgeo.Pos) *Image {
	var pos = i.Size().Pos().DividedByD(2)
	if around != nil {
		pos = *around
	}
	transform := zgeo.MatrixForRotatingAroundPoint(pos, deg)
	fmt.Println("Image.Rotated not made yet:", transform)
	return i
}

func (i *Image) FixedOrientation() *Image {
	return i
}
