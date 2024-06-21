package mapping

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
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
		require.Equal(t, 1, pow(58, 0))
		require.Equal(t, 58, pow(58, 1))
		require.Equal(t, 3364, pow(58, 2))
		require.Equal(t, 195112, pow(58, 3))
		require.Equal(t, 11316496, pow(58, 4))
		require.Equal(t, 656356768, pow(58, 5))
		require.Equal(t, 38068692544, pow(58, 6))
		require.Equal(t, 2207984167552, pow(58, 7))
		require.Equal(t, 128063081718016, pow(58, 8))
	})

	t.Run("TestPowArray", func(t *testing.T) {
		require.Equal(t, 1, pow58[0])
		require.Equal(t, 58, pow58[1])
		require.Equal(t, 3364, pow58[2])
		require.Equal(t, 195112, pow58[3])
		require.Equal(t, 11316496, pow58[4])
		require.Equal(t, 656356768, pow58[5])
		require.Equal(t, 38068692544, pow58[6])
		require.Equal(t, 2207984167552, pow58[7])
		require.Equal(t, 128063081718016, pow58[8])
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
			require.Equal(t, i, indices[char])
		}

		for i := range byte(0xff) {
			if !slices.Contains(chars, i) {
				require.Equalf(t, -1, index(i), "index(%c) = %d", i, index(i))
			}
		}
	})
}

func BenchmarkBase58Encode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		Base58Encode(uint64(i + 1e9))
	}
}

func BenchmarkBase58Decode(b *testing.B) {
	b.ReportAllocs()

	arr := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		arr[i] = Base58Encode(uint64(i + 1e9))
	}

	b.ResetTimer()
	for i := range b.N {
		Base58Decode(arr[i])
	}
}
