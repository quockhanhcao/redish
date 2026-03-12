package data_structure

import (
	"math"

	"github.com/spaolacci/murmur3"
)

const Ln2 float64 = 0.693147180559945
const Ln2Square float64 = 0.480453013918201
const ABigSeed uint32 = 0x9747b28c

type Bloom struct {
	hashes      int
	bloomFilter []bool
	bits        uint64
}

type HashValue struct {
	a uint64
	b uint64
}

func calculateBitPerEntry(error float64) float64 {
	num := math.Log(error)
	bpe := math.Abs(-(num / Ln2Square))
	return bpe
}

/*
http://en.wikipedia.org/wiki/Bloom_filter
- Optimal number of bits is: bits = (entries * ln(error)) / ln(2)^2
- bitPerEntry = bits/entries
- Optimal number of hash functions is: hashes = bitPerEntry * ln(2)
*/
func NewBloomFilter(entries uint64, errorRate float64) *Bloom {
	bitPerEntry := calculateBitPerEntry(errorRate)
	bits := entries * uint64(bitPerEntry)
	if bits%64 != 0 {
		bits = ((bits / 64) + 1) * 64
	}
	hashes := math.Ceil(bitPerEntry * Ln2)
	bf := make([]bool, bits)
	bloom := Bloom{
		hashes:      int(hashes),
		bloomFilter: bf,
		bits:        bits,
	}
	return &bloom
}

func (b *Bloom) CalcHash(entry string) HashValue {
	hasher := murmur3.New128WithSeed(ABigSeed)
	hasher.Write([]byte(entry))
	x, y := hasher.Sum128()
	return HashValue{
		a: x,
		b: y,
	}
}

func (b *Bloom) Add(entry string) {
	var bitPos uint64
	initHash := b.CalcHash(entry)
	for i := 0; i < b.hashes; i++ {
		bitPos = (initHash.a + initHash.b*uint64(i)) % b.bits
		b.bloomFilter[bitPos] = true
	}
}

func (b *Bloom) Exist(entry string) bool {
	var bitPos uint64
	initHash := b.CalcHash(entry)
	for i := 0; i < b.hashes; i++ {
		bitPos = (initHash.a + initHash.b*uint64(i)) % b.bits
		if !b.bloomFilter[bitPos] {
			return false
		}
	}
	return true
}
