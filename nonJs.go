// +build !js

package zgo

import (
	"syscall"
	"time"
)

func getCreatedTimeFromStatT(fstat *syscall.Stat_t) Time {
	return Time{time.Unix(int64(fstat.Ctimespec.Sec), int64(fstat.Ctimespec.Nsec))}
}
