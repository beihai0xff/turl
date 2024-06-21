// Package mapping provides URL shortening functions
// base58.go provides functions to encode and decode numbers to base58
package mapping

import (
	"errors"
	"fmt"
)

const (
	carry    = 58
	maxBytes = 8
)

var (
	// ErrInvalidInput base error for invalid input
	ErrInvalidInput = errors.New("base58 invalid input")
	// ErrBase58Overflow is returned when the number to decode is too large, greater than eight bytes
	ErrBase58Overflow = fmt.Errorf("%w: number is too large", ErrInvalidInput)
	// ErrorInvalidCharacter is returned when an invalid character is found in the input
	ErrorInvalidCharacter = fmt.Errorf("%w: invalid character in input", ErrInvalidInput)

	chars = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

	// indices is an array of indices of characters in the base58 alphabet

	// according to the below benchmarks, this implementation is nearly ten faster than using bytes.IndexByte, and it doesn't allocate memory
	// BenchmarkIndexArray
	// BenchmarkIndexArray-8          	1000000000	         0.3158 ns/op	       0 B/op	       0 allocs/op
	// BenchmarkIndex
	// BenchmarkIndex-8               	902548872	         1.320 ns/op	       0 B/op	       0 allocs/op
	// BenchmarkStandardIndexByte
	// BenchmarkStandardIndexByte-8   	427854204	         2.876 ns/op	       0 B/op	       0 allocs/op
	// BenchmarkIndexMap
	// BenchmarkIndexMap-8            	146482124	         8.188 ns/op	       0 B/op	       0 allocs/op
	indices = [256]int{}

	// pow58 returns 58^n
	// according to the below benchmarks, this implementation is sixteen times faster than calculating the power
	// and thirty times faster than using math.Pow
	// BenchmarkPowArray
	// BenchmarkPowArray-8            	1000000000	         0.3565 ns/op	       0 B/op	       0 allocs/op
	// BenchmarkPowCompute
	// BenchmarkPowCompute-8          	204542810	         5.839 ns/op	       0 B/op	       0 allocs/op
	// BenchmarkMathPow
	// BenchmarkMathPow-8             	100000000	        10.09 ns/op	       0 B/op	       0 allocs/op
	pow58 = [9]int{}
)

func init() {
	for i := range indices {
		indices[i] = -1
	}

	for i, char := range chars {
		indices[char] = i
	}

	for i := range pow58 {
		pow58[i] = pow(carry, i)
	}
}

// Base58Encode encodes a number to base58
func Base58Encode(num uint64) []byte {
	b := make([]byte, 0, maxBytes) // 58^8 = 1.28e14, enough for our use case

	for ; num > 0; num /= carry {
		b = append(b, chars[num%carry])
	}
	reverse(b)

	return b
}

// Base58Decode decodes a base58 encoded string to a number
func Base58Decode(b []byte) (uint64, error) {
	n := len(b)
	if n > maxBytes { // 58^8 = 1.28e14, less than math.MaxUint64, and it's enough for our use case
		return 0, ErrBase58Overflow
	}

	var num uint64

	for i := range n {
		pos := indices[b[i]]
		if pos == -1 {
			return 0, ErrorInvalidCharacter
		}

		num += uint64(pow58[n-i-1] * pos)
	}

	return num, nil
}

func reverse(a []byte) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

func pow(x, n int) int {
	if n == 0 {
		return 1
	}

	if n%2 == 0 {
		return pow(x*x, n/2) //nolint:mnd
	}

	return x * pow(x, n-1)
}
