// +build !js

package zui

import (
	"github.com/torlangballe/zutil/zcommand"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"strconv"
	"strings"
)

func DeviceWasmBrowser() string {
	return ""
}

func DeviceBrowserLocation() (protocol, hostname string, port int) {
	return "", "", 0
}

func DeviceHardwareTypeAndVersion() (string, float32) {
	str, err := zcommand.RunCommand("sysctl", 0, "-n", "hw.model")
	if err != nil {
		zlog.Error(err)
		return "", 0
	}
	i := strings.IndexAny(str, zstr.Digits)
	if i == -1 {
		return str, 1
	}
	name := zstr.Head(str, i)
	num, _ := strconv.ParseFloat(zstr.Body(str, i, -1), 32)

	return name, float32(num)
}
