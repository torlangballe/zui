// +build !js
// +build !windows

package zui

// See https://github.com/d-tsuji/clipboard/blob/master/clipboard_darwin.go for some implementations

func PasteboardSetString(str string) {}
