//go:build server

package zitems

import (
	"path/filepath"
	"reflect"

	"github.com/torlangballe/zutil/zjson"
	"github.com/torlangballe/zutil/zlog"
	zrpc "github.com/torlangballe/zutil/zrpc"
)

var (
	FileStoreFolder string
)

func Init() {
	zrpc.Register(Calls)
}

func RegisterItemForServer(dataPtr any, resourceID, name string) *Item {
	item := RegisterItem(dataPtr, resourceID, name)
	file := filepath.Join(FileStoreFolder, resourceID+".json")
	err := zjson.UnmarshalFromFile(dataPtr, file, true)
	// zlog.Info("Loaded:", resourceID, err, dataPtr, file)
	if err != nil {
		zlog.Error(err, "load", resourceID)
	}
	return item
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
