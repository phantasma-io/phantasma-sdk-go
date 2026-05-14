package cryptography

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFrom(t *testing.T) {
	// HashFromBytes hashes non-32-byte input, while HashFromString hashes text input.
	hash, _ := HashFromBytes([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19})
	assert.Equal(t, hash.String(), "4bcb1b15332489764c289b51b1119b1057c9e6dbca85b1ebc6553eae7c69e5e4")

	hash = HashFromString("asjdhweiurhwiuthedkgsdkfjh4otuiheriughdfjkgnsdçfjherslighjsghnoçiljhoçitujgpe8rotu89pearthkjdf.")
	assert.Equal(t, hash.String(), "9b93849b43a088f6d0add08f8ebfd4cd4ba8040515f281926c44954dbf65567d")
}

func TestHashIsNull(t *testing.T) {
	// The zero-value hash is the null hash used by older SDK code paths.
	hash := Hash{}
	assert.Equal(t, true, hash.IsNull())
}

func TestHashBytes(t *testing.T) {
	// Bytes returns the internal little-endian representation used by serializers.
	hash, _ := HashFromBytes([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19})
	result := []byte{228, 229, 105, 124, 174, 62, 85, 198, 235, 177, 133, 202, 219, 230, 201, 87, 16, 155, 17, 177, 81, 155, 40, 76, 118, 137, 36, 51, 21, 27, 203, 75}
	assert.Equal(t, result, hash.Bytes())
}

func TestHashBytesReturnsDefensiveCopy(t *testing.T) {
	input := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 48, 49}
	hash, err := HashFromBytes(input)
	assert.NoError(t, err)

	input[0] = 99
	bytes := hash.Bytes()
	bytes[1] = 88

	assert.Equal(t, []byte{0, 1, 2, 3}, hash.Bytes()[:4])
}

func TestHashParse(t *testing.T) {
	// ParseHash accepts the canonical 64-character public hex form and preserves String roundtrip.
	hash, _ := ParseHash("e4e5697cae3e55c6ebb185cadbe6c957109b11b1519b284c76892433151bcb4b")
	assert.Equal(t, "e4e5697cae3e55c6ebb185cadbe6c957109b11b1519b284c76892433151bcb4b", hash.String())
}

func TestHashParseRejectsInvalidInputWithoutPanic(t *testing.T) {
	// Invalid user-supplied hash text must be reported as an error, not as a panic.
	assert.NotPanics(t, func() {
		_, err := ParseHash("abc")
		assert.Error(t, err)
	})
	assert.NotPanics(t, func() {
		_, err := ParseHash(strings.Repeat("z", HashLength*2))
		assert.Error(t, err)
	})
}
