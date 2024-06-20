// Package mapping provides URL shortening functions
// base58.go provides functions to encode and decode numbers to base58
package mapping

import (
	"errors"
)

const (
	carry    = 58
	maxBytes = 8
)

var (
	// ErrBase58Overflow is returned when the number to decode is too large, greater than eight bytes
	ErrBase58Overflow = errors.New("base58: number is too large to decode")
	// ErrorInvalidCharacter is returned when an invalid character is found in the input
	ErrorInvalidCharacter = errors.New("base58: invalid character in input")

	chars = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
)

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
		pos := index(b[i])
		if pos == -1 {
			return 0, ErrorInvalidCharacter
		}

		num += uint64(pow(n-i-1) * pos)
	}

	return num, nil
}

func reverse(a []byte) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}

// pow returns 58^n
// according to the below benchmarks, this implementation is sixteen times faster than calculating the power
// and thirty times faster than using math.Pow
// BenchmarkPow
// BenchmarkPow-8            	1000000000	         0.3520 ns/op	       0 B/op	       0 allocs/op
// BenchmarkPowC
// BenchmarkPowC-8           	203575606	         5.851 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMathPow
// BenchmarkMathPow-8        	100000000	        10.25 ns/op	       0 B/op	       0 allocs/op

//nolint:mnd, gocyclo
func pow(n int) int {
	switch {
	case n < 4:
		if n < 2 {
			if n == 0 {
				return 1
			}
			return 58
		}

		if n == 2 {
			return 3364
		}

		return 195112
	case n < 9:
		if n < 6 {
			if n == 4 {
				return 11316496
			}

			return 656356768
		} else {
			if n < 8 {
				if n == 6 {
					return 38068692544

				}

				return 2207984167552
			}

			return 128063081718016
		}
	default:
		return 0
	}
}

// index returns the index of a character in the base58 alphabet

// according to the below benchmarks, this implementation is twice faster than using bytes.IndexByte
// and six times faster than using a map
// BenchmarkIndex-8          	870152160	         1.314 ns/op	       0 B/op	       0 allocs/op
// BenchmarkIndexByte
// BenchmarkIndexByte-8      	434694481	         2.758 ns/op	       0 B/op	       0 allocs/op
// BenchmarkIndexMap
// BenchmarkIndexMap-8       	147732596	         8.102 ns/op	       0 B/op	       0 allocs/op

func index(char byte) int {
	if char >= '1' && char <= '9' {
		return int(char - '1')
	}

	if char >= 'A' && char <= 'Z' {
		if char < 'I' {
			return int(char - 'A' + 9)
		}

		if char > 'O' {
			return int(char - 'A' + 7)
		}

		if char > 'I' && char < 'O' {
			return int(char - 'A' + 8)
		}
	}

	if char >= 'a' && char <= 'z' {
		if char < 'l' {
			return int(char - 'a' + 33)
		} else if char > 'l' {
			return int(char - 'm' + 44)
		}
	}

	return -1
}
