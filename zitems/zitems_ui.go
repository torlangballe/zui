//go:build zui

package zitems

import (
	zrpc "github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

var (
	firstGets = true
)

func RepeatGetItems() {
	ztimer.RepeatNow(2, func() bool {
		// zlog.Info("RepeatGetItems:", zrpc.MainClient.UseAuth, zrpc.MainClient.AuthToken, len(AllItems))
		// if zrpc.MainClient.UseAuth && zrpc.MainClient.AuthToken == "" {
		// 	return true
		// }
		var resIDs []string
		if !firstGets {
			zrpc.MainClient.Call("RPCCalls.GetUpdatedResourcesAndSetSent", nil, &resIDs)
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
