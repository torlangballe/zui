// Copyright 2022 Tor Langballe. All rights reserved. Created by Tor Langballe on /20/10/15.
// Package image implements a pixel image; On a browser it a javascript image,
// otherwise a wrapper to a go image.Image (and other native images in future).
// It also has numerous general image manipulation functions.

package zimage

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/bamiaux/rez"
	"github.com/disintegration/imaging"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
)

type Image struct {
	imageBase
	Scale   int
	Path    string
	Loading bool
}

type SetableImage interface {
	image.Image
	Set(x, y int, c color.Color)
}

type Loader interface {
	IsLoading() bool
}

type Owner interface {
	GetImage() *Image
	SetImage(image *Image, path string, got func(*Image))
}

var (
	GlobalURLPrefix     string
	MainScreenScaleFunc = func() float64 {
		return 1
	}
)

func (i *Image) ForPixels(got func(x, y int, color zgeo.Color)) {
	gi := i.ToGo()
	b := gi.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := gi.At(x, y)
			got(x, y, zgeo.ColorFromGo(c))
		}
	}
}

func (i *Image) SetCapInsetsCorner(c zgeo.Size) *Image {
	r := zgeo.RectFromMinMax(c.Pos(), c.Pos().Negative())
	return i.SetCapInsets(r)
}

func MakeImagePathWithAddedScale(spath string, scale int) string {
	dir, _, stub, ext := zfile.Split(spath)
	size := fmt.Sprintf("@%dx", scale)
	zlog.Assert(!strings.HasSuffix(stub, "@2x"))
	return path.Join(dir, stub+size+ext)
}

func FromPathAddScreenScaleSuffix(spath string, useCache bool, got func(*Image)) {
	dir, _, stub, ext := zfile.Split(spath)
	size := ""
	zlog.Assert(!strings.HasSuffix(stub, "@2x"))
	if MainScreenScaleFunc() >= 2 {
		size = "@2x"
	}
	spath = path.Join(dir, stub+size+ext)
	zlog.Info("FromPathAddScreenScaleSuffix:", spath, MainScreenScaleFunc())
	FromPath(spath, useCache, got)
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
	return zgeo.SizeD(float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))
}

// GGoImageShrunkInto scales down the image to fit inside size.
// It must be a subset of standard libarary image types, as it uses rez
// package to downsample, which works on underlying image types.
func GoImageShrunkInto(goImage image.Image, size zgeo.Size, proportional bool) (image.Image, error) {
	// zlog.Info("GoImageShrunkInto:", goImage != nil, size)
	var vsize = size
	s := GoImageZSize(goImage)
	if proportional {
		vsize = zgeo.Rect{Size: size}.Align(s, zgeo.Center|zgeo.Shrink|zgeo.Proportional, zgeo.SizeNull).Size
	}
	//	this didn't work for large image?
	width := int(vsize.W)
	height := int(vsize.H)
	var newImage image.Image
	// zlog.Info("** Shrink:", reflect.ValueOf(goImage).Type(), reflect.ValueOf(goImage).Kind(), zlog.GetCallingStackString())
	if s.Max() > 1000 {
		//		nrgba := NRGBAImage()
		goRect := zgeo.Rect{Size: vsize}.GoRect()
		newImage = image.NewRGBA(goRect)
		biLin := rez.NewBilinearFilter() // NewBicubicFilter
		err := rez.Convert(newImage, goImage, biLin)
		if err != nil {
			goImage = GoImageToGoRGBA(goImage)
			err = rez.Convert(newImage, goImage, biLin)
		}
		if err != nil {
			return nil, zlog.Error(err, "rez resize")
		}
	} else {
		newImage = imaging.Resize(goImage, width, height, imaging.Lanczos)
	}
	// zlog.Info("GoImageShrunkInto2:", newImage != nil)
	return newImage, nil
}

func GoImageFromFile(path string) (img image.Image, format string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	return image.Decode(file)
}

func GoImageFromURL(path string) (img image.Image, format string, err error) {
	params := zhttp.MakeParameters()
	params.Method = http.MethodGet
	resp, err := zhttp.GetResponse(path, params)
	if err != nil {
		zlog.Error(err, path)
		return nil, "", err
	}
	return image.Decode(resp.Body)
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

func GoImageToPNGFile(img image.Image, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return zlog.Error(err, zlog.StackAdjust(1), "os.create", filepath)
	}
	defer out.Close()
	err = png.Encode(out, img)
	if err != nil {
		return zlog.Error(err, "encode", filepath)
	}
	return nil
}

func GoImageToGoRGBA(i image.Image) image.Image {
	r := i.Bounds()
	n := image.NewRGBA(r)
	draw.Draw(n, r, i, image.Point{}, draw.Over)
	return n
}

