package zui

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"

	//"github.com/bamiaux/rez"

	//"github.com/nfnt/resize"

	"github.com/disintegration/imaging"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
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

func (i *Image) SetCapInsetsCorner(c zgeo.Size) *Image {
	r := zgeo.RectFromMinMax(c.Pos(), c.Pos().Negative())
	return i.SetCapInsets(r)
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

// goImageShrunkInto scales down the image to fit inside size.
// It must be a subset of standard libarary image types, as it uses rez
// package to downsample, which works on underlying image types.

func goImageShrunkInto(goImage image.Image, scale float64, size zgeo.Size, proportional bool) image.Image {
	var vsize = size
	b := goImage.Bounds()
	s := zgeo.Size{float64(b.Dx()), float64(b.Dy())}
	if proportional {
		vsize = zgeo.Rect{Size: size}.Align(s, zgeo.Center|zgeo.Shrink|zgeo.Proportional, zgeo.Size{0, 0}, zgeo.Size{0, 0}).Size
	}
	nSize := vsize.TimesD(scale)
	width := int(nSize.W)
	height := int(nSize.H)
	//	newImage := resize.Resize(width, height, goImage, resize.Lanczos3) //Bicubic)

	newImage := imaging.Resize(goImage, width, height, imaging.Lanczos)

	// goRect := zgeo.Rect{Size: nSize}.GoRect()
	// newImage := image.NewRGBA(goRect)
	// err := rez.Convert(newImage, goImage, rez.NewBilinearFilter()) //NewBicubicFilter
	// if err != nil {
	// 	zlog.Error(err, "rez Resample")
	// }
	return newImage
}

func goImageFromPath(path string, isFile bool) image.Image {
	var err error
	var reader io.Reader
	if isFile {
		file, err := os.Open(path)
		if err != nil {
			zlog.Error(err, "open", path)
			return nil
		}
		reader = file
	} else {
		params := zhttp.MakeParameters()
		params.Method = http.MethodGet
		resp, err := zhttp.GetResponse(path, params)
		if err != nil {
			zlog.Error(err, "get", path)
			return nil
		}
		reader = resp.Body
	}
	goImage, _, err := image.Decode(reader)
	if err != nil {
		zlog.Error(err, "decode", path)
		return nil
	}
	return goImage
}

func goImagePNGData(goImage image.Image) ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	err := png.Encode(out, goImage)
	if err != nil {
		err = zlog.Error(err, "encode")
		return []byte{}, err
	}
	return out.Bytes(), nil
}
