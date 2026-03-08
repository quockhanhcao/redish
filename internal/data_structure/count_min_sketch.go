package data_structure

import (
	"math"

	"github.com/spaolacci/murmur3"
)

type CountMinSketch struct {
	depth  uint32
	width  uint32
	matrix [][]uint64
}

func NewCountMinSketch(errorRate, probabilityRate float64) *CountMinSketch {
	depth, width := calcCMSDim(errorRate, probabilityRate)
	cms := &CountMinSketch{
		depth: depth,
		width: width,
	}
	matrix := make([][]uint64, depth)
	for i := range depth {
		matrix[i] = make([]uint64, width)
	}
	cms.matrix = matrix
	return cms
}

func hash(key string, seed uint32) uint32 {
	hasher := murmur3.New32WithSeed(seed)
	hasher.Write([]byte(key))
	return hasher.Sum32()
}

func calcCMSDim(errorRate, probabilityRate float64) (uint32, uint32) {
	width := uint32(math.Ceil(2.0 / errorRate))

	depth := uint32(math.Ceil(math.Log10(probabilityRate) / math.Log10(0.5)))
	return width, depth
}

func (cms *CountMinSketch) Increase(key string, value uint64) uint64 {
	minCount := uint64(math.MaxUint64)
	// hash the key to find the position in every row, add to it
	// return the estimated value for the key
	for i := range cms.depth {
		hashedKey := hash(key, i)
		pos := hashedKey % cms.width
		// avoid overflow
		if value > uint64(math.MaxUint64)-cms.matrix[i][pos] {
			cms.matrix[i][pos] = uint64(math.MaxUint64)
		} else {
			cms.matrix[i][pos] += value
		}
		if cms.matrix[i][pos] < minCount {
			minCount = cms.matrix[i][pos]
		}
	}

	return minCount
}