func (i *Image) ShrunkInto(size zgeo.Size, proportional bool, got func(*Image)) {
	// this can be better, use canvas.Image()
	goImage := i.ToGo()
	if goImage == nil {
		zlog.Error("ToGo")
		got(nil)
	}
	newGoImage, err := GoImageShrunkInto(goImage, size, proportional)
	if err != nil {
		return
	}
	FromGo(newGoImage, func(img *Image) {
		if img != nil {
			img.Path = i.Path + fmt.Sprint("|", size)
		}
		got(img)
	})
}

func GoImageShrunkCroppedToFillSize(img image.Image, size zgeo.Size, proportional bool) (image.Image, error) {
	is := GoImageZSize(img)
	ns := zgeo.Rect{Size: size}.Align(is, zgeo.Shrink|zgeo.Center|zgeo.Out|zgeo.Proportional, zgeo.SizeNull).Size
	ni, err := GoImageShrunkInto(img, ns, true)
	if err != nil {
		return nil, err
	}
	r := zgeo.Rect{Size: ns}.Align(size, zgeo.Center|zgeo.Proportional, zgeo.SizeNull)
	niCropped, err := GoImageCropped(ni, r, false)
	return niCropped, err
}

func GoImagesAreIdentical(img1, img2 image.Image) bool {
	b := img1.Bounds()
	if b != img2.Bounds() {
		return false
	}
	xlen := b.Dx()
	ylen := b.Dy()
	// xparts := zmath.LengthIntoDividePoints(xlen)
	// yparts := zmath.LengthIntoDividePoints(ylen)
	count := 0
	for x := 0; x < xlen; x++ {
		yend := ylen
		if x != xlen-1 {
			yend = zint.Min(x+1, ylen)
		}
		zlog.Info("c:", x, yend)
		for y := 0; y < yend; y++ {
			count++
		}
	}
	zlog.Info("done", xlen, ylen, xlen*ylen, count)
	return true
}

func GoImagesChangeDifference(img1, img2 image.Image, change func(imgSet SetableImage, amount float32, c1, c2 color.Color, x, y int)) (diffImage SetableImage, changedAmount float32) {
	w := img1.Bounds().Dx()
	h := img2.Bounds().Dy()
	if w != img2.Bounds().Dx() || h != img2.Bounds().Dy() {
		return nil, 1
	}
	di := image.NewNRGBA(image.Rectangle{Max: image.Pt(w, h)})
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c1 := img1.At(x, y)
			c2 := img1.At(x, y)
			change(di, -1, c1, c2, x, y) // this is for initing dest image
		}
	}
	var count int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 == c2 {
				change(di, 0, c1, c2, x, y)
				continue
			}
			d := zgeo.ColorFromGo(c1).Difference(zgeo.ColorFromGo(c2))
			if d < 0.02 {
				change(di, 0, c1, c2, x, y)
				continue
			}
			count++
			change(di, d, c1, c2, x, y)
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c1 := img1.At(x, y)
			c2 := img1.At(x, y)
			change(di, -2, c1, c2, x, y) // this is for initing dest image
		}
	}
	if count == 0 {
		return nil, 0
	}
	return di, float32(count) / float32(w*h)
}

func DrawCircle(img SetableImage, circle zgeo.Circle, col zgeo.Color) {
	// gcol := col.GoColor()
	alpha := col.Colors.A
	col.Colors.A = 1
	x1 := int(math.Max(0, circle.Center.X-circle.Radius))
	dx := float64(img.Bounds().Dx())
	dy := float64(img.Bounds().Dy())
	x2 := int(math.Ceil(math.Min(dx-1, circle.Center.X+circle.Radius)))
	y1 := int(math.Max(0, circle.Center.Y-circle.Radius))
	y2 := int(math.Ceil(math.Min(dy-1, circle.Center.Y+circle.Radius)))
	iradius := math.Floor(circle.Radius)
	// zlog.Info("circle:", x1, y1, x2, y2)
	for j := y1; j <= y2; j++ {
		var p zgeo.Pos
		p.Y = float64(j) - circle.Center.Y
		for i := x1; i <= x2; i++ {
			a := alpha
			c := col
			p.X = float64(i) - circle.Center.X
			len := p.Length()
			ilen := math.Floor(len)
			fract := float32(len - ilen)
			if ilen == iradius {
				a *= (1 - fract)
			}
			if ilen <= iradius {
				if a != 1 {
					old := zgeo.ColorFromGo(img.At(i, j))
					c = c.Mixed(old, 1-a)
				}
				img.Set(i, j, c.GoColor())
			}
		}
	}
}

