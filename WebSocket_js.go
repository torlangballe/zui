package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zlog"
)

type WebSocket struct {
	ws js.Value
}

func WebSocketConnect(address, id string) *WebSocket {
	w := &WebSocket{}
	w.ws = js.Global().Get("WebSocket").New("wss://127.0.0.1:9998/ws")
	// zlog.Info("zui.WebSocketConnect", ws)
	w.ws.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		w.Send(id)
		zlog.Info("open ws")
		// var a js.Value
		// ws.Call("send", js.CopyBytesToJS(a, []byte{123}))
		return nil
	}))
	return w
}

func (w *WebSocket) Send(str string) {
	w.ws.Call("send", str)
}
