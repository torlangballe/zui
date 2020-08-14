package zui

import (
	"encoding/json"
	"time"

	"github.com/torlangballe/zutil/zdict"
	"github.com/torlangballe/zutil/zlog"
)

//  Created by Tor Langballe on /30/10/15.

// For storage:
// https://github.com/peterbourgon/diskv
// https://github.com/recoilme/pudge
// https://github.com/nanobox-io/golang-scribble

type KeyValueStore struct {
	Local     bool   // if true, only for single browser or device, otherwise for user anywhere
	Secure    bool   // true if key/value stored in secure key chain
	KeyPrefix string // this can be a user id. Not used if key starts with /
}

var DefaultLocalKeyValueStore *KeyValueStore

func KeyValueStoreNew(local bool) *KeyValueStore {
	return &KeyValueStore{Local: local}
}

func (k KeyValueStore) GetObject(key string, objectPtr interface{}) (got bool) {
	var rawjson string
	got = k.getItem(key, &rawjson)
	if got {
		err := json.Unmarshal([]byte(rawjson), objectPtr)
		if zlog.OnError(err, "unmarshal") {
			return
		}
	}
	return
}

func (k KeyValueStore) GetString(key string) (str string, got bool) {
	got = k.getItem(key, &str)
	return
}

func (k KeyValueStore) GetDict(key string) (dict zdict.Dict, got bool) {
	got = k.getItem(key, &dict)
	return
}

func (k KeyValueStore) GetInt64(key string, def int64) (val int64, got bool) {
	got = k.getItem(key, &val)
	if got {
		return val, true
	}
	return def, true
}

func (k KeyValueStore) GetInt(key string, def int) (int, bool) {
	n, got := k.GetInt64(key, int64(def))
	return int(n), got
}

func (k KeyValueStore) GetDouble(key string, def float64) (val float64, got bool) {
	got = k.getItem(key, &val)
	if got {
		return val, true
	}
	return def, true
}

func (k KeyValueStore) GetTime(key string) (time.Time, bool) {
	return time.Time{}, false
}

func (k KeyValueStore) GetBool(key string, def bool) (val bool, got bool) {
	got = k.getItem(key, &val)
	if got {
		return val, true
	}
	return def, true
}

func (k KeyValueStore) IncrementInt(key string, sync bool, inc int) int {
	return 0
}

func (k KeyValueStore) RemoveForKey(key string, sync bool) {

}

func (k KeyValueStore) SetObject(object interface{}, key string, sync bool) {
	data, err := json.Marshal(object)
	if zlog.OnError(err, "marshal") {
		return
	}
	k.setItem(key, string(data), sync)
}
func (k KeyValueStore) SetString(value string, key string, sync bool)  { k.setItem(key, value, sync) }
func (k KeyValueStore) SetDict(dict zdict.Dict, key string, sync bool) { k.setItem(key, dict, sync) }
func (k KeyValueStore) SetInt64(value int64, key string, sync bool)    { k.setItem(key, value, sync) }
func (k KeyValueStore) SetInt(value int, key string, sync bool)        { k.setItem(key, value, sync) }
func (k KeyValueStore) SetDouble(value float64, key string, sync bool) { k.setItem(key, value, sync) }
func (k KeyValueStore) SetBool(value bool, key string, sync bool)      { k.setItem(key, value, sync) }
func (k KeyValueStore) SetTime(value time.Time, key string, sync bool) { k.setItem(key, value, sync) }
func (k KeyValueStore) ForAllKeys(got func(key string))                {}

func (k KeyValueStore) prefixKey(key *string) {
	if (*key)[0] != '/' && k.KeyPrefix != "" {
		*key = k.KeyPrefix + "/" + *key
	}
}
