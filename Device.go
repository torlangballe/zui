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

// DeviceCPUUsage returns a slice of 0-100% of how much each CPU is utilized. Order unknown, but hopefully doesn't change
func DeviceCPUUsage() (out []int) {
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
	out = make([]int, coresPhysical, coresPhysical)
	for i := 0; i < threads; i++ {
		for j := 0; j < coresPhysical; j++ {			
			out[j] += int(vals[n]) / threads
			n++
		}
	}
	return
}
