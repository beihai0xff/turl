package mapping

import (
	"bytes"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBase58Encode(t *testing.T) {
	t.Run("EncodeSmallNum", func(t *testing.T) {
		require.Equal(t, "", string(Base58Encode(0)))
		require.Equal(t, "2", string(Base58Encode(1)))
		require.Equal(t, "z", string(Base58Encode(57)))
		require.Equal(t, "21", string(Base58Encode(58)))
	})

	t.Run("EncodePositiveNumber", func(t *testing.T) {
		require.Equal(t, "BukQL", string(Base58Encode(123456789)))
	})

	t.Run("EncodeLargeNumber", func(t *testing.T) {
		require.Equal(t, "NQm6nKp8qFC", string(Base58Encode(9223372036854775807)))
	})
}

func TestBase58Decode(t *testing.T) {
	t.Run("DecodeSmallNum", func(t *testing.T) {
		num, err := Base58Decode([]byte(""))
		require.NoError(t, err)
		require.Equal(t, uint64(0), num)

		num, err = Base58Decode([]byte("2"))
		require.NoError(t, err)
		require.Equal(t, uint64(1), num)

		num, err = Base58Decode([]byte("z"))
		require.NoError(t, err)
		require.Equal(t, uint64(57), num)

		num, err = Base58Decode([]byte("21"))
		require.NoError(t, err)
		require.Equal(t, uint64(58), num)
	})

	t.Run("DecodePositiveNumber", func(t *testing.T) {
		num, err := Base58Decode([]byte("BukQL"))
		require.NoError(t, err)
		require.Equal(t, uint64(123456789), num)
	})

	t.Run("DecodeLargeNumber", func(t *testing.T) {
		num, err := Base58Decode([]byte("zzzzzzzz"))
		require.NoError(t, err)
		require.Equal(t, uint64(128063081718015), num)
	})

	t.Run("DecodeInvalidNumber", func(t *testing.T) {
		num, err := Base58Decode([]byte("zzzzzzzzz"))
		require.ErrorIs(t, err, ErrBase58Overflow)
		require.Equal(t, uint64(0), num)

		num, err = Base58Decode([]byte("000"))
		require.ErrorIs(t, err, ErrorInvalidCharacter)
		require.Equal(t, uint64(0), num)

		num, err = Base58Decode([]byte("12345l"))
		require.ErrorIs(t, err, ErrorInvalidCharacter)
		require.Equal(t, uint64(0), num)
	})
}

func Test_pow(t *testing.T) {
	t.Run("TestPow", func(t *testing.T) {
		require.Equal(t, 1, pow(0))
		require.Equal(t, 58, pow(1))
		require.Equal(t, 3364, pow(2))
		require.Equal(t, 195112, pow(3))
		require.Equal(t, 11316496, pow(4))
		require.Equal(t, 656356768, pow(5))
		require.Equal(t, 38068692544, pow(6))
		require.Equal(t, 2207984167552, pow(7))
		require.Equal(t, 128063081718016, pow(8))
		require.Equal(t, 0, pow(9))
	})
}

func Test_reverse(t *testing.T) {
	t.Run("TestReverse", func(t *testing.T) {
		a := []byte("hello")
		reverse(a)
		require.Equal(t, "olleh", string(a))
	})
}

func Test_index(t *testing.T) {
	t.Run("TestIndex", func(t *testing.T) {
		for i, char := range chars {
			require.Equal(t, i, index(char))
		}
	})
}

func BenchmarkBase58Decode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Base58Decode([]byte("zzzzzz"))
	}
}

func BenchmarkBase58Encode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		Base58Encode(1e9)
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

func BenchmarkIndexByte(b *testing.B) {
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

func BenchmarkPow(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := range 9 {
		for range b.N / 8 {
			pow(i)
		}
	}
}

func BenchmarkPowC(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := range 9 {
		for range b.N / 8 {
			powC(carry, i)
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

func powC(x, n int) int {
	if n == 0 {
		return 1
	}

	if n%2 == 0 {
		return powC(x*x, n/2)
	}

	return x * powC(x, n-1)
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
