package hmap

import (
	"math/rand"
	"testing"

	. "github.com/apmckinlay/gsuneido/util/hamcrest"
)

type ik int

func (x ik) Hash() uint32 {
	return uint32(x)
}

func (x ik) Equals(y interface{}) bool {
	return x == y.(ik)
}

func TestRandom(t *testing.T) {
	const N = 1000
	hm := NewHmap(0)
	Assert(t).That(hm.Size(), Equals(0))
	nums := map[int32]int{}
	for i := 0; i < N; i++ {
		n := rand.Int31n(N)
		hm.Put(ik(n), i)
		nums[n] = i
	}
	rand.Seed(1)
	for i := 0; i < N; i++ {
		n := rand.Int31n(N)
		Assert(t).That(hm.Get(ik(n)), Equals(nums[n]))
	}
	rand.Seed(1)
	for i := 0; i < N; i++ {
		n := rand.Int31n(N)
		v := hm.Del(ik(n))
		if nums[n] == -1 {
			Assert(t).That(v, Equals(nil))
		} else {
			Assert(t).That(v, Equals(nums[n]))
		}
		nums[n] = -1
	}
	Assert(t).That(hm.Size(), Equals(0))
}

func BenchmarkAdd(b *testing.B) {
	for n := 0; n < b.N; n++ {
		hm := NewHmap(0)
		for i := 0; i < 100; i++ {
			hm.Put(mix(i), i)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	hm := NewHmap(100)
	for i := 0; i < 100; i++ {
		hm.Put(ik(i), i)
	}
	for n := 0; n < b.N; n++ {
		hm.Get(ik(n % 100))
	}
}

func mix(n int) ik {
	n = ^n + (n << 15)
	n = n ^ (n >> 12)
	n = n + (n << 2)
	n = n ^ (n >> 4)
	n = n * 2057
	n = n ^ (n >> 16)
	return ik(n)
}

func TestCopy(t *testing.T) {
	hm := NewHmap(10)
	for i := 0; i < 10; i++ {
		hm.Put(mix(i), i)
	}
	hm2 := hm.Copy()
	Assert(t).That(hm2.Size(), Equals(hm.Size()))
	Assert(t).That(hm2.String(), Equals(hm.String()))
}
