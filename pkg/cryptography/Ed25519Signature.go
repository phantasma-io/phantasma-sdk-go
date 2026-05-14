package cryptography

import (
	"crypto/ed25519"
	"fmt"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/io"
)

// Ed25519Signature is a Phantasma Ed25519 transaction signature.
type Ed25519Signature struct {
	bytes []byte
	kind  SignatureKind
}

// NewEd25519Signature creates a signature object from raw 64-byte signature data.
func NewEd25519Signature(bytes []byte) (*Ed25519Signature, error) {
	if len(bytes) != ed25519.SignatureSize {
		return nil, fmt.Errorf("Ed25519 signature length must be %d but length is %d", ed25519.SignatureSize, len(bytes))
	}
	return &Ed25519Signature{bytes: append([]byte(nil), bytes...), kind: Ed25519}, nil
}

func signEd25519(keyPair KeyPair, message []byte) (*Ed25519Signature, error) {
	if keyPair == nil {
		return nil, fmt.Errorf("key pair is required")
	}

	privateKey := keyPair.ExpandedPrivateKey()
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("expanded private key length must be %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	return NewEd25519Signature(ed25519.Sign(ed25519.PrivateKey(privateKey), message))
}

// Kind returns the type of the signature.
func (sig Ed25519Signature) Kind() SignatureKind {
	return sig.kind
}

// Verify checks whether message was signed by one of addresses.
func (sig Ed25519Signature) Verify(message []byte, addresses []Address) bool {
	if len(sig.bytes) != ed25519.SignatureSize {
		return false
	}

	for _, address := range addresses {
		if !address.IsUser() {
			continue
		}

		address := address
		pubKey, err := address.GetPublicKey()
		if err != nil || len(pubKey) != ed25519.PublicKeySize {
			continue
		}

		if ed25519.Verify(pubKey, message, sig.bytes) {
			return true
		}
	}

	return false
}

// Bytes returns the serialized byte representation of the signature.
func (sig Ed25519Signature) Bytes() []byte {
	bw := *io.NewBufBinWriter()
	sig.Serialize(bw.BinWriter)
	if bw.Err != nil {
		return nil
	}
	return bw.Bytes()
}

// Serialize implements the Serializable interface.
func (sig Ed25519Signature) Serialize(writer *io.BinWriter) {
	if len(sig.bytes) != ed25519.SignatureSize {
		writer.Err = fmt.Errorf("Ed25519 signature length must be %d but length is %d", ed25519.SignatureSize, len(sig.bytes))
		return
	}
	writer.WriteVarBytes(sig.bytes)
}

// Deserialize implements the Serializable interface.
func (sig *Ed25519Signature) Deserialize(reader *io.BinReader) {
	data := reader.ReadVarBytes(ed25519.SignatureSize)
	if reader.Err != nil {
		return
	}
	if len(data) != ed25519.SignatureSize {
		reader.Err = fmt.Errorf("Ed25519 signature length must be %d but length is %d", ed25519.SignatureSize, len(data))
		return
	}
	sig.bytes = append([]byte(nil), data...)
	sig.kind = Ed25519
}
