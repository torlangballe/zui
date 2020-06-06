// +build !js

package zui

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/disintegration/imaging"

	"github.com/nfnt/resize"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

// Use instead:
// https://github.com/disintegration/imaging

type imageBase struct {
	goimage image.Image
	//	size      Size `json:"size"`
	//	scale     int  `json:"scale"`
	//	capInsets Rect `json:"capInsets"`
	//	hasAlpha  bool `json:"hasAlpha"`
}

func ImageFromNative(n image.Image) *Image {
	i := &Image{}
	i.goimage = n
	i.scale = 1
	return i
}

func ImageFromPath(path string, got func(*Image)) *Image {
	i := &Image{}
	if got != nil {
		defer got(i)
	}
	file, err := os.Open(path)
	if err != nil {
		zlog.Error(err, "open", path)
		return nil
	}
	i.goimage, _, err = image.Decode(file)
	if err != nil {
		zlog.Error(err, "decode", path)
		return nil
	}
	i.scale = imageGetScaleFromPath(path)
	return i
}

func (i *Image) Colored(color zgeo.Color, size zgeo.Size) *Image {
	//	rect := NewRect(0, 0, size.W, size.H)
	return i
}

func (i *Image) Size() zgeo.Size {
	if i.goimage == nil {
		return zgeo.Size{}
	}
	s := i.goimage.Bounds().Size()
	return zgeo.Size{float64(s.X), float64(s.Y)}
}

func (i *Image) CapInsets(capInsets zgeo.Rect) *Image {
	//	i.capInsets = capInsets
	return i
}

func (i *Image) GetCapInsets() zgeo.Rect {
	return zgeo.Rect{}
}

func (i *Image) HasAlpha() bool {
	return true
}

func (i *Image) TintedWithColor(color zgeo.Color) *Image {
	return i
}

func (i *Image) ShrunkInto(size zgeo.Size, proportional bool) *Image {
	var vsize = size
	if proportional {
		vsize = zgeo.Rect{Size: size}.Align(i.Size(), zgeo.Center|zgeo.Shrink|zgeo.Proportional, zgeo.Size{0, 0}, zgeo.Size{0, 0}).Size
	}
	scale := float64(i.scale)
	width := uint(vsize.W * scale)
	height := uint(vsize.H * scale)

	newImage := resize.Resize(width, height, i.goimage, resize.Lanczos3)

	ni := &Image{}
	ni.scale = i.scale
	ni.goimage = newImage
	return ni
}

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
	// newImage, err := cutter.Crop(i.goimage, config)
	// if err != nil {
	// 	zlog.Error(err, "cutter.Crop")
	// 	return i
	// }

	r := image.Rect(int(crop.Min().X), int(crop.Min().Y), int(crop.Max().X), int(crop.Max().Y))
	newImage := imaging.Crop(i.goimage, r)

	ni := &Image{}
	ni.scale = i.scale
	ni.goimage = newImage
	return ni
}

func (i *Image) SaveToPNG(filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return zlog.Error(err, "os.create", filepath)
	}
	defer out.Close()
	err = png.Encode(out, i.goimage)
	if err != nil {
		return zlog.Error(err, "encode")
	}
	return nil
}

func (i *Image) SaveToJPEG(filepath string, qualityPercent int) error {
	out, err := os.Create(filepath)
	if err != nil {
		return zlog.Error(err, "os.create", filepath)
	}
	defer out.Close()
	err = jpeg.Encode(out, i.goimage, &jpeg.Options{Quality: 98})
	if err != nil {
		return zlog.Error(err, "encode")
	}
	return nil
}

func (i *Image) PNGData() ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	err := png.Encode(out, i.goimage)
	if err != nil {
		err = zlog.Error(err, "encode")
		return []byte{}, err
	}
	return out.Bytes(), nil
}

func (i *Image) JPEGData(qualityPercent int) ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	options := jpeg.Options{Quality: qualityPercent}
	err := jpeg.Encode(out, i.goimage, &options)
	if err != nil {
		err = zlog.Error(err, "encode")
		return []byte{}, err
	}
	return out.Bytes(), nil
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
	zlog.Info("Image.Rotated not made yet:", transform)
	return i
}

func (i *Image) FixedOrientation() *Image {
	return i
}
