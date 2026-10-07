package data_structure

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
)

// Pairs are deliberately asymmetric. A pair whose width and depth happen to be
// equal would pass even with the two transposed, so it would prove nothing.
var cmsDimensions = []struct {
	name      string
	errorRate float64
	probRate  float64
	wantWidth uint32
	wantDepth uint32
}{
	{"1% error, 1% probability", 0.01, 0.01, 200, 7},
	{"0.1% error, 1% probability", 0.001, 0.01, 2000, 7},
	{"10% error, 10% probability", 0.1, 0.1, 20, 4},
	{"1% error, 0.1% probability", 0.01, 0.001, 200, 10},
	{"10% error, 0.1% probability", 0.1, 0.001, 20, 10},
	{"50% error, 50% probability", 0.5, 0.5, 4, 1},
}

// TestCalcCMSDim covers the formulas on their own: width = 2/error sets how
// large the overcount can be, depth = log2(1/probability) sets how often it is
// exceeded.
func TestCalcCMSDim(t *testing.T) {
	for _, tt := range cmsDimensions {
		t.Run(tt.name, func(t *testing.T) {
			width, depth := calcCMSDim(tt.errorRate, tt.probRate)
			if width != tt.wantWidth || depth != tt.wantDepth {
				t.Errorf("calcCMSDim(%v, %v) = (%d, %d); want (%d, %d)", tt.errorRate, tt.probRate, width, depth, tt.wantWidth, tt.wantDepth)
			}
		})
	}
}

func TestNewCountMinSketch(t *testing.T) {
	t.Run("the sketch is built with the dimensions the rates ask for", func(t *testing.T) {
		for _, tt := range cmsDimensions {
			t.Run(tt.name, func(t *testing.T) {
				cms := newTestCMS(t, tt.errorRate, tt.probRate)
				if cms.GetWidth() != tt.wantWidth {
					t.Errorf("NewCountMinSketch(%v, %v).GetWidth() = %d; want %d", tt.errorRate, tt.probRate, cms.GetWidth(), tt.wantWidth)
				}
				if cms.GetDepth() != tt.wantDepth {
					t.Errorf("NewCountMinSketch(%v, %v).GetDepth() = %d; want %d", tt.errorRate, tt.probRate, cms.GetDepth(), tt.wantDepth)
				}
			})
		}
	})

	t.Run("the matrix has one row per hash and one column per bucket", func(t *testing.T) {
		for _, tt := range cmsDimensions {
			t.Run(tt.name, func(t *testing.T) {
				cms := newTestCMS(t, tt.errorRate, tt.probRate)
				if uint32(len(cms.matrix)) != cms.GetDepth() {
					t.Fatalf("matrix has %d rows; GetDepth() reports %d", len(cms.matrix), cms.GetDepth())
				}
				for i, row := range cms.matrix {
					if uint32(len(row)) != cms.GetWidth() {
						t.Fatalf("matrix row %d has %d columns; GetWidth() reports %d", i, len(row), cms.GetWidth())
					}
				}
			})
		}
	})

	t.Run("a fresh sketch counts nothing", func(t *testing.T) {
		cms := newTestCMS(t, 0.01, 0.01)
		if total := cms.GetTotalCount(); total != 0 {
			t.Errorf("GetTotalCount() on a fresh sketch = %d; want 0", total)
		}
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key-%d", i)
			if got := cms.GetMember(key); got != 0 {
				t.Errorf("GetMember(%q) on a fresh sketch = %d; want 0", key, got)
			}
		}
	})
}

