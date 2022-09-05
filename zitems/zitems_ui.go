//go:build zui

package zitems

import (
	"github.com/torlangballe/zutil/zrpc2"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

var (
	firstGets = true
)

func RepeatGetItems() {
	ztimer.RepeatNow(2, func() bool {
		// zlog.Info("RepeatGetItems:", zrpc2.MainClient.UseAuth, zrpc2.MainClient.AuthToken)
		if zrpc2.MainClient.UseAuth && zrpc2.MainClient.AuthToken == "" {
			return true
		}
		var resIDs []string
		if !firstGets {
			zrpc2.MainClient.Call("RPCCalls.GetUpdatedResourcesAndSetSent", nil, &resIDs)
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
