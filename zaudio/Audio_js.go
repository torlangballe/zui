package zaudio

import (
	"io"
	"net/http"
	"syscall/js"

	"github.com/torlangballe/zui/zalert"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
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

func (a *Audio) Play() {
	promise := a.audio.Call("play")
	zdom.Resolve(promise, func(resolved js.Value, err error) {
		if err != nil {
			zalert.ShowError(err)
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

type Recording struct {
	mediaRecorder js.Value
	stopped       bool
}

type Options struct {
	BitsPerSecond int
	TimeSliceMS   int
	MimeFormat    string
}

func (r *Recording) Stop() {
	// zlog.Info("STOP!!")
	r.mediaRecorder.Call("stop")
}

func NewAudioRecording(opts Options, w io.Writer, started func(r *Recording), finished func()) {
	r := &Recording{}
	mediaDevs := getMediaDevices()
	constraints := map[string]any{
		"audio": true,
	}
	mime := opts.MimeFormat
	if mime == "" {
		mime = "audio/mp4"
	}
	streamCall := mediaDevs.Call("getUserMedia", constraints)
	zdom.Resolve(streamCall, func(stream js.Value, err error) {
		if err != nil {
			zlog.Error(err, "getUserMedia")
			return
		}
		options := map[string]any{
			"audioBitsPerSecond": opts.BitsPerSecond,
			"mimeType":           mime,
		}
		r.mediaRecorder = js.Global().Get("MediaRecorder").New(stream, options)
		// r.audioChunks = js.ValueOf([]any{})
		r.mediaRecorder.Call("addEventListener", "dataavailable", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			event := args[0]
			// r.audioChunks.Call("push", event.Get("data"))
			// audioBlob := js.Global().Get("Blob").New(r.audioChunks)
			// zlog.Assert(r.GotAudio != nil)
			// zview.JSFileToGo(audioBlob, func(data []byte, name string) {
			// 	r.GotAudio(data)
			// }, nil)
			audioBlob := event.Get("data")
			// if audioBlob.Get("size").Int() == 0 {
			// 	return nil
			// }
			// zlog.Info("DataAvailable")
			zdom.JSFileToGo(audioBlob, func(data []byte, name string) {
				if len(data) != 0 {
					// zlog.Info("Bytes:", len(data), name)
					w.Write(data)
				}
				if r.stopped {
					c, _ := w.(io.Closer)
					if c != nil {
						c.Close()
					}
					if finished != nil {
						finished()
					}
				}
				// zlog.Info("DataAvailable end")
			}, nil)
			return nil
		}))
		r.mediaRecorder.Set("onstop", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			r.stopped = true // need to set this here, as we can get a dataavailable after stop
			// zlog.Info("Event Stopped")
			return nil
		}))
		r.mediaRecorder.Call("start", opts.TimeSliceMS)
		started(r)
	})
}

func PostAudioRecording(opts Options, surl, userTokenForHeader string, started func(r *Recording), finished func()) {
	reader, writer := io.Pipe()
	go func() {
		params := zhttp.MakeParameters()
		params.Reader = reader
		params.Method = http.MethodPost
		if opts.MimeFormat == "" {
			opts.MimeFormat = "audio/mp4"
		}
		params.ContentType = opts.MimeFormat
		if userTokenForHeader != "" {
			params.Headers[zrest.UserAuthTokenHeaderKey] = userTokenForHeader
		}
		_, err := zhttp.GetResponse(surl, params)
		zlog.OnError(err, "post")
	}()
	NewAudioRecording(opts, writer, started, finished)
}
