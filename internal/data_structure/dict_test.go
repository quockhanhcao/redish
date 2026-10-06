package data_structure

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"
)

func TestDictionarySet(t *testing.T) {
	t.Run("after Set(k, v, _), Get(k).Value == v, last write wins", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, -1)
		dict.Set("a", 2, -1)
		dict.Set("b", 3, -1)

		objA := dict.Get("a")
		if objA == nil {
			t.Fatalf("Set(a, 2, -1) return nil pointer")
		}
		if objA.Value != 2 {
			t.Errorf("Set(a, 2, -1) after Set(a, 1, -1) = %d; want 2", objA.Value)
		}
		objB := dict.Get("b")
		if objB == nil {
			t.Fatalf("Set(b, 3, -1) return nil pointer")
		}
		if objB.Value != 3 {
			t.Errorf("Set(b, 3, -1) = %d; want 3", objB.Value)
		}
	})

	t.Run("a new key is counted once", func(t *testing.T) {
		dict := InitDictionary()
		before := Stats.Key

		dict.Set("a", 1, -1)

		assertStatsDelta(t, before, 1)
	})

	t.Run("overwriting a key is not counted again", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, -1)
		before := Stats.Key

		dict.Set("a", 2, -1)

		assertStatsDelta(t, before, 0)
	})

	t.Run("records an expiry when given a TTL", func(t *testing.T) {
		dict := InitDictionary()
		const ttl = 10

		before := time.Now().UnixMilli()
		dict.Set("a", 1, ttl)
		after := time.Now().UnixMilli()

		expireAt, ok := dict.GetExpiry("a")
		if !ok {
			t.Fatalf("GetExpiry(a) after Set(a, 1, %d) reported no expiry; want one", ttl)
		}
		lo, hi := before+ttl*1000, after+ttl*1000
		if expireAt < lo || expireAt > hi {
			t.Errorf("GetExpiry(a) = %d; want within [%d, %d]", expireAt, lo, hi)
		}
	})

	t.Run("records no expiry when exp is -1", func(t *testing.T) {
		dict := InitDictionary()

		dict.Set("a", 1, -1)

		assertNoExpiry(t, dict, "a")
	})

	t.Run("clears an existing expiry when exp is -1", func(t *testing.T) {
		t.Skip("known bug (Part 4.4): SET without EX keeps the old TTL, see dict.go:41")

		dict := InitDictionary()
		dict.Set("a", 1, 10)

		dict.Set("a", 2, -1)

		assertNoExpiry(t, dict, "a")
	})
}

func TestDictionaryGet(t *testing.T) {
	t.Run("missing key returns nil", func(t *testing.T) {
		dict := InitDictionary()

		if obj := dict.Get("missing"); obj != nil {
			t.Errorf("Get(missing) on an empty dictionary = %v; want nil", obj.Value)
		}
	})

	t.Run("key whose deadline has not passed is still alive", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, 60)

		obj := dict.Get("a")
		if obj == nil {
			t.Fatal("Get(a) before its deadline = nil; want the stored object")
		}
		if obj.Value != 1 {
			t.Errorf("Get(a).Value = %v; want 1", obj.Value)
		}
	})

	t.Run("expired key returns nil, deletes the key and decrements Stats", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, 10)
		expireNow(t, dict, "a")
		before := Stats.Key

		obj := dict.Get("a")

		if obj != nil {
			t.Errorf("Get(a) on an expired key = %v; want nil", obj.Value)
		}
		// Expiry is lazy and Get is what performs it, so the key and its
		// deadline must both be gone afterwards and the counter must drop.
		assertNoExpiry(t, dict, "a")
		assertStatsDelta(t, before, -1)
	})

	// A key is expired only when now > expireAt, strictly (dict.go:50), so at
	// exactly its deadline it is still alive. That boundary can't be pinned
	// down while time.Now() is hard-coded; add the case in Part 5.4 once the
	// clock is a parameter.
}

func TestDictionaryDel(t *testing.T) {
	t.Run("removes the key, its expiry and decrements Stats", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, 10)
		before := Stats.Key

		dict.Del("a")

		if obj := dict.Get("a"); obj != nil {
			t.Errorf("Get(a) after Del(a) = %v; want nil", obj.Value)
		}
		assertNoExpiry(t, dict, "a")
		assertStatsDelta(t, before, -1)
	})

	t.Run("missing key is a no-op and leaves Stats unchanged", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, -1)
		before := Stats.Key

		dict.Del("missing")

		if obj := dict.Get("a"); obj == nil {
			t.Error("Get(a) after Del(missing) = nil; want the stored object")
		}
		assertStatsDelta(t, before, 0)
	})
}

