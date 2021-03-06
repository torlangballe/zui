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
var fpath string

func KeyValueStoreFileNew(path string) *KeyValueStore {
	fpath = path
	if zfile.Exists(path) {
		err := zjson.UnmarshalFromFile(&dict, fpath)
		if err != nil {
			zlog.Error(err, "unmarshal")
			return nil
		}
	}
	return &KeyValueStore{Local: true}
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
		zjson.MarshalToFile(dict, fpath)
		lock.Unlock()
	}()
	return nil
}