// TestNewCountMinSketchRejectsBadRates covers the rates the formulas cannot be
// run on. Each one used to produce a sketch rather than an error, and each
// failed in its own way: a zero depth returned MaxUint64 for every key, a zero
// width divided by zero on first use, and a rate above 1 turned into a width of
// 4294967295 and an attempt to allocate several terabytes.
func TestNewCountMinSketchRejectsBadRates(t *testing.T) {
	rates := []struct {
		name                string
		errorRate, probRate float64
	}{
		{"error rate of zero", 0.0, 0.01},
		{"error rate of one", 1.0, 0.01},
		{"error rate above one", 2.0, 0.01},
		{"negative error rate", -0.01, 0.01},
		{"error rate not a number", math.NaN(), 0.01},
		{"probability of zero", 0.01, 0.0},
		{"probability of one", 0.01, 1.0},
		{"probability above one", 0.01, 2.0},
		{"negative probability", 0.01, -0.01},
		{"probability not a number", 0.01, math.NaN()},
		{"both out of range", 0.0, 0.0},
	}
	for _, tt := range rates {
		t.Run(tt.name, func(t *testing.T) {
			cms, err := NewCountMinSketch(tt.errorRate, tt.probRate)
			if err == nil {
				t.Fatalf("NewCountMinSketch(%v, %v) = (a %dx%d sketch, nil); want an error", tt.errorRate, tt.probRate, cms.GetDepth(), cms.GetWidth())
			}
			if cms != nil {
				t.Errorf("NewCountMinSketch(%v, %v) returned a sketch alongside its error; want nil", tt.errorRate, tt.probRate)
			}
		})
	}
}

func TestNewCountMinSketchAcceptsRatesInRange(t *testing.T) {
	// The extremes of the valid range, where the formulas produce their
	// smallest dimensions and are most likely to round down to zero.
	rates := []struct{ errorRate, probRate float64 }{
		{0.999, 0.999},
		{0.5, 0.5},
		{0.9, 0.1},
		{0.1, 0.9},
	}
	for _, tt := range rates {
		cms, err := NewCountMinSketch(tt.errorRate, tt.probRate)
		if err != nil {
			t.Errorf("NewCountMinSketch(%v, %v) = error %v; want a sketch", tt.errorRate, tt.probRate, err)
			continue
		}
		if cms.GetWidth() == 0 || cms.GetDepth() == 0 {
			t.Errorf("NewCountMinSketch(%v, %v) built a %dx%d sketch; want both dimensions at least 1", tt.errorRate, tt.probRate, cms.GetDepth(), cms.GetWidth())
		}
		// A sketch with a usable shape has to survive being used.
		cms.Increase("key", 1)
		if got := cms.GetMember("key"); got < 1 {
			t.Errorf("NewCountMinSketch(%v, %v): GetMember(key) after Increase(key, 1) = %d; want at least 1", tt.errorRate, tt.probRate, got)
		}
	}
}

func TestCountMinSketchIncrease(t *testing.T) {
	t.Run("a key on its own is counted exactly", func(t *testing.T) {
		// Nothing else has been added, so no other key can have collided into
		// this one's counters and the estimate has to be exact.
		cms := newTestCMS(t, 0.01, 0.01)
		if got := cms.Increase("solo", 5); got != 5 {
			t.Errorf("Increase(solo, 5) on a fresh sketch = %d; want 5", got)
		}
		if got := cms.GetMember("solo"); got != 5 {
			t.Errorf("GetMember(solo) after Increase(solo, 5) = %d; want 5", got)
		}
	})

	t.Run("repeated increases accumulate", func(t *testing.T) {
		cms := newTestCMS(t, 0.01, 0.01)
		for i := 0; i < 10; i++ {
			cms.Increase("solo", 3)
		}
		if got := cms.GetMember("solo"); got != 30 {
			t.Errorf("GetMember(solo) after ten Increase(solo, 3) = %d; want 30", got)
		}
	})

	t.Run("Increase returns what GetMember would", func(t *testing.T) {
		cms := newTestCMS(t, 0.01, 0.01)
		for i := 0; i < 200; i++ {
			key := fmt.Sprintf("key-%d", i%20)
			returned := cms.Increase(key, uint64(i%5)+1)
			if queried := cms.GetMember(key); returned != queried {
				t.Fatalf("Increase(%q, ...) = %d but GetMember(%q) = %d; want the same estimate", key, returned, key, queried)
			}
		}
	})

	t.Run("totalCount is the sum of every increment", func(t *testing.T) {
		cms := newTestCMS(t, 0.01, 0.01)
		var want uint64
		for i := 0; i < 500; i++ {
			value := uint64(i%7) + 1
			cms.Increase(fmt.Sprintf("key-%d", i%50), value)
			want += value
		}
		if got := cms.GetTotalCount(); got != want {
			t.Errorf("GetTotalCount() = %d; want %d", got, want)
		}
	})

	t.Run("counters saturate instead of wrapping", func(t *testing.T) {
		// Wrapping would turn an overestimate into an underestimate, which is
		// the one thing a count-min sketch promises never to do.
		cms := newTestCMS(t, 0.01, 0.01)
		cms.Increase("big", math.MaxUint64)
		if got := cms.Increase("big", 10); got != math.MaxUint64 {
			t.Errorf("Increase(big, 10) after saturating = %d; want %d", got, uint64(math.MaxUint64))
		}
		if got := cms.GetMember("big"); got != math.MaxUint64 {
			t.Errorf("GetMember(big) after saturating = %d; want %d", got, uint64(math.MaxUint64))
		}
		if got := cms.GetTotalCount(); got != math.MaxUint64 {
			t.Errorf("GetTotalCount() after saturating = %d; want %d", got, uint64(math.MaxUint64))
		}
	})
}

