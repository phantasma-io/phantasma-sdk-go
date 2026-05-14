package cryptography

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/io"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/util"
	hashing "github.com/phantasma-io/phantasma-sdk-go/pkg/util/hashing"
)

// HashLength is the number of bytes in a Phantasma hash.
const HashLength = 32

// HexPrefix is the optional prefix accepted by ParseHash.
const HexPrefix = "0x"

// Hash stores a 32-byte transaction, block or content hash.
type Hash struct {
	_data []byte
}

// HashFromBytes returns a Hash from a 32-byte slice or hashes arbitrary data into one.
func HashFromBytes(data []byte) (Hash, error) {

	if len(data) != HashLength {
		data = hashing.Sha256(data)
	}

	return Hash{_data: append([]byte(nil), data...)}, nil
}

// HashFromString creates an instance of Hash from a string.
func HashFromString(s string) Hash {
	data := hashing.Sha256([]byte(s))
	return Hash{append([]byte(nil), data...)}
}

// ParseHash parses a public hex hash string.
func ParseHash(s string) (Hash, error) {

	if strings.HasPrefix(s, HexPrefix) {
		s = s[2:]
	}

	if len(s) != HashLength*2 {
		return Hash{}, errors.New("hash string must contain exactly 64 hex characters")
	}

	data, err := hex.DecodeString(s)
	if err != nil {
		return Hash{}, err
	}

	slices.Reverse(data)
	return Hash{append([]byte(nil), data...)}, nil
}

// Size returns the length of the underlying byte slice
func (h Hash) Size() int {
	return len(h._data)
}

// IsNull checks if the Hash represents a nil hash
func (h Hash) IsNull() bool {
	if h._data == nil {
		return true
	}

	empty := make([]byte, HashLength)
	return bytes.Equal(h._data, empty)
}

// String returns the canonical public hex representation of the hash.
func (h Hash) String() string {
	data := util.ArrayCloneAndReverse(h._data)
	return hex.EncodeToString(data)
}

// Bytes returns a copy of the internal little-endian hash bytes.
func (h Hash) Bytes() []byte {
	return append([]byte(nil), h._data...)
}

// FromUnpaddedHex creates an instance of Hash from an unpadded hex string.
func (h Hash) FromUnpaddedHex(s string) (Hash, error) {

	if strings.HasPrefix(s, HexPrefix) {
		s = s[2:]
	}

	var sb strings.Builder
	sb.WriteString(s)

	for sb.Len() < 64 {
		sb.WriteString("0")
		sb.WriteString("0")
	}

	return ParseHash(sb.String())
}

// GetDifficulty retrieves the current difficulty of the hash
func (h Hash) GetDifficulty() int {
	var result int = 0
	for i := 0; i < len(h._data); i++ {
		var n = h._data[i]

		for j := 0; j < 8; j++ {
			if (n & (1 << j)) != 0 {
				result = 1 + (i << 3) + j
			}
		}
	}

	return 256 - result
}

// Serialize implements the Serializable interface.
func (h *Hash) Serialize(writer *io.BinWriter) {
	writer.WriteVarBytes(h._data)
}

// Deserialize implements the Serializable interface.
func (h *Hash) Deserialize(reader *io.BinReader) {
	data := reader.ReadVarBytes(HashLength)
	if reader.Err != nil {
		return
	}
	if len(data) != 0 && len(data) != HashLength {
		reader.Err = fmt.Errorf("invalid hash byte length: got %d, want %d", len(data), HashLength)
		return
	}
	h._data = append([]byte(nil), data...)
}
