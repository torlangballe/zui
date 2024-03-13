package zimage

import (
	"image"
	"path"
	"strings"
	"syscall/js"
	"time"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zcache"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztime"
)

var (
	remoteCache      = zcache.NewWithExpiry(10*60, false)
	localCache       = zcache.New() // cache for included-in-app images, no expiry.
	DrawInCanvasFunc func(size zgeo.Size, draw func(e js.Value)) image.Image
)

type imageBase struct {
	size      zgeo.Size `json:"size"`
	capInsets zgeo.Rect `json:"capInsets"`
	hasAlpha  bool      `json:"hasAlpha"`
	ImageJS   js.Value
}

func GetSynchronous(timeoutSecs float64, useCache bool, imagePaths ...interface{}) bool {
	added := make(chan struct{}, 100)
	for i := 0; i < len(imagePaths); i++ {
		imgPtr := imagePaths[i].(**Image)
		i++
		spath := imagePaths[i].(string)
		FromPath(spath, useCache, func(image *Image) {
			*imgPtr = image
			added <- struct{}{}
			// zlog.Info("GetSynchronous got", spath, image != nil)
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
	zlog.Fatal("GetSynchronous can't get here")
	return false
}

func FromPath(spath string, useCache bool, got func(*Image)) {
	var cache *zcache.Cache

	if useCache {
		cache = remoteCache
		if strings.HasPrefix(spath, "images/") {
			cache = localCache
		}
	}

	if spath == "" {
		if got != nil {
			got(nil)
		}
		return
	}
	// zlog.Info("FromPath:", GlobalURLPrefix, "#", spath)
	if !strings.HasPrefix(spath, "data:") && !zhttp.StringStartsWithHTTPX(spath) {
		spath = path.Join(GlobalURLPrefix, spath)
	}
	i := &Image{}
	if cache != nil {
		if cache.Get(&i, spath) {
			if got != nil {
				got(i)
			}
			return
		}
	}
	i.load(spath, func(success bool) {
		// zlog.Info("FromPath loaded:", success, got != nil, spath)
		if !success {
			i = nil
		}
		if cache != nil {
			cache.Put(spath, i)
		}
		if got != nil {
			got(i)
		}
	})
}

func (i *Image) RGBAImage() *Image {
	zlog.Fatal("Not implemented")
	return nil
}

func (i *Image) load(spath string, done func(success bool)) {
	if !strings.HasPrefix(spath, "data:") {
		i.Path = spath
	}
	i.Loading = true
	i.Loading = strings.HasPrefix(spath, "images/")
	i.Scale = imageGetScaleFromPath(spath)

	imageF := js.Global().Get("Image")
	i.ImageJS = imageF.New()
	i.ImageJS.Set("crossOrigin", "Anonymous")

	// i.ImageJS.Set("onload", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	i.ImageJS.Set("onerror", zdom.MakeSingleCallJSCallback(func(js.Value, []js.Value) interface{} {
		i.Loading = false
		i.size.W = 5
		i.size.H = 5
		// zlog.Info("Image Load fail:", spath)
		if done != nil {
			done(false)
		}
		return nil
	}))
	f := zdom.MakeSingleCallJSCallback(func(this js.Value, args []js.Value) any {
		i.Loading = false
		i.size.W = i.ImageJS.Get("width").Float()
		i.size.H = i.ImageJS.Get("height").Float()
		if done != nil {
			done(true)
		}
		return nil
	})
	i.ImageJS.Call("addEventListener", "load", f)
	i.ImageJS.Set("src", spath)
}

func (i *Image) Release() {
	i.ImageJS.Set("src", "")
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
	FromPath(surl, false, got)
}

func (i *Image) ToGo() image.Image {
	s := i.Size()
	return DrawInCanvasFunc(s, func(canvasContext js.Value) {
		canvasContext.Call("drawImage", i.ImageJS, 0, 0, s.W, s.H)
	})
}

func (i *Image) IsLoaded() bool {
	return i.ImageJS.Get("complete").Bool() && i.ImageJS.Get("naturalHeight").Float() > 0
}
