package zui

import (
	"syscall"
	"time"

	"github.com/torlangballe/zutil/zlog"
)

func init() {
	zlog.Info("zui init()")
	// runtime.LockOSThread() // ! skip for now
}

func getCreatedTimeFromStatT(fstat *syscall.Stat_t) time.Time {
	return time.Unix(int64(fstat.Ctimespec.Sec), int64(fstat.Ctimespec.Nsec))
}
