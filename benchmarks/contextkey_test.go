package httpwarebenchmarks

import "testing"

// How bad is the performance of using strings as context keys?
func BenchmarkStringKey(b *testing.B) {
	m := map[string]string{
		"github.com/nstogner/oneware":   "one",
		"github.com/nstogner/twoware":   "two",
		"github.com/nstogner/threeware": "three",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := m["github.com/nstogner/twoware"]
		_ = s
	}
}

// What is the performance of using ints as context keys?
func BenchmarkIntKey(b *testing.B) {
	m := map[int]string{
		1000: "one",
		2000: "two",
		3000: "three",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := m[2000]
		_ = s
	}
}
