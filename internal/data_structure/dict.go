package data_structure

import (
	"time"
)

type Dictionary struct {
	dataStore  map[string]string
	expireKeys map[string]int64
}

func InitSet() *Dictionary {
	dictionary := &Dictionary{
		dataStore:  make(map[string]string),
		expireKeys: make(map[string]int64),
	}
	return dictionary
}

func (d *Dictionary) AddToSet(key, value string, exp int64) {
	d.dataStore[key] = value
	if (exp != -1) {
		d.expireKeys[key] = time.Now().UnixMilli() + exp * 1000
	}
}

func (d *Dictionary) GetFromSet(key string) (string, bool) {
	val, ok := d.dataStore[key]
	return val, ok
}
