package zgo

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

func (k KeyValueStore) ObjectForKey(key string) (object interface{}, got bool) {
	got = k.getItem(key, &object)
	return
}

func (k KeyValueStore) StringForKey(key string) (str string, got bool) {
	got = k.getItem(key, &str)
	return
}

func (k KeyValueStore) DictionaryForKey(key string) (dict Dictionary, got bool) {
	got = k.getItem(key, &dict)
	return
}

func (k KeyValueStore) DataForKey(key string) (data *Data, got bool) {
	got = k.getItem(key, &data)
	return
}

func (k KeyValueStore) IntForKey(key string) (int64, bool) {
	return 0, false
}

func (k KeyValueStore) DoubleForKey(key string) (float64, bool) {
	return 0, false
}

func (k KeyValueStore) TimeForKey(key string) (Time, bool) {
	return TimeNull, false
}

func (k KeyValueStore) BoolForKey(key string) (bool, bool) {
	return false, false
}

func (k KeyValueStore) IncrementInt(key string, sync bool, inc int) int {
	return 0
}

func (k KeyValueStore) RemoveForKey(key string, sync bool) {

}

func (k KeyValueStore) SetObject(object interface{}, key string)                         {}
func (k KeyValueStore) SetString(string, key string, sync bool)                          {}
func (k KeyValueStore) SetData(data *Data, key string, sync bool)                        {}
func (k KeyValueStore) SetDictionary(dict map[string]interface{}, key string, sync bool) {}
func (k KeyValueStore) SetInt(value int64, key string, sync bool)                        {}
func (k KeyValueStore) SetDouble(value float64, key string, sync bool)                   {}
func (k KeyValueStore) Setbool(value bool, key string, sync bool)                        {}
func (k KeyValueStore) SetTime(value Time, key string, sync bool)                        {}
func (k KeyValueStore) ForAllKeys(got func(key string))                                  {}

func (k KeyValueStore) prefixKey(key *string) {
	if (*key)[0] != '/' && k.KeyPrefix != "" {
		*key = k.KeyPrefix + "/" + *key
	}
}
