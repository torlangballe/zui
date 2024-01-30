//go:build !js && !windows && catalyst

package zapp

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa
// void* SharedApplication(void);
// void Run(void *app);
import "C"
import (
	"net/url"
	"unsafe"
)

type nativeApp struct {
	sharedPtr unsafe.Pointer
}

func appNew(a *App) {
	a.sharedPtr = C.SharedApplication()
}

func (a *App) Run() {
	C.Run(a.sharedPtr)
}

func URLStub() string {
	return ""
}

func URL() *url.URL {
	return nil
}
