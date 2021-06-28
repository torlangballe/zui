package zui

import (
	"image"
	"strings"
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/zcache"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztime"
)

var remoteCache = zcache.New(3600, false)
var localCache = zcache.New(0, false) // cache for with-app images, no expiry

type imageBase struct {
	size      zgeo.Size `json:"size"`
	capInsets zgeo.Rect `json:"capInsets"`
	hasAlpha  bool      `json:"hasAlpha"`
	imageJS   js.Value
}

func ImagesGetSynchronous(timeoutSecs float64, imagePaths ...interface{}) bool {
	added := make(chan struct{}, 100)
	for i := 0; i < len(imagePaths); i++ {
		imgPtr := imagePaths[i].(**Image)
		i++
		path := imagePaths[i].(string)
		ImageFromPath(path, func(image *Image) {
			*imgPtr = image
			added <- struct{}{}
			// zlog.Info("ImagesGetSynchronous got", path, image != nil)
		})
	}
	var count int
	for {
		select {
		case <-added:
			count++
			if count >= len(imagePaths)/2 {
				return true
			}
		case <-time.After(ztime.SecondsDur(timeoutSecs)):
			zlog.Info("ImagesGetSynchronous bail:", count)
			return false
		}
	}
	zlog.Fatal(nil, "ImagesGetSynchronous can't get here")
	return false
}

func ImageFromPath(path string, got func(*Image)) *Image {
	// if !strings.HasSuffix(path, ".png") {
	// 	zlog.Info("ImageFromPath:", path, zlog.GetCallingStackString())
	// }
	if path == "" {
		if got != nil {
			got(nil)
		}
		return nil
	}
	cache := remoteCache
	if strings.HasPrefix(path, "images/") {
		cache = localCache
	}
	i := &Image{}
	if cache.Get(&i, path) {
		if got != nil {
			got(i)
		}
		return i
	}
	// zlog.Info("ImageFromPath before load:", path)
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

func (i *Image) ToGo() image.Image {
	canvas := CanvasNew()
	canvas.element.Set("id", "render-canvas")
	s := i.Size()
	canvas.SetSize(s)
	canvas.context.Call("drawImage", i.imageJS, 0, 0, s.W, s.H)
	goImage := canvas.GoImage(zgeo.Rect{})
	return goImage
}

func (i *Image) RGBAImage() *Image {
	zlog.Fatal(nil, "Not implemented")
	return nil
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
	// i.imageJS.Set("crossOrigin", "Anonymous")

	// i.imageJS.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	i.imageJS.Call("addEventListener", "load", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		i.loading = false
		i.size.W = i.imageJS.Get("width").Float()
		i.size.H = i.imageJS.Get("height").Float()
		// zlog.Info("Image Loaded", this, args, len(args))
		if done != nil {
			done(true)
		}
		return nil
	}))
	i.imageJS.Set("onerror", js.FuncOf(func(js.Value, []js.Value) interface{} {
		i.loading = false
		i.size.W = 5
		i.size.H = 5
		zlog.Info("Image Load fail:", path)
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

func CanvasFromGoImage(i image.Image) *Canvas {
	canvas := CanvasNew()
	canvas.SetSize(GoImageZSize(i))
	canvas.SetGoImage(i, zgeo.Pos{0, 0})
	return canvas
}

func ImageFromGo(img image.Image, got func(image *Image)) {
	data, err := GoImagePNGData(img)
	if err != nil {
		zlog.Error(err)
		got(nil)
	}
	surl := zhttp.MakeDataURL(data, "image/png")
	ImageFromPath(surl, got)
}
