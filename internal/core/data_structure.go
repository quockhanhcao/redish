package core

import "github.com/quockhanhcao/redish/internal/data_structure"

var Dictionary *data_structure.Dictionary

func init() {
	Dictionary = data_structure.InitSet()
}
