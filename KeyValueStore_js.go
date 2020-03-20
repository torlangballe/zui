package zui

import (
	"syscall/js"

	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zlog"
)

func getLocalStorage() js.Value {
	return js.Global().Get("localStorage")
}

func (k KeyValueStore) getItem(key string, v interface{}) bool {
	k.prefixKey(&key)
	local := getLocalStorage()
	o := local.Get(key)

	switch o.Type() {
	case js.TypeUndefined:
		zlog.Debug(nil, zlog.StackAdjust(1), "KeyValueStore getItem item undefined:", key)
		return false

	case js.TypeNumber:
		zfloat.SetAny(v, o.Float())
		return true

	case js.TypeBoolean:
		*v.(*bool) = o.Bool()
		return true

	case js.TypeString:
		*v.(*string) = o.String()
		return true
	}
	zlog.Debug("bad type:", o.Type())
	return false
}

func (k *KeyValueStore) setItem(key string, v interface{}, sync bool) error {
	k.prefixKey(&key)
	local := getLocalStorage()
	local.Set(key, v)
	return nil
}
