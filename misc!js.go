//go:build !js
// +build !js

package zui

import (
	"os"
)

// // TextInfo
// func (ti *TextInfo) getTextSize(noWidth bool) zgeo.Size {
// 	return zgeo.Size{}
// }

// func zViewSetRect(view View, rect zgeo.Rect, layout bool) { // layout only used on android
// }

// App:

// MainArgs returns the path of the executable and arguments given
func AppArgs() (path string, args []string) {
	path = os.Args[0]
	if len(os.Args) != 1 {
		args = os.Args[1:]
	}
	return
}
