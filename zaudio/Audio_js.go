package zaudio

import (
	"syscall/js"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zlog"
)

// https://developer.mozilla.org/en-US/docs/Web/API/MediaRecorder/MediaRecorder
type audioNative struct {
	audio js.Value
}

func AudioNew(path string) *Audio {
	a := &Audio{}
	audioF := js.Global().Get("Audio")
	a.audio = audioF.New(path)
	return a
}

func (a *Audio) Play(fail func(err error)) {
	promise := a.audio.Call("play")
	zdom.Resolve(promise, func(resolved js.Value, err error) {
		if err != nil {
			zlog.Error(err, "play")
			if fail != nil {
				fail(err)
			}
		}
	})
}

func (a *Audio) Stop() {
	a.audio.Call("pause")
	a.audio.Set("currentTime", 0)
}

func (a *Audio) SetVolume(v float32) {
	a.audio.Set("volume", v)
}

func (a *Audio) SetHandleFinished(f func()) {
	a.audio.Set("onended", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if f != nil {
			f()
		}
		return nil
	}))
}

func getMediaDevices() js.Value {
	return zdom.CreateDotSeparatedObject("navigator.mediaDevices")
}
