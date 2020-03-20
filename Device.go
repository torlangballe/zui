package zui

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/torlangballe/zutil/zlog"
)

type DeviceCellularNetworkType int

const (
	DeviceCellularUnknown DeviceCellularNetworkType = iota
	DeviceCellularWifiMax
	DeviceCellular2G
	DeviceCellular3G
	DeviceCellular4G
	DeviceCellular5G
	DeviceCellularXG
)

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
