//go:build server

package zitems

import (
	"path/filepath"
	"reflect"

	"github.com/torlangballe/zutil/zjson"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zwrpc"
)

var (
	FileStoreFolder string
)

func Init() {
	// zrpc.Register(Calls)
	zwrpc.Register(Calls)
}

/*
func Init(rpcPort int) {
	// zwrpc.InitServer("", rpcPort)
	zwrpc.Register(Calls)
	zwrpc.NewClientHandler = func(id string) {
		for _, item := range AllItems {
			if item.ResourceID == "xxcontainers" {
				continue
			}
			zlog.Info("Get", item.ResourceID)
			err := zwrpc.CallAllClientsFromServer("ZItemsCalls.ReceiveUpdateItem", &item, id)
			// zlog.Info("zitems.Init call receive update", err, item.ResourceID)
			if err != nil {
				zlog.Error(err, "zitems.Init call receive update err")
			}
		}
	}
}
*/

func RegisterItemForServer(dataPtr any, resourceID string) {
	var item Item
	item.DataPtr = dataPtr
	item.ResourceID = resourceID
	AllItems = append(AllItems, item)
	file := filepath.Join(FileStoreFolder, resourceID+".json")
	err := zjson.UnmarshalFromFile(dataPtr, file, true)
	// zlog.Info("Loaded:", resourceID, err, dataPtr, file)
	if err != nil {
		zlog.Error(err, "load", resourceID)
	}
}

func SaveItem(resourceID string) error {
	item, _ := FindItem(resourceID)
	zlog.Assert(item != nil, resourceID)
	data := reflect.ValueOf(item.DataPtr).Elem().Interface()
	file := filepath.Join(FileStoreFolder, resourceID+".json")
	// zlog.Info("Save:", reflect.ValueOf(item.DataPtr).Elem().Len(), data)
	err := zjson.MarshalToFile(data, file)
	return err
}

// func (ic *ZItemsCalls) GetItem(req *http.Request, getItem *Item, newDataPtr any) error {
// 	findItem, _ := FindItem(getItem.ResourceID)
// 	if findItem == nil {
// 		return zlog.Error(nil, "zitems.GetItem failed, no item registered with resourceId:", getItem.ResourceID)
// 	}
// 	data := reflect.ValueOf(findItem.DataPtr).Elem()
// 	ndp := reflect.ValueOf(newDataPtr)
// 	zlog.Info("KIND:", ndp.Kind()) //, ndp.Type())
// 	reflect.ValueOf(newDataPtr).Elem().Set(data)
// 	return nil
// }

func (ic *ZItemsCalls) GetItem(id *string, item *Item) error {
	findItem, _ := FindItem(*id)
	if findItem == nil {
		return zlog.Error(nil, "zitems.UpdateItem failed, no item registered with resourceId:", id)
	}
	*item = *findItem
	// data = reflect.ValueOf(findItem.DataPtr).Elem() // we set it from pointer again just in case changed somehow
	// reflect.ValueOf(newDataPtr).Elem().Set(data)
	return nil
}
