package data_structure

import (
	"time"
)

type Dictionary struct {
	dataDict  map[string]string
	expireKeyDictStore map[string]int64
}

func (d *Dictionary) GetExpireKeyDict() map[string]int64 {
	return d.expireKeyDictStore
}

func (d *Dictionary) GetDataDict() map[string]string {
	return d.dataDict
}

func InitSet() *Dictionary {
	dictionary := &Dictionary{
		dataDict:  make(map[string]string),
		expireKeyDictStore: make(map[string]int64),
	}
	return dictionary
}

func (d *Dictionary) Set(key, value string, exp int64) {
	d.dataDict[key] = value
	if exp != -1 {
		d.expireKeyDictStore[key] = time.Now().UnixMilli() + exp*1000
	}
}

func (d *Dictionary) Get(key string) (string, bool) {
	expireTime, ok := d.expireKeyDictStore[key]
	if ok && time.Now().UnixMilli() > expireTime {
		d.Del(key)
		return "", false
	}
	val, ok := d.dataDict[key]
	return val, ok
}

func (d *Dictionary) GetExpiry(key string) (int64, bool) {
	expireTime, ok := d.expireKeyDictStore[key]
	return expireTime, ok
}

func (d *Dictionary) Del(key string) {
	delete(d.dataDict, key)
	delete(d.expireKeyDictStore, key)
}
