package zui

import (
	"fmt"
	"syscall/js"
)

func WebSocketConnect(address string) {
	ws := js.Global().Get("WebSocket").New("ws://localhost:8040/ws")

	ws.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("open ws")
		// var a js.Value
		// ws.Call("send", js.CopyBytesToJS(a, []byte{123}))
		return nil
	}))
}