func IsImageExtensionInName(surl string) bool {
	str := zstr.HeadUntil(surl, "?")
	for _, ext := range []string{"png", "jpeg", "jpg"} {
		if strings.HasSuffix(str, "."+ext) {
			return true
		}
	}
	return false
}

func CloneGoImage(src image.Image) draw.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}

func (i *Image) TintedWithColor(color zgeo.Color, amount float32, got func(i *Image)) {
	gi := i.ToGo()
	amount = float32(zfloat.Clamped(float64(amount), 0, 1))
	out := image.NewRGBA(gi.Bounds())
	i.ForPixels(func(x, y int, c zgeo.Color) {
		n := c.MixedColor(color, amount).GoColor()
		out.Set(x, y, n)
	})
	FromGo(out, got)
}

type ImageGetter struct {
	Path       string
	Image      *Image
	Tint       zgeo.Color
	TintAmount float32
	MaxSize    zgeo.Size

	Alignment zgeo.Alignment
	Margin    zgeo.Size
	Opacity   float32
}

func ImageGetterNew() *ImageGetter {
	return &ImageGetter{TintAmount: 1, Opacity: 1, Alignment: zgeo.Center}
}

func tryChangeTint(ig *ImageGetter, wg *sync.WaitGroup) {
	if ig.Tint.Valid {
		ig.Image.TintedWithColor(ig.Tint, ig.TintAmount, func(ti *Image) {
			ig.Image = ti
			wg.Done()
		})
	} else {
		wg.Done()
	}
}

// GetImages goes thu the ImageGetter's scaling/tinting images, returning when all are done. Release() is NOT called on input or output images
func GetImages(images []*ImageGetter, useCache bool, got func(all bool)) {
	var wg sync.WaitGroup
	var count int
	for _, ig := range images {
		wg.Add(1)
		FromPath(ig.Path, useCache, func(img *Image) {
			if img != nil {
				count++
				ig.Image = img
				if !ig.MaxSize.IsNull() {
					img.ShrunkInto(ig.MaxSize, true, func(shrunk *Image) {
						tryChangeTint(ig, &wg)
					})
				}
				tryChangeTint(ig, &wg)
			} else {
				wg.Done()
			}
		})
	}
	wg.Wait()
	got(count == len(images))
}

type SolidImage struct {
	Size  zgeo.Size
	color color.Color
}

func (c SolidImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (c SolidImage) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: int(c.Size.W), Y: int(c.Size.H)}}
}

func (c SolidImage) At(x, y int) color.Color {
	return c.color
}

func MakeSolidImage(size zgeo.Size, col zgeo.Color) SolidImage {
	return SolidImage{Size: size, color: col.GoColor()}
}

func GoImageBlurred(img image.Image, sigma float64) *image.NRGBA {
	return imaging.Blur(img, math.Abs(sigma))
}

func NewGoWithNoise(img image.Image, noiseCoverage float32, noiseMix float32) *image.NRGBA {
	return GoTransformed(img, func(x, y int, c zgeo.Color) zgeo.Color {
		if rand.Float32() >= noiseCoverage {
			return zgeo.Color{}
		}
		randCol := zgeo.ColorRandom()
		return c.Mixed(randCol, noiseMix)
	})
}

func GoForPixels(img image.Image, got func(x, y int, color zgeo.Color)) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.At(x, y)
			got(x, y, zgeo.ColorFromGo(c))
		}
	}
}

// GoAlterPixels calls set for the pixel at set's x, y, and sets it if the color returned is Valid
func GoAlterPixels(img SetableImage, set func(x, y int, c zgeo.Color) zgeo.Color) {
	GoForPixels(img, func(x, y int, color zgeo.Color) {
		setCol := set(x, y, color)
		if setCol.Valid {
			img.Set(x, y, setCol)
		}
	})
}

// GoTransformed creates a new image where if set returns a valid color, it uses that, otherwise old.
func GoTransformed(img image.Image, set func(x, y int, c zgeo.Color) zgeo.Color) *image.NRGBA {
	out := image.NewNRGBA(img.Bounds())
	GoForPixels(img, func(x, y int, c zgeo.Color) {
		ncol := set(x, y, c)
		if ncol.Valid {
			c = ncol
		}
		out.Set(x, y, c.GoColor())
	})
	return out
}

func GoImageFlippedHorizontal(img image.Image) *image.NRGBA {
	return imaging.FlipH(img)
}

func GoImageFlippedVertical(img image.Image) *image.NRGBA {
	return imaging.FlipV(img)
}

func GoImageCropped(img image.Image, crop zgeo.Rect, copy bool) (image.Image, error) {
	r := image.Rect(int(crop.Min().X), int(crop.Min().Y), int(crop.Max().X), int(crop.Max().Y))
	ni := imaging.Crop(img, r)
	return ni, nil
}
