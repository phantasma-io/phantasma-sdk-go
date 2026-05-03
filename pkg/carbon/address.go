package carbon

import (
	"fmt"

	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
)

var (
	// SystemAddressNull is the Carbon system null address.
	SystemAddressNull = systemAddress(0)
	// SystemAddressGasPool is the Carbon system gas pool address.
	SystemAddressGasPool = systemAddress(1)
	// SystemAddressDataPool is the Carbon system data pool address.
	SystemAddressDataPool = systemAddress(2)
)

// Bytes32FromPublicKey converts a 32-byte public key into a Carbon address value.
func Bytes32FromPublicKey(publicKey []byte) (Bytes32, error) {
	if len(publicKey) != 32 {
		return Bytes32{}, fmt.Errorf("public key length must be 32, got %d", len(publicKey))
	}
	return NewBytes32(publicKey)
}

// MustBytes32FromPublicKey converts a public key into a Carbon address value and panics on error.
func MustBytes32FromPublicKey(publicKey []byte) Bytes32 {
	out, err := Bytes32FromPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	return out
}

// Bytes32FromPhantasmaAddress converts a Phantasma address into a Carbon address value.
func Bytes32FromPhantasmaAddress(address cryptography.Address) (Bytes32, error) {
	switch address.Kind() {
	case cryptography.User, cryptography.System:
		bytes := address.Bytes()
		if len(bytes) != cryptography.Length {
			return Bytes32{}, fmt.Errorf("address length must be %d, got %d", cryptography.Length, len(bytes))
		}
		return NewBytes32(bytes[2:])
	default:
		return Bytes32{}, fmt.Errorf("unsupported address kind %d", address.Kind())
	}
}

// Bytes32FromPhantasmaAddressText parses a Phantasma address and converts it into a Carbon address value.
func Bytes32FromPhantasmaAddressText(text string) (Bytes32, error) {
	address, err := cryptography.FromString(text)
	if err != nil {
		return Bytes32{}, err
	}
	return Bytes32FromPhantasmaAddress(address)
}

// MustBytes32FromPhantasmaAddress converts a Phantasma address into a Carbon address value and panics on error.
func MustBytes32FromPhantasmaAddress(address cryptography.Address) Bytes32 {
	out, err := Bytes32FromPhantasmaAddress(address)
	if err != nil {
		panic(err)
	}
	return out
}

// MustBytes32FromPhantasmaAddressText parses a Phantasma address into a Carbon address value and panics on error.
func MustBytes32FromPhantasmaAddressText(text string) Bytes32 {
	out, err := Bytes32FromPhantasmaAddressText(text)
	if err != nil {
		panic(err)
	}
	return out
}

func systemAddress(lastByte byte) Bytes32 {
	var address Bytes32
	address[31] = lastByte
	return address
}
