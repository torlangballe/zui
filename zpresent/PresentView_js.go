package zpresent

import (
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zdom"
	"github.com/torlangballe/zui/zview"
)

func init() {
	zcustom.IsPresentingFunc = func() bool {
		return Presenting
	}
	zview.SetPresentReadyFunc = CallReady
}
func SetURL(surl string) {
	zdom.WindowJS.Get("location").Call("assign", surl)
}

// https://bubblin.io/blog/fullscreen-api-ipad
// https://medium.com/@firt/iphone-11-ipados-and-ios-13-for-pwas-and-web-development-5d5d9071cc49
