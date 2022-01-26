package zui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bamiaux/rez"
	"github.com/disintegration/imaging"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
	"github.com/torlangballe/zutil/zstr"
)

//  Created by Tor Langballe on /20/10/15.

type Image struct {
	imageBase
	scale   int
	Path    string
	loading bool
}

type SetableImage interface {
	image.Image
	Set(x, y int, c color.Color)
}

type ImageLoader interface {
	IsLoading() bool
}

type ImageOwner interface {
	GetImage() *Image
	SetImage(image *Image, path string, got func(*Image))
}

var ImageGlobalURLPrefix string

func (i *Image) ForPixels(got func(pos zgeo.Pos, color zgeo.Color)) {
}

func (i *Image) SetCapInsetsCorner(c zgeo.Size) *Image {
	r := zgeo.RectFromMinMax(c.Pos(), c.Pos().Negative())
	return i.SetCapInsets(r)
}

func ImagePathAddedScale(spath string, scale int) string {
	dir, _, stub, ext := zfile.Split(spath)
	size := fmt.Sprintf("@%dx", scale)
	zlog.Assert(!strings.HasSuffix(stub, "@2x"))
	return path.Join(dir, stub+size+ext)
}

func ImageFromPathAddScale(spath string, got func(*Image)) {
	dir, _, stub, ext := zfile.Split(spath)
	size := ""
	zlog.Assert(!strings.HasSuffix(stub, "@2x"))
	if zscreen.MainScale >= 2 {
		size = "@2x"
	}
	spath = path.Join(dir, stub+size+ext)
	zlog.Info("ImageFromPathAddScale:", spath, zscreen.MainScale)
	ImageFromPath(spath, got)
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
func GoImageShrunkInto(goImage image.Image, size zgeo.Size, proportional bool) image.Image {
	var vsize = size
	s := GoImageZSize(goImage)
	if proportional {
		vsize = zgeo.Rect{Size: size}.Align(s, zgeo.Center|zgeo.Shrink|zgeo.Proportional, zgeo.Size{}).Size
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

func GoImageFromFile(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	return image.Decode(file)
}

func GoImageFromURL(path string) (image.Image, error) {
	params := zhttp.MakeParameters()
	params.Method = http.MethodGet
	params.Headers["Origin"] = "https://192.168.0.30:443"
	resp, err := zhttp.GetResponse(path, params)
	if err != nil {
		return nil, err
	}
	goImage, _, err := image.Decode(resp.Body)
	return goImage, err
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
		return zlog.Error(err, "encode")
	}
	return nil
}

func (i *Image) Merge(myMaxSize zgeo.Size, with *Image, align zgeo.Alignment, marg, withMaxSize zgeo.Size, done func(img *Image)) {
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
	downsampleCache := false
	canvas.DrawImage(i, downsampleCache, mr, 1, zgeo.Rect{})
	canvas.DrawImage(with, downsampleCache, wr, 1, zgeo.Rect{})
	canvas.ZImage(false, done)
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
		zlog.Error(nil, "GoImageFromPath")
		got(nil)
	}
	newGoImage := GoImageShrunkInto(goImage, size, proportional)
	ImageFromGo(newGoImage, func(img *Image) {
		if img != nil {
			img.Path = i.Path + fmt.Sprint("|", size)
		}
		got(img)
	})
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

func ImageExtensionInName(surl string) bool {
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
