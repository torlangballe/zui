package zaudio

import (
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zlog"
	"syscall/js"
)

type VoiceJS js.Value

var allVoices []Voice = nil

func (v VoiceJS) Name() string {
	return js.Value(v).Get("name").String()
}

func (v VoiceJS) Language() string {
	return js.Value(v).Get("lang").String()
}

func (v VoiceJS) Female() bool {
	return false
}

func getSynth() js.Value {
	return zdom.WindowJS.Get("speechSynthesis")
}

// if (speechSynthesis.onvoiceschanged !== undefined) {

func AllVoices() (vs []Voice) {
	voices := getSynth().Call("getVoices")
	if voices.IsUndefined() {
		return
	}

	vs = []Voice{}
	for i := 0; i < voices.Length(); i++ {
		v := VoiceJS(voices.Index(i))
		vs = append(vs, v)
	}
	return
}

func GetVoice(name string) (Voice, bool) {
	if allVoices == nil {
		allVoices = AllVoices()
	}
	for _, v := range allVoices {
		if v.Name() == name {
			return v, true
		}
	}
	return VoiceJS{}, false
}

func IsSpeaking() bool {
	return getSynth().Get("speaking").Bool()
}

// SpeakText speaks the string *text* using *voice*.
// pitch can be -1 to 1, where 0 is "normal".
// A rate of 1 is normal 0.1 slowest (10%) and 10 highest (1000%)
// Volume is 0-1, 1 is default/highest.
func SpeakText(text string, voice Voice, pitch, rate, volume float64) {
	if IsSpeaking() || text == "" {
		return
	}
	utterance := js.Global().Get("SpeechSynthesisUtterance").New(text)
	vjs, _ := voice.(VoiceJS)
	utterance.Set("voice", js.Value(vjs))
	utterance.Set("pitch", pitch+1)
	utterance.Set("rate", rate)
	utterance.Set("volume", volume)
	utterance.Set("onend", js.FuncOf(func(this js.Value, args []js.Value) any {
		zlog.Info("Speech end")
		return nil
	}))
	utterance.Set("onerror", js.FuncOf(func(this js.Value, args []js.Value) any {
		zlog.Info("Speech error")
		return nil
	}))
	getSynth().Call("speak", utterance)
}

// https://github.com/mdn/dom-examples/blob/main/web-speech-api/speak-easy-synthesis/script.js
