package cryptography

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/encoding/base58"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/io"
)

// Length is the length of data
const Length = 34

// AddressKind type
type AddressKind byte

const (
	// Invalid address
	Invalid AddressKind = 0x00
	// User address
	User AddressKind = 0x01
	// System address
	System AddressKind = 0x02
	// Interop address
	Interop AddressKind = 0x03
)

// Address stores the binary and text forms of a Phantasma address.
type Address struct {
	data []byte
	text string
	kind AddressKind
}

var nullAddressBytes = [Length]byte{}

// Text returns (and initializes if needed) raw text representation of the address. For null address it returns nil. In most cases use String() instead to get text address, which returns "NULL" string for null addresses.
func (a Address) Text() string {
	if a.text != "" {
		return a.text
	}

	var prefix string
	switch kind := a.Kind(); kind {
	case User:
		prefix = "P"
	case Interop:
		prefix = "X"
	default:
		prefix = "S"
	}

	a.text = prefix + base58.Encode(a.data)
	return a.text
}

// NewAddress returns a new Address object from raw address bytes.
func NewAddress(pubKey []byte) (Address, error) {
	if pubKey == nil {
		return Address{}, fmt.Errorf("public key is required")
	}

	if len(pubKey) != Length {
		return Address{}, fmt.Errorf("public key length must be %d but length is %d", Length, len(pubKey))
	}

	address := Address{}
	address.data = append([]byte(nil), pubKey...)

	return address, nil
}

// NullAddress returns a new null Address object.
func NullAddress() Address {
	address, err := NewAddress(nullAddressBytes[:])
	if err != nil {
		panic(err)
	}
	return address
}

// FromString creates an instance of an Address from a string
func FromString(s string) (Address, error) {

	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return Address{}, fmt.Errorf("invalid address string: empty or too short")
	}

	data, err := base58.Decode(s[1:])
	if err != nil {
		return Address{}, err
	}

	address, err := NewAddress(data)
	if err != nil {
		return Address{}, err
	}

	switch prefix := s[:1]; prefix {
	case "P":
		if address.Kind() != User {
			return Address{}, fmt.Errorf("address has to be of type User")
		}
	case "X":
		if address.Kind() != Interop {
			return Address{}, fmt.Errorf("address has to be of type Interop")
		}
	case "S":
		if address.Kind() != System {
			return Address{}, fmt.Errorf("address has to be of type System")
		}
	default:
		return Address{}, fmt.Errorf("unknown address prefix: %s", prefix)
	}

	return address, nil
}

// MustAddressFromString parses a Phantasma address and panics on invalid input.
func MustAddressFromString(s string) Address {
	address, err := FromString(s)
	if err != nil {
		panic(err)
	}
	return address
}

// IsValidAddress verifies if a string is a valid address
func IsValidAddress(text string) bool {
	_, err := FromString(text)
	return err == nil
}

// IsUser verifies if the passed in address is a user address
func (a Address) IsUser() bool {
	return a.Kind() == User
}

// FromKey generates an address from a KeyPair.
func FromKey(keyPair KeyPair) (Address, error) {
	if keyPair == nil {
		return Address{}, fmt.Errorf("key pair is required")
	}

	data := make([]byte, Length)
	data[0] = byte(User)

	publicKey := keyPair.PublicKey()
	if len(publicKey) == 32 {

		copy(data[2:], publicKey[0:32])

	} else if len(publicKey) == 33 {

		copy(data[1:], publicKey[0:33])

	} else {
		return Address{}, fmt.Errorf("public key length must be 32 or 33 bytes but length is %d", len(publicKey))
	}

	return NewAddress(data)
}

// IsNull checks if the Address represents a nil Address
func (a Address) IsNull() bool {
	if a.data == nil {
		return true
	}

	empty := make([]byte, Length)
	return bytes.Equal(a.data, empty)
}

// Kind returns the kind of an address
func (a Address) Kind() AddressKind {
	if a.IsNull() {
		a.kind = System
		return a.kind
	}

	if a.data[0] >= 3 {
		a.kind = Interop
	} else if a.data[0] == 2 {
		a.kind = System
	} else {
		a.kind = User
	}

	return a.kind
}

// String returns the base58 encoded address with its text prefix.
func (a Address) String() string {
	if a.IsNull() {
		return "NULL"
	}
	return a.Text()
}

// Bytes returns a copy of the raw address bytes.
func (a Address) Bytes() []byte {
	return append([]byte(nil), a.data...)
}

// BytesPrefixed returns address bytes with the length prefix used by binary serialization.
func (a Address) BytesPrefixed() []byte {
	prefixed := make([]byte, 1+len(a.data))
	prefixed[0] = Length
	copy(prefixed[1:], a.data)
	return prefixed
}

// GetPublicKey returns a copy of the Ed25519 public key embedded in a user address.
func (a *Address) GetPublicKey() ([]byte, error) {
	if a.data == nil {
		return []byte{}, nil
	}
	if len(a.data) != Length {
		return nil, fmt.Errorf("invalid address byte length: got %d, want %d", len(a.data), Length)
	}

	p := make([]byte, 32)
	copy(p, a.data[2:])

	return p, nil
}

// Serialize implements the Serializable interface.
func (a *Address) Serialize(writer *io.BinWriter) {
	writer.WriteVarBytes(a.data)
}

// Deserialize implements the Serializable interface.
func (a *Address) Deserialize(reader *io.BinReader) {
	data := reader.ReadVarBytes(Length)
	if reader.Err != nil {
		return
	}
	if len(data) != Length {
		reader.Err = fmt.Errorf("invalid address byte length: got %d, want %d", len(data), Length)
		return
	}
	a.data = append([]byte(nil), data...)
	a.text = ""
	a.kind = Invalid
}
