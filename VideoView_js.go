//go:build zui

package zui

import (
	"context"
	"image"
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/ztimer"
)

// https://www.twilio.com/blog/2018/04/choosing-cameras-javascript-mediadevices-api.html
// https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Taking_still_photos
// https://webrtchacks.com/guide-to-safari-webrtc/
// https://webkit.org/blog/6784/new-video-policies-for-ios/

type baseVideoView struct {
}

func (v *VideoView) Init(view View, maxSize zgeo.Size) {
	v.MakeJSElement(v, "video")
	v.SetObjectName("video-input")
	v.maxSize = maxSize
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
				// zlog.Info("ids:", id, name, devs)
			}
		}
		got(devs)
		return nil
	}))
	return
}

func (v *VideoView) getNextImage(continuousImageHandler func(image.Image) bool) {
	if v.renderCanvas == nil {
		v.makeRenderCanvas()
	}
	v.renderCanvas.context.Call("drawImage", v.Element, 0, 0, v.StreamSize.W, v.StreamSize.H)
	goImage := v.renderCanvas.GoImage(zgeo.Rect{})
	continuousImageHandler(goImage)
}

// CreateStream creates a stream, perhaps withAudio, and hooks it up with video.
// Needs to be called in a goroutine.v.Element
func (v *VideoView) CreateStream(withAudio, selfie bool, continuousImageHandler func(image.Image) bool) {
	mediaDevs := getMediaDevices()
	facing := "environment"
	if selfie {
		facing = "user"
	}
	constraints := map[string]interface{}{
		"video": map[string]interface{}{
			"facingMode": facing, // user, environment, left and right -- what is camera pointing at?
		}, "audio": withAudio,
	}
	v.Element.Set("autoplay", "true")
	v.Element.Set("muted", "true")
	v.Element.Set("playsInline", "true") // playsInline must be camel case!!!
	stream := mediaDevs.Call("getUserMedia", constraints)
	then := stream.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		stream := args[0]
		v.Element.Set("srcObject", stream)
		v.Element.Set("oncanplay", js.FuncOf(func(js.Value, []js.Value) interface{} {
			if !v.streaming {
				//				v.Element.Call("play") // do we need this???
				v.StreamSize.W = v.Element.Get("videoWidth").Float()
				v.StreamSize.H = v.Element.Get("videoHeight").Float()
				v.streaming = true
				if v.StreamingStarted != nil {
					v.StreamingStarted()
				}
			}
			return nil
		}))
		v.Element.Set("onloadedmetadata", js.FuncOf(func(js.Value, []js.Value) interface{} {
			// v.Element.Set("width", 360)
			// v.Element.Set("height", 640)
			//		zlog.Info("loaded meta data")
			return nil
		}))
		if continuousImageHandler != nil {
			_, cancel := context.WithCancel(context.Background())
			v.AddOnRemoveFunc(cancel)
			// zlog.Info("video get image3:", s)
			timer := ztimer.RepeatIn(0.1, func() bool {
				// if ctx.Err() != nil {
				// 	return false
				// }
				if v.streaming {
					v.getNextImage(continuousImageHandler)
				}
				return true
			})
			v.AddOnRemoveFunc(timer.Stop)
		}
		return nil
	}))
	then.Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		zlog.Info("CATCH!")
		return nil
	}))
}

func (v *VideoView) makeRenderCanvas() {
	v.renderCanvas = CanvasNew()
	v.renderCanvas.element.Set("id", "render-canvas")
	v.renderCanvas.SetSize(v.StreamSize)
	//			v.renderCanvas.context.Call("scale", scale, scale) // this must be AFTER setElementRect, doesn't do anything!
	v.Element.Call("appendChild", v.renderCanvas.element)
	v.renderCanvas.element.Get("style").Set("visible", "hidden")
	setElementRect(v.renderCanvas.element, zgeo.Rect{Size: v.StreamSize})
}

// func (v *VideoView) Capture() {
// 	canvas := CanvasNew()
// 	canvas.SetSize(zgeo.Size{500, 100})
// }

func getMediaDevices() js.Value {
	return jsCreateDotSeparatedObject("navigator.mediaDevices")
}
