package cryptography

import "github.com/phantasma-io/phantasma-sdk-go/pkg/io"

// SignatureKind identifies the signature algorithm encoded in a transaction.
type SignatureKind uint

const (
	// None means no signature algorithm is present.
	None SignatureKind = iota
	// Ed25519 identifies Phantasma Ed25519 signatures.
	Ed25519
	// ECDSA identifies ECDSA signatures.
	ECDSA
	// Ring identifies ring signatures.
	Ring
)

// Signature is the common interface for transaction signatures.
type Signature interface {
	Kind() SignatureKind
	Verify(message []byte, addresses []Address) bool
	Serialize(*io.BinWriter)
	Deserialize(*io.BinReader)
	Bytes() []byte
}
