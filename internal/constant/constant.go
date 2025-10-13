package constant

import "time"

const SamplingEvictionSize = 20
const SamplingEvictionThreshold = 0.1
const SamplingEvictionFrequency = 100 * time.Millisecond
