// Package mapping provides URL shortening functions
package mapping

import (
	"math"
	"strings"
)

const (
	chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	carry = 62
)

func encode(num int64) string {
	bytes := []byte{}
	for num > 0 {
		bytes = append(bytes, chars[num%carry])
		num /= carry
	}
	reverse(bytes)

	return string(bytes)
}

func decode(str string) int64 {
	var num int64

	n := len(str)
	for i := 0; i < n; i++ {
		pos := strings.IndexByte(chars, str[i])
		num += int64(math.Pow(carry, float64(n-i-1)) * float64(pos))
	}

	return num
}

func reverse(a []byte) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}
