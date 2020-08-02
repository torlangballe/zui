// +build !js

package zui

import (
	"os"
	"strings"

	"github.com/torlangballe/zutil/zgeo"
)

// TextInfo
func (ti *TextInfo) getTextSize(noWidth bool) zgeo.Size {
	return zgeo.Size{}
}

// CustomView

type baseCustomView struct {
}

func CustomViewInit() *NativeView {
	return nil
}

func zViewAddView(parent View, child View, index int) {
}

// CustomView

func zViewSetRect(view View, rect zgeo.Rect, layout bool) { // layout only used on android
}

// App:

// AppURL returns the url/command that invoked this app
func AppURL() string {
	return strings.Join(os.Args, " ")
}

// MainArgs returns the path of the executable and arguments given
func AppArgs() (path string, args []string) {
	path = os.Args[0]
	if len(os.Args) != 1 {
		args = os.Args[1:]
	}
	return
}