func TestCountMinSketchGetMember(t *testing.T) {
	t.Run("an estimate is never below the true count", func(t *testing.T) {
		// The one guarantee the structure makes. On its own it proves very
		// little: a GetMember that returned MaxUint64 for everything would also
		// pass, which is why the error bound below is the test that matters.
		const errorRate, probRate = 0.01, 0.01
		cms := newTestCMS(t, errorRate, probRate)
		truth := make(map[string]uint64)

		for i := 0; i < 2000; i++ {
			key := fmt.Sprintf("key-%d", i%100)
			value := uint64(i%9) + 1
			cms.Increase(key, value)
			truth[key] += value
		}
		for key, want := range truth {
			if got := cms.GetMember(key); got < want {
				t.Errorf("GetMember(%q) = %d; want at least the true count %d", key, got, want)
			}
		}
	})

	t.Run("an estimate stays within the error rate of the total", func(t *testing.T) {
		// The promise is estimate <= trueCount + errorRate*totalCount, for all
		// but a probRate fraction of keys. Note it scales with the grand total,
		// not with the key's own count: a count-min sketch is accurate about
		// heavy hitters and noisy about rare keys, by design. So the mix below
		// is skewed on purpose rather than uniform.
		const (
			errorRate, probRate = 0.01, 0.01
			heavyKeys           = 5
			heavyCount          = 1000
			tailKeys            = 1000
		)

		cms := newTestCMS(t, errorRate, probRate)
		truth := make(map[string]uint64, heavyKeys+tailKeys)
		for i := 0; i < heavyKeys; i++ {
			key := fmt.Sprintf("heavy-%d", i)
			cms.Increase(key, heavyCount)
			truth[key] = heavyCount
		}
		for i := 0; i < tailKeys; i++ {
			key := fmt.Sprintf("tail-%d", i)
			cms.Increase(key, 1)
			truth[key] = 1
		}

		slack := uint64(math.Ceil(errorRate * float64(cms.GetTotalCount())))
		// The bound is probabilistic, so a probRate fraction of keys is allowed
		// to exceed it. Every key here is a fixed string and the hashes are
		// seeded by row index, so the number of breaches is the same on every
		// run and can be asserted exactly.
		allowed := int(math.Ceil(probRate * float64(len(truth))))

		var breached int
		var worstKey string
		var worstEstimate, worstTruth uint64
		for key, want := range truth {
			got := cms.GetMember(key)
			if got > want+slack {
				breached++
				if got-want > worstEstimate-worstTruth {
					worstKey, worstEstimate, worstTruth = key, got, want
				}
			}
		}
		if breached > allowed {
			t.Errorf("%d of %d keys exceed trueCount+%d (errorRate %v of totalCount %d); want at most %d. Worst: GetMember(%q) = %d for a true count of %d",
				breached, len(truth), slack, errorRate, cms.GetTotalCount(), allowed, worstKey, worstEstimate, worstTruth)
		}
	})
}

