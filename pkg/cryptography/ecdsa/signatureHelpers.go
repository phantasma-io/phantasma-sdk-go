package ecdsa

import (
	"fmt"
	"math/big"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/util"
)

const (
	// SignatureSize is the byte length of a raw R||S ECDSA signature.
	SignatureSize = 64
	// RecoverableSignatureSize is the byte length of an R||S||V ECDSA signature.
	RecoverableSignatureSize = 65
)

// SignatureDropRecoveryID removes the recovery ID from a 65-byte R||S||V signature.
func SignatureDropRecoveryID(signature []byte) []byte {

	if len(signature) != RecoverableSignatureSize {
		return util.ArrayClone(signature)
	}

	return util.ArrayClone(signature[:SignatureSize])
}

// RSToSignatureWithoutRecoveryID serializes an R/S pair as a fixed-width R||S signature.
func RSToSignatureWithoutRecoveryID(r, s *big.Int) []byte {
	if r == nil || s == nil {
		return nil
	}
	if r.Sign() < 0 || s.Sign() < 0 || r.BitLen() > 256 || s.BitLen() > 256 {
		return nil
	}

	signature := make([]byte, SignatureSize)
	r.FillBytes(signature[:32])
	s.FillBytes(signature[32:])
	return signature
}

// SignatureToRS splits a 64-byte R||S or 65-byte R||S||V signature into R and S.
func SignatureToRS(signature []byte) (*big.Int, *big.Int, error) {
	signature = SignatureDropRecoveryID(signature)
	if len(signature) != SignatureSize {
		return nil, nil, fmt.Errorf("signature length must be %d but length is %d", SignatureSize, len(signature))
	}

	return big.NewInt(0).SetBytes(signature[:32]), big.NewInt(0).SetBytes(signature[32:]), nil
}