func TestDictionaryExpiry(t *testing.T) {
	t.Run("GetExpiry on a key with no TTL reports none", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, -1)

		assertNoExpiry(t, dict, "a")
	})

	t.Run("GetExpiry on a missing key reports none", func(t *testing.T) {
		dict := InitDictionary()

		assertNoExpiry(t, dict, "missing")
	})

	t.Run("GetExpiry does not expire the key", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, 10)
		expireNow(t, dict, "a")
		before := Stats.Key

		if _, ok := dict.GetExpiry("a"); !ok {
			t.Error("GetExpiry(a) on an expired key reported no expiry; want one")
		}
		assertStatsDelta(t, before, 0)
	})

	t.Run("SetExpiry overwrites an existing expiry", func(t *testing.T) {
		dict := InitDictionary()
		dict.Set("a", 1, 10)
		first, ok := dict.GetExpiry("a")
		if !ok {
			t.Fatal("GetExpiry(a) after Set(a, 1, 10) reported no expiry; want one")
		}

		const ttl = 60
		before := time.Now().UnixMilli()
		dict.SetExpiry("a", ttl)
		after := time.Now().UnixMilli()

		second, ok := dict.GetExpiry("a")
		if !ok {
			t.Fatal("GetExpiry(a) after SetExpiry(a, 60) reported no expiry; want one")
		}
		if second == first {
			t.Errorf("GetExpiry(a) = %d, unchanged by SetExpiry; want a new deadline", second)
		}
		lo, hi := before+ttl*1000, after+ttl*1000
		if second < lo || second > hi {
			t.Errorf("GetExpiry(a) after SetExpiry(a, %d) = %d; want within [%d, %d]", ttl, second, lo, hi)
		}
	})

	t.Run("SetExpiry on a missing key records an orphan expiry", func(t *testing.T) {
		dict := InitDictionary()

		dict.SetExpiry("missing", 10)

		// SetExpiry does not check that the key exists, so it leaves a deadline
		// behind for a key holding no data. Documented deliberately: this is
		// the one operation that breaks "every expiry entry has a data entry".
		// Decide in Part 5.4 whether SetExpiry should reject it.
		if _, ok := dict.GetExpiry("missing"); !ok {
			t.Error("GetExpiry(missing) after SetExpiry(missing, 10) reported no expiry; want one")
		}
		if obj := dict.Get("missing"); obj != nil {
			t.Errorf("Get(missing) = %v; want nil", obj.Value)
		}
	})
}

// TestDictionaryInvariants checks the two rules that must hold after *any*
// call, no matter what came before. The tests above each pick their own steps
// and check one promise; this one lets the random number generator pick the
// steps and checks the same two rules after every single one. That is what
// catches bugs that only show up in an order nobody would think to write
// down by hand.
//
// SetExpiry is deliberately left out of the operation mix: it records a
// deadline without checking that the key exists, which breaks rule (b) on its
// own. See "SetExpiry on a missing key records an orphan expiry" above, which
// pins that behavior down as a contract instead.
func TestDictionaryInvariants(t *testing.T) {
	t.Run("random sequence keeps Stats and the expiry map consistent", func(t *testing.T) {
		const (
			steps   = 1000
			keyPool = 10 // small on purpose, so the same keys are hit over and over
		)

		dict := InitDictionary()
		baseline := Stats.Key
		// A fixed seed means these 1000 operations are the same on every run,
		// so a failure can be reproduced exactly instead of once in a while.
		r := rand.New(rand.NewPCG(1, 2))

		for step := range steps {
			key := fmt.Sprintf("k%d", r.IntN(keyPool))

			// op records what just happened, so a failure says which call broke
			// the rule rather than only which step number did.
			var op string
			switch r.IntN(5) {
			case 0:
				op = fmt.Sprintf("Set(%s, v, -1)", key)
				dict.Set(key, "v", -1)
			case 1:
				op = fmt.Sprintf("Set(%s, v, 60)", key)
				dict.Set(key, "v", 60)
			case 2:
				// Set a key and immediately back-date it, so later Gets have
				// something to expire. Without this the lazy-expiry path in
				// Get never runs and the most interesting case goes untested.
				op = fmt.Sprintf("Set(%s, v, 60) then back-date it", key)
				dict.Set(key, "v", 60)
				expireNow(t, dict, key)
			case 3:
				op = fmt.Sprintf("Get(%s)", key)
				dict.Get(key)
			case 4:
				op = fmt.Sprintf("Del(%s)", key)
				dict.Del(key)
			}

			// Rule (a): the counter equals the number of keys actually stored.
			// Breaks if Set counts an overwrite twice, if Del decrements for a
			// key that was not there, or if Get expires a key without counting
			// the removal.
			live := int64(len(dict.GetDataDict()))
			if counted := Stats.Key - baseline; counted != live {
				t.Fatalf("step %d, after %s: Stats.Key counts %d keys; dictionary holds %d", step, op, counted, live)
			}

			// Rule (b): no deadline is left behind for a key that holds no data.
			// Breaks if Del forgets the expiry map, or if lazy expiry in Get
			// removes the data but not the deadline.
			for expiringKey := range dict.GetExpireKeyDict() {
				if _, ok := dict.GetDataDict()[expiringKey]; !ok {
					t.Fatalf("step %d, after %s: %q has an expiry but no data", step, op, expiringKey)
				}
			}
		}
	})
}

// assertNoExpiry checks key has no deadline recorded.
func assertNoExpiry(t *testing.T, d *Dictionary, key string) {
	t.Helper()
	if expireAt, ok := d.GetExpiry(key); ok {
		t.Errorf("GetExpiry(%q) = (%d, true); want (0, false)", key, expireAt)
	}
}

// assertStatsDelta checks Stats.Key moved by want since before was captured.
// Stats is a package-level global shared by every test in this package, so its
// absolute value means nothing here and only the delta can be asserted. For the
// same reason no test in this file may call t.Parallel().
// TODO(Part 5.3): drop this once the keyspace counter stops being global.
func assertStatsDelta(t *testing.T, before, want int64) {
	t.Helper()
	if got := Stats.Key - before; got != want {
		t.Errorf("Stats.Key moved by %d; want %d", got, want)
	}
}

// expireNow back-dates key's deadline so it is already expired, which avoids a
// sleep of over a second (Set only accepts whole seconds). It reaches into the
// private map on purpose: that makes it visibly a test back door rather than
// something mistakable for supported API.
// TODO(Part 5.4): replace with the injected clock.
func expireNow(t *testing.T, d *Dictionary, key string) {
	t.Helper()
	d.expireKeyDictStore[key] = time.Now().UnixMilli() - 1
}
