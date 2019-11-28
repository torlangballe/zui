// +build !js

package zgo

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/disintegration/imaging"

	"github.com/nfnt/resize"
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

func ImageFromPath(path string, got func()) *Image {
	i := &Image{}
	if got != nil {
		defer got()
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

func (i *Image) Colored(color Color, size Size) *Image {
	//	rect := NewRect(0, 0, size.W, size.H)
	return i
}

func (i *Image) Size() Size {
	if i.goimage == nil {
		return Size{}
	}
	s := i.goimage.Bounds().Size()
	return Size{float64(s.X), float64(s.Y)}
}

func (i *Image) CapInsets(capInsets Rect) *Image {
	//	i.capInsets = capInsets
	return i
}

func (i *Image) GetCapInsets() Rect {
	return Rect{}
}

func (i *Image) HasAlpha() bool {
	return true
}

func (i *Image) TintedWithColor(color Color) *Image {
	return i
}

func (i *Image) ShrunkInto(size Size, proportional bool) *Image {
	var vsize = size
	if proportional {
		vsize = Rect{Size: size}.Align(i.Size(), AlignmentCenter|AlignmentShrink|AlignmentScaleToFitProportionally, Size{0, 0}, Size{0, 0}).Size
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

func (i *Image) Cropped(crop Rect, copy bool) *Image {
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

func (i *Image) SaveToPNG(file FilePath) error {
	out, err := os.Create(file.String())
	if err != nil {
		return zlog.Error(err, "os.create")
	}
	defer out.Close()
	err = png.Encode(out, i.goimage)
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
