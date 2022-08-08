package zaudio

import (
	"syscall/js"
)

type audioNative struct {
	audio js.Value
}

func AudioNew(path string) *Audio {
	a := &Audio{}
	audioF := js.Global().Get("Audio")
	a.audio = audioF.New(path)
	return a
}

func (a *Audio) Play() {
	a.audio.Call("play")
}

func (a *Audio) Stop() {
	// a.audio.Call("stop")
}

func (a *Audio) SetVolume(v float32) {
	a.audio.Set("volume", v)
}
