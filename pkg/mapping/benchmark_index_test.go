package mapping

import (
	"bytes"
	"testing"
)

func BenchmarkIndexArray(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for _, char := range chars {
		b.StartTimer()
		for range b.N / carry {
			_ = indices[char]
		}
		b.StopTimer()
	}
}

func BenchmarkIndex(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for _, char := range chars {
		b.StartTimer()
		for range b.N / carry {
			index(char)
		}
		b.StopTimer()
	}
}

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

func BenchmarkStandardIndexByte(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for _, char := range chars {
		b.StartTimer()
		for range b.N / carry {
			bytes.IndexByte(chars, char)
		}
		b.StopTimer()
	}
}

func BenchmarkIndexMap(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for _, char := range chars {
		b.StartTimer()
		for range b.N / carry {
			indexMap(char)
		}
		b.StopTimer()
	}
}

var cache = map[byte]int{
	'1': 0, '2': 1, '3': 2, '4': 3, '5': 4, '6': 5, '7': 6, '8': 7, '9': 8,
	'A': 9, 'B': 10, 'C': 11, 'D': 12, 'E': 13, 'F': 14, 'G': 15, 'H': 16, 'J': 17, 'K': 18, 'L': 19, 'M': 20,
	'N': 21, 'P': 22, 'Q': 23, 'R': 24, 'S': 25, 'T': 26, 'U': 27, 'V': 28, 'W': 29, 'X': 30, 'Y': 31, 'Z': 32,
	'a': 33, 'b': 34, 'c': 35, 'd': 36, 'e': 37, 'f': 38, 'g': 39, 'h': 40, 'i': 41, 'j': 42, 'k': 43, 'm': 44,
	'n': 45, 'o': 46, 'p': 47, 'q': 48, 'r': 49, 's': 50, 't': 51, 'u': 52, 'v': 53, 'w': 54, 'x': 55, 'y': 56,
	'z': 57,
}

func indexMap(char byte) int {
	if pos, ok := cache[char]; ok {
		return pos
	}

	return -1
}
