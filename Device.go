package zui

import (
	"runtime"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/torlangballe/zutil/zlog"
)

// Interesting: https://github.com/jaypipes/ghw

type DeviceCellularNetworkType int
type DeviceOSType string

const (
	DeviceCellularUnknown DeviceCellularNetworkType = iota
	DeviceCellularWifiMax
	DeviceCellular2G
	DeviceCellular3G
	DeviceCellular4G
	DeviceCellular5G
	DeviceCellularXG

	DeviceMacOSType   DeviceOSType = "macos"
	DeviceWindowsType DeviceOSType = "windows"
	DeviceJSType      DeviceOSType = "js"
)

func DeviceOS() DeviceOSType {
	switch runtime.GOOS {
	case "windows":
		return DeviceWindowsType
	case "darwin":
		return DeviceMacOSType
	case "js":
		return DeviceJSType
	}
	zlog.Fatal(nil, "other type")
	return DeviceOSType("")
}

func DeviceOSVersion() string {
	info, err := host.Info()
	zlog.OnError(err)
	return info.PlatformVersion
}

// DeviceCPUUsage returns a slice of 0-1 where 1 is 100% of how much each CPU is utilized. Order unknown, but hopefully doesn't change
func DeviceCPUUsage() (out []float64) {
	coresVirtual, _ := cpu.Counts(true)
	coresPhysical, _ := cpu.Counts(false)

	threads := coresVirtual / coresPhysical
	percpu := true
	vals, err := cpu.Percent(0, percpu)
	if err != nil {
		zlog.Error(err)
		return
	}

	n := 0
	out = make([]float64, coresPhysical, coresPhysical)
	for i := 0; i < threads; i++ {
		for j := 0; j < coresPhysical; j++ {
			out[j] += float64(int(vals[n]) / threads)
			n++
		}
	}
	for j := 0; j < coresPhysical; j++ {
		out[j] /= 100
	}
	return
}
