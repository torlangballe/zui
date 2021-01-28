package zui

import (
	"image"
	"io"
	"strings"
	"sync"
	"syscall/js"

	"github.com/torlangballe/zutil/zcache"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
)

var cache = zcache.New(3600)

type imageBase struct {
	size      zgeo.Size `json:"size"`
	capInsets zgeo.Rect `json:"capInsets"`
	hasAlpha  bool      `json:"hasAlpha"`
	imageJS   js.Value
}

func ImageFromPath(path string, got func(*Image)) *Image {
	// zlog.Info("ImageFromPath:", path)
	if path == "" {
		if got != nil {
			got(nil)
		}
		return nil
	}
	i := &Image{}
	if cache.GetTo(&i, path) {
		if got != nil {
			got(i)
		}
		return i
	}
	i.load(path, func(success bool) {
		// zlog.Info("ImageFromPath loaded:", path, success)
		if !success {
			i = nil
		}
		cache.Put(path, i)
		if got != nil {
			got(i)
		}
	})
	return i
}

func ImageFromNative(n image.Image) *Image {
	zlog.Fatal(nil, "Not implemented")
	return nil
}

func (i *Image) ShrunkInto(size zgeo.Size, proportional bool) *Image {
	isFile := false
	goImage := goImageFromPath(i.Path, isFile)
	if goImage == nil {
		zlog.Error(nil, "goImageFromPath")
		return nil
	}
	scale := float64(i.scale)
	newGoImage := goImageShrunkInto(goImage, scale, size, proportional)
	data, err := goImagePNGData(newGoImage)
	if err != nil {
		zlog.Error(err)
		return nil
	}
	surl := zhttp.MakeDataURL(data, "image/png")
	//	zlog.Info("image url:\n", surl)
	var wg sync.WaitGroup
	var newImage *Image
	wg.Add(1)
	ImageFromPath(surl, func(image *Image) {
		newImage = image
		wg.Done()
	})
	wg.Wait()
	return newImage
}

func (i *Image) Encode(w io.Writer, qualityPercent int) error {
	zlog.Fatal(nil, "Not implemented")
	return nil
}

func (i *Image) RGBAImage() *Image {
	zlog.Fatal(nil, "Not implemented")
	return nil
}

func (i *Image) JPEGData(qualityPercent int) ([]byte, error) {
	zlog.Fatal(nil, "Not implemented")
	return nil, nil
}

func (i *Image) load(path string, done func(success bool)) {
	// if !strings.HasPrefix(path, "http:") && !strings.HasPrefix(path, "https:") {
	// 	if !strings.HasPrefix(path, "images/") {
	// 		path = "images/" + path
	// 	}
	// }
	// zlog.Info("Image Load:", path)

	i.Path = path
	i.loading = true
	i.loading = strings.HasPrefix(path, "images/")
	i.scale = imageGetScaleFromPath(path)

	imageF := js.Global().Get("Image")
	i.imageJS = imageF.New()
	i.imageJS.Set("onload", js.FuncOf(func(js.Value, []js.Value) interface{} {
		i.loading = false
		i.size.W = i.imageJS.Get("width").Float()
		i.size.H = i.imageJS.Get("height").Float()
		// zlog.Info("Image Load scale:", i.scale, path, i.Size(), i.loading)
		if done != nil {
			done(true)
		}
		return nil
	}))
	i.imageJS.Set("onerror", js.FuncOf(func(js.Value, []js.Value) interface{} {
		i.loading = false
		i.size.W = 5
		i.size.H = 5
		// zlog.Info("Image Load fail:", path)
		if done != nil {
			done(false)
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

func (i *Image) SetCapInsets(capInsets zgeo.Rect) *Image {
	i.capInsets = capInsets
	return i
}

func (i *Image) CapInsets() zgeo.Rect {
	return i.capInsets
}

func (i *Image) HasAlpha() bool {
	return i.hasAlpha
}

func (i *Image) TintedWithColor(color zgeo.Color) *Image {
	return i
}

// See
// func (i *Image) GetScaledInSize(size zgeo.Size, proportional bool) *Image {
// 	var vsize = size
// 	if proportional {
// 		vsize = zgeo.Rect{Size: size}.Align(i.Size(), zgeo.Center|zgeo.Shrink|zgeo.ScaleToFitProp, zgeo.Size{0, 0}, zgeo.Size{0, 0}).Size
// 	}
// 	width := int(vsize.W) / int(i.scale)
// 	height := int(vsize.H) / int(i.scale)
// 	zlog.Info("GetScaledInSize not made yet:", width, height)
// 	return nil
// }

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
	zlog.Info("Image.Rotated not made yet:", transform)
	return i
}

func (i *Image) FixedOrientation() *Image {
	return i
}
