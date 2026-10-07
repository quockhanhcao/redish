package data_structure

import (
	"fmt"
	"math"

	"github.com/spaolacci/murmur3"
)

type CountMinSketch struct {
	depth      uint32
	width      uint32
	matrix     [][]uint64
	totalCount uint64
}

// NewCountMinSketch builds a sketch whose estimates overshoot the true count by
// at most errorRate of the total ever added, for all but probabilityRate of the
// keys. Both rates must be between 0 and 1, exclusive.
//
// The bounds are written as !(x > 0 && x < 1) rather than x <= 0 || x >= 1 so
// that NaN is rejected too: every comparison against NaN is false, so the
// negated form catches it while the direct form would let it through and leave
// calcCMSDim converting NaN to a uint32.
func NewCountMinSketch(errorRate, probabilityRate float64) (*CountMinSketch, error) {
	if !(errorRate > 0 && errorRate < 1) {
		return nil, fmt.Errorf("error rate must be between 0 and 1, got %v", errorRate)
	}
	if !(probabilityRate > 0 && probabilityRate < 1) {
		return nil, fmt.Errorf("probability rate must be between 0 and 1, got %v", probabilityRate)
	}
	width, depth := calcCMSDim(errorRate, probabilityRate)
	cms := &CountMinSketch{
		depth:      depth,
		width:      width,
		totalCount: 0,
	}
	matrix := make([][]uint64, depth)
	for i := range depth {
		matrix[i] = make([]uint64, width)
	}
	cms.matrix = matrix
	return cms, nil
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
	// increase total increment
	if value > uint64(math.MaxUint64)-cms.totalCount {
		cms.totalCount = uint64(math.MaxUint64)
	} else {
		cms.totalCount += value
	}
	return minCount
}

func (cms *CountMinSketch) GetWidth() uint32 {
	return cms.width
}

func (cms *CountMinSketch) GetDepth() uint32 {
	return cms.depth
}

func (cms *CountMinSketch) GetTotalCount() uint64 {
	return cms.totalCount
}

func (cms *CountMinSketch) GetMember(key string) uint64 {
	minCount := uint64(math.MaxUint64)
	// hash the key to find the position in every row, add to it
	// return the estimated value for the key
	for i := range cms.depth {
		hashedKey := hash(key, i)
		pos := hashedKey % cms.width
		// avoid overflow
		if cms.matrix[i][pos] < minCount {
			minCount = cms.matrix[i][pos]
		}
	}
	return minCount
}
