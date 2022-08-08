package zui

import (
	"github.com/torlangballe/zutil/zlog"
)

var DebugMode bool
var DebugOwnerMode bool // OwnerMode allows the software's developer to edit manuals and more from gui.

func IsRelease() bool {
	return false
}

func ErrorOnRelease() {
	if IsRelease() {
		var n = 100
		for n > 0 {
			zlog.Fatal(nil, "Should not run on ")
			n--
		}
	}
}
