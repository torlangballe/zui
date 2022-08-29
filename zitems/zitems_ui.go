//go:build zui

package zitems

import (
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
	"github.com/torlangballe/zutil/zwrpc"
)

var (
	firstGets = true
)

func Init() {
	zwrpc.Register(Calls)
}

func RepeatGetItems() {
	ztimer.RepeatNow(2, func() bool {
		var resIDs []string
		if !firstGets {
			zwrpc.MainHTTPClient.Call("RPCCalls.GetUpdatedResourcesAndSetSent", nil, &resIDs)
		}
		for _, item := range AllItems {
			// zlog.Info("Get1:", item.ResourceID, firstGets, resIDs)
			if item.GetFunc != nil && (firstGets || zstr.IndexOf(item.ResourceID, resIDs) != -1) {
				// zlog.Info("Get:", item.ResourceID)
				item.GetFunc()
			}
		}
		firstGets = false
		return true
	})
}
