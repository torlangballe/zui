// +build !js

package zui

import (
	"github.com/garyburd/redigo/redis"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zredis"
)

var redisPool *redis.Pool

func (k KeyValueStore) getItem(key string, v interface{}) bool {
	if key[0] != '/' && k.KeyPrefix != "" {
		key = k.KeyPrefix + "/" + key
	}
	got, err := zredis.Get(redisPool, v, key)
	if err != nil {
		zlog.Error(err, "keyvalstore redis get")
	}
	return got
}
