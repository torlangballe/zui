package zimage

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

var (
	remoteCache      = zcache.New(3600, false)
	localCache       = zcache.New(0, false) // cache for with-app images, no expiry
	DrawInCanvasFunc func(size zgeo.Size, draw func(e js.Value)) image.Image
)

type imageBase struct {
	size      zgeo.Size `json:"size"`
	capInsets zgeo.Rect `json:"capInsets"`
	hasAlpha  bool      `json:"hasAlpha"`
	ImageJS   js.Value
}

func GetSynchronous(timeoutSecs float64, imagePaths ...interface{}) bool {
	added := make(chan struct{}, 100)
	for i := 0; i < len(imagePaths); i++ {
		imgPtr := imagePaths[i].(**Image)
		i++
		path := imagePaths[i].(string)
		FromPath(path, func(image *Image) {
			*imgPtr = image
			added <- struct{}{}
			// zlog.Info("GetSynchronous got", path, image != nil)
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
			zlog.Info("GetSynchronous bail:", count)
			return false
		}
	}
	zlog.Fatal(nil, "GetSynchronous can't get here")
	return false
}

func FromPath(path string, got func(*Image)) {
	// if !strings.HasSuffix(path, ".png") {
	// zlog.Info("FromPath:", path, zlog.GetCallingStackString())
	// }
	if path == "" {
		if got != nil {
			got(nil)
		}
		return
	}
	// zlog.Info("FromPath:", GlobalURLPrefix, "#", path)
	if !strings.HasPrefix(path, "data:") && !zhttp.StringStartsWithHTTPX(path) {
		path = GlobalURLPrefix + path
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
		return
	}
	// zlog.Info("FromPath before load:", path)
	i.load(path, func(success bool) {
		// zlog.Info("FromPath loaded:", success, got != nil, path)
		if !success {
			i = nil
		}
		cache.Put(path, i)
		if got != nil {
			got(i)
		}
	})
}

func (i *Image) ToGo() image.Image {
	s := i.Size()
	return DrawInCanvasFunc(s, func(canvasContext js.Value) {
		canvasContext.Call("drawImage", i.ImageJS, 0, 0, s.W, s.H)
	})
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
	i.Loading = true
	i.Loading = strings.HasPrefix(path, "images/")
	i.Scale = imageGetScaleFromPath(path)

	imageF := js.Global().Get("Image")
	i.ImageJS = imageF.New()
	i.ImageJS.Set("crossOrigin", "Anonymous")

	// i.ImageJS.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	i.ImageJS.Call("addEventListener", "load", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		i.Loading = false
		i.size.W = i.ImageJS.Get("width").Float()
		i.size.H = i.ImageJS.Get("height").Float()
		// zlog.Info("Image Loaded", this, args, len(args))
		if done != nil {
			done(true)
		}
		return nil
	}))
	i.ImageJS.Set("onerror", js.FuncOf(func(js.Value, []js.Value) interface{} {
		i.Loading = false
		i.size.W = 5
		i.size.H = 5
		zlog.Info("Image Load fail:", path)
		if done != nil {
			done(false)
		}
		return nil
	}))
	i.ImageJS.Set("src", path)
}

func (i *Image) Colored(color zgeo.Color, size zgeo.Size) *Image {
	//	rect := NewRect(0, 0, size.W, size.H)
	return i
}

func (i *Image) Size() zgeo.Size {
	return i.size.DividedByD(float64(i.Scale))
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

func FromGo(img image.Image, got func(image *Image)) {
	data, err := GoImagePNGData(img)
	if err != nil {
		zlog.Error(err)
		got(nil)
	}
	surl := zhttp.MakeDataURL(data, "image/png")
	FromPath(surl, got)
}

func (i *Image) IsLoaded() bool {
	return i.ImageJS.Get("complete").Bool() && i.ImageJS.Get("naturalHeight").Float() > 0
}
