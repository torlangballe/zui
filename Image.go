package zui

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"

	//"github.com/nfnt/resize"

	"github.com/bamiaux/rez"
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
	SetImage(image *Image, path string, got func(*Image))
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

func GoImageZSize(img image.Image) zgeo.Size {
	return zgeo.Size{float64(img.Bounds().Dx()), float64(img.Bounds().Dy())}
}

// GGoImageShrunkInto scales down the image to fit inside size.
// It must be a subset of standard libarary image types, as it uses rez
// package to downsample, which works on underlying image types.
func GoImageShrunkInto(goImage image.Image, screenScale float64, size zgeo.Size, proportional bool) image.Image {
	var vsize = size
	s := GoImageZSize(goImage)
	if proportional {
		vsize = zgeo.Rect{Size: size}.Align(s, zgeo.Center|zgeo.Shrink|zgeo.Proportional, zgeo.Size{}).Size
	}
	nSize := vsize.TimesD(screenScale)

	//	this didn't work for large image?
	width := int(nSize.W)
	height := int(nSize.H)
	var newImage image.Image
	// zlog.Info("** Shrink:", reflect.ValueOf(goImage).Type(), reflect.ValueOf(goImage).Kind(), zlog.GetCallingStackString())
	if s.Max() > 1000 {
		//		nrgba := NRGBAImage()
		goRect := zgeo.Rect{Size: nSize}.GoRect()
		newImage = image.NewRGBA(goRect)
		err := rez.Convert(newImage, goImage, rez.NewBilinearFilter()) //NewBicubicFilter
		if err != nil {
			zlog.Error(err, "rez Resample")
			return nil
		}
	} else {
		newImage = imaging.Resize(goImage, width, height, imaging.Lanczos)
	}
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
		params.Headers["Origin"] = "https://192.168.0.30:443"
		resp, err := zhttp.GetResponse(path, params)
		if err != nil {
			zlog.Error(err, "get", path, zlog.GetCallingStackString())
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

func GoImagePNGData(goImage image.Image) ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	err := png.Encode(out, goImage)
	if err != nil {
		err = zlog.Error(err, "encode")
		return []byte{}, err
	}
	return out.Bytes(), nil
}

func GoImageJPEGData(goImage image.Image, qualityPercent int) ([]byte, error) {
	out := bytes.NewBuffer([]byte{})
	options := jpeg.Options{Quality: qualityPercent}
	err := jpeg.Encode(out, goImage, &options)
	if err != nil {
		err = zlog.Error(err, "encode")
		return []byte{}, err
	}
	return out.Bytes(), nil
}

func GoImageToJPEGFile(img image.Image, filepath string, qualityPercent int) error {
	out, err := os.Create(filepath)
	if err != nil {
		return zlog.Error(err, "os.create", filepath)
	}
	defer out.Close()
	options := jpeg.Options{Quality: qualityPercent}
	err = jpeg.Encode(out, img, &options)
	if err != nil {
		return zlog.Error(err, "encode")
	}
	return nil
}

func (i *Image) Merge(myMaxSize zgeo.Size, with *Image, align zgeo.Alignment, marg, withMaxSize zgeo.Size) *Image {
	zlog.Assert(i != nil)
	zlog.Assert(with != nil)
	s := i.Size()
	if !myMaxSize.IsNull() {
		s = s.ShrunkInto(myMaxSize)
	}
	ws := with.Size()
	if !withMaxSize.IsNull() {
		ws = ws.ShrunkInto(withMaxSize)
	}
	mr := zgeo.Rect{Size: s}
	wr := mr.Align(ws, align, marg)
	box := mr.UnionedWith(wr)
	canvas := CanvasNew()
	// canvas.DownsampleImages = false
	delta := zgeo.Pos{-math.Min(wr.Pos.X, mr.Pos.X), -math.Min(wr.Pos.Y, mr.Pos.Y)}
	// zlog.Info("image.Merge1:", box, mr, wr, delta)
	mr.AddPos(delta)
	wr.AddPos(delta)
	// zlog.Info("image.Merge2:", box, mr, wr)
	canvas.SetSize(box.Size)
	// canvas.SetColor(zgeo.ColorBlue)
	// canvas.DrawRect(mr)
	// canvas.SetColor(zgeo.ColorRed)
	// canvas.DrawRect(wr)
	synchronous := true
	downsampleCache := false
	canvas.DrawImage(i, synchronous, downsampleCache, mr, 1, zgeo.Rect{})
	canvas.DrawImage(with, synchronous, downsampleCache, wr, 1, zgeo.Rect{})
	return canvas.ZImage()
}

func GoImageToGoRGBA(i image.Image) image.Image {
	r := i.Bounds()
	n := image.NewRGBA(r)
	draw.Draw(n, r, i, image.Point{}, draw.Over)
	return n
}

func (i *Image) ShrunkInto(size zgeo.Size, proportional bool) *Image {
	// this can be better, use canvas.Image()
	goImage := ImageToGo(i)
	if goImage == nil {
		zlog.Error(nil, "goImageFromPath")
		return nil
	}
	screenScale := float64(i.scale)
	newGoImage := GoImageShrunkInto(goImage, screenScale, size, proportional)
	img := ImageFromGo(newGoImage)
	return img
}
