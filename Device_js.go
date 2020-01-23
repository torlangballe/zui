package zui

import (
	"strconv"
	"syscall/js"
	"time"

	ua "github.com/mileusna/useragent"
)

// https://developer.mozilla.org/en-US/docs/Web/API/Navigator - info about browser/dev

var userAgent *ua.UserAgent

func getUserAgentInfo() *ua.UserAgent {
	if userAgent == nil {
		ustr := js.Global().Get("navigator").Get("userAgent").String()
		u := ua.Parse(ustr)
		userAgent = &u
	}
	return userAgent
}

func DeviceIsIPad() bool {
	return false
}

func DeviceBrowserLocation() (protocol, hostname string, port int) {
	loc := WindowJS.Get("location")
	protocol = loc.Get("protocol").String()
	hostname = loc.Get("hostname").String()
	port, _ = strconv.Atoi(loc.Get("port").String())
	return
}
func DeviceIsBrowser() bool {
	return true
}

func DeviceWasmBrowser() string {
	return getUserAgentInfo().Name
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

func DeviceTimeZone() *time.Location {
	return time.UTC
}

func DeviceDeviceType() string {
	return ""
}

func DeviceDeviceCodeNumbered() (string, int, string) {
	return "", 0, ""
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

func DeviceHardwareTypeAndVersion() (string, float32) {
	return getUserAgentInfo().Device, 1
}