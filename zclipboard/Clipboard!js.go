//go:build !js && !windows

package zclipboard

// See https://github.com/d-tsuji/clipboard/blob/master/clipboard_darwin.go for some implementations

func SetString(str string) {}
