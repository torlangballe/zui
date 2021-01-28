package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type baseVideoView struct {
	videoElement js.Value
}

func (v *VideoView) Init(view View, minSize zgeo.Size) {
	v.Element = DocumentJS.Call("createElement", "div")
	v.setjs("className", "camera")
	v.Element.Set("style", "position:absolute")
	v.View = view
	v.SetObjectName("camera")
	v.minSize = minSize

	v.videoElement = DocumentJS.Call("createElement", "video")
	v.videoElement.Set("id", "video")
}

// CreateStream creates a stream, perhaps withAudio, and hooks it up with video.
// Needs to be called in a goroutine.
func (v *VideoView) CreateStream(withAudio bool) {
	mediaDevs := jsCreateDotSeparatedObject("navigator.mediaDevices")
	constraints := map[string]interface{}{
		"video": true,
		"audio": withAudio,
	}
	stream := mediaDevs.Call("getUserMedia", constraints)
	v.videoElement.Set("srcObject", stream)
	v.videoElement.Set("canplay", js.FuncOf(func(js.Value, []js.Value) interface{} {
		zlog.Info("can play")
		return nil
	}))

}

func (v *VideoView) Capture() {
	canvas := CanvasNew()
	canvas.SetSize(zgeo.Size{500, 100})
}
