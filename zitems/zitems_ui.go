//go:build zui

package zitems

import (
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zreflect"
	"github.com/torlangballe/zutil/zrpc"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
	"github.com/torlangballe/zutil/zwrpc"
	// "github.com/torlangballe/zutil/zwrpc"
)

var (
	firstGets = true

// RPCReceiver *zwrpc.Client
)

func Init() {
	var err error

	// id := zstr.GenerateRandomHexBytes(12)
	// zwrpc.MainSendClient, err = zwrpc.NewClient(rpcAddress, rpcPort, rpcSSL, id, 0.8)
	// zlog.Info("zitems.Init:", rpcAddress, rpcPort, zwrpc.MainSendClient != nil, err)
	if err != nil {
		zlog.Error(err, "create rpc client/receiver")
	}
	zwrpc.Register(Calls)
}

func RegisterItemInGUI(data any, resourceID, name string, update func(any)) {
	item := Item{DataPtr: data, Name: name, UpdateFunc: update, ResourceID: resourceID}
	item.UpdateFunc = update
	AllItems = append(AllItems, item)
}

func CallGetItem(resourceID string) {
	item, _ := FindItem(resourceID)
	zlog.Assert(item != nil)
	// newDataPtr := zreflect.NewOfAny(item.DataPtr)
	err := zwrpc.MainHTTPClient.Call("ZItemsCalls.GetItem", item, item.DataPtr)
	if err != nil {
		zlog.Error(err, "call GetItem failed")
		return
	}
	item.UpdateFunc(item.DataPtr)
	// AllItems[i].DataPtr = newDataPtr
}

func CallUpdateItem(resourceID string) {
	item, _ := FindItem(resourceID)
	zlog.Assert(item != nil)
	newDataPtr := zreflect.NewOfAny(item.DataPtr)
	err := zrpc.ToServerClient.CallRemote("ZItemsCalls.UpdateItem", &item, newDataPtr)
	if err != nil {
		zlog.Error(err, "callToUpdateItem failed")
		return
	}
	item.UpdateFunc(newDataPtr)
}

func RepeatGetItems() {
	ztimer.RepeatNow(2, func() bool {
		var resIDs []string
		if !firstGets {
			zrpc.ToServerClient.CallRemote("RPCCalls.GetUpdatedResourcesAndSetSent", nil, &resIDs)
		}
		for _, item := range AllItems {
			zlog.Info("Get1:", item.ResourceID, firstGets, resIDs)
			if firstGets || zstr.IndexOf(item.ResourceID, resIDs) != -1 {
				zlog.Info("Get:", item.ResourceID)
				go CallGetItem(item.ResourceID)
			}
		}
		firstGets = false
		return true
	})
}

// func (tc *ZItemsCalls) ReceiveUpdateItem(item *Item, result *zwrpc.Any) error {
// 	zlog.Info("ReceiveUpdateItem", item.ResourceID)
// 	return nil
// }
