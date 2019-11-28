package zgo

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
