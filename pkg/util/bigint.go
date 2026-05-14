package util

import (
	"math/big"
	"slices"
)

// BigIntToCsharpByteArray converts a big.Int to C# BigInteger little-endian two's-complement bytes.
func BigIntToCsharpByteArray(n *big.Int) []byte {
	if n.BitLen() == 0 { // Check if big int is zero
		return []byte{0x00}
	}

	var bytes = n.Bytes()

	if n.Sign() == -1 {
		TwosComplementConvertTo(bytes)
	}

	slices.Reverse(bytes) // Converting to little-endian format

	if n.Sign() == -1 {
		if bytes[len(bytes)-1] < 128 {
			bytes = append(bytes, 0xff)
		}
	} else {
		if bytes[len(bytes)-1] >= 128 && bytes[len(bytes)-1] != 0x00 {
			bytes = append(bytes, 0x00)
		}
	}

	return bytes
}

// BigIntToPhantasmaByteArray converts a big.Int to Phantasma's signed integer byte format.
func BigIntToPhantasmaByteArray(n *big.Int) []byte {
	var bytes = BigIntToCsharpByteArray(n)

	if n.Sign() == -1 { // Big int is negative
		if len(bytes) == 1 {
			bytes = append(bytes, 0xff, 0xff)
		} else if len(bytes) > 1 && bytes[len(bytes)-1] == 0xff {
			bytes = append(bytes, 0xff)
		}
	} else {
		if bytes[len(bytes)-1] != 0x00 {
			bytes = append(bytes, 0x00)
		}
	}

	return bytes
}

// BigIntBytesFromCsharpOrPhantasmaByteArray returns magnitude bytes and sign from a signed integer byte slice.
func BigIntBytesFromCsharpOrPhantasmaByteArray(bytes []byte) ([]byte, int) {
	n := make([]byte, len(bytes))
	copy(n, bytes)

	if len(n) == 1 && n[0] == 0x00 {
		return []byte{}, 0
	}

	if n[len(n)-1] < 128 {
		// It's a positive number.

		slices.Reverse(n)
		return n, 1
	}

	// It's a negative number.

	slices.Reverse(n)
	TwosComplementConvertFrom(n)

	return n, -1
}

// BigIntFromCsharpOrPhantasmaByteArray converts C# or Phantasma signed integer bytes to a big.Int.
func BigIntFromCsharpOrPhantasmaByteArray(bytes []byte) *big.Int {
	b, sign := BigIntBytesFromCsharpOrPhantasmaByteArray(bytes)

	n := big.NewInt(0)
	if sign < 0 {
		n = n.Mul(n.SetBytes(b), big.NewInt(-1))
	} else {
		n = n.SetBytes(b)
	}

	return n
}
