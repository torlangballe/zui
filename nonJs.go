// +build !js

package zui

import (
	"syscall"
	"time"
)

func getCreatedTimeFromStatT(fstat *syscall.Stat_t) time.Time {
	return time.Unix(int64(fstat.Ctimespec.Sec), int64(fstat.Ctimespec.Nsec))
}
