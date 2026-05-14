package proofOfAddress

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/dustinxie/ecc"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography/ecdsa"
	"github.com/phantasma-io/phantasma-sdk-go/pkg/cryptography/neoLegacy"
)

type ProofOfAddressesVerifier struct {
	Message            string
	SignedMessage      string
	SignedMessageBytes []byte

	PhaAddress        string
	PhaPublicKeyBytes []byte

	EthAddress        string
	EthPublicKey      string
	EthPublicKeyBytes []byte

	Neo2Address        string
	Neo2PublicKey      string
	Neo2PublicKeyBytes []byte

	PhaSignature       string
	PhaSignatureBytes  []byte
	EthSignature       string
	EthSignatureBytes  []byte
	Neo2Signature      string
	Neo2SignatureBytes []byte
}

// NewProofOfAddressesVerifier parses a signed proof-of-addresses message.
func NewProofOfAddressesVerifier(message string) (*ProofOfAddressesVerifier, error) {
	v := &ProofOfAddressesVerifier{}
	v.Message = message

	split := strings.Split(strings.Replace(v.Message, "\r", "", -1), "\n")
	if len(split) < 10 {
		return nil, fmt.Errorf("proof message has %d lines, expected at least 10", len(split))
	}

	v.SignedMessage = strings.Join(split[:6], "\n")
	v.SignedMessageBytes = []byte(v.SignedMessage)

	var err error
	v.PhaAddress, err = proofField(split[1], "Phantasma address: ")
	if err != nil {
		return nil, err
	}
	phaAddress, err := cryptography.FromString(v.PhaAddress)
	if err != nil {
		return nil, err
	}
	v.PhaPublicKeyBytes, err = phaAddress.GetPublicKey()
	if err != nil {
		return nil, err
	}

	v.EthAddress, err = proofField(split[2], "Ethereum address: ")
	if err != nil {
		return nil, err
	}
	v.EthPublicKey, err = proofField(split[3], "Ethereum public key: ")
	if err != nil {
		return nil, err
	}
	v.EthPublicKeyBytes, err = hex.DecodeString(v.EthPublicKey)
	if err != nil {
		return nil, err
	}

	v.Neo2Address, err = proofField(split[4], "Neo Legacy address: ")
	if err != nil {
		return nil, err
	}
	v.Neo2PublicKey, err = proofField(split[5], "Neo Legacy public key: ")
	if err != nil {
		return nil, err
	}
	v.Neo2PublicKeyBytes, err = hex.DecodeString(v.Neo2PublicKey)
	if err != nil {
		return nil, err
	}
	v.PhaSignature, err = proofField(split[7], "Phantasma signature: ")
	if err != nil {
		return nil, err
	}
	v.PhaSignatureBytes, err = hex.DecodeString(v.PhaSignature)
	if err != nil {
		return nil, err
	}
	v.EthSignature, err = proofField(split[8], "Ethereum signature: ")
	if err != nil {
		return nil, err
	}
	v.EthSignatureBytes, err = hex.DecodeString(v.EthSignature)
	if err != nil {
		return nil, err
	}

	v.Neo2Signature, err = proofField(split[9], "Neo Legacy signature: ")
	if err != nil {
		return nil, err
	}
	v.Neo2SignatureBytes, err = hex.DecodeString(v.Neo2Signature)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func proofField(line string, prefix string) (string, error) {
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("proof line %q does not start with %q", line, prefix)
	}
	return strings.TrimSpace(strings.TrimPrefix(line, prefix)), nil
}

// VerifyMessage verifies all signatures and address derivations in the proof.
func (v *ProofOfAddressesVerifier) VerifyMessage() (bool, string, error) {
	success := true
	errorMessage := ""

	if len(v.PhaPublicKeyBytes) != ed25519.PublicKeySize {
		return false, errorMessage, fmt.Errorf("phantasma public key length must be %d but length is %d", ed25519.PublicKeySize, len(v.PhaPublicKeyBytes))
	}
	if len(v.PhaSignatureBytes) != ed25519.SignatureSize {
		return false, errorMessage, fmt.Errorf("phantasma signature length must be %d but length is %d", ed25519.SignatureSize, len(v.PhaSignatureBytes))
	}

	if !ed25519.Verify(v.PhaPublicKeyBytes, v.SignedMessageBytes, v.PhaSignatureBytes) {
		success = false
		errorMessage += "Phantasma signature is incorrect!\n"
	}

	ethRes, err := ecdsa.Verify(v.SignedMessageBytes, v.EthSignatureBytes, v.EthPublicKeyBytes, ecdsa.Secp256k1)
	if err != nil {
		return false, errorMessage, err
	}
	if !ethRes {
		success = false
		errorMessage += "Ethereum signature is incorrect!\n"
	}

	neoRes, err := ecdsa.Verify(v.SignedMessageBytes, v.Neo2SignatureBytes, v.Neo2PublicKeyBytes, ecdsa.Secp256r1)
	if err != nil {
		return false, errorMessage, err
	}
	if !neoRes {
		success = false
		errorMessage += "Neo Legacy signature is incorrect!\n"
	}

	pubEth, err := ecdsa.PublicKeyUnmarshal(v.EthPublicKeyBytes, ecc.P256k1())
	if err != nil {
		return false, errorMessage, err
	}
	ethAddressFromPublicKey := crypto.PubkeyToAddress(*pubEth).Hex()

	if v.EthAddress != ethAddressFromPublicKey {
		success = false
		errorMessage += "Ethereum address is incorrect: " + ethAddressFromPublicKey + "\n"
	}

	neo2AddressFromPublicKey := neoLegacy.Address([]byte(v.Neo2PublicKeyBytes))
	if v.Neo2Address != neo2AddressFromPublicKey {
		success = false
		errorMessage += "Neo Legacy address is incorrect: " + neo2AddressFromPublicKey + "\n"
	}

	return success, errorMessage, nil
}
