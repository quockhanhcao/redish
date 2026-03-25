package core

import "github.com/quockhanhcao/redish/internal/data_structure"

var Dictionary *data_structure.Dictionary
var StoredSet map[string]*data_structure.SimpleSet
var StoredCountMinSketch map[string]*data_structure.CountMinSketch
var StoredBloomFilter map[string]*data_structure.Bloom

func init() {
	Dictionary = data_structure.InitSet()
	StoredSet = make(map[string]*data_structure.SimpleSet)
	StoredCountMinSketch = make(map[string]*data_structure.CountMinSketch)
}
