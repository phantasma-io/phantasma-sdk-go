package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/dustinxie/ecc"
	hash "github.com/phantasma-io/phantasma-sdk-go/pkg/util/hashing"
)

// Sign signs message with prikey using the requested ECDSA curve.
func Sign(message, prikey []byte, curve Curve) ([]byte, error) {
	if len(message) == 0 {
		return nil, errors.New("message length is 0")
	}
	if len(prikey) != PrivateKeySize {
		return nil, fmt.Errorf("private key length must be %d but length is %d", PrivateKeySize, len(prikey))
	}

	hash := hash.Sha256(message)

	if curve == Secp256k1 {
		pk, err := PrivateKeyUnmarshal(prikey, ecc.P256k1())
		if err != nil {
			return nil, err
		}

		signature, err := ecc.SignBytes(pk, hash, ecc.LowerS)
		if err != nil {
			return nil, err
		}

		return SignatureDropRecoveryID(signature), nil
	} else if curve == Secp256r1 {
		pk, err := PrivateKeyUnmarshal(prikey, elliptic.P256())
		if err != nil {
			return nil, err
		}

		r, s, err := ecdsa.Sign(rand.Reader, pk, hash)
		if err != nil {
			return nil, err
		}

		signature := RSToSignatureWithoutRecoveryID(r, s)

		return signature, nil
	}

	return nil, errors.New("unsupported curve")
}

// Verify checks an ECDSA signature against message and pubkey.
func Verify(message, signature, pubkey []byte, curve Curve) (bool, error) {
	if len(message) == 0 {
		return false, errors.New("message length is 0")
	}
	if len(signature) != SignatureSize && len(signature) != RecoverableSignatureSize {
		return false, fmt.Errorf("signature length must be %d or %d but length is %d", SignatureSize, RecoverableSignatureSize, len(signature))
	}
	if len(pubkey) == 0 {
		return false, errors.New("public key length is 0")
	}

	hash := hash.Sha256(message)
	if curve == Secp256k1 {
		pub, err := PublicKeyUnmarshal(pubkey, ecc.P256k1())
		if err != nil {
			return false, err
		}

		return ecc.VerifyBytes(pub, hash, SignatureDropRecoveryID(signature), ecc.Normal), nil
	}
	if curve == Secp256r1 {
		pub, err := PublicKeyUnmarshal(pubkey, elliptic.P256())
		if err != nil {
			return false, err
		}

		r, s, err := SignatureToRS(signature)
		if err != nil {
			return false, err
		}
		return ecdsa.Verify(pub, hash, r, s), nil
	}

	return false, nil
}
