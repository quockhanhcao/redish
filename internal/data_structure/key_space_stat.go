package data_structure

type KeySpaceStat struct {
	Key    int64
	Expire int64
}

func InitKeySpaceStat() *KeySpaceStat {
	return &KeySpaceStat{
		Key:    0,
		Expire: 0,
	}
}

func (k *KeySpaceStat) AddKey() {
	k.Key += 1
}

func (k *KeySpaceStat) RemoveKey() {
	k.Key -= 1
}

var Stats = InitKeySpaceStat()
