package data_structure

import (
	"time"
)

type Obj struct {
	Value      interface{}
	AccessTime uint32
}

type Dictionary struct {
	dataDict           map[string]*Obj
	expireKeyDictStore map[string]int64
}

func (d *Dictionary) GetExpireKeyDict() map[string]int64 {
	return d.expireKeyDictStore
}

func (d *Dictionary) GetDataDict() map[string]*Obj {
	return d.dataDict
}

func InitDictionary() *Dictionary {
	dictionary := &Dictionary{
		dataDict:           make(map[string]*Obj),
		expireKeyDictStore: make(map[string]int64),
	}
	return dictionary
}

func (d *Dictionary) Set(key string, value interface{}, exp int64) {
	_, exist := d.dataDict[key]
	if !exist {
		Stats.AddKey()
	}
	d.dataDict[key] = &Obj{
		Value: value,
	}
	if exp != -1 {
		d.expireKeyDictStore[key] = time.Now().UnixMilli() + exp*1000
	}
}

func (d *Dictionary) Get(key string) *Obj {
	val, ok := d.dataDict[key]
	if ok {
		expireTime, ok := d.expireKeyDictStore[key]
		if ok && time.Now().UnixMilli() > expireTime {
			d.Del(key)
			return nil
		}
	}
	return val
}

func (d *Dictionary) GetExpiry(key string) (int64, bool) {
	expireTime, ok := d.expireKeyDictStore[key]
	return expireTime, ok
}

func (d *Dictionary) SetExpiry(key string, exp int64) {
	d.expireKeyDictStore[key] = time.Now().UnixMilli() + exp*1000
}

func (d *Dictionary) Del(key string) {
	if _, exist := d.dataDict[key]; exist {
		delete(d.dataDict, key)
		delete(d.expireKeyDictStore, key)
		Stats.RemoveKey()
	}
}
