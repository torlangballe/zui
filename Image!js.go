//go:build !js
// +build !js

package zui

import (
	"image"
	"image/png"
	"os"

	"github.com/disintegration/imaging"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
)

// Use instead:
// https://github.com/disintegration/imaging

type imageBase struct {
	GoImage image.Image
	//	size      Size `json:"size"`
	//	scale     int  `json:"scale"`
	//	capInsets Rect `json:"capInsets"`
	//	hasAlpha  bool `json:"hasAlpha"`
}

func ImageNewNRGBA(size zgeo.Size) *Image {
	i := &Image{}
	i.GoImage = image.NewNRGBA(zgeo.Rect{Size: size}.GoRect())
	return i
}

func ImageFromGo(img image.Image, got func(image *Image)) {
	i := &Image{}
	i.GoImage = img
	i.scale = 1
	got(i)
}

func (i *Image) ToGo() image.Image {
	return i.GoImage
}

func ImageFromPath(path string, got func(*Image)) {
	var goImage image.Image
	if zhttp.StringStartsWithHTTPX(path) {
		goImage, _ = GoImageFromURL(path)
	} else {
		goImage, _, _ = GoImageFromFile(path)
	}
	if goImage == nil {
		if got != nil {
			got(nil)
		}
	}
	ImageFromGo(goImage, func(i *Image) {
		i.scale = imageGetScaleFromPath(path)
		if got != nil {
			got(i)
		}
	})
}

func (i *Image) Colored(color zgeo.Color, size zgeo.Size) *Image {
	//	rect := NewRect(0, 0, size.W, size.H)
	return i
}

func (i *Image) Size() zgeo.Size {
	if i.GoImage == nil {
		return zgeo.Size{}
	}
	s := i.GoImage.Bounds().Size()
	return zgeo.Size{float64(s.X), float64(s.Y)}
}

func (i *Image) SetCapInsets(capInsets zgeo.Rect) *Image {
	//	i.capInsets = capInsets
	return i
}

func (i *Image) CapInsets() zgeo.Rect {
	return zgeo.Rect{}
}

func (i *Image) HasAlpha() bool {
	return true
}

// func (i *Image) ShrunkInto(size zgeo.Size, proportional bool) *Image {
// 	scale := float64(i.scale)
// 	newImage := GoImageShrunkInto(i.GoImage, scale, size, proportional)
// 	zlog.Assert(newImage != nil, "shunk image nil")
// 	img := ImageFromGo(newImage)
// 	return img
// }

func (i *Image) Cropped(crop zgeo.Rect, copy bool) *Image {
	// config := cutter.Config{
	// 	Width:  int(crop.Size.W),
	// 	Height: int(crop.Size.H),
	// 	Anchor: image.Point{int(crop.Pos.X), int(crop.Pos.Y)},
	// 	Mode:   cutter.TopLeft,
	// }
	// if copy {
	// 	config.Options = cutter.Copy
	// }
	// newImage, err := cutter.Crop(i.GoImage, config)
	// if err != nil {
	// 	zlog.Error(err, "cutter.Crop")
	// 	return i
	// }

	r := image.Rect(int(crop.Min().X), int(crop.Min().Y), int(crop.Max().X), int(crop.Max().Y))
	newImage := imaging.Crop(i.GoImage, r)

	ni := &Image{}
	ni.scale = i.scale
	ni.GoImage = newImage
	return ni
}

func (i *Image) SaveToPNG(filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return zlog.Error(err, "os.create", filepath)
	}
	defer out.Close()
	err = png.Encode(out, i.GoImage)
	if err != nil {
		return zlog.Error(err, "encode")
	}
	return nil
}

func ImageFromFile(filepath string, got func(i *Image, format string, err error)) {
	gi, format, err := GoImageFromFile(filepath)
	if gi == nil {
		got(nil, format, err)
		return
	}
	ImageFromGo(gi, func(zi *Image) {
		got(zi, format, err)
	})
}

// func (i *Image) PNGData() ([]byte, error) {
// 	return GoImagePNGData(i.GoImage)
// }

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
	zlog.Info("Image.Rotated not made yet:", transform)
	return i
}

func (i *Image) FixedOrientation() *Image {
	return i
}

func (i *Image) RGBAImage(got func(img *Image)) {
	n := GoImageToGoRGBA(i.GoImage)
	ImageFromGo(n, got)
}
