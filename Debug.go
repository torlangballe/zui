package zui

import "github.com/torlangballe/zutil/zlog"

var DebugMode bool

func DebugIsRelease() bool {
	return false
}

func IsMinIOS11() bool {
	return false
}

func ErrorOnRelease() {
	if DebugIsRelease() {
		var n = 100
		for n > 0 {
			zlog.Fatal(nil, "Should not run on ")
			n--
		}
	}
}
