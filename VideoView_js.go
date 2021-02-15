package zui

import (
	"context"
	"fmt"
	"syscall/js"
	"time"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type baseVideoView struct {
}

func (v *VideoView) Init(view View, minSize zgeo.Size) {
	v.MakeJSElement(v, "video")
	v.SetObjectName("video-input")
	v.Element.Set("autoplay", true)
	v.minSize = minSize
}

func VideoViewGetInputDevices(got func(devs map[string]string)) {
	mediaDevs := getMediaDevices()
	mediaDevs.Call("enumerateDevices").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		devs := map[string]string{}
		enumerated := args[0]
		elen := enumerated.Length()
		for i := 0; i < elen; i++ {
			e := enumerated.Index(i)
			if e.Get("kind").String() == "videoinput" {
				id := e.Get("deviceId").String()
				name := e.Get("label").String()
				devs[id] = name
				zlog.Info("ids:", id, name, devs)
			}
		}
		got(devs)
		return nil
	}))
	return
}

func (v *VideoView) keepGettingImage(ctx context.Context, canvas *Canvas, continuousImageHandler func(*Image) bool) {
	zlog.Info("keepGettingImag:", ctx.Err())

	for ctx.Err() == nil {
		time.Sleep(time.Second)

		s := v.Rect().Size
		canvas.context.Call("drawImage", v.Element, 0, 0, s.W, s.H)
		zlog.Info("draw image")
	}
}

// CreateStream creates a stream, perhaps withAudio, and hooks it up with video.
// Needs to be called in a goroutine.v.Element
func (v *VideoView) CreateStream(withAudio bool, continuousImageHandler func(*Image) bool) {
	// https://www.twilio.com/blog/2018/04/choosing-cameras-javascript-mediadevices-api.html
	mediaDevs := getMediaDevices()
	constraints := map[string]interface{}{
		"video": true,
		"audio": withAudio,
		//     facingMode: "user" -- handy
	}
	stream := mediaDevs.Call("getUserMedia", constraints)
	stream.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		stream := args[0]
		v.Element.Set("srcObject", stream)
		fmt.Println("Got Stream33")
		//		v.Element.Call("play")
		zlog.Info("HERE1!")
		v.Element.Set("canplay", js.FuncOf(func(js.Value, []js.Value) interface{} {
			zlog.Info("can play")
			return nil
		}))
		zlog.Info("HERE2", continuousImageHandler != nil)
		if continuousImageHandler != nil {
			ctx, cancel := context.WithCancel(context.Background())
			v.AddStopper(cancel)
			canvas := CanvasNew()
			canvas.SetSize(zgeo.Size{500, 100})
			zlog.Info("HERE!")
			go v.keepGettingImage(ctx, canvas, continuousImageHandler)
		}
		return nil
	}))
}

func (v *VideoView) Capture() {
	canvas := CanvasNew()
	canvas.SetSize(zgeo.Size{500, 100})
}

func getMediaDevices() js.Value {
	return jsCreateDotSeparatedObject("navigator.mediaDevices")
}
