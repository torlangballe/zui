package zui

import (
	"syscall"
	"time"

	"github.com/torlangballe/zutil/zdevice"
	"github.com/torlangballe/zutil/zgeo"
)

// https://github.com/siongui/godom
// https://medium.zenika.com/go-1-11-webassembly-for-the-gophers-ae4bb8b1ee03
// https://github.com/golang/go/wiki/WebAssembly

// type css js.Value

func init() {
	if zdevice.OS() == zdevice.MacOSType && zdevice.WasmBrowser() == "safari" {
		zgeo.FontDefaultName = "-apple-system"
	}
}

func getCreatedTimeFromStatT(fstat *syscall.Stat_t) time.Time {
	return time.Time{}
}
