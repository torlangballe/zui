// +build !js

package zui

import (
	"github.com/denisbrodbeck/machineid"
	"github.com/torlangballe/zutil/zlog"
)

func DeviceWasmBrowser() string {
	return ""
}

func DeviceUniqueID() string {
	id, err := machineid.ID()
	zlog.Assert(err == nil)
	return id
}
