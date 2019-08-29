package zgo

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	"github.com/torlangballe/zutil/ustr"
)

type imageBase struct {
	size      Size `json:"size"`
	capInsets Rect `json:"capInsets"`
	hasAlpha  bool `json:"hasAlpha"`
	imageJS   js.Value
}

func ImageFromPath(path string, got func()) *Image {
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
	i.path = path
	i.loading = true
	i.scale = getScaleFromPath(path)
	imageF := js.Global().Get("Image")
	i.imageJS = imageF.New()
	i.imageJS.Set("onload", js.FuncOf(func(js.Value, []js.Value) interface{} {
		i.loading = false
		i.size.W = i.imageJS.Get("width").Float()
		i.size.H = i.imageJS.Get("height").Float()
		fmt.Println("image loaded:", path)
		if done != nil {
			done()
		}
		return nil
	}))

	if !strings.HasPrefix(path, "http:") && !strings.HasPrefix(path, "https:") {
		path = "www/images/" + path
	}
	i.imageJS.Set("src", path)
}

func (i *Image) Colored(color Color, size Size) *Image {
	//	rect := NewRect(0, 0, size.W, size.H)
	return i
}

func (i *Image) Size() Size {
	return i.size.DividedByD(float64(i.scale))
}

func (i *Image) CapInsets(capInsets Rect) *Image {
	i.capInsets = capInsets
	return i
}

func (i *Image) GetCapInsets() Rect {
	return i.capInsets
}

func (i *Image) HasAlpha() bool {
	return i.hasAlpha
}

func (i *Image) TintedWithColor(color Color) *Image {
	return i
}

func (i *Image) GetScaledInSize(size Size, proportional bool) *Image {
	var vsize = size
	if proportional {
		vsize = Rect{Size: size}.Align(i.Size(), AlignmentCenter|AlignmentShrink|AlignmentScaleToFitProportionally, Size{0, 0}, Size{0, 0}).Size
	}
	width := int(vsize.W) / int(i.scale)
	height := int(vsize.H) / int(i.scale)
	fmt.Println("GetScaledInSize not made yet:", width, height)
	return nil
}

func (i *Image) GetCropped(crop Rect) *Image {
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

func (i *Image) Rotated(deg float64, around *Pos) *Image {
	var pos = i.Size().Pos().DividedByD(2)
	if around != nil {
		pos = *around
	}
	transform := MatrixForRotatingAroundPoint(pos, deg)
	fmt.Println("Image.Rotated not made yet:", transform)
	return i
}

func (i *Image) FixedOrientation() *Image {
	return i
}
