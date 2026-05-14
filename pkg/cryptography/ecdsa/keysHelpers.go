package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"

	"github.com/dustinxie/ecc"
)

const (
	// PrivateKeySize is the byte length of supported ECDSA private keys.
	PrivateKeySize = 32
	// CompressedPublicKeySize is the byte length of compressed ECDSA public keys.
	CompressedPublicKeySize = 33
	// UncompressedPublicKeySize is the byte length of unprefixed X||Y public keys.
	UncompressedPublicKeySize = 64
	// PrefixedUncompressedPublicKeySize is the byte length of 0x04||X||Y public keys.
	PrefixedUncompressedPublicKeySize = 65
)

// UncompressedPublicKeyTo65Bytes returns pubkey with the uncompressed-key prefix.
func UncompressedPublicKeyTo65Bytes(pubkey []byte) ([]byte, error) {
	switch len(pubkey) {
	case PrefixedUncompressedPublicKeySize:
		if pubkey[0] != 0x04 {
			return nil, fmt.Errorf("uncompressed public key prefix must be 0x04")
		}
		return append([]byte(nil), pubkey...), nil
	case UncompressedPublicKeySize:
		result := make([]byte, PrefixedUncompressedPublicKeySize)
		result[0] = 0x04
		copy(result[1:], pubkey)
		return result, nil
	default:
		return nil, fmt.Errorf("uncompressed public key length must be %d or %d but length is %d", UncompressedPublicKeySize, PrefixedUncompressedPublicKeySize, len(pubkey))
	}
}

// CompressPublicKey converts an uncompressed public key into compressed form.
func CompressPublicKey(uncompressedPublicKey []byte) ([]byte, error) {
	raw, err := uncompressedPublicKeyBytes(uncompressedPublicKey)
	if err != nil {
		return nil, err
	}

	var prefix byte = 0x02
	if raw[63]&1 == 1 {
		prefix = 0x03
	}

	result := make([]byte, CompressedPublicKeySize)
	result[0] = prefix
	copy(result[1:], raw[:32])
	return result, nil
}

// PrivateKeyUnmarshal converts a raw private key into an ECDSA private key.
func PrivateKeyUnmarshal(privKey []byte, curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	if curve == nil {
		return nil, fmt.Errorf("curve is required")
	}
	if len(privKey) != PrivateKeySize {
		return nil, fmt.Errorf("private key length must be %d but length is %d", PrivateKeySize, len(privKey))
	}

	d := new(big.Int).SetBytes(privKey)
	if d.Sign() <= 0 {
		return nil, fmt.Errorf("private key must be positive")
	}
	if params := curve.Params(); params != nil && params.N != nil && d.Cmp(params.N) >= 0 {
		return nil, fmt.Errorf("private key is outside curve order")
	}

	pk := new(ecdsa.PrivateKey)
	pk.Curve = curve
	pk.D = d
	// Go's ECDSA implementation expects the public affine point to be populated on private keys.
	//lint:ignore SA1019 secp256k1 support still depends on the elliptic-compatible dustinxie/ecc curve.
	pk.X, pk.Y = curve.ScalarBaseMult(privKey)
	if pk.X == nil || pk.Y == nil {
		return nil, fmt.Errorf("private key does not produce a valid public point")
	}

	return pk, nil
}

// PublicKeyUnmarshal converts compressed or uncompressed bytes into an ECDSA public key.
func PublicKeyUnmarshal(pubKey []byte, curve elliptic.Curve) (*ecdsa.PublicKey, error) {
	if curve == nil {
		return nil, fmt.Errorf("curve is required")
	}

	compressed, err := compressedPublicKeyBytes(pubKey)
	if err != nil {
		return nil, err
	}

	pub := new(ecdsa.PublicKey)
	pub.Curve = curve

	if curve == ecc.P256k1() {
		pub.X, pub.Y = ecc.UnmarshalCompressed(ecc.P256k1(), compressed)
	} else {
		pub.X, pub.Y = elliptic.UnmarshalCompressed(curve, compressed)
	}
	if pub.X == nil || pub.Y == nil {
		return nil, fmt.Errorf("public key is not a valid point on %s", curve.Params().Name)
	}

	return pub, nil
}

func compressedPublicKeyBytes(pubKey []byte) ([]byte, error) {
	switch len(pubKey) {
	case CompressedPublicKeySize:
		if pubKey[0] != 0x02 && pubKey[0] != 0x03 {
			return nil, fmt.Errorf("compressed public key prefix must be 0x02 or 0x03")
		}
		return append([]byte(nil), pubKey...), nil
	case UncompressedPublicKeySize, PrefixedUncompressedPublicKeySize:
		return CompressPublicKey(pubKey)
	default:
		return nil, fmt.Errorf("public key length must be %d, %d or %d but length is %d", CompressedPublicKeySize, UncompressedPublicKeySize, PrefixedUncompressedPublicKeySize, len(pubKey))
	}
}

func uncompressedPublicKeyBytes(pubKey []byte) ([]byte, error) {
	switch len(pubKey) {
	case UncompressedPublicKeySize:
		return append([]byte(nil), pubKey...), nil
	case PrefixedUncompressedPublicKeySize:
		if pubKey[0] != 0x04 {
			return nil, fmt.Errorf("uncompressed public key prefix must be 0x04")
		}
		return append([]byte(nil), pubKey[1:]...), nil
	default:
		return nil, fmt.Errorf("uncompressed public key length must be %d or %d but length is %d", UncompressedPublicKeySize, PrefixedUncompressedPublicKeySize, len(pubKey))
	}
}