// TestCountMinSketchInvariants checks the rules that hold after any sequence of
// calls, with the random number generator picking the sequence.
func TestCountMinSketchInvariants(t *testing.T) {
	t.Run("random increments keep every estimate at or above the truth", func(t *testing.T) {
		const (
			steps   = 300
			keyPool = 20
		)

		cms := newTestCMS(t, 0.01, 0.01)
		truth := make(map[string]uint64)
		// A fixed seed means these 300 operations are the same on every run, so
		// a failure can be reproduced exactly instead of once in a while.
		r := rand.New(rand.NewPCG(1, 2))
		counterSum := cmsCounterSum(cms)
		var expectedTotal uint64

		for step := range steps {
			key := fmt.Sprintf("k%d", r.IntN(keyPool))

			// op records what just happened, so a failure says which call broke
			// the rule rather than only which step number did.
			var op string
			switch r.IntN(2) {
			case 0:
				value := uint64(r.IntN(100)) + 1
				op = fmt.Sprintf("Increase(%s, %d)", key, value)
				returned := cms.Increase(key, value)
				truth[key] += value
				expectedTotal += value
				if queried := cms.GetMember(key); returned != queried {
					t.Fatalf("step %d, %s returned %d but GetMember(%s) = %d", step, op, returned, key, queried)
				}
			case 1:
				op = fmt.Sprintf("GetMember(%s)", key)
				cms.GetMember(key)
			}

			// Rule (a): no estimate is ever below the true count. Breaks if a
			// counter is decremented, if a counter wraps, or if Increase and
			// GetMember disagree about where a key's counters live.
			for trackedKey, want := range truth {
				if got := cms.GetMember(trackedKey); got < want {
					t.Fatalf("step %d, after %s: GetMember(%s) = %d, below the true count %d", step, op, trackedKey, got, want)
				}
			}

			// Rule (b): no call ever lowers a counter.
			current := cmsCounterSum(cms)
			if current < counterSum {
				t.Fatalf("step %d, after %s: counters sum to %d, down from %d; want no decrease", step, op, current, counterSum)
			}
			counterSum = current

			// Rule (c): totalCount tracks every increment.
			if got := cms.GetTotalCount(); got != expectedTotal {
				t.Fatalf("step %d, after %s: GetTotalCount() = %d; want %d", step, op, got, expectedTotal)
			}

			// Rule (d): a single counter cannot hold more than everything that
			// was ever added. Breaks if a key writes outside its own column.
			if max := cmsMaxCounter(cms); max > expectedTotal {
				t.Fatalf("step %d, after %s: a counter holds %d, more than the %d ever added", step, op, max, expectedTotal)
			}
		}
	})
}

// newTestCMS builds a sketch from rates that are expected to be valid, so the
// tests that are about counting do not each have to handle a construction error
// that should never happen.
func newTestCMS(t *testing.T, errorRate, probRate float64) *CountMinSketch {
	t.Helper()
	cms, err := NewCountMinSketch(errorRate, probRate)
	if err != nil {
		t.Fatalf("NewCountMinSketch(%v, %v) = error %v; want a sketch", errorRate, probRate, err)
	}
	return cms
}

func cmsCounterSum(cms *CountMinSketch) uint64 {
	var sum uint64
	for _, row := range cms.matrix {
		for _, counter := range row {
			sum += counter
		}
	}
	return sum
}

func cmsMaxCounter(cms *CountMinSketch) uint64 {
	var max uint64
	for _, row := range cms.matrix {
		for _, counter := range row {
			if counter > max {
				max = counter
			}
		}
	}
	return max
}
