package mapping

import (
	"math"
	"testing"
)

func BenchmarkPowArray(b *testing.B) {
	b.ReportAllocs()

	b.ResetTimer()
	for i := range 9 {
		for range b.N / 8 {
			_ = pow58[i]
		}
	}
}

func BenchmarkPowCompute(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := range 9 {
		for range b.N / 8 {
			pow(carry, i)
		}
	}
}

func BenchmarkMathPow(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := range 9 {
		for range b.N / 8 {
			_ = int(math.Pow(float64(carry), float64(i)))
		}
	}
}
