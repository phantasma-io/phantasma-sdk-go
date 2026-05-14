package cryptography

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/encoding/base58"
)

const (
	wifVersion          = byte(0x80)
	wifCompressedSuffix = byte(0x01)
)

// KeyPair signs Phantasma messages and exposes the public identity associated
// with the private key. Byte slices returned by implementations must be copies
// so callers cannot mutate key material owned by the SDK.
type KeyPair interface {
	PrivateKey() []byte
	ExpandedPrivateKey() []byte
	PublicKey() []byte
	Address() Address

	Sign(msg []byte) (Signature, error)
}

// PrivateKeyLength defines the length of a private key seed.
const PrivateKeyLength = 32

// PhantasmaKeys holds the Ed25519 key material used by Phantasma accounts.
type PhantasmaKeys struct {
	privateKey         []byte
	expandedPrivateKey []byte
	publicKey          []byte
	address            Address
}

// NewPhantasmaKeys instantiates a new PhantasmaKeys object based on the given seed.
func NewPhantasmaKeys(seed []byte) (PhantasmaKeys, error) {
	if len(seed) != PrivateKeyLength {
		return PhantasmaKeys{}, fmt.Errorf("private key seed length must be %d but length is %d", PrivateKeyLength, len(seed))
	}

	keys := PhantasmaKeys{}
	pk := ed25519.NewKeyFromSeed(seed)
	keys.publicKey = append([]byte(nil), pk[32:]...)
	keys.privateKey = append([]byte(nil), pk[:32]...)
	keys.expandedPrivateKey = append([]byte(nil), pk...)
	address, err := FromKey(keys)
	if err != nil {
		return PhantasmaKeys{}, err
	}
	keys.address = address

	return keys, nil
}

// GeneratePhantasmaKeys creates a new Phantasma keypair.
func GeneratePhantasmaKeys() (PhantasmaKeys, error) {
	seed := make([]byte, PrivateKeyLength)
	if _, err := rand.Read(seed); err != nil {
		return PhantasmaKeys{}, err
	}
	return NewPhantasmaKeys(seed)
}

// FromWIF creates a new key pair from a Bitcoin-style WIF private key.
func FromWIF(wif string) (PhantasmaKeys, error) {

	if len(wif) == 0 {
		return PhantasmaKeys{}, fmt.Errorf("WIF needs to be set")
	}

	data, err := base58.CheckDecode(wif)
	if err != nil {
		return PhantasmaKeys{}, err
	}

	if len(data) != PrivateKeyLength+1 && len(data) != PrivateKeyLength+2 {
		return PhantasmaKeys{}, fmt.Errorf("invalid WIF payload length: got %d, want %d or %d", len(data), PrivateKeyLength+1, PrivateKeyLength+2)
	}
	if data[0] != wifVersion {
		return PhantasmaKeys{}, fmt.Errorf("invalid WIF version: got 0x%02x, want 0x%02x", data[0], wifVersion)
	}
	if len(data) == PrivateKeyLength+2 && data[PrivateKeyLength+1] != wifCompressedSuffix {
		return PhantasmaKeys{}, fmt.Errorf("invalid compressed WIF suffix: got 0x%02x, want 0x%02x", data[PrivateKeyLength+1], wifCompressedSuffix)
	}

	privateKey := make([]byte, PrivateKeyLength)
	copy(privateKey[0:], data[1:33])

	return NewPhantasmaKeys(privateKey)
}

func (k PhantasmaKeys) String() string {
	return k.address.String()
}

// Sign generates a signature for the passed in message.
func (k PhantasmaKeys) Sign(msg []byte) (Signature, error) {
	signature, err := signEd25519(k, msg)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// WIF returns the compressed WIF representation of the private key.
func (k PhantasmaKeys) WIF() string {
	bytes := make([]byte, PrivateKeyLength+2)
	bytes[0] = wifVersion

	copy(bytes[1:], k.PrivateKey()[0:32])
	bytes[33] = wifCompressedSuffix
	encoded := base58.CheckEncode(bytes)

	return string(encoded)
}

// ExpandedPrivateKey returns a copy of the Ed25519 expanded private key.
func (k PhantasmaKeys) ExpandedPrivateKey() []byte {
	return append([]byte(nil), k.expandedPrivateKey...)
}

// PrivateKey returns a copy of the private key seed.
func (k PhantasmaKeys) PrivateKey() []byte {
	return append([]byte(nil), k.privateKey...)
}

// PublicKey returns a copy of the public key.
func (k PhantasmaKeys) PublicKey() []byte {
	return append([]byte(nil), k.publicKey...)
}

// Address returns the associated address.
func (k PhantasmaKeys) Address() Address {
	return k.address
}
