package zgo

import (
	ua "github.com/mileusna/useragent"
)

var userAgent *ua.UserAgent

func getUserAgentInfo() *ua.UserAgent {
	if userAgent == nil {
		u := ua.Parse("")
		userAgent = &u
	}
	return nil
}

func DeviceIsIPad() bool {
	return false
}

func DeviceIsIPhone() bool {
	return false
}

func DeviceName() string {
	return ""
}

func DeviceFingerPrint() string {
	return ""
}

func DeviceIdentifierForVendor() string {
	return ""
}

func DeviceManufacturer() string {
	return ""
}

func DeviceBatteryLevel() float32 {
	return 0
}

func DeviceIsCharging() int {
	return 0
}

func DeviceOSVersionstring() string {
	return ""
}

func DeviceTimeZone() TimeZone {
	return TimeZone{}
}

func DeviceDeviceType() string {
	return ""
}

func DeviceDeviceCodeNumbered() (string, int, string) {
	return "", 0, ""
}

func DeviceHardwareType() string {
	return ""
}

func DeviceHardwareModel() string {
	return ""
}

func DeviceHardwareBrand() string {
	return ""
}

func DeviceOSPlatform() string {
	return ""
}

func DeviceCpuUsage() []float32 {
	return []float32{}
}

func DeviceFreeAndUsedDiskSpace() (int64, int64) {
	return 0, 0
}

func DeviceMemoryFreeAndUsed() (int64, int64) {
	return 0, 0
}

func DeviceIsWifiEnabled() bool {
	return false
}

func DeviceWifiIPv4Address() string {
	return ""
}

func DeviceIpv4Address() string {
	return ""
}

func DeviceIPv6Address() string {
	return ""
}

func DeviceGetMainMACint64() string {
	return ""
}

func DeviceLanMACint64() string {
	return ""
}

func DeviceWifiMint64() string {
	return ""
}

func DeviceWifiLinkSpeed() string {
	return ""
}

func DeviceCellularNetwork() DeviceCellularNetworkType {
	return DeviceCellularUnknown
}
