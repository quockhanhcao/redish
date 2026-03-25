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
	bloomFilter []uint8
	bits        uint64
	bytes       uint64
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
	// each entry not only need 1 position in the array, but multiple one
	bits := entries * uint64(bitPerEntry)
	var bytes uint64
	if bits%64 != 0 {
		bytes = ((bits / 64) + 1) * 64
	} else {
		bytes = bits / 64
	}
	hashes := math.Ceil(bitPerEntry * Ln2)
	bf := make([]uint8, bytes)
	bloom := Bloom{
		hashes:      int(hashes),
		bloomFilter: bf,
		bits:        bits,
		bytes:       bytes,
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
	initHash := b.CalcHash(entry)
	for i := 0; i < b.hashes; i++ {
		bitIdx := (initHash.a + initHash.b*uint64(i)) % b.bits
		// we need to find the position of the bit in the array of bytes
		// shift right by 3 (divide by 8) to find
		byteIdx := bitIdx >> 3
		// now we need the offset of the bit in the byte
		idx := bitIdx % 8
		// turn the bit at offset idx to 1, we create a mask
		mask := uint8(1) << idx
		// use OR operator to turn on that bit
		b.bloomFilter[byteIdx] = b.bloomFilter[byteIdx] | mask
	}
}

func (b *Bloom) Exist(entry string) bool {
	initHash := b.CalcHash(entry)
	for i := 0; i < b.hashes; i++ {
		bitIdx := (initHash.a + initHash.b*uint64(i)) % b.bits
		byteIdx := bitIdx >> 3
		// create a mask
		mask := 1 << (bitIdx % 8)
		if (b.bloomFilter[byteIdx] & uint8(mask)) == 0 {
			return false
		}
	}
	return true
}
