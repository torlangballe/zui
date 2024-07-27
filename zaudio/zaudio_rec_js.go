package zaudio

import (
	"io"
	"net/http"
	"syscall/js"
	"time"

	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zutil/zhttp"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zrest"
	"github.com/torlangballe/zutil/ztimer"
)

type Recording struct {
	mediaRecorder js.Value // shouldn't be in general struct
	stopped       bool
	progTimer     *ztimer.Repeater
}

type RecOptions struct {
	BitsPerSecond     int
	TimeSliceMS       int
	MimeFormat        string
	Progress          func(dur time.Duration)
	Started           func(r *Recording)
	FinishedRecording func()
	Posted            func(resultMessage string)
}

func (r *Recording) Stop() {
	// zlog.Info("STOP!!")
	r.mediaRecorder.Call("stop")
	if r.progTimer != nil {
		r.progTimer.Stop()
		r.progTimer = nil
	}
}

func NewAudioRecording(opts RecOptions, w io.Writer) *Recording {
	r := &Recording{}
	mediaDevs := getMediaDevices()
	constraints := map[string]any{
		"audio": true,
	}
	if opts.BitsPerSecond == 0 {
		opts.BitsPerSecond = 128000
	}
	mime := opts.MimeFormat
	if mime == "" {
		mime = "audio/mp4"
	}
	streamCall := mediaDevs.Call("getUserMedia", constraints)
	zdom.Resolve(streamCall, func(stream js.Value, err error) {
		if err != nil {
			zlog.Error("getUserMedia", err)
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
					if opts.FinishedRecording != nil {
						opts.FinishedRecording()
					}
					if r.progTimer != nil {
						r.progTimer.Stop()
						r.progTimer = nil
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
		if opts.TimeSliceMS == 0 {
			r.mediaRecorder.Call("start")
		} else {
			r.mediaRecorder.Call("start", opts.TimeSliceMS)
		}
		if opts.Progress != nil {
			start := time.Now()
			r.progTimer = ztimer.RepeatForeverNow(0.43, func() {
				opts.Progress(time.Since(start))
			})
		}
		if opts.Started != nil {
			opts.Started(r)
		}
	})
	return r
}

func PostAudioRecording(opts RecOptions, surl, userTokenForHeader string) {
	reader, writer := io.Pipe()
	NewAudioRecording(opts, writer)
	go func() {
		params := zhttp.MakeParameters()
		params.Reader = reader
		params.Method = http.MethodPost
		params.TimeoutSecs = 60 * 30
		if opts.MimeFormat == "" {
			opts.MimeFormat = "audio/mp4"
		}
		params.ContentType = opts.MimeFormat
		if userTokenForHeader != "" {
			params.Headers[zrest.UserAuthTokenHeaderKey] = userTokenForHeader
		}
		resp, err := zhttp.GetResponse(surl, params)
		if zlog.OnError(err, "post") {
			return
		}
		zlog.Info("REC POSTED:", resp.Header)
		message := resp.Header.Get(ResultMessageHeaderKey)
		if opts.Posted != nil {
			opts.Posted(message)
		}
	}()
}
