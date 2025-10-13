package core

import (
	"time"

	"github.com/quockhanhcao/redish/internal/constant"
)

func SamplingEviction() {
	for {
		expireCount := 0
		samplingSize := constant.SamplingEvictionSize
		expireKeys := Dictionary.GetExpireKeyDict()
		for key := range expireKeys {
			samplingSize--
			if time.Now().UnixMilli() > expireKeys[key] {
				expireCount++
				Dictionary.Del(key)
			}
			if samplingSize == 0 {
				break
			}
		}
		if float64(expireCount)/float64(constant.SamplingEvictionSize) <= constant.SamplingEvictionThreshold {
			break
		}
	}
}
