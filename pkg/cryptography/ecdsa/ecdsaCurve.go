package ecdsa

// Curve identifies the ECDSA curve used for signing and verification.
type Curve uint

const (
	// Secp256r1 is used for Neo signatures.
	Secp256r1 Curve = 0
	// Secp256k1 is used for Ethereum and BSC signatures.
	Secp256k1 Curve = 1
)
