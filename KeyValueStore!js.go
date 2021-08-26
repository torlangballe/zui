// +build !js

package zui

import (
	"reflect"
	"sync"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zjson"
	"github.com/torlangballe/zutil/zlog"
)

var lock sync.Mutex
var dict = zdict.Dict{}

func KeyValueStoreFileNew(path string) *KeyValueStore {
	k := &KeyValueStore{Local: true}
	k.filepath = zfile.ChangedExtension(path, ".json")
	err := zjson.UnmarshalFromFile(&dict, k.filepath, true)
	if err != nil {
		zlog.Error(err, "unmarshal")
		return nil
	}
	return k
}

func (k KeyValueStore) getItem(key string, pointer interface{}) bool {
	if key[0] != '/' && k.KeyPrefix != "" {
		key = k.KeyPrefix + "/" + key
	}
	lock.Lock()
	defer lock.Unlock()
	gval, got := dict[key]
	if got {
		reflect.ValueOf(pointer).Elem().Set(reflect.ValueOf(gval))
		return true
	}
	return false
}

func (k *KeyValueStore) setItem(key string, v interface{}, sync bool) error {
	go func() {
		lock.Lock()
		dict[key] = v
		zjson.MarshalToFile(dict, k.filepath)
		lock.Unlock()
	}()
	return nil
}
