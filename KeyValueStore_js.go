package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zlog"
)

func getLocalStorage() js.Value {
	return js.Global().Get("localStorage")
}

func (k KeyValueStore) getItem(key string, v interface{}) bool {
	k.prefixKey(&key)
	local := getLocalStorage()
	o := local.Get(key)
	if o.Type() == js.TypeUndefined {
		zlog.Error(nil, "KeyValueStore getItem item undefined:", key)
		return false
	}
	v = &o
	return true
}

func (k *KeyValueStore) setitem(key string, v interface{}) error {
	k.prefixKey(&key)
	local := getLocalStorage()
	local.Set(key, v)
	return nil
}
